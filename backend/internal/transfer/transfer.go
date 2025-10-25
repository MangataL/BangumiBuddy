package transfer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mholt/archives"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/magnet"
	"github.com/MangataL/BangumiBuddy/internal/notice"
	"github.com/MangataL/BangumiBuddy/internal/scrape"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
	"github.com/MangataL/BangumiBuddy/pkg/utils"
)

var _ Interface = (*Transfer)(nil)

func NewTransfer(dep Dependency) *Transfer {
	ctx, cancel := context.WithCancel(context.Background())

	transfer := &Transfer{
		stop:            cancel,
		config:          dep.Config,
		torrentOperator: dep.TorrentOperator,
		downloader:      dep.Downloader,
		subscriber:      dep.Subscriber,
		episodeParser:   dep.EpisodeParser,
		transferFiles:   dep.TransferFiles,
		notifier:        dep.Notifier,
		magnetManager:   dep.MagnetManager,
		fontSubsetter:   dep.FontOperator,
		scraper:         dep.Scraper,
	}

	go transfer.run(ctx)
	return transfer
}

type Dependency struct {
	Config
	downloader.TorrentOperator
	EpisodeParser
	Downloader    downloader.Interface
	Subscriber    subscriber.Interface
	TransferFiles TransferFilesRepo
	Notifier      notice.Notifier
	MagnetManager magnet.Interface
	FontOperator  subtitle.Subsetter
	Scraper       scrape.Interface
}

type EpisodeParser interface {
	Parse(ctx context.Context, fileName string) (int, error)
}

type FileTransfer interface {
	Transfer(ctx context.Context, src, dst string) (originFile string, err error)
}

type TransferFilesRepo interface {
	Set(ctx context.Context, fileTransferred FileTransferred) error
	Get(ctx context.Context, req GetFileTransferredReq) (FileTransferred, error)
	List(ctx context.Context, req ListFileTransferredReq) ([]FileTransferred, error)
	Del(ctx context.Context, req DeleteFileTransferredReq) error
}

var ErrFileTransferredNotFound = errors.New("文件转移记录未找到")

type Config struct {
	Interval             int                  `mapstructure:"interval" json:"interval" default:"1"`
	TVPath               string               `mapstructure:"tv_path" json:"tvPath"`
	TVFormat             string               `mapstructure:"tv_format" json:"tvFormat" default:"{name}/Season {season}/{name} {season_episode}"`
	MoviePath            string               `mapstructure:"movie_path" json:"moviePath"`
	MovieFormat          string               `mapstructure:"movie_format" json:"movieFormat" default:"{name} ({year})"`
	TransferType         string               `mapstructure:"transfer_type" json:"transferType"`
	SubtitleRename       SubtitleRenameConfig `mapstructure:"subtitle_rename" json:"subtitleRename"`
	EnableSubtitleSubset bool                 `mapstructure:"enable_subtitle_subset" json:"enableSubtitleSubset"`
	IgnoreSubsetError    bool                 `mapstructure:"ignore_subset_error" json:"ignoreSubsetError"`
}

type SubtitleRenameConfig struct {
	Enabled                     bool     `mapstructure:"enabled" json:"enabled"`
	SimpleChineseRenameExt      string   `mapstructure:"simple_chinese_rename_ext" json:"simpleChineseRenameExt" default:".zh"`
	SimpleChineseExts           []string `mapstructure:"simple_chinese_exts" json:"simpleChineseExts" default:"[\".zh-cn\",\".zh-hans\",\".sc\"]"`
	TraditionalChineseExts      []string `mapstructure:"traditional_chinese_exts" json:"traditionalChineseExts" default:"[\".zh-tw\",\".zh-hk\",\".zh-hant\",\".tc\"]"`
	TraditionalChineseRenameExt string   `mapstructure:"traditional_chinese_rename_ext" json:"traditionalChineseRenameExt" default:".zh-hant"`
}

