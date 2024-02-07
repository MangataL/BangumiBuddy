package subscriber

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/meta"
	"github.com/MangataL/BangumiBuddy/internal/notice"
	"github.com/MangataL/BangumiBuddy/pkg/errs"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

//go:generate mockgen -destination subscriber_mock.go -source $GOFILE -package $GOPACKAGE

var _ Interface = (*Subscriber)(nil)

// NewSubscriber 创建订阅器
func NewSubscriber(dep Dependency) *Subscriber {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Subscriber{
		rssParser:       dep.RSSParser,
		metaParser:      dep.MetaParser,
		repo:            dep.Repository,
		rssRecord:       dep.RSSRecordRepository,
		config:          dep.Config,
		downloader:      dep.Downloader,
		torrentOperator: dep.TorrentOperator,
		notifier:        dep.Notifier,
		stop:            cancel,
	}

	// 启动RSS检查器
	go s.runRSSChecker(ctx)

	return s
}

// Dependency subscriber初始化依赖
type Dependency struct {
	RSSParser
	Repository
	RSSRecordRepository
	downloader.TorrentOperator
	notice.Notifier
	Config
	Downloader downloader.Interface
	MetaParser meta.Parser
}

// RSSParser RSS解析器
type RSSParser interface {
	Parse(ctx context.Context, link string) (RSS, error)
}

// Repository 存储层
type Repository interface {
	Save(ctx context.Context, bangumi Bangumi) error
	List(ctx context.Context, req ListBangumiReq) ([]Bangumi, error)
	Get(ctx context.Context, rssLink string) (Bangumi, error)
	Delete(ctx context.Context, id string) error
	UpdateLastAirEpisode(ctx context.Context, subscriptionID string, episode int) error
	StopSubscription(ctx context.Context, id string) error
}

// RSSRecordRepository 用于存储RSS处理记录
type RSSRecordRepository interface {
	// IsProcessed 检查RSS条目是否已处理
	IsProcessed(ctx context.Context, subscriptionID string, guid string) (bool, error)
	// MarkProcessed 标记RSS条目为已处理
	MarkProcessed(ctx context.Context, subscriptionID string, guid ...string) error
	// ListProcessedGUIDs 获取订阅下所有已处理的GUID
	ListProcessedGUIDs(ctx context.Context, subscriptionID string) ([]string, error)
	// DeleteProcessed 删除RSS条目处理记录
	DeleteProcessed(ctx context.Context, subscriptionID string, guid ...string) error
}

// Config 配置项
type Config struct {
	RSSCheckInterval int      `mapstructure:"rss_check_interval" json:"rssCheckInterval" default:"30"`
	IncludeRegs      []string `mapstructure:"include_regs" json:"includeRegs"`
	ExcludeRegs      []string `mapstructure:"exclude_regs" json:"excludeRegs"`
	AutoStop         bool     `mapstructure:"auto_stop" json:"autoStop"`
}

type Subscriber struct {
	rssParser       RSSParser
	metaParser      meta.Parser
	repo            Repository
	rssRecord       RSSRecordRepository
	config          Config
	downloader      downloader.Interface
	torrentOperator downloader.TorrentOperator
	notifier        notice.Notifier
	stop            func()

	rssTicker *time.Ticker
}

func (s *Subscriber) ParseRSS(ctx context.Context, rssLink string) (ParseRSSRsp, error) {
	if rssLink == "" {
		return ParseRSSRsp{}, errs.NewBadRequest("RSS链接不能为空")
	}
	rss, err := s.rssParser.Parse(ctx, rssLink)
	if err != nil {
		return ParseRSSRsp{}, err
	}
	meta, err := s.metaParser.Search(ctx, rss.BangumiName)
	if err != nil {
		return ParseRSSRsp{}, err
	}
	return ParseRSSRsp{
		Name:            meta.ChineseName,
		Season:          meta.Season,
		Year:            meta.Year,
		TMDBID:          meta.TMDBID,
		RSSLink:         rssLink,
		ReleaseGroup:    rss.ReleaseGroup,
		EpisodeTotalNum: meta.EpisodeTotalNum,
		AirWeekday:      meta.AirWeekday,
	}, nil
}

func (s *Subscriber) Subscribe(ctx context.Context, req SubscribeReq) (Bangumi, error) {
	meta, err := s.metaParser.Parse(ctx, req.TMDBID)
	if err != nil {
		return Bangumi{}, fmt.Errorf("解析元数据失败: %w", err)
	}
	bangumi := Bangumi{
		SubscriptionID:  uuid.New().String(),
		Name:            meta.ChineseName,
		RSSLink:         req.RSSLink,
		Active:          true, // 默认为活跃状态
		IncludeRegs:     req.IncludeRegs,
		ExcludeRegs:     req.ExcludeRegs,
		EpisodeOffset:   req.EpisodeOffset,
		Priority:        req.Priority,
		Season:          req.Season,
		Year:            meta.Year,
		TMDBID:          req.TMDBID,
		ReleaseGroup:    req.ReleaseGroup,
		EpisodeLocation: req.EpisodeLocation,
		PosterURL:       meta.PosterURL,
		BackdropURL:     meta.BackdropURL,
		Overview:        meta.Overview,
		Genres:          meta.Genres,
		AirWeekday:      req.AirWeekday,
		EpisodeTotalNum: req.EpisodeTotalNum,
	}
	if err := s.repo.Save(ctx, bangumi); err != nil {
		return Bangumi{}, fmt.Errorf("保存失败: %w", err)
	}
	return bangumi, nil
}

// Get implements Subscriber.
func (s *Subscriber) Get(ctx context.Context, id string) (Bangumi, error) {
	if id == "" {
		return Bangumi{}, errs.NewBadRequest("订阅ID不能为空")
	}
	return s.repo.Get(ctx, id)
}

// List implements Subscriber.
func (s *Subscriber) List(ctx context.Context, req ListBangumiReq) ([]Bangumi, error) {
	return s.repo.List(ctx, req)
}

// StopSubscribe implements Subscriber.
func (s *Subscriber) UpdateSubscription(ctx context.Context, req UpdateSubscribeReq) error {
	if err := validateUpdateSubscribeReq(req); err != nil {
		return err
	}

	oldBangumi, err := s.Get(ctx, req.SubscriptionID)
	if err != nil {
		return fmt.Errorf("获取旧订阅失败: %w", err)
	}
	var (
		overview = oldBangumi.Overview
		genres   = oldBangumi.Genres
	)
	log.Debugf(ctx, "当前订阅: %+v", oldBangumi)
	bm, err := s.metaParser.Parse(ctx, oldBangumi.TMDBID)
	if err == nil {
		overview = bm.Overview
		genres = bm.Genres
	}
	bangumi := Bangumi{
		SubscriptionID:  oldBangumi.SubscriptionID,
		RSSLink:         oldBangumi.RSSLink,
		Active:          req.Active,
		IncludeRegs:     req.IncludeRegs,
		ExcludeRegs:     req.ExcludeRegs,
		EpisodeOffset:   req.EpisodeOffset,
		Priority:        req.Priority,
		Season:          oldBangumi.Season,
		TMDBID:          oldBangumi.TMDBID,
		ReleaseGroup:    oldBangumi.ReleaseGroup,
		EpisodeLocation: req.EpisodeLocation,
		EpisodeTotalNum: req.EpisodeTotalNum,
		AirWeekday:      req.AirWeekday,
		LastAirEpisode:  oldBangumi.LastAirEpisode,
		PosterURL:       oldBangumi.PosterURL,
		BackdropURL:     oldBangumi.BackdropURL,
		Overview:        overview,
		Genres:          genres,
		Name:            oldBangumi.Name,
		Year:            oldBangumi.Year,
		CreatedAt:       oldBangumi.CreatedAt,
	}
	return errors.WithMessage(s.repo.Save(ctx, bangumi), "更新订阅失败")
}

func validateUpdateSubscribeReq(req UpdateSubscribeReq) error {
	if req.SubscriptionID == "" {
		return errs.NewBadRequest("订阅ID不能为空")
	}
	return nil
}

func (s *Subscriber) DeleteSubscription(ctx context.Context, id string) error {
	if id == "" {
		return errs.NewBadRequest("订阅ID不能为空")
	}
	return s.repo.Delete(ctx, id)
}

// Close 关闭订阅器，停止所有后台任务
func (s *Subscriber) Close() error {
	if s.stop != nil {
		s.stop()
	}
	if s.rssTicker != nil {
		s.rssTicker.Stop()
	}
	return nil
}

// runRSSChecker 启动RSS检查器，定期检查RSS更新
func (s *Subscriber) runRSSChecker(ctx context.Context) {
	interval := s.config.RSSCheckInterval
	s.rssTicker = time.NewTicker(time.Duration(interval) * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.rssTicker.C:
			s.handleSubscriptions(ctx)
		}
	}
}

