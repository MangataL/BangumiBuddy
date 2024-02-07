package adapter

import (
	"context"
	"errors"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/downloader/qbittorrent"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

var _ downloader.Downloader = &Adapter{}

type Adapter struct {
	downloader.Downloader
}

func NewAdapter(config Config) *Adapter {
	adapter := &Adapter{}
	if err := adapter.Reload(&config); err != nil {
		log.Errorf(context.Background(), "初始化下载器失败: %v", err)
		adapter.Downloader = &downloader.Empty{}
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
		a.Downloader = qbittorrent.NewQBittorrent(cfg.QBitTorrent)
	default:
		a.Downloader = &downloader.Empty{}
	}
	return nil
}