type Transfer struct {
	ticker          *time.Ticker
	stop            func()
	config          Config
	torrentOperator downloader.TorrentOperator
	downloader      downloader.Interface
	magnetManager   magnet.Interface
	subscriber      subscriber.Interface
	episodeParser   EpisodeParser
	transferFiles   TransferFilesRepo
	notifier        notice.Notifier
	fontSubsetter   subtitle.Subsetter
	scraper         scrape.Interface
}

func (t *Transfer) run(ctx context.Context) {
	interval := t.config.Interval
	t.ticker = time.NewTicker(time.Duration(interval) * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.ticker.C:
			t.transferDownloaded(ctx)
		}
	}
}

func (t *Transfer) transferDownloaded(ctx context.Context) {
	torrents, _, err := t.torrentOperator.List(ctx, downloader.TorrentFilter{
		Statuses: []downloader.TorrentStatus{downloader.TorrentStatusDownloaded, downloader.TorrentStatusTransferredError},
	})
	if err != nil {
		log.Errorf(ctx, "transfer: list torrents failed: %v", err)
		return
	}
	for _, torrent := range torrents {
		if err := t.transferTorrent(ctx, torrent); err != nil {
			log.Errorf(ctx, "转移种子(%s)失败: %v", torrent.Hash, err)
		}
	}
}

type transferError struct {
	errs    []error
	fileNum int
}

func (t *transferError) Append(err error, fileName string) {
	if t.fileNum == 1 {
		t.errs = append(t.errs, err)
	} else {
		t.errs = append(t.errs, fmt.Errorf("文件 %s 转移失败: %v", fileName, err))
	}
}

func (t *transferError) ErrorOrNil() error {
	if len(t.errs) == 0 {
		return nil
	}
	errMsg := ""
	for _, err := range t.errs {
		errMsg += err.Error() + ";"
	}
	return errors.New(errMsg)
}

func (t *Transfer) transferTorrent(ctx context.Context, torrent downloader.Torrent) (err error) {
	if torrent.SubscriptionID != "" {
		err = t.transferTorrentForSubscribe(ctx, torrent)
	} else {
		err = t.transferTorrentForTask(ctx, torrent)
	}
	transferState := downloader.TorrentStatusTransferred
	var transferDetail string
	if err != nil {
		transferState = downloader.TorrentStatusTransferredError
		transferDetail = err.Error()
	}
	if err := t.torrentOperator.SetTorrentStatus(ctx, torrent.Hash, transferState, transferDetail, &downloader.SetTorrentStatusOptions{
		TransferType: t.config.TransferType,
	}); err != nil {
		return errors.WithMessagef(err, "设置种子(%s)转移状态失败", torrent.Hash)
	}
	return err
}

func (t *Transfer) transferTorrentForSubscribe(ctx context.Context, torrent downloader.Torrent) error {
	fileNames := torrent.FileNames
	transferErr := &transferError{fileNum: len(fileNames)}
	for _, fileName := range fileNames {
		path := filepath.Join(torrent.Path, fileName)
		if !utils.IsMediaFile(path) {
			continue
		}
		if err := t.transferFileForSubscribe(ctx, torrent, path, fileName); err != nil {
			log.Errorf(ctx, "文件 %s 转移失败： %v", fileName, err)
			transferErr.Append(err, fileName)
		}
	}
	return transferErr.ErrorOrNil()
}

func (t *Transfer) transferFileForSubscribe(ctx context.Context, torrent downloader.Torrent, path, fileName string) error {
	bangumi, err := t.subscriber.Get(ctx, torrent.SubscriptionID)
	if err != nil {
		return errors.WithMessage(err, "文件转移时获取番剧信息失败")
	}
	meta := Meta{
		ChineseName:     bangumi.Name,
		Year:            bangumi.Year,
		Season:          bangumi.Season,
		EpisodeLocation: bangumi.EpisodeLocation,
		EpisodeOffset:   bangumi.EpisodeOffset,
		FileName:        fileName,
		FilePath:        path,
		SubscriptionID:  torrent.SubscriptionID,
		ReleaseGroup:    bangumi.ReleaseGroup,
	}
	checkPriority := func(ctx context.Context, newFileID string) (bool, error) {
		return t.checkPriority(ctx, newFilePriority{
			newFileID: newFileID,
			fileName:  fileName,
			priority:  bangumi.Priority,
		})
	}

	episode, newFilePath, transferd, err := t.transferFileWithCheckers(ctx, meta, checkPriority)
	if (torrent.Status != downloader.TorrentStatusTransferredError && err != nil) || transferd {
		if err := t.notifier.NoticeSubscriptionTransferred(ctx, notice.NoticeSubscriptionTransferredReq{
			RSSGUID:       torrent.RSSGUID,
			FileName:      fileName,
			BangumiName:   bangumi.Name,
			Season:        bangumi.Season,
			ReleaseGroup:  bangumi.ReleaseGroup,
			Poster:        bangumi.PosterURL,
			MediaFilePath: newFilePath,
			Error:         err,
		}); err != nil {
			log.Warnf(ctx, "通知转移失败: %v", err)
		}
	}
	if err != nil {
		return err
	}
	if !transferd {
		return nil
	}
	if err := t.subscriber.UpdateLastAirEpisode(ctx, torrent.SubscriptionID, episode); err != nil {
		log.Warnf(ctx, "更新订阅信息失败: %v", err)
	}

	if !t.subscriber.AutoStopSubscription(ctx, torrent.SubscriptionID) {
		return nil
	}

	if episode == bangumi.EpisodeTotalNum {
		log.Infof(ctx, "番剧(%s S%s)订阅已更新完毕(更新至%s集)，自动停止订阅", bangumi.Name, utils.FormatNumber(bangumi.Season), utils.FormatNumber(episode))
		if err := t.subscriber.StopSubscription(ctx, torrent.SubscriptionID); err != nil {
			log.Warnf(ctx, "停止订阅(%s)失败: %v", torrent.SubscriptionID, err)
		}
	}

	if t.scraper.Enable() {
		if err := t.scraper.AddMetadataFillTask(ctx, scrape.AddMetadataFillTaskReq{
			FilePath:     newFilePath,
			TMDBID:       bangumi.TMDBID,
			DownloadType: downloader.DownloadTypeTV,
		}); err != nil {
			log.Warnf(ctx, "添加元数据填充任务失败: %v", err)
		}
	}
	return nil
}

type newFilePriority struct {
	fileName  string
	priority  int
	newFileID string
}

// 检查优先级，返回是否应该进行转移以及可能的错误
func (t *Transfer) checkPriority(ctx context.Context, newFilePriority newFilePriority) (bool, error) {
	// 从缓存中获取优先级信息
	var priorityToCompare int
	transferred, err := t.transferFiles.Get(ctx, GetFileTransferredReq{
		NewFileID: newFilePriority.newFileID,
	})
	if err != nil {
		if !errors.Is(err, ErrFileTransferredNotFound) {
			return false, errors.WithMessage(err, "获取转移记录失败")
		}
		return true, nil
	} else {
		bangumi, err := t.subscriber.Get(ctx, transferred.SubscriptionID)
		if err != nil {
			if !errors.Is(err, subscriber.ErrSubscriberNotFound) {
				return false, errors.WithMessage(err, "获取番剧信息失败")
			}
			_ = t.transferFiles.Del(ctx, DeleteFileTransferredReq{
				SubscriptionID: transferred.SubscriptionID,
				NewFileID:      transferred.NewFileID,
			})
			return true, nil
		}
		priorityToCompare = bangumi.Priority
	}

	// 比较优先级
	if priorityToCompare > newFilePriority.priority {
		// 缓存的优先级更高，不进行转移
		log.Infof(ctx, "文件 %s 已存在更高优先级的版本，跳过转移", newFilePriority.fileName)
		return false, nil
	}

	// 当前文件优先级更高，删除现有文件
	log.Infof(ctx, "文件 %s 优先级(%d)不低于现有文件优先级(%d)，将覆盖现有文件",
		newFilePriority.fileName, newFilePriority.priority, priorityToCompare)

	if transferred.NewFile == "" {
		log.Warnf(ctx, "转移记录 %s 中没有新文件路径，跳过删除", transferred.NewFileID)
		return true, nil
	}
	// 删除所有前缀相同的文件
	matches, err := utils.FindSameBaseFiles(transferred.NewFile)
	if err != nil {
		log.Warnf(ctx, "删除低优先级文件时查找文件出错: %v", err)
	} else {
		log.Infof(ctx, "由于将转移更高优先级的文件 %s，执行删除低优先级文件 %v", newFilePriority.fileName, matches)
		deleteFiles := make([]string, 0, len(matches))
		for _, file := range matches {
			if filepath.Ext(file) == ".nfo" {
				// 元数据信息没必要删除
				continue
			}
			if err := os.Remove(file); err != nil {
				log.Warnf(ctx, "删除低优先级文件 %s 失败: %v", file, err)
				continue
			}
			deleteFiles = append(deleteFiles, file)
		}
		log.Infof(ctx, "删除低优先级文件 %v 成功", deleteFiles)
	}

	return true, nil
}

type transferChecker func(ctx context.Context, newFileID string) (bool, error)

func (t *Transfer) transferFileWithCheckers(ctx context.Context, meta Meta, checkers ...transferChecker) (int, string, bool, error) {
	log.Infof(ctx, "开始转移文件 %s", meta.FileName)
	var episode int
	if meta.EpisodeLocation == "" {
		var err error
		episode, err = t.episodeParser.Parse(ctx, meta.FileName)
		if err != nil {
			return 0, "", false, errors.WithMessage(err, "解析文件集数失败")
		}
	} else {
		var err error
		episode, err = utils.ParseEpisodeWithLocation(meta.FileName, meta.EpisodeLocation)
		if err != nil {
			return 0, "", false, errors.WithMessage(err, "通过集数定位解析文件集数失败")
		}
	}
	episode += meta.EpisodeOffset

	newFileID := fmt.Sprintf("%s/%s/%s", meta.ChineseName, strconv.Itoa(meta.Season), strconv.Itoa(episode))

	for _, checker := range checkers {
		shouldTransfer, err := checker(ctx, newFileID)
		if err != nil {
			return 0, "", false, errors.WithMessage(err, "检查优先级失败")
		}
		if !shouldTransfer {
			return 0, "", false, nil
		}
	}

	_, newFilePath, err := t.transferFileForTV(ctx, meta, episode, newFileID)
	if err != nil {
		return 0, "", false, errors.WithMessage(err, "转移文件失败")
	}
	return episode, newFilePath, true, nil
}

func (t *Transfer) transferFileForTV(ctx context.Context, meta Meta, episode int, newFileID string) (originFile string, newFilePath string, err error) {
	// 生成新文件路径
	newFilePathWithoutExt := t.generateNewFilePath(t.config.TVFormat, meta, episode)
	originFile, newFilePath, err = t.transferFile(ctx, newFilePathWithoutExt, meta, newFileID)
	return
}

func (t *Transfer) transferFile(ctx context.Context, newFilePathWithoutExt string, meta Meta, newFileID string) (originFile, newFilePath string, err error) {
	newFilePath = newFilePathWithoutExt + filepath.Ext(meta.FileName)
	originFile, err = GetFileTransfer(t.config.TransferType).Transfer(ctx, meta.FilePath, newFilePath)
	if err != nil {
		return "", "", errors.WithMessage(err, "文件转移失败")
	}
	// 查找并转移相关的字幕和音频文件
	if err := t.transferRelatedFiles(ctx, meta, newFilePathWithoutExt); err != nil {
		return "", "", errors.WithMessage(err, "转移相关文件失败")
	}
	// 设置转移记录
	if err := t.transferFiles.Set(ctx, FileTransferred{
		OriginFile:     originFile,
		BangumiName:    meta.ChineseName,
		Season:         meta.Season,
		SubscriptionID: meta.SubscriptionID,
		NewFile:        newFilePath,
		NewFileID:      newFileID,
	}); err != nil {
		log.Warnf(ctx, "更新转移记录失败: %v", err)
	}
	return originFile, newFilePath, nil
}

