package downloader

import (
	"context"
	"errors"
)

var _ Downloader = (*Empty)(nil)

var ErrDownloaderNotSet = errors.New("下载器未设置，如已设置，请检查下载器连接情况")

type Empty struct{
}

// AddTorrent implements Downloader.
func (e *Empty) AddTorrent(ctx context.Context, torrentLink string, savePath string, stopCondition string) error {
	return ErrDownloaderNotSet
}

// DeleteTorrent implements Downloader.
func (e *Empty) DeleteTorrent(ctx context.Context, hash string) error {
	return ErrDownloaderNotSet
}

// GetDownloadStatuses implements Downloader.
func (e *Empty) GetDownloadStatuses(ctx context.Context, hashes []string) ([]DownloadStatus, error) {
	return nil, ErrDownloaderNotSet
}

// GetTorrentFileNames implements Downloader.
func (e *Empty) GetTorrentFileNames(ctx context.Context, hash string) ([]string, error) {
	return nil, ErrDownloaderNotSet
}

// GetTorrentName implements Downloader.
func (e *Empty) GetTorrentName(ctx context.Context, hash string) (string, error) {
	return "", ErrDownloaderNotSet
}

// ListTorrentsStatus implements Downloader.
func (e *Empty) ListTorrentsStatus(ctx context.Context) ([]DownloadStatus, error) {
	return nil, nil
}

// SetLocation implements Downloader.
func (e *Empty) SetLocation(ctx context.Context, hash string, savePath string) error {
	return ErrDownloaderNotSet
}

// ContinueDownload implements Downloader.
func (e *Empty) ContinueDownload(ctx context.Context, hash string) error {
	return ErrDownloaderNotSet
}

// GetTorrentSavePath implements Downloader.
func (e *Empty) GetTorrentSavePath(ctx context.Context, hash string) (string, error) {
	return "", ErrDownloaderNotSet
}