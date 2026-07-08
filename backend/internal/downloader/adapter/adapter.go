package adapter

import (
	"context"
	"errors"
	"sync"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/downloader/qbittorrent"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

var _ downloader.Downloader = &Adapter{}

type Adapter struct {
	mu         sync.RWMutex
	downloader downloader.Downloader
}

func NewAdapter(config Config) *Adapter {
	adapter := &Adapter{}
	if err := adapter.Reload(&config); err != nil {
		log.Errorf(context.Background(), "初始化下载器失败: %v", err)
		adapter.downloader = &downloader.Empty{}
	}
	return adapter
}

type Config struct {
	DownloadType string             `mapstructure:"download_type" json:"downloadType"`
	QBitTorrent  qbittorrent.Config `mapstructure:"qbittorrent" json:"qbittorrent"`
}

func (a *Adapter) Reload(config interface{}) error {
	cfg, ok := config.(*Config)
	if !ok {
		return errors.New("配置类型错误")
	}
	switch cfg.DownloadType {
	case "qbittorrent":
		a.setDownloader(qbittorrent.NewQBittorrent(cfg.QBitTorrent))
	default:
		a.setDownloader(&downloader.Empty{})
	}
	return nil
}

func (a *Adapter) GetDownloadStatuses(ctx context.Context, hashes []string) ([]downloader.DownloadStatus, error) {
	return a.currentDownloader().GetDownloadStatuses(ctx, hashes)
}

func (a *Adapter) GetTorrentFileNames(ctx context.Context, hash string) ([]string, error) {
	return a.currentDownloader().GetTorrentFileNames(ctx, hash)
}

func (a *Adapter) AddTorrent(ctx context.Context, torrentLink, savePath, stopCondition string) error {
	return a.currentDownloader().AddTorrent(ctx, torrentLink, savePath, stopCondition)
}

func (a *Adapter) SetLocation(ctx context.Context, hash, savePath string) error {
	return a.currentDownloader().SetLocation(ctx, hash, savePath)
}

func (a *Adapter) GetTorrentName(ctx context.Context, hash string) (string, error) {
	return a.currentDownloader().GetTorrentName(ctx, hash)
}

func (a *Adapter) ListTorrentsStatus(ctx context.Context) ([]downloader.DownloadStatus, error) {
	return a.currentDownloader().ListTorrentsStatus(ctx)
}

func (a *Adapter) SetTorrentFilePriorities(
	ctx context.Context,
	hash string,
	files []downloader.TorrentFileSelection,
) error {
	return a.currentDownloader().SetTorrentFilePriorities(ctx, hash, files)
}

func (a *Adapter) DeleteTorrent(ctx context.Context, hash string) error {
	return a.currentDownloader().DeleteTorrent(ctx, hash)
}

func (a *Adapter) ContinueDownload(ctx context.Context, hash string) error {
	return a.currentDownloader().ContinueDownload(ctx, hash)
}

func (a *Adapter) setDownloader(d downloader.Downloader) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.downloader = d
}

func (a *Adapter) currentDownloader() downloader.Downloader {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.downloader == nil {
		return &downloader.Empty{}
	}
	return a.downloader
}
