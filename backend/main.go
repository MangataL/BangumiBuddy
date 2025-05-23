package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	sqlitedriver "github.com/glebarez/sqlite"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"

	"github.com/MangataL/BangumiBuddy/internal/auth"
	"github.com/MangataL/BangumiBuddy/internal/auth/crypto/pbkdf2"
	"github.com/MangataL/BangumiBuddy/internal/auth/token/jwt"
	"github.com/MangataL/BangumiBuddy/internal/downloader"
	downloadadapter "github.com/MangataL/BangumiBuddy/internal/downloader/adapter"
	"github.com/MangataL/BangumiBuddy/internal/meta/tmdb"
	noticeadapter "github.com/MangataL/BangumiBuddy/internal/notice/adapter"
	"github.com/MangataL/BangumiBuddy/internal/repository/viper"
	ginrouter "github.com/MangataL/BangumiBuddy/internal/router/gin"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
	subscriberrepo "github.com/MangataL/BangumiBuddy/internal/subscriber/repository"
	"github.com/MangataL/BangumiBuddy/internal/subscriber/rss/mikan"
	"github.com/MangataL/BangumiBuddy/internal/transfer"
	_ "github.com/MangataL/BangumiBuddy/internal/transfer/hadrlink"
	episodeparser "github.com/MangataL/BangumiBuddy/internal/transfer/parser"
	transferrepo "github.com/MangataL/BangumiBuddy/internal/transfer/repository"
	_ "github.com/MangataL/BangumiBuddy/internal/transfer/softlink"
	"github.com/MangataL/BangumiBuddy/internal/web"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

//go:embed web
var html embed.FS

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	initLogger(ctx)
	var r *gin.Engine
	var dev bool
	if os.Getenv("DEV") == "true" {
		dev = true
	}
	if dev {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r = gin.Default()
	if dev {
		r.Use(log.GinLogger())
	}
	r.Use(serverStatic("web", html))
	r.GET("/favicon.ico", func(c *gin.Context) {
		file, err := html.ReadFile("web/favicon.ico")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "image/x-icon", file)
	})
	r.Use(log.GinRecovery())

	db, err := gorm.Open(sqlitedriver.Open(getDBPath()), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Fatalf(ctx, "open db failed %s", err)
	}

	configPath := getConfigPath()
	initConfig(ctx, configPath)
	conf, err := viper.NewRepo(configPath)
	if err != nil {
		log.Fatalf(ctx, "init config failed %s", err)
	}
	authenticator := auth.New(auth.Dependencies{
		Config:        conf,
		Cipher:        pbkdf2.NewCipher(),
		TokenOperator: jwt.NewTokenOperator(),
	})

	rssParser := mikan.NewParser()
	tmdbConfig, err := conf.GetTMDBConfig()
	if err != nil {
		log.Fatalf(ctx, "get tmdb config failed %s", err)
	}
	metaParser := tmdb.NewParser(tmdbConfig)
	conf.RegisterReloadable(viper.ComponentNameTMDB, metaParser)

	noticeConfig, err := conf.GetNoticeConfig()
	if err != nil {
		log.Fatalf(ctx, "get notice config failed %s", err)
	}
	noticeAdapter := noticeadapter.NewAdapter(noticeConfig)
	conf.RegisterReloadable(viper.ComponentNameNotice, noticeAdapter)

	downloaderConfig, err := conf.GetDownloaderConfig()
	if err != nil {
		log.Fatalf(ctx, "get downloader config failed %s", err)
	}
	torrentOperator := downloader.NewTorrentOperator(db)
	downloadAdapter := downloadadapter.NewAdapter(downloaderConfig)
	conf.RegisterReloadable(viper.ComponentNameDownloader, downloadAdapter)

	downloadManagerConfig, err := conf.GetDownloadManagerConfig()
	if err != nil {
		log.Fatalf(ctx, "get download manager config failed %s", err)
	}
	downloadManager := downloader.NewManager(downloader.Dependency{
		Downloader:      downloadAdapter,
		TorrentOperator: torrentOperator,
		Config:          downloadManagerConfig,
		Notifier:        noticeAdapter,
	})
	conf.RegisterReloadable(viper.ComponentNameDownloadManager, downloadManager)
	subscriberConfig, err := conf.GetSubscriberConfig()
	if err != nil {
		log.Fatalf(ctx, "get subscriber config failed %s", err)
	}
	subscriberRepo := subscriberrepo.New(db)

	subscriberDep := subscriber.Dependency{
		RSSParser:           rssParser,
		MetaParser:          metaParser,
		Repository:          subscriberRepo,
		Downloader:          downloadManager,
		TorrentOperator:     torrentOperator,
		Config:              subscriberConfig,
		RSSRecordRepository: subscriberRepo,
		Notifier:            noticeAdapter,
	}
	subscriber := subscriber.NewSubscriber(subscriberDep)
	conf.RegisterReloadable(viper.ComponentNameSubscriber, subscriber)

	transferConfig, err := conf.GetTransferConfig()
	if err != nil {
		log.Fatalf(ctx, "get transfer config failed %s", err)
	}
	transfer := transfer.NewTransfer(transfer.Dependency{
		Config:          transferConfig,
		TorrentOperator: torrentOperator,
		Downloader:      downloadManager,
		EpisodeParser:   episodeparser.NewEpisodeParser(),
		Subscriber:      subscriber,
		TransferFiles:   transferrepo.NewTransferFilesRepo(db),
		Notifier:        noticeAdapter,
	})
	conf.RegisterReloadable(viper.ComponentNameTransfer, transfer)

	webService := web.New(web.Dependency{
		Subscriber:      subscriber,
		Downloader:      downloadManager,
		TorrentOperator: torrentOperator,
		Transfer:        transfer,
	})

	// 注册路由

	// 注册认证相关路由
	router := ginrouter.New(ginrouter.Dependency{
		Authenticator:   authenticator,
		TorrentOperator: torrentOperator,
		ConfigRepo:      conf,
		Subscriber:      subscriber,
		Web:             webService,
		Transfer:        transfer,
	})
	r.POST("/apis/v1/token", router.Token)
	apisRouter := r.Group("/apis/v1", router.CheckToken)
	apisRouter.PUT("/user", router.UpdateUser)

	// 注册配置相关路由
	apisRouter.GET("/config/tmdb", router.GetTMDBConfig)
	apisRouter.PUT("/config/tmdb", router.SetTMDBConfig)
	apisRouter.GET("/config/download/manager", router.GetDownloadManagerConfig)
	apisRouter.PUT("/config/download/manager", router.SetDownloadManagerConfig)
	apisRouter.GET("/config/download/downloader", router.GetDownloaderConfig)
	apisRouter.PUT("/config/download/downloader", router.SetDownloaderConfig)
	apisRouter.GET("/config/subscriber", router.GetSubscriberConfig)
	apisRouter.PUT("/config/subscriber", router.SetSubscriberConfig)
	apisRouter.GET("/config/transfer", router.GetTransferConfig)
	apisRouter.PUT("/config/transfer", router.SetTransferConfig)
	apisRouter.GET("/config/notice", router.GetNoticeConfig)
	apisRouter.PUT("/config/notice", router.SetNoticeConfig)

	// 注册番剧相关路由
	apisRouter.GET("/bangumis/rss", router.ParseRSS)
	apisRouter.POST("/bangumis", router.Subscribe)
	apisRouter.GET("/bangumis/:id", router.GetBangumi)
	apisRouter.GET("/bangumis/base", router.ListBangumisBase)
	apisRouter.GET("/bangumis", router.ListBangumis)
	apisRouter.PUT("/bangumis/:id", router.UpdateSubscription)
	apisRouter.DELETE("/bangumis/:id", router.DeleteSubscription)
	apisRouter.GET("/bangumis/:id/rss_match", router.GetRSSMatch)
	apisRouter.POST("/bangumis/:id/rss_match", router.MarkRSSRecord)
	apisRouter.POST("/bangumis/:id/download", router.HandleBangumiSubscription)
	apisRouter.GET("/bangumis/:id/torrents", router.GetBangumiTorrents)
	apisRouter.GET("/bangumis/calendar", router.GetSubscriptionCalendar)

	// 注册torrent相关路由
	apisRouter.DELETE("/torrents/:hash", router.DeleteTorrent)
	apisRouter.POST("/torrents/:hash/transfer", router.Transfer)
	apisRouter.GET("/torrents/:hash/files", router.GetTorrentFiles)
	apisRouter.GET("/torrents/recent", router.ListRecentUpdatedTorrents)

	// 注册qbittorrent相关路由
	apisRouter.POST("/downloader/qbittorrent/check", router.CheckQBittorrentConnection)

	// 注册日志相关路由
	apisRouter.GET("/logs", ginrouter.GetLogContent)

	r.NoRoute(func(c *gin.Context) {
		log.Debugf(context.Background(), "no route %s", c.Request.URL.Path)
		c.FileFromFS("/web/index.html", http.FS(html))
	})

	if err := r.Run("[::]:6937"); err != nil {
		log.Fatalf(ctx, "run server failed %s", err)
	}
}

