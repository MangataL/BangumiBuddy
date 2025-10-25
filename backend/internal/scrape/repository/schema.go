package repository

import (
	"github.com/MangataL/BangumiBuddy/internal/scrape"
)

// metadataCheckSchema 元数据检查任务数据库模型
type metadataCheckSchema struct {
	ID           uint   `gorm:"type:int;primaryKey;autoIncrement"`
	TMDBID       int    `gorm:"type:int;not null"`
	FilePath     string `gorm:"type:varchar(512);not null;uniqueIndex"` // 媒体文件路径
	DownloadType string `gorm:"type:varchar(10);not null"`
	ImageChecked bool   `gorm:"type:boolean;default:false"` // 图片是否已检查过
}

// TableName 设置表名
func (metadataCheckSchema) TableName() string {
	return "metadata_checks"
}

// fromTask 从业务模型转换为数据库模型
func fromTask(task scrape.MetadataCheckTask) metadataCheckSchema {
	return metadataCheckSchema{
		TMDBID:       task.TMDBID,
		FilePath:     task.FilePath,
		DownloadType: string(task.DownloadType),
		ImageChecked: task.ImageChecked,
	}
}

// toTask 从数据库模型转换为业务模型
func toTask(schema metadataCheckSchema) scrape.MetadataCheckTask {
	return scrape.MetadataCheckTask{
		TMDBID:       schema.TMDBID,
		FilePath:     schema.FilePath,
		DownloadType: scrape.DownloadType(schema.DownloadType),
		ImageChecked: schema.ImageChecked,
	}
}
