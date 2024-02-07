package repository

import (
	"encoding/json"
	"time"

	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

// bangumiSchema 番剧数据库模型
type bangumiSchema struct {
	ID              uint         `gorm:"type:int;primaryKey;autoIncrement"`
	SubscriptionID  string       `gorm:"type:char(36);not null;uniqueIndex"`
	Name            string       `gorm:"type:varchar(255);not null;index"`
	RSSLink         string       `gorm:"type:varchar(2048);not null;uniqueIndex"`
	Active          bool         `gorm:"type:boolean;not null"`
	IncludeRegs     string       `gorm:"type:text"` // JSON 格式存储多个正则表达式
	ExcludeRegs     string       `gorm:"type:text"` // JSON 格式存储多个正则表达式
	Priority        int          `gorm:"type:int;default:0"`
	EpisodeOffset   int          `gorm:"type:int;default:0"`
	Season          int          `gorm:"type:int;default:1"`
	Year            string       `gorm:"type:varchar(10)"`
	TMDBID          int          `gorm:"type:int"`
	ReleaseGroup    string       `gorm:"type:varchar(255)"`
	EpisodeLocation string       `gorm:"type:varchar(512)"`
	CreatedAt       time.Time    `gorm:"type:datetime;autoCreateTime"`
	UpdatedAt       time.Time    `gorm:"type:datetime;autoUpdateTime"`
	LastAirEpisode  int          `gorm:"type:int;default:0"`
	PosterURL       string       `gorm:"type:varchar(2048)"`
	BackdropURL     string       `gorm:"type:varchar(2048)"`
	Overview        string       `gorm:"type:text"`
	Genres          string       `gorm:"type:text"`
	AirWeekday      time.Weekday `gorm:"type:int"`
	EpisodeTotalNum int          `gorm:"type:int;default:0"`
}

// TableName 设置表名
func (bangumiSchema) TableName() string {
	return "bangumis"
}

// rssRecordSchema RSS处理记录
type rssRecordSchema struct {
	ID             uint      `gorm:"type:int;primaryKey;autoIncrement"`
	SubscriptionID string    `gorm:"type:varchar(36);uniqueIndex:idx_subscription_guid,priority:1"`
	GUID           string    `gorm:"type:varchar(255);uniqueIndex:idx_subscription_guid,priority:2"`
	CreatedAt      time.Time `gorm:"type:datetime;autoCreateTime"`
}

// TableName 表名
func (rssRecordSchema) TableName() string {
	return "rss_records"
}

// 模型转换函数
func fromBangumi(b subscriber.Bangumi) bangumiSchema {
	includeRegsJSON, _ := json.Marshal(b.IncludeRegs)
	excludeRegsJSON, _ := json.Marshal(b.ExcludeRegs)

	return bangumiSchema{
		SubscriptionID:  b.SubscriptionID,
		Name:            b.Name,
		RSSLink:         b.RSSLink,
		Active:          b.Active,
		IncludeRegs:     string(includeRegsJSON),
		ExcludeRegs:     string(excludeRegsJSON),
		Priority:        b.Priority,
		EpisodeOffset:   b.EpisodeOffset,
		Season:          b.Season,
		Year:            b.Year,
		TMDBID:          b.TMDBID,
		ReleaseGroup:    b.ReleaseGroup,
		EpisodeLocation: b.EpisodeLocation,
		PosterURL:       b.PosterURL,
		BackdropURL:     b.BackdropURL,
		Overview:        b.Overview,
		Genres:          b.Genres,
		AirWeekday:      b.AirWeekday,
		EpisodeTotalNum: b.EpisodeTotalNum,
		CreatedAt:       b.CreatedAt,
		LastAirEpisode:  b.LastAirEpisode,
	}
}

func toBangumi(m bangumiSchema) subscriber.Bangumi {
	var includeRegs []string
	var excludeRegs []string

	// 解析 JSON 字符串为字符串数组
	if m.IncludeRegs != "" {
		_ = json.Unmarshal([]byte(m.IncludeRegs), &includeRegs)
	}

	if m.ExcludeRegs != "" {
		_ = json.Unmarshal([]byte(m.ExcludeRegs), &excludeRegs)
	}

	return subscriber.Bangumi{
		SubscriptionID:  m.SubscriptionID,
		Name:            m.Name,
		RSSLink:         m.RSSLink,
		Active:          m.Active,
		IncludeRegs:     includeRegs,
		ExcludeRegs:     excludeRegs,
		Priority:        m.Priority,
		EpisodeOffset:   m.EpisodeOffset,
		Season:          m.Season,
		Year:            m.Year,
		TMDBID:          m.TMDBID,
		ReleaseGroup:    m.ReleaseGroup,
		EpisodeLocation: m.EpisodeLocation,
		PosterURL:       m.PosterURL,
		BackdropURL:     m.BackdropURL,
		Overview:        m.Overview,
		Genres:          m.Genres,
		AirWeekday:      m.AirWeekday,
		EpisodeTotalNum: m.EpisodeTotalNum,
		LastAirEpisode:  m.LastAirEpisode,
		CreatedAt:       m.CreatedAt,
	}
}
