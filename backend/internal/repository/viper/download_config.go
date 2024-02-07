package viper

import (
	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/downloader/adapter"
)

const (
	ComponentNameDownloader      = ComponentName("download.downloader")
	ComponentNameDownloadManager = ComponentName("download.manager")
)

func (r *Repo) GetDownloaderConfig() (adapter.Config, error) {
	var config adapter.Config
	if err := r.GetComponentConfig(ComponentNameDownloader, &config); err != nil {
		return adapter.Config{}, err
	}
	return config, nil
}

func (r *Repo) SetDownloaderConfig(config *adapter.Config) error {
	return r.SetComponentConfig(ComponentNameDownloader, config)
}

func (r *Repo) GetDownloadManagerConfig() (downloader.Config, error) {
	var config downloader.Config
	if err := r.GetComponentConfig(ComponentNameDownloadManager, &config); err != nil {
		return downloader.Config{}, err
	}
	return config, nil
}

func (r *Repo) SetDownloadManagerConfig(config *downloader.Config) error {
	return r.SetComponentConfig(ComponentNameDownloadManager, config)
}
