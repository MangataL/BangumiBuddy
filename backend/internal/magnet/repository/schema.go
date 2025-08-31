package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/magnet"
)

// taskSchema 下载任务数据库模型
type taskSchema struct {
	ID           uint      `gorm:"type:int;primaryKey;autoIncrement"`
	TaskID       string    `gorm:"type:char(36);not null;uniqueIndex"`
	MagnetLink   string    `gorm:"type:varchar(255);not null;default:''"`
	TorrentHash  string    `gorm:"type:char(40);not null;index"`
	TorrentName  string    `gorm:"type:varchar(255);not null;default:''"`
	TorrentFiles string    `gorm:"type:text"` // JSON存储
	CreatedAt    time.Time `gorm:"type:datetime;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"type:datetime;autoUpdateTime"`
	DownloadType string    `gorm:"type:varchar(16);not null;default:''"`
	Status       string    `gorm:"type:varchar(32);not null;default:''"`
	Size         int64     `gorm:"type:bigint;not null;default:0"`

	// Meta信息
	ChineseName  string `gorm:"type:varchar(255)"`
	Year         string `gorm:"type:varchar(10)"`
	TMDBID       int    `gorm:"type:int;index"`
	ReleaseGroup string `gorm:"type:varchar(255)"`
}

// TableName 设置表名
func (taskSchema) TableName() string {
	return "magnet_tasks"
}

// fromTask 将业务模型转换为数据库模型
func fromTask(task magnet.Task) taskSchema {
	// 序列化TorrentFiles
	filesJSON, _ := json.Marshal(task.Torrent.Files)

	return taskSchema{
		TaskID:       task.TaskID,
		TorrentHash:  task.Torrent.Hash,
		TorrentName:  task.Torrent.Name,
		TorrentFiles: string(filesJSON),
		Size:         task.Torrent.Size,
		MagnetLink:   task.MagnetLink,
		CreatedAt:    task.CreatedAt,
		DownloadType: string(task.DownloadType),
		Status:       string(task.Status),
		ChineseName:  task.Meta.ChineseName,
		Year:         task.Meta.Year,
		TMDBID:       task.Meta.TMDBID,
		ReleaseGroup: task.Meta.ReleaseGroup,
	}
}

// toTask 将数据库模型转换为业务模型
func toTask(model taskSchema) (magnet.Task, error) {
	// 反序列化TorrentFiles
	var files []magnet.TorrentFile
	if model.TorrentFiles != "" {
		if err := json.Unmarshal([]byte(model.TorrentFiles), &files); err != nil {
			return magnet.Task{}, fmt.Errorf("反序列化种子文件失败: %w", err)
		}
	}

	return magnet.Task{
		TaskID: model.TaskID,
		Torrent: magnet.Torrent{
			Hash:  model.TorrentHash,
			Name:  model.TorrentName,
			Files: files,
			Size:  model.Size,
		},
		MagnetLink:   model.MagnetLink,
		CreatedAt:    model.CreatedAt,
		DownloadType: downloader.DownloadType(model.DownloadType),
		Meta: magnet.Meta{
			ChineseName:  model.ChineseName,
			Year:         model.Year,
			TMDBID:       model.TMDBID,
			ReleaseGroup: model.ReleaseGroup,
		},
		Status: magnet.TaskStatus(model.Status),
	}, nil
}
