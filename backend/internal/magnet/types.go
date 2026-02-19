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
	FileName string           `json:"fileName"`
	Season   int              `json:"season"`
	Episode  int              `json:"episode"`
	Media    bool             `json:"media"`
	Download bool             `json:"download"`
	LinkFile string           `json:"linkFile"`
	Meta     *TorrentFileMeta `json:"meta,omitempty"` // 文件级别元数据，为空时使用任务级别
}

// TorrentFileMeta 文件级别元数据
type TorrentFileMeta struct {
	MediaType   downloader.DownloadType `json:"mediaType"` // tv 或 movie
	ChineseName string                  `json:"chineseName"`
	Year        string                  `json:"year"`
	TMDBID      int                     `json:"tmdbID"`
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
	SubtitleFiles    map[string]string `json:"subtitleFiles"`
	PreserveOriginal bool              `json:"preserveOriginal"`
}

// AddSubtitlesResp 添加字幕响应
type AddSubtitlesResp struct {
	SuccessCount  int               `json:"successCount"`
	FailedDetails map[string]string `json:"failedDetails"`
}

// PreviewAddSubtitlesReq 预览添加字幕请求
type PreviewAddSubtitlesReq struct {
	SubtitlePath    string `json:"subtitlePath"`
	EpisodeLocation string `json:"episodeLocation"`
	EpisodeOffset   *int   `json:"episodeOffset,omitempty"`
	Season          *int   `json:"season"`
	TaskID          string `json:"-"`
	DstPath         string `json:"dstPath"`
	ExtensionLevel  *int   `json:"extensionLevel"`
}

// PreviewAddSubtitlesResp 预览添加字幕响应
type PreviewAddSubtitlesResp struct {
	SubtileFiles map[string]AddSubtitlesResult `json:"subtitleFiles"`
}

type AddSubtitlesResult struct {
	SubtitleFile  string `json:"subtitleFile"`
	NewFileName   string `json:"newFileName,omitempty"`
	TargetPath    string `json:"targetPath,omitempty"`
	Error         string `json:"error,omitempty"`
	MediaFileName string `json:"mediaFileName,omitempty"`
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