// 生成新文件的路径，不包含扩展名
func (t *Transfer) generateNewFilePath(format string, meta Meta, episode int) string {
	result := t.replaceCommonVar(format, meta)

	episodeStr := strconv.Itoa(episode)
	result = strings.ReplaceAll(result, "{episode}", episodeStr)

	seasonStr := strconv.Itoa(meta.Season)
	result = strings.ReplaceAll(result, "{season}", seasonStr)

	seasonEpisode := fmt.Sprintf("S%sE%s", utils.FormatNumber(meta.Season), utils.FormatNumber(episode))
	result = strings.ReplaceAll(result, "{season_episode}", seasonEpisode)

	return filepath.Join(t.config.TVPath, result)
}

func (t *Transfer) replaceCommonVar(format string, meta Meta) string {
	result := strings.ReplaceAll(format, "{name}", meta.ChineseName)
	result = strings.ReplaceAll(result, "{year}", meta.Year)
	result = strings.ReplaceAll(result, "{release_group}", meta.ReleaseGroup)
	originName := utils.GetFileBaseName(meta.FileName)
	result = strings.ReplaceAll(result, "{origin_name}", originName)
	return result
}

var audioExtensions = map[string]struct{}{
	".aac":  {},
	".ac3":  {},
	".dts":  {},
	".flac": {},
	".m4a":  {},
	".mka":  {},
	".mp3":  {},
	".ogg":  {},
}

// 转移相关的字幕和音频文件
func (t *Transfer) transferRelatedFiles(ctx context.Context, meta Meta, newFilePathWithoudExt string) error {
	files, err := utils.FindSameBaseFiles(meta.FilePath)
	if err != nil {
		return errors.WithMessage(err, "获取相关文件列表失败")
	}

	fontSubsetter := t.fontSubsetter
	newFiles := make([]string, 0, len(files))
	for _, file := range files {
		if utils.IsSubtitleFile(file) && t.config.EnableSubtitleSubset && notSubsetFile(file) {
			if meta.FontSubsetter != nil {
				fontSubsetter = meta.FontSubsetter
			}
			newFile, err := fontSubsetter.SubsetFont(ctx, file)
			if err != nil {
				if !t.config.IgnoreSubsetError {
					return errors.WithMessagef(err, "子集化字幕文件 %s 失败", file)
				}
				log.Warnf(ctx, "子集化字幕文件 %s 失败: %v，跳过", file, err)
			} else {
				newFiles = append(newFiles, newFile)
			}
		}
		newFiles = append(newFiles, file)
	}

	for _, file := range newFiles {
		if file == meta.FilePath {
			continue
		}
		fileExt := strings.TrimPrefix(file, utils.GetFileBaseName(meta.FilePath))

		// 检查是否是字幕或音频文件
		isSubtitle := utils.IsSubtitleFile(fileExt)
		_, isAudio := audioExtensions[fileExt]

		if !isSubtitle && !isAudio {
			continue
		}

		// 创建新的文件路径
		fileExt = t.makeFileExt(fileExt, isSubtitle)
		newRelatedFilePath := newFilePathWithoudExt + fileExt
		log.Infof(ctx, "转移相关文件 %s 到 %s", file, newRelatedFilePath)
		// 转移相关文件
		if _, err := GetFileTransfer(t.config.TransferType).Transfer(ctx, file, newRelatedFilePath); err != nil {
			return errors.WithMessagef(err, "转移 %s 文件失败", file)
		}
	}

	return nil
}

func notSubsetFile(file string) bool {
	return !strings.Contains(file, subtitle.SubtitleSubsetExt)
}

func (t *Transfer) makeFileExt(fileExt string, isSubtitle bool) string {
	if !isSubtitle || !t.config.SubtitleRename.Enabled {
		return fileExt
	}
	fileExt = strings.ToLower(fileExt)
	for _, ext := range t.config.SubtitleRename.SimpleChineseExts {
		if strings.Contains(fileExt, ext) {
			return strings.ReplaceAll(fileExt, ext, t.config.SubtitleRename.SimpleChineseRenameExt)
		}
	}
	for _, ext := range t.config.SubtitleRename.TraditionalChineseExts {
		if strings.Contains(fileExt, ext) {
			return strings.ReplaceAll(fileExt, ext, t.config.SubtitleRename.TraditionalChineseRenameExt)
		}
	}
	return fileExt
}