// Reload 重载配置
func (s *Subscriber) Reload(config interface{}) error {
	cfg, ok := config.(*Config)
	if !ok {
		return errors.New("配置类型错误")
	}
	if s.rssTicker != nil {
		s.rssTicker.Reset(time.Duration(cfg.RSSCheckInterval) * time.Minute)
	}
	s.config = *cfg
	return nil
}

// handleSubscriptions 检查RSS更新并下载新的种子
func (s *Subscriber) handleSubscriptions(ctx context.Context) {
	// 获取所有活跃的订阅
	active := true
	bangumis, err := s.repo.List(ctx, ListBangumiReq{
		Active: &active, // 只获取活跃的订阅
	})
	if err != nil {
		log.Errorf(ctx, "获取订阅列表失败: %v", err)
		return
	}

	for _, bangumi := range bangumis {
		// 处理单个订阅
		err = s.handleBangumiSubscription(ctx, &bangumi)
		if err != nil {
			log.Error(ctx, err)
			continue
		}
	}
}

func (s *Subscriber) handleBangumiSubscription(ctx context.Context, bangumi *Bangumi) error {
	// 解析RSS
	rss, err := s.rssParser.Parse(ctx, bangumi.RSSLink)
	if err != nil {
		return fmt.Errorf("解析RSS失败 [%s]: %w", bangumi.Name, err)
	}

	var errs *multierror.Error
	// 处理RSS中的每个item
	for _, item := range rss.Items {
		// 根据包含和排除规则过滤
		if !s.matchesFilters(ctx, item.GUID, bangumi.IncludeRegs, bangumi.ExcludeRegs) {
			continue
		}

		// 检查是否已经下载过
		downloaded, err := s.isAlreadyDownloaded(ctx, bangumi.SubscriptionID, item.GUID)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("检查下载状态失败 [%s]: %w", item.GUID, err))
			continue
		}

		if downloaded {
			log.Infof(ctx, "种子已经处理，跳过 [%s]", item.GUID)
			continue
		}

		hash, err := extractHashFromMagnet(item.TorrentLink)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("提取哈希值失败 [%s]: %w", item.GUID, err))
			continue
		}

		// 执行下载
		err = s.downloader.Download(ctx, downloader.DownloadReq{
			TorrentLink:    item.TorrentLink,
			SavePath:       fmt.Sprintf("/%s/Season %d/", bangumi.Name, bangumi.Season),
			DownloadType:   downloader.DownloadTypeTV,
			SubscriptionID: bangumi.SubscriptionID,
			TMDBID:         bangumi.TMDBID,
			Hash:           hash,
			RSSGUID:        item.GUID,
		})
		if nerr := s.notifier.NoticeSubscriptionUpdated(ctx, notice.NoticeSubscriptionUpdatedReq{
			BangumiName:  bangumi.Name,
			Season:       bangumi.Season,
			ReleaseGroup: bangumi.ReleaseGroup,
			RSSGUID:      item.GUID,
			Poster:       bangumi.PosterURL,
			Error:        err,
		}); nerr != nil {
			log.Warnf(ctx, "通知订阅更新失败 [%s]: %v", item.GUID, nerr)
		}
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("下载失败 [%s]: %w", item.GUID, err))
			continue
		}

		// 标记RSS条目为已处理
		if err := s.rssRecord.MarkProcessed(ctx, bangumi.SubscriptionID, item.GUID); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("标记RSS条目为已处理失败 [%s]: %w", item.GUID, err))
			continue
		}

		log.Infof(ctx, "成功添加下载任务 [%s]", item.GUID)
	}
	return errs.ErrorOrNil()
}

// matchesFilters 检查是否符合过滤规则
func (s *Subscriber) matchesFilters(ctx context.Context, guid string, includeRegs, excludeRegs []string) bool {
	includeRegs = append(includeRegs, s.config.IncludeRegs...)
	excludeRegs = append(excludeRegs, s.config.ExcludeRegs...)
	// 优先检查排除规则
	for _, reg := range excludeRegs {
		matched, err := regexp.MatchString(reg, guid)
		if err == nil && matched {
			log.Debugf(ctx, "%s 符合排除规则: %s", guid, reg)
			return false // 符合排除规则，直接返回false
		}
	}

	// 如果没有包含规则，默认包含所有
	if len(includeRegs) == 0 {
		return true
	}

	// 有包含规则，必须全部匹配
	for _, reg := range includeRegs {
		matched, err := regexp.MatchString(reg, guid)
		if err != nil || !matched {
			log.Debugf(ctx, "%s 不符合包含规则: %s", guid, reg)
			return false // 不符合包含规则，返回false
		}
	}

	return true
}

