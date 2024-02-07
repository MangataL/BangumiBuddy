package downloader

import "time"

// DownloadReq 下载请求参数
type DownloadReq struct {
	TorrentLink    string       // 种子链接
	Hash           string       // 种子哈希值
	SavePath       string       // 保存路径
	SubscriptionID string       // 订阅ID，通过订阅下载的会设置这个字段
	TMDBID         int          // TMDBID，直接下载的会设置这个字段
	DownloadType   DownloadType // 下载类型
	RSSGUID        string       // 标记是哪个RSS项
}

// DownloadType 下载类型
type DownloadType string

const (
	DownloadTypeTV    DownloadType = "tv"
	DownloadTypeMovie DownloadType = "movie"
)

// DownloadStatus 下载状态信息
type DownloadStatus struct {
	Hash          string        // 种子哈希值
	Name          string        // 种子名称
	Progress      float64       // 下载进度，0-1
	DownloadSpeed int64         // 下载速度，单位 bytes/s
	Status        TorrentStatus // 下载状态，如downloading, completed等
	Error         string        // 错误信息，如果有的话
	Cost          time.Duration // 下载耗时
	Size          int64         // 种子大小，单位 bytes
}

// TorrentFilter 种子过滤条件
type TorrentFilter struct {
	Statuses       []TorrentStatus // 状态过滤
	SubscriptionID string          // 按订阅ID过滤
	TMDBID         int             // 按TMDBID过滤
	Page           Page            // 分页
	StartTime      time.Time       // 按创建时间过滤
	EndTime        time.Time       // 按更新时间过滤
	Order          Order           // 排序
}

type Page struct {
	Num  int
	Size int
}

func (p *Page) Empty() bool {
	return p.Num == 0 && p.Size == 0
}

type Order struct {
	Field string
	Way   string
}

func (o *Order) Empty() bool {
	return o.Field == "" && o.Way == ""
}

// Torrent 种子文件
type Torrent struct {
	Hash           string        // 哈希值
	Name           string        // 种子名称
	Path           string        // 文件路径
	Status         TorrentStatus // 种子状态
	StatusDetail   string        // 种子状态详情，一般用于存储错误信息
	SubscriptionID string        // 订阅ID，通过订阅下载的会设置这个字段
	TMDBID         int           // TMDBID，直接下载的，会设置这个字段，用于获取种子的元数据信息
	TransferType   string        // 转移类型，用于获取转移文件
	RSSGUID        string        // 标记是哪个RSS项
	CreatedAt      time.Time     // 创建时间
	UpdatedAt      time.Time     // 更新时间
}

// TorrentStatus 种子状态
type TorrentStatus string

const (
	// TorrentStatusDownloading 下载中
	TorrentStatusDownloading TorrentStatus = "downloading"
	// TorrentStatusDownloadPaused 下载暂停
	TorrentStatusDownloadPaused TorrentStatus = "downloadPaused"
	// TorrentStatusDownloadError 下载错误
	TorrentStatusDownloadError TorrentStatus = "downloadError"
	// TorrentStatusDownloaded 已下载
	TorrentStatusDownloaded TorrentStatus = "downloaded"
	// TorrentStatusTransferred 已转移
	TorrentStatusTransferred TorrentStatus = "transferred"
	// TorrentStatusTransferredError 转移错误
	TorrentStatusTransferredError TorrentStatus = "transferredError"
)
