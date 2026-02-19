package repository

import (
	"encoding/json"

	"github.com/MangataL/BangumiBuddy/internal/scrape"
)

// metadataCheckSchema 元数据检查任务数据库模型
type metadataCheckSchema struct {
	ID          uint   `gorm:"type:int;primaryKey;autoIncrement"`
	TMDBID      int    `gorm:"type:int;not null"`
	FilePath    string `gorm:"type:varchar(512);not null;uniqueIndex"` // 媒体文件路径
	BangumiName string `gorm:"type:varchar(255)"`                     // 番剧名称
	PosterURL   string `gorm:"type:varchar(2048)"`                    // 海报URL
	Season      int    `gorm:"type:int;default:0"`                    // 季度
	Episode     int    `gorm:"type:int;default:0"`                    // 集数
	Statuses    string `gorm:"type:text"`                             // 刮削状态列表（JSON）
}

// TableName 设置表名
func (metadataCheckSchema) TableName() string {
	return "metadata_checks"
}

// fromTask 从业务模型转换为数据库模型
func fromTask(task scrape.MetadataCheckTask) metadataCheckSchema {
	statusesJSON, _ := json.Marshal(task.Statuses)
	return metadataCheckSchema{
		TMDBID:      task.TMDBID,
		FilePath:    task.FilePath,
		BangumiName: task.BangumiName,
		PosterURL:   task.PosterURL,
		Season:      task.Season,
		Episode:     task.Episode,
		Statuses:    string(statusesJSON),
	}
}

// toTask 从数据库模型转换为业务模型
func toTask(schema metadataCheckSchema) scrape.MetadataCheckTask {
	var statuses []scrape.ScrapeStatus
	if schema.Statuses != "" {
		_ = json.Unmarshal([]byte(schema.Statuses), &statuses)
	}
	if len(statuses) == 0 {
		statuses = []scrape.ScrapeStatus{scrape.ScrapeStatusPending}
	}
	return scrape.MetadataCheckTask{
		ID:          schema.ID,
		TMDBID:      schema.TMDBID,
		FilePath:    schema.FilePath,
		BangumiName: schema.BangumiName,
		PosterURL:   schema.PosterURL,
		Season:      schema.Season,
		Episode:     schema.Episode,
		Statuses:    statuses,
	}
}
