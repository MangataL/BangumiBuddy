package scrape

import (
	"github.com/pkg/errors"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
)

type DownloadType = downloader.DownloadType

type AddMetadataFillTaskReq struct {
	TMDBID       int
	DownloadType DownloadType
	FilePath     string
}

type Config struct {
	Enable        bool          `mapstructure:"enable" json:"enable" default:"false"`
	CheckInterval int           `mapstructure:"check_interval" json:"checkInterval" default:"24"` // 默认1天
}

type MetadataCheckTask struct {
	TMDBID       int
	FilePath     string
	DownloadType DownloadType
	ImageChecked bool // 图片是否已检查过（避免重复下载）
}

var ErrTaskNotFound = errors.New("任务未找到")
