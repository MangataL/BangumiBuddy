package magnet

import (
	"time"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/types"
)

// ParseTorrentRsp 解析种子响应
type ParseTorrentRsp struct {
	Torrent Torrent `json:"torrent"`
	TMDBID  string  `json:"tmdbID"`
}

// Torrent 种子文件
type Torrent struct {
	Hash  string        `json:"hash"`
	Name  string        `json:"name"`
	Files []TorrentFile `json:"files"`
	Size  int64         `json:"size"`
}

// TorrentFile 种子文件
type TorrentFile struct {
	FileName string `json:"fileName"`
	Season   int    `json:"season"`
	Episode  int    `json:"episode"`
	Media    bool   `json:"media"`
	Download bool   `json:"download"`
	LinkFile string `json:"linkFile"`
}

// AddTaskReq 添加下载任务请求
type AddTaskReq struct {
	MagnetLink string                  `json:"magnetLink"`
	Type       downloader.DownloadType `json:"type"`
}

// UpdateTaskReq 确认下载任务请求
type UpdateTaskReq struct {
	TaskID           string  `json:"taskID"`
	TMDBID           int     `json:"tmdbID"`
	ReleaseGroup     string  `json:"releaseGroup"`
	Torrent          Torrent `json:"torrent"`
	ContinueDownload *bool   `json:"continueDownload"`
}

// AddSubtitlesReq 添加字幕请求
type AddSubtitlesReq struct {
	SubtitleDir     string `json:"subtitleDir"`
	EpisodeLocation string `json:"episodeLocation"`
	TaskID          string `json:"-"`
	DstDir          string `json:"dstDir"`
}

// Task 下载任务概览
type Task struct {
	TaskID       string                  `json:"taskID"`
	MagnetLink   string                  `json:"magnetLink"`
	Torrent      Torrent                 `json:"torrent"`
	CreatedAt    time.Time               `json:"createdAt"`
	DownloadType downloader.DownloadType `json:"downloadType"`
	Meta         Meta                    `json:"meta"`
	Status       TaskStatus              `json:"status"`
}

// TaskStatus 下载任务状态
type TaskStatus string

const (
	TaskStatusWaitingForParsing      TaskStatus = "waiting for parsing"
	TaskStatusWaitingForConfirmation TaskStatus = "waiting for confirmation"
	TaskStatusInitSuccess            TaskStatus = "init success"
)

// Meta 元数据信息
type Meta struct {
	ChineseName  string `json:"chineseName"`
	Year         string `json:"year"`
	TMDBID       int    `json:"tmdbID"`
	ReleaseGroup string `json:"releaseGroup"`
}

// ListTasksReq 列出下载任务请求
type ListTasksReq struct {
	TaskIDs     []string    `json:"taskIDs"`
	TorrentName string      `json:"torrentName"`
	Page        types.Page  `json:"page"`
	Order       types.Order `json:"order"`
}
