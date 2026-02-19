package scrape

import (
	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/pkg/errs"
)

type DownloadType = downloader.DownloadType

// ScrapeStatus 刮削状态
type ScrapeStatus string

const (
	// ScrapeStatusPending 未刮削
	ScrapeStatusPending ScrapeStatus = "pending"
	// ScrapeStatusMissingTitle 缺少标题
	ScrapeStatusMissingTitle ScrapeStatus = "missingTitle"
	// ScrapeStatusMissingPlot 缺少简介
	ScrapeStatusMissingPlot ScrapeStatus = "missingPlot"
	// ScrapeStatusMissingImage 缺少海报
	ScrapeStatusMissingImage ScrapeStatus = "missingImage"
)

type AddMetadataFillTaskReq struct {
	TMDBID      int
	FilePath    string
	BangumiName string
	PosterURL   string
	Season      int
	Episode     int
}

type Config struct {
	Enable        bool `mapstructure:"enable" json:"enable" default:"false"`
	CheckInterval int  `mapstructure:"check_interval" json:"checkInterval" default:"24"` // 默认1天
}

type MetadataCheckTask struct {
	ID          uint           `json:"id"`
	TMDBID      int            `json:"tmdbID"`
	FilePath    string         `json:"filePath"`
	BangumiName string         `json:"bangumiName"`
	PosterURL   string         `json:"posterURL"`
	Season      int            `json:"season"`
	Episode     int            `json:"episode"`
	Statuses    []ScrapeStatus `json:"statuses"`
}

var ErrTaskNotFound = errs.NewNotFound("任务未找到")
