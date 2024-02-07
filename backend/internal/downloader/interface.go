package downloader

import "context"

//go:generate mockgen -destination interface_mock.go -source $GOFILE -package $GOPACKAGE

// Interface 下载器接口，负责执行下载任务
type Interface interface {
	// Download 下载种子文件
	// req 包含下载所需的信息，如种子链接、保存路径等
	Download(ctx context.Context, req DownloadReq) error

	// GetDownloadStatuses 批量获取下载任务状态
	// 如果 hashes 为空，则返回所有下载任务的状态
	GetDownloadStatuses(ctx context.Context, hashes []string) ([]DownloadStatus, error)

	// GetTorrentFileNames 获取种子文件内容
	GetTorrentFileNames(ctx context.Context, hash string) ([]string, error)

	// DeleteTorrent 删除种子文件
	DeleteTorrent(ctx context.Context, hash string) error
}

type Closer interface {
	Close()
}

// TorrentOperator 种子文件操作接口，负责管理种子文件的元数据和状态
type TorrentOperator interface {
	// Save 保存种子文件信息
	Save(ctx context.Context, torrent Torrent) error

	// SetTorrentStatus 设置种子文件状态
	SetTorrentStatus(ctx context.Context, hash string, status TorrentStatus, detail string, opts *SetTorrentStatusOptions) error

	// Get 获取种子文件信息
	Get(ctx context.Context, hash string) (Torrent, error)

	// List 列出所有种子文件
	// filter 可以指定过滤条件，如状态、番剧名等
	List(ctx context.Context, filter TorrentFilter) ([]Torrent, int, error)

	// Delete 删除种子文件
	Delete(ctx context.Context, hash string) error
}

// SetTorrentStatusOptions 设置种子文件状态选项
type SetTorrentStatusOptions struct {
	TransferType string
}
