package downloader

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// torrentSchema 是Torrent的数据库模型
type torrentSchema struct {
	ID             int       `gorm:"type:int;primaryKey;autoIncrement"`
	Hash           string    `gorm:"type:char(32);uniqueIndex"`
	Name           string    `gorm:"type:varchar(255)"`
	BangumiName    string    `gorm:"type:varchar(255)"`
	Path           string    `gorm:"type:varchar(2048)"`
	Status         string    `gorm:"type:varchar(255)"`
	StatusDetail   string    `gorm:"type:text"`
	SubscriptionID string    `gorm:"type:varchar(36);not null;index;default:''"`
	TaskID         string    `gorm:"type:varchar(36);not null;index;default:''"`
	TransferType   string    `gorm:"type:varchar(16)"`
	RSSGUID        string    `gorm:"type:varchar(255)"`
	CreatedAt      time.Time `gorm:"type:datetime;autoCreateTime;index"`
	UpdatedAt      time.Time `gorm:"type:datetime;autoUpdateTime"`
	FileNames      string    `gorm:"type:text"`
}

// TableName 指定表名
func (torrentSchema) TableName() string {
	return "torrents"
}

const fileNamesSeparator = "|#"

// ToTorrent 将数据库模型转换为业务模型
func (m *torrentSchema) ToTorrent() Torrent {
	return Torrent{
		Hash:           m.Hash,
		Path:           m.Path,
		Status:         TorrentStatus(m.Status),
		StatusDetail:   m.StatusDetail,
		SubscriptionID: m.SubscriptionID,
		TaskID:         m.TaskID,
		Name:           m.Name,
		TransferType:   m.TransferType,
		RSSGUID:        m.RSSGUID,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		FileNames:      strings.Split(m.FileNames, fileNamesSeparator),
	}
}

// FromTorrent 将业务模型转换为数据库模型
func (m *torrentSchema) FromTorrent(t Torrent) {
	m.Hash = t.Hash
	m.Path = t.Path
	m.Status = string(t.Status)
	m.StatusDetail = t.StatusDetail
	m.SubscriptionID = t.SubscriptionID
	m.TaskID = t.TaskID
	m.TransferType = t.TransferType
	m.Name = t.Name
	m.RSSGUID = t.RSSGUID
	m.FileNames = strings.Join(t.FileNames, fileNamesSeparator)
}

func NewTorrentOperator(db *gorm.DB) TorrentOperator {
	// 自动迁移表结构
	db.AutoMigrate(&torrentSchema{})
	return &torrentOperator{db: db}
}

type torrentOperator struct {
	db *gorm.DB
}

// SetTorrentStatus 设置种子文件状态
func (t *torrentOperator) SetTorrentStatus(ctx context.Context, hash string, status TorrentStatus, detail string, opts *SetTorrentStatusOptions) error {
	updates := map[string]interface{}{
		"status":        status,
		"status_detail": detail,
	}
	if opts != nil && opts.TransferType != "" {
		updates["transfer_type"] = opts.TransferType
	}
	if opts != nil && len(opts.FileNames) > 0 {
		updates["file_names"] = strings.Join(opts.FileNames, fileNamesSeparator)
	}
	return t.db.WithContext(ctx).Model(&torrentSchema{}).Where("hash = ?", hash).Updates(updates).Error
}

// Save 保存种子文件信息
func (t *torrentOperator) Save(ctx context.Context, torrent Torrent) error {
	if torrent.Hash == "" {
		return errors.New("torrent hash cannot be empty")
	}

	model := &torrentSchema{}
	model.FromTorrent(torrent)

	// 使用Upsert操作，如果存在则更新，不存在则创建
	return t.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "hash"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"name":         model.Name,
			"bangumi_name": model.BangumiName,
			"path":         model.Path,
			"status": gorm.Expr(`CASE 
				WHEN status IN (?, ?) THEN status 
				ELSE ? 
				END`, TorrentStatusTransferred, TorrentStatusTransferredError, model.Status),
			"status_detail":   model.StatusDetail,
			"subscription_id": model.SubscriptionID,
			"task_id":         model.TaskID,
			"transfer_type":   model.TransferType,
			"rss_guid":        model.RSSGUID,
		}),
	}).Create(model).Error
}

// Get 获取种子文件信息
func (t *torrentOperator) Get(ctx context.Context, hash string) (Torrent, error) {
	if hash == "" {
		return Torrent{}, errors.New("hash cannot be empty")
	}

	var model torrentSchema
	err := t.db.WithContext(ctx).Where("hash = ?", hash).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Torrent{}, fmt.Errorf("torrent with hash %s not found", hash)
		}
		return Torrent{}, err
	}

	return model.ToTorrent(), nil
}

// List 列出所有种子文件
func (t *torrentOperator) List(ctx context.Context, filter TorrentFilter) ([]Torrent, int, error) {
	var models []torrentSchema
	query := t.db.WithContext(ctx)

	// 应用过滤条件
	if len(filter.Statuses) > 0 {
		statuses := make([]string, len(filter.Statuses))
		for i, status := range filter.Statuses {
			statuses[i] = string(status)
		}
		query = query.Where("status IN ?", statuses)
	}

	if filter.SubscriptionID != "" {
		query = query.Where("subscription_id = ?", filter.SubscriptionID)
	}

	if filter.TaskID != "" {
		query = query.Where("task_id = ?", filter.TaskID)
	}

	if !filter.StartTime.IsZero() {
		query = query.Where("created_at >= ?", filter.StartTime)
	}

	if !filter.EndTime.IsZero() {
		query = query.Where("created_at <= ?", filter.EndTime)
	}

	if filter.MagnetTask != nil {
		if *filter.MagnetTask {
			query = query.Where("task_id <> ''")
		} else {
			query = query.Where("task_id = ''")
		}
	}

	// 获取总数
	var total int64
	err := query.WithContext(ctx).Model(&torrentSchema{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	query = filter.Page.Apply(query)
	query = filter.Order.Apply(query)

	// 执行查询
	err = query.Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	// 转换为业务模型
	torrents := make([]Torrent, len(models))
	for i, model := range models {
		torrents[i] = model.ToTorrent()
	}

	return torrents, int(total), nil
}

// Delete 删除种子文件
func (t *torrentOperator) Delete(ctx context.Context, hash string) error {
	return t.db.WithContext(ctx).Where("hash = ?", hash).Delete(&torrentSchema{}).Error
}

func (t *torrentOperator) DeleteBySubscriptionID(ctx context.Context, subscriptionID string) error {
	return t.db.WithContext(ctx).Where("subscription_id = ?", subscriptionID).Delete(&torrentSchema{}).Error
}

func (t *torrentOperator) DeleteByTaskID(ctx context.Context, taskID string) error {
	return t.db.WithContext(ctx).Where("task_id = ?", taskID).Delete(&torrentSchema{}).Error
}