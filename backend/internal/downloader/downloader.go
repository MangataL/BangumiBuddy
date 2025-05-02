package downloader

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/MangataL/BangumiBuddy/internal/notice"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

var _ Interface = (*Manager)(nil)

func NewManager(dep Dependency) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		downloader: dep.Downloader,
		torrentOp:  dep.TorrentOperator,
		stop:       cancel,
		config:     dep.Config,
		notifier:   dep.Notifier,
	}

	go m.runMonitor(ctx)
	return m
}

type Manager struct {
	downloader Downloader
	torrentOp  TorrentOperator
	notifier   notice.Notifier
	config     Config

	stop func()
}

type Dependency struct {
	Downloader
	TorrentOperator
	Config
	notice.Notifier
}

type Config struct {
	TVSavePath    string `mapstructure:"tv_save_path" json:"tvSavePath"`
	MovieSavePath string `mapstructure:"movie_save_path" json:"movieSavePath"`
}

type Downloader interface {
	GetTorrentFileNames(ctx context.Context, hash string) ([]string, error)
	GetDownloadStatuses(ctx context.Context, hashes []string) ([]DownloadStatus, error)
	AddTorrent(ctx context.Context, torrentLink, savePath string) error
	SetLocation(ctx context.Context, hash, savePath string) error
	GetTorrentName(ctx context.Context, hash string) (string, error)
	ListTorrentsStatus(ctx context.Context) ([]DownloadStatus, error)
	DeleteTorrent(ctx context.Context, hash string) error
}

func (m *Manager) GetTorrentFileNames(ctx context.Context, hash string) ([]string, error) {
	return m.downloader.GetTorrentFileNames(ctx, hash)
}

func (m *Manager) Download(ctx context.Context, req DownloadReq) error {
	// 获取种子当前状态
	statuses, err := m.GetDownloadStatuses(ctx, []string{req.Hash})
	if err != nil {
		return fmt.Errorf("获取种子状态失败: %w", err)
	}

	// 如果没有找到种子，添加新种子
	if len(statuses) == 0 {
		return m.addNewTorrent(ctx, req)
	}
	savePath := m.getSavePath(req.SavePath, req.DownloadType)
	if err := m.downloader.SetLocation(ctx, req.Hash, savePath); err != nil {
		return fmt.Errorf("更新种子保存路径失败: %w", err)
	}

	statuses, err = m.GetDownloadStatuses(ctx, []string{req.Hash})
	if err != nil {
		return fmt.Errorf("更新种子保存路径后，获取种子状态失败: %w", err)
	}
	status := statuses[0]

	// 创建种子信息
	torrent := Torrent{
		Hash:           req.Hash,
		Path:           savePath,
		SubscriptionID: req.SubscriptionID,
		TMDBID:         req.TMDBID,
		Name:           status.Name,
		RSSGUID:        req.RSSGUID,
	}

	// 根据下载状态设置种子状态
	torrent.Status = status.Status
	if status.Status == TorrentStatusDownloadError {
		torrent.StatusDetail = status.Error
	}

	// 保存种子信息
	if err := m.torrentOp.Save(ctx, torrent); err != nil {
		return fmt.Errorf("保存种子信息失败: %w", err)
	}

	return nil
}

// addNewTorrent 添加新的种子下载任务
func (m *Manager) addNewTorrent(ctx context.Context, req DownloadReq) error {
	if req.TorrentLink == "" {
		return errors.New("未提供种子链接")
	}
	savePath := m.getSavePath(req.SavePath, req.DownloadType)
	if err := m.downloader.AddTorrent(ctx, req.TorrentLink, savePath); err != nil {
		return fmt.Errorf("添加种子下载任务失败: %w", err)
	}

	// 获取种子信息
	name, err := m.downloader.GetTorrentName(ctx, req.Hash)
	if err != nil {
		return fmt.Errorf("获取种子信息失败: %w", err)
	}

	// 创建种子信息
	torrent := Torrent{
		Hash:           req.Hash,
		Path:           savePath,
		Status:         TorrentStatusDownloading,
		SubscriptionID: req.SubscriptionID,
		TMDBID:         req.TMDBID,
		Name:           name,
		RSSGUID:        req.RSSGUID,
	}

	// 保存种子信息
	if err := m.torrentOp.Save(ctx, torrent); err != nil {
		return fmt.Errorf("保存种子信息失败: %w", err)
	}

	return nil
}

func (m *Manager) getSavePath(savePath string, downloadType DownloadType) string {
	switch downloadType {
	case DownloadTypeTV:
		return m.config.TVSavePath + savePath
	case DownloadTypeMovie:
		return m.config.MovieSavePath + savePath
	default:
		return savePath
	}
}

// GetDownloadStatuses 实现Downloader接口的GetDownloadStatuses方法
func (m *Manager) GetDownloadStatuses(ctx context.Context, hashes []string) ([]DownloadStatus, error) {
	return m.downloader.GetDownloadStatuses(ctx, hashes)
}

// runMonitor 启动监控任务
func (m *Manager) runMonitor(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkDownloadStatus()
		case <-ctx.Done():
			return
		}
	}
}

// Close 关闭监控任务
func (m *Manager) Close() {
	m.stop()
}

var (
	downloadStatusTranslate = map[TorrentStatus]struct{}{
		TorrentStatusDownloading:    {},
		TorrentStatusDownloadError:  {},
		TorrentStatusDownloaded:     {},
		TorrentStatusDownloadPaused: {},
	}
	needNoticeStatus = map[TorrentStatus]struct{}{
		TorrentStatusDownloadError: {},
		TorrentStatusDownloaded:    {},
	}
)

// checkDownloadStatus 检查下载状态并更新数据库
func (m *Manager) checkDownloadStatus() {
	ctx := log.NewContext()
	statuses, err := m.downloader.ListTorrentsStatus(ctx)
	if err != nil {
		log.Errorf(ctx, "获取种子列表失败: %v", err)
		return
	}
	if len(statuses) == 0 {
		return
	}
	// 使用WaitGroup等待所有goroutine完成
	var wg sync.WaitGroup

	// 更新每个种子的状态
	for _, status := range statuses {
		wg.Add(1)

		// 为每个种子创建一个goroutine处理状态更新
		go func(status DownloadStatus) {
			defer wg.Done()

			// 获取种子信息
			torrent, err := m.torrentOp.Get(ctx, status.Hash)
			if err != nil {
				log.Errorf(ctx, "获取种子信息失败 [%s]: %v", status.Hash, err)
				return
			}

			if _, ok := downloadStatusTranslate[torrent.Status]; !ok {
				return
			}

			// 确定当前状态
			currentStatus := status.Status
			statusDetail := status.Error

			if torrent.Status != currentStatus ||
				(currentStatus == TorrentStatusDownloadError && torrent.StatusDetail != statusDetail) {
				var fileNames []string
				if currentStatus == TorrentStatusDownloaded {
					fileNames, err = m.downloader.GetTorrentFileNames(ctx, torrent.Hash)
					if err != nil {
						log.Errorf(ctx, "获取种子文件名失败 [%s]: %v", status.Hash, err)
						return
					}
				}
				if err := m.torrentOp.SetTorrentStatus(ctx, torrent.Hash, currentStatus, statusDetail, &SetTorrentStatusOptions{
					FileNames: fileNames,
				}); err != nil {
					log.Errorf(ctx, "更新种子状态失败 [%s]: %v", status.Hash, err)
				}
				if _, ok := needNoticeStatus[currentStatus]; ok {
					if err := m.notifier.NoticeDownloaded(ctx, notice.NoticeDownloadedReq{
						RSSGUID:     torrent.RSSGUID,
						TorrentName: torrent.Name,
						Failed:      currentStatus == TorrentStatusDownloadError,
						FailDetail:  statusDetail,
						Cost:        status.Cost,
						Size:        status.Size,
					}); err != nil {
						log.Errorf(ctx, "通知种子状态失败 [%s-%s]: %v", status.Hash, torrent.Name, err)
					}
				}
			}
		}(status)
	}

	// 等待所有goroutine完成
	wg.Wait()
}

// DeleteTorrent 删除种子文件
func (m *Manager) DeleteTorrent(ctx context.Context, hash string) error {
	if err := m.downloader.DeleteTorrent(ctx, hash); err != nil {
		return fmt.Errorf("删除种子文件失败: %w", err)
	}
	return m.torrentOp.Delete(ctx, hash)
}

func (m *Manager) Reload(config interface{}) error {
	cfg, ok := config.(*Config)
	if !ok {
		return errors.New("配置类型错误")
	}
	m.config = *cfg
	return nil
}