// extractHashFromMagnet 从磁力链接中提取哈希值
func extractHashFromMagnet(magnetLink string) (string, error) {
	// 获取URL的最后一部分
	base := path.Base(magnetLink)

	// 去除.torrent后缀
	hash := strings.TrimSuffix(base, ".torrent")

	return hash, nil
}

// isAlreadyDownloaded 检查种子是否已经下载过
func (s *Subscriber) isAlreadyDownloaded(ctx context.Context, subscriptionID string, guid string) (bool, error) {
	return s.rssRecord.IsProcessed(ctx, subscriptionID, guid)
}

// UpdateLastAirEpisode 更新番剧的最新集数
func (s *Subscriber) UpdateLastAirEpisode(ctx context.Context, subscriptionID string, episode int) error {
	return s.repo.UpdateLastAirEpisode(ctx, subscriptionID, episode)
}

// GetRSSMatch 获取RSS匹配
func (s *Subscriber) GetRSSMatch(ctx context.Context, subscriptionID string) ([]RSSMatch, error) {
	bangumi, err := s.repo.Get(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("获取订阅失败: %w", err)
	}
	rss, err := s.rssParser.Parse(ctx, bangumi.RSSLink)
	if err != nil {
		return nil, fmt.Errorf("解析RSS失败: %w", err)
	}

	// 获取所有已处理的GUID
	processedGUIDs, err := s.rssRecord.ListProcessedGUIDs(ctx, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("获取已处理GUID失败: %w", err)
	}

	// 转换为map便于查找
	processedMap := make(map[string]bool)
	for _, guid := range processedGUIDs {
		processedMap[guid] = true
	}

	matches := make([]RSSMatch, 0, len(rss.Items))
	for _, item := range rss.Items {
		matches = append(matches, RSSMatch{
			GUID:      item.GUID,
			Match:     s.matchesFilters(ctx, item.GUID, bangumi.IncludeRegs, bangumi.ExcludeRegs),
			Processed: processedMap[item.GUID],
		})
	}
	return matches, nil
}

// HandleBangumiSubscription 运行订阅下载任务
func (s *Subscriber) HandleBangumiSubscription(ctx context.Context, subscriptionID string) error {
	bangumi, err := s.repo.Get(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("获取订阅失败: %w", err)
	}
	return s.handleBangumiSubscription(ctx, &bangumi)
}

// MarkRSSRecord 标记RSS记录
func (s *Subscriber) MarkRSSRecord(ctx context.Context, req MarkRSSRecordReq) error {
	if req.Processed {
		return s.rssRecord.MarkProcessed(ctx, req.SubscriptionID, req.GUIDs...)
	}
	return s.rssRecord.DeleteProcessed(ctx, req.SubscriptionID, req.GUIDs...)
}

// GetSubscriptionCalendar 获取订阅日历
func (s *Subscriber) GetSubscriptionCalendar(ctx context.Context) (map[time.Weekday][]CalendarItem, error) {
	active := true
	bangumis, err := s.repo.List(ctx, ListBangumiReq{
		Active: &active, // 只获取活跃的订阅
	})
	if err != nil {
		return nil, fmt.Errorf("获取订阅列表失败: %w", err)
	}

	calendar := make(map[time.Weekday][]CalendarItem)
	for _, bangumi := range bangumis {
		calendar[bangumi.AirWeekday] = appendIfNotExists(calendar[bangumi.AirWeekday], CalendarItem{
			BangumiName: bangumi.Name,
			PosterURL:   bangumi.PosterURL,
			Season:      bangumi.Season,
		})
	}
	return calendar, nil
}

// 这里的数据不会太多，使用map嵌套的性能提升也不明显，反而增加了代码复杂度
func appendIfNotExists(slice []CalendarItem, item CalendarItem) []CalendarItem {
	for _, existing := range slice {
		if existing.BangumiName == item.BangumiName && existing.Season == item.Season { // 比较 BangumiName 和 Season 是否完全相同
			return slice // 如果已存在，直接返回原切片
		}
	}
	return append(slice, item) // 如果不存在，追加并返回新切片
}

// StopSubscription implements Interface.
func (s *Subscriber) StopSubscription(ctx context.Context, id string) error {
	return s.repo.StopSubscription(ctx, id)
}

// AutoStopSubscription 是否自动停止订阅
func (s *Subscriber) AutoStopSubscription(ctx context.Context, id string) bool {
	return s.config.AutoStop
}