func (t *Transfer) transferTorrentForTask(ctx context.Context, torrent downloader.Torrent) error {
	task, err := t.magnetManager.GetTask(ctx, torrent.TaskID)
	if err != nil {
		return errors.WithMessage(err, "获取任务失败")
	}
	taskTorrents := lo.SliceToMap(task.Torrent.Files, func(file magnet.TorrentFile) (string, magnet.TorrentFile) {
		return file.FileName, file
	})
	transferErr := &transferError{fileNum: len(torrent.FileNames)}
	successFilePaths := make(map[string]string)
	fontPath, closeFunc, err := t.getFontPath(ctx, task, torrent.Path)
	defer closeFunc() // 确保临时目录被清理
	if err != nil {
		log.Warnf(ctx, "尝试获取字体库失败，使用默认字体库: %v", err)
	}
	var fontSubsetter subtitle.Subsetter
	if fontPath != "" {
		fontSubsetter, err = t.fontSubsetter.UsingTempFontDir(ctx, fontPath)
		if err != nil {
			log.Warnf(ctx, "使用临时字体目录创建字体库失败: %v", err)
		}
	}

	for _, fileName := range torrent.FileNames {
		file, ok := taskTorrents[fileName]
		if !ok {
			log.Warnf(ctx, "任务中没有找到文件 %s", fileName)
			continue
		}
		if !file.Download {
			log.Infof(ctx, "文件 %s 未下载，跳过转移", fileName)
			continue
		}
		if !file.Media {
			log.Infof(ctx, "文件 %s 不是待入库文件，跳过转移", fileName)
			continue
		}
		meta := Meta{
			ChineseName:   task.Meta.ChineseName,
			Year:          task.Meta.Year,
			FileName:      fileName,
			FilePath:      filepath.Join(torrent.Path, fileName),
			ReleaseGroup:  task.Meta.ReleaseGroup,
			FontSubsetter: fontSubsetter,
		}
		newFileID := fmt.Sprintf("%s-%s", torrent.Hash, fileName)
		var originFile, newFilePath string
		switch task.DownloadType {
		case downloader.DownloadTypeTV:
			originFile, newFilePath, err = t.transferForTaskTV(ctx, meta, file, newFileID)
		case downloader.DownloadTypeMovie:
			originFile, newFilePath, err = t.transferForTaskMovie(ctx, meta, newFileID)
		}

		if err != nil {
			log.Errorf(ctx, "文件 %s 转移失败： %v", fileName, err)
			transferErr.Append(err, fileName)
		}
		successFilePaths[originFile] = newFilePath
	}
	err = transferErr.ErrorOrNil()
	if torrent.Status == downloader.TorrentStatusTransferredError && err != nil {
		return err
	}

	if nerr := t.notifier.NoticeTaskTransferred(ctx, notice.NoticeTaskTransferredReq{
		BangumiName:    task.Meta.ChineseName,
		TorrentName:    torrent.Name,
		Error:          err,
		MediaFilePaths: successFilePaths,
	}); nerr != nil {
		log.Warnf(ctx, "通知任务转移结果失败: %v", nerr)
	}

	return err
}

// isFontArchive 判断文件名是否为包含 "fonts" 的压缩包
func isFontArchive(filename string) bool {
	// 确保只处理文件名，不包含路径
	baseName := filepath.Base(filename)

	// 检查文件名是否包含 "fonts"（不区分大小写）
	if !strings.Contains(strings.ToLower(baseName), "fonts") {
		return false
	}

	// 检查文件扩展名是否为压缩包格式
	ext := strings.ToLower(filepath.Ext(baseName))
	archiveExtensions := map[string]struct{}{
		".zip":     {},
		".rar":     {},
		".7z":      {},
		".tar":     {},
		".gz":      {},
		".tar.gz":  {},
		".bz2":     {},
		".tar.bz2": {},
		".xz":      {},
		".tar.xz":  {},
	}

	// 检查 .tar.gz 等复合扩展名
	if strings.HasSuffix(strings.ToLower(baseName), ".tar.gz") {
		return true
	}
	if strings.HasSuffix(strings.ToLower(baseName), ".tar.bz2") {
		return true
	}
	if strings.HasSuffix(strings.ToLower(baseName), ".tar.xz") {
		return true
	}

	_, isArchive := archiveExtensions[ext]
	return isArchive
}

// extractArchive 解压压缩包到指定目录（支持多种格式）
func extractArchive(ctx context.Context, archivePath, extractDir string) error {
	// 打开压缩文件
	file, err := os.Open(archivePath)
	if err != nil {
		return errors.WithMessagef(err, "打开压缩文件 %s 失败", archivePath)
	}
	defer file.Close()

	// 识别压缩格式
	format, stream, err := archives.Identify(ctx, archivePath, file)
	if err != nil {
		return errors.WithMessagef(err, "识别压缩格式失败: %s", archivePath)
	}

	// 检查是否为支持的解压缩器
	extractor, ok := format.(archives.Extractor)
	if !ok {
		return errors.Errorf("不支持的压缩格式: %s", archivePath)
	}

	// 解压缩文件
	err = extractor.Extract(ctx, stream, func(ctx context.Context, f archives.FileInfo) error {
		targetPath := filepath.Join(extractDir, f.NameInArchive)
		if f.IsDir() {
			return os.MkdirAll(targetPath, os.ModePerm)
		}
		// 创建目标文件所在的目录
		if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
			return err
		}
		// 创建并写入文件
		outFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer outFile.Close()
		srcFile, err := f.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()
		_, err = io.Copy(outFile, srcFile)
		return err
	})
	if err != nil {
		return errors.WithMessagef(err, "解压文件 %s 失败", archivePath)
	}

	return nil
}

// getCommonParent 计算多个目录的公共父目录
func getCommonParent(dirs []string, basePath string) string {
	dirs = lo.Uniq(dirs)
	if len(dirs) == 0 {
		return ""
	}
	if len(dirs) == 1 {
		return dirs[0]
	}

	// 清理basePath
	basePath = filepath.Clean(basePath)

	// 计算每个目录相对于basePath的路径
	var relPaths []string
	for _, dir := range dirs {
		dir = filepath.Clean(dir)

		// 检查目录是否在basePath下
		relPath, err := filepath.Rel(basePath, dir)
		if err != nil || strings.HasPrefix(relPath, "..") {
			// 如果目录不在basePath下，跳过
			continue
		}

		relPaths = append(relPaths, relPath)
	}

	if len(relPaths) == 0 {
		return basePath
	}

	// 找到相对路径的最长公共前缀
	commonPrefix := findLongestCommonPrefix(relPaths)

	// 将公共前缀与basePath组合
	if commonPrefix == "" {
		return basePath
	}

	return filepath.Join(basePath, commonPrefix)
}

// findLongestCommonPrefix 找到字符串数组的最长公共前缀
func findLongestCommonPrefix(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	if len(paths) == 1 {
		return filepath.Dir(paths[0])
	}

	// 分割所有路径
	parts := make([][]string, len(paths))
	for i, path := range paths {
		if path == "." {
			parts[i] = []string{}
		} else {
			parts[i] = strings.Split(path, string(filepath.Separator))
		}
	}

	// 找最短路径长度
	minLen := len(parts[0])
	for _, p := range parts[1:] {
		if len(p) < minLen {
			minLen = len(p)
		}
	}

	// 逐段比较，找到最长公共前缀
	var common []string
	for i := 0; i < minLen; i++ {
		segment := parts[0][i]
		allMatch := true
		for j := 1; j < len(parts); j++ {
			if parts[j][i] != segment {
				allMatch = false
				break
			}
		}
		if allMatch {
			common = append(common, segment)
		} else {
			break
		}
	}

	return filepath.Join(common...)
}