func serverStatic(prefix string, embedFS embed.FS) gin.HandlerFunc {
	return func(c *gin.Context) {
		fsys, err := fs.Sub(embedFS, prefix)
		if err != nil {
			log.Fatalf(context.Background(), "create sub fs failed %s", err)
		}
		fs2 := http.FS(fsys)
		f, err := fs2.Open(c.Request.URL.Path)
		if err != nil {
			c.Next()
			return
		}
		defer f.Close()
		http.FileServer(fs2).ServeHTTP(c.Writer, c.Request)
		c.Abort()
	}
}

const (
	defaultDBPath = "/data/data.db"
)

func getDBPath() string {
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		fmt.Println("dbPath", dbPath)
		return dbPath
	}
	return defaultDBPath
}

const (
	defaultConfigPath = "/config/config.yaml"
)

func getConfigPath() string {
	if configPath := os.Getenv("CONFIG_FILE_PATH"); configPath != "" {
		return configPath
	}
	return defaultConfigPath
}

func initConfig(ctx context.Context, path string) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			configFile, err := os.Create(path)
			if err != nil {
				log.Fatalf(ctx, "create config file failed %s", err)
				return
			}
			_ = configFile.Close()
			return
		}
		log.Fatalf(ctx, "open config file failed %s", err)
	}
	_ = file.Close()
}

var logConfig = log.Config{
	Level:    zapcore.DebugLevel,
	Filename: "/data/log/log.log",
}

func initLogger(ctx context.Context) {
	if logPath := os.Getenv("LOG_FILE_PATH"); logPath != "" {
		logConfig.Filename = logPath
	}
	logger, err := logConfig.Build()
	if err != nil {
		log.Fatalf(ctx, "init logger failed %s", err)
	}
	log.SetLogger(logger)
}
