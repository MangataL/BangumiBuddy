package transfer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/notice"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
	"github.com/MangataL/BangumiBuddy/internal/utils"
	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/Tnze/go.num/v2/zh"
	"github.com/pkg/errors"
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
	Interval     int    `mapstructure:"interval" json:"interval" default:"1"`
	TVPath       string `mapstructure:"tv_path" json:"tvPath"`
	TVFormat     string `mapstructure:"tv_format" json:"tvFormat" default:"{name}/Season {season}/{name} {season_episode}"`
	TransferType string `mapstructure:"transfer_type" json:"transferType"`
}

type Transfer struct {
	ticker          *time.Ticker
	stop            func()
	config          Config
	torrentOperator downloader.TorrentOperator
	downloader      downloader.Interface
	subscriber      subscriber.Interface
	episodeParser   EpisodeParser
	transferFiles   TransferFilesRepo
	notifier        notice.Notifier
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
	fileNames := torrent.FileNames
	transferErr := &transferError{fileNum: len(fileNames)}
	for _, fileName := range fileNames {
		path := torrent.Path + fileName
		if !utils.IsMediaFile(path) {
			continue
		}
		var err error
		if torrent.SubscriptionID != "" {
			err = t.transferForSubscribe(ctx, torrent, path, fileName)
		}
		if err != nil {
			log.Errorf(ctx, "文件 %s 转移失败： %v", fileName, err)
			transferErr.Append(err, fileName)
		}
	}
	transferState := downloader.TorrentStatusTransferred
	var transferDetail string
	if err := transferErr.ErrorOrNil(); err != nil {
		transferState = downloader.TorrentStatusTransferredError
		transferDetail = err.Error()
	}
	if err := t.torrentOperator.SetTorrentStatus(ctx, torrent.Hash, transferState, transferDetail, &downloader.SetTorrentStatusOptions{
		TransferType: t.config.TransferType,
	}); err != nil {
		return errors.WithMessagef(err, "设置种子(%s)转移状态失败", torrent.Hash)
	}
	return transferErr.ErrorOrNil()
}

func (t *Transfer) transferForSubscribe(ctx context.Context, torrent downloader.Torrent, path, fileName string) error {
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

	episode, newFilePath, transferd, err := t.transfer(ctx, meta, checkPriority)
	if (torrent.Status != downloader.TorrentStatusTransferredError && err != nil) || transferd {
		if err := t.notifier.NoticeTransferred(ctx, notice.NoticeTransferredReq{
			RSSGUID:       torrent.RSSGUID,
			FileName:      fileName,
			BangumiName:   bangumi.Name,
			Season:        bangumi.Season,
			ReleaseGroup:  bangumi.ReleaseGroup,
			Poster:        bangumi.PosterURL,
			MediaFilePath: strings.TrimPrefix(newFilePath, t.config.TVPath),
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

func (t *Transfer) transfer(ctx context.Context, meta Meta, checkers ...transferChecker) (int, string, bool, error) {
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
		episode, err = t.parseEpisodeWithLocation(meta.FileName, meta.EpisodeLocation)
		if err != nil {
			return 0, "", false, errors.WithMessage(err, "通过集数定位解析文件集数失败")
		}
	}
	episode += meta.EpisodeOffset

	// 生成新文件路径
	newFilePathWithoutExt := t.generateNewFilePath(t.config.TVFormat, meta, episode)
	newFilePath := newFilePathWithoutExt + filepath.Ext(meta.FileName)
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

	// 转移主媒体文件
	originFile, err := GetFileTransfer(t.config.TransferType).Transfer(ctx, meta.FilePath, newFilePath)
	if err != nil {
		return 0, "", false, errors.WithMessage(err, "文件转移失败")
	}
	// 查找并转移相关的字幕和音频文件
	if err := t.transferRelatedFiles(ctx, meta, newFilePathWithoutExt); err != nil {
		return 0, "", false, errors.WithMessage(err, "转移相关文件失败")
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
	log.Infof(ctx, "转移文件 %s 成功", meta.FileName)
	return episode, newFilePath, true, nil
}

func (t *Transfer) parseEpisodeWithLocation(name string, location string) (int, error) {
	pattern := regexp.QuoteMeta(location)
	pattern = strings.ReplaceAll(pattern, `\{ep\}`, `(\d+|[一二三四五六七八九十百千]+)`)

	reg := regexp.MustCompile(pattern)
	matches := reg.FindStringSubmatch(name)
	if len(matches) < 2 {
		return 0, errors.New("无法从文件名中解析出集数信息")
	}

	epStr := matches[1]

	// 如果是数字直接转换
	if num, err := strconv.Atoi(epStr); err == nil {
		return num, nil
	}

	var ep zh.Uint64
	if _, err := fmt.Sscan(epStr, &ep); err != nil {
		return 0, errors.WithMessage(err, "转换中文集数失败")
	}

	return int(ep), nil
}

// 生成新文件的路径，不包含扩展名
func (t *Transfer) generateNewFilePath(format string, meta Meta, episode int) string {
	result := strings.ReplaceAll(format, "{name}", meta.ChineseName)
	result = strings.ReplaceAll(result, "{year}", meta.Year)
	result = strings.ReplaceAll(result, "{release_group}", meta.ReleaseGroup)

	episodeStr := strconv.Itoa(episode)
	result = strings.ReplaceAll(result, "{episode}", episodeStr)

	seasonStr := strconv.Itoa(meta.Season)
	result = strings.ReplaceAll(result, "{season}", seasonStr)

	seasonEpisode := fmt.Sprintf("S%sE%s", utils.FormatNumber(meta.Season), utils.FormatNumber(episode))
	result = strings.ReplaceAll(result, "{season_episode}", seasonEpisode)

	originName := utils.GetFileBaseName(meta.FileName)
	result = strings.ReplaceAll(result, "{origin_name}", originName)

	return filepath.Join(t.config.TVPath, result)
}

var subtitleExtensions = map[string]struct{}{
	".srt": {},
	".ass": {},
	".ssa": {},
	".sub": {},
	".idx": {},
	".vtt": {},
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
	log.Infof(ctx, "获取相关文件列表 %v, file: %s", files, meta.FilePath)

	for _, file := range files {
		fileName := filepath.Base(file)
		fileExt := filepath.Ext(fileName)

		// 检查是否是字幕或音频文件
		_, isSubtitle := subtitleExtensions[fileExt]
		_, isAudio := audioExtensions[fileExt]

		if !isSubtitle && !isAudio {
			continue
		}

		// 创建新的文件路径
		newRelatedFilePath := newFilePathWithoudExt + fileExt
		log.Infof(ctx, "转移相关文件 %s", newRelatedFilePath)
		// 转移相关文件
		if _, err := GetFileTransfer(t.config.TransferType).Transfer(ctx, file, newRelatedFilePath); err != nil {
			return errors.WithMessagef(err, "转移 %s 文件失败", fileExt)
		}
	}

	return nil
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