func (t *Transfer) getFontPath(ctx context.Context, task magnet.Task, basePath string) (string, func(), error) {
	var fontDirs []string
	var tempDirs []string
	closeFunc := func() {} // 默认空实现

	// 遍历 task.Torrent.Files
	for _, file := range task.Torrent.Files {
		// 构建完整文件路径
		filePath := filepath.Join(basePath, file.FileName)

		// 检查是否为字体文件
		if utils.IsFontFile(file.FileName) {
			fontDir := filepath.Dir(filePath)
			fontDirs = append(fontDirs, fontDir)
			continue
		}

		// 检查是否为包含 "fonts" 的压缩包
		if isFontArchive(file.FileName) {
			// 在压缩包所在目录创建临时目录
			tempDir, err := os.MkdirTemp(filepath.Dir(filePath), ".fonts_extract_")
			if err != nil {
				log.Warnf(ctx, "创建临时目录失败: %v", err)
				continue
			}
			tempDirs = append(tempDirs, tempDir)

			// 解压到临时目录
			if err := extractArchive(ctx, filePath, tempDir); err != nil {
				log.Warnf(ctx, "解压文件 %s 失败: %v", file.FileName, err)
				// 清理失败的临时目录
				os.RemoveAll(tempDir)
				tempDirs = tempDirs[:len(tempDirs)-1]
				continue
			}

			// 将临时目录本身作为字体目录记录
			fontDirs = append(fontDirs, tempDir)
		}
	}

	// 计算公共父目录
	commonParent := getCommonParent(fontDirs, basePath)

	// 如果有临时目录需要清理，更新 closeFunc
	if len(tempDirs) > 0 {
		closeFunc = func() {
			for _, dir := range tempDirs {
				if err := os.RemoveAll(dir); err != nil {
					log.Warnf(ctx, "清理临时目录 %s 失败: %v", dir, err)
				}
			}
		}
	}

	return commonParent, closeFunc, nil
}

func (t *Transfer) transferForTaskTV(ctx context.Context, meta Meta, file magnet.TorrentFile, newFileID string) (string, string, error) {
	meta.Season = file.Season
	originFile, newFilePath, err := t.transferFileForTV(ctx, meta, file.Episode, newFileID)
	return originFile, newFilePath, err
}

func (t *Transfer) transferForTaskMovie(ctx context.Context, meta Meta, newFileID string) (string, string, error) {
	newPath := filepath.Join(t.config.MoviePath, t.replaceCommonVar(t.config.MovieFormat, meta))
	originFile, newFilePath, err := t.transferFile(ctx, newPath, meta, newFileID)
	return originFile, newFilePath, err
}

func (t *Transfer) Close() {
	t.stop()
}

func (t *Transfer) Reload(config interface{}) error {
	cfg, ok := config.(*Config)
	if !ok {
		return errors.New("配置类型错误")
	}
	if t.ticker != nil {
		t.ticker.Reset(time.Duration(cfg.Interval) * time.Minute)
	}
	t.config = *cfg
	return nil
}

func (t *Transfer) Transfer(ctx context.Context, hash string) error {
	torrent, err := t.torrentOperator.Get(ctx, hash)
	if err != nil {
		return errors.WithMessage(err, "获取种子失败")
	}
	return t.transferTorrent(ctx, torrent)
}

func (t *Transfer) DeleteTransferFile(ctx context.Context, file string) error {
	files, err := utils.FindSameBaseFiles(file)
	if err != nil {
		return errors.WithMessage(err, "获取相关文件列表失败")
	}
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return errors.WithMessagef(err, "删除文件 %s 失败", file)
		}
	}
	if err := t.transferFiles.Del(ctx, DeleteFileTransferredReq{
		NewFile: file,
	}); err != nil {
		log.Warnf(ctx, "删除转移缓存失败: %v", err)
	}
	return nil
}

func (t *Transfer) GetTransferFile(ctx context.Context, filePath string) (string, error) {
	tf, err := t.transferFiles.Get(ctx, GetFileTransferredReq{
		OriginFile: filePath,
	})
	if err != nil {
		return "", errors.WithMessage(err, "获取转移文件失败")
	}
	return tf.NewFile, nil
}

func (t *Transfer) DeleteTransferCache(ctx context.Context, req DeleteFileTransferredReq) error {
	return t.transferFiles.Del(ctx, req)
}
