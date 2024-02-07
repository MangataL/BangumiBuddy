package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/glebarez/go-sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

var (
	_ subscriber.Repository          = &Repository{}
	_ subscriber.RSSRecordRepository = &Repository{}
)

// Repository 实现 subscriber.Repository 接口
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建存储层实例
func NewRepository(db *gorm.DB) *Repository {
	db.AutoMigrate(&bangumiSchema{})
	db.AutoMigrate(&rssRecordSchema{})
	return &Repository{db: db}
}

// New 创建一个新的 Repository 实例
func New(db *gorm.DB) *Repository {
	db.AutoMigrate(&bangumiSchema{})
	db.AutoMigrate(&rssRecordSchema{})
	return &Repository{
		db: db,
	}
}

// Save 保存番剧信息
func (r *Repository) Save(ctx context.Context, bangumi subscriber.Bangumi) error {
	model := fromBangumi(bangumi)

	// 使用 upsert 操作，如果存在则更新，不存在则插入
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "subscription_id"}},
		UpdateAll: true,
	}).Create(&model)

	if result.Error != nil {
		if err := (&sqlite.Error{}); errors.As(result.Error, &err) {
			if err.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				return errors.New("订阅已存在")
			}
		}
		return fmt.Errorf("保存番剧信息失败: %w", result.Error)
	}

	return nil
}

// List 列出番剧信息
func (r *Repository) List(ctx context.Context, req subscriber.ListBangumiReq) ([]subscriber.Bangumi, error) {
	var models []bangumiSchema

	db := r.db.WithContext(ctx).Model(&bangumiSchema{})

	if req.Active != nil {
		db = db.Where("active = ?", *req.Active)
	}
	if req.FuzzName != "" {
		db = db.Where("name LIKE ?", "%"+req.FuzzName+"%")
	}
	if req.Name != "" {
		db = db.Where("name = ?", req.Name)
	}
	if req.Season != 0 {
		db = db.Where("season = ?", req.Season)
	}
	if len(req.SubscriptionIDs) > 0 {
		db = db.Where("subscription_id IN ?", req.SubscriptionIDs)
	}

	if err := db.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("获取番剧列表失败: %w", err)
	}

	// 转换为业务模型
	bangumis := make([]subscriber.Bangumi, 0, len(models))
	for _, model := range models {
		bangumis = append(bangumis, toBangumi(model))
	}

	return bangumis, nil
}

// Get 获取单个番剧信息
func (r *Repository) Get(ctx context.Context, subscriptionID string) (subscriber.Bangumi, error) {
	var model bangumiSchema

	result := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		First(&model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return subscriber.Bangumi{}, subscriber.ErrSubscriberNotFound
		}
		return subscriber.Bangumi{}, fmt.Errorf("获取番剧信息失败: %w", result.Error)
	}

	return toBangumi(model), nil
}

// Delete 删除番剧信息
func (r *Repository) Delete(ctx context.Context, subscriptionID string) error {
	tx := r.db.WithContext(ctx).Begin()	
	if err := tx.Where("subscription_id = ?", subscriptionID).Delete(&bangumiSchema{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("subscription_id = ?", subscriptionID).Delete(&rssRecordSchema{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// UpdateLastAirEpisode 更新番剧的最新集数
func (r *Repository) UpdateLastAirEpisode(ctx context.Context, subscriptionID string, episode int) error {
	return r.db.WithContext(ctx).Model(&bangumiSchema{}).Where("subscription_id = ? AND last_air_episode < ?", subscriptionID, episode).Update("last_air_episode", episode).Error
}

// IsProcessed 检查RSS条目是否已处理
func (r *Repository) IsProcessed(ctx context.Context, subscriptionID string, guid string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&rssRecordSchema{}).
		Where("subscription_id = ? AND guid = ?", subscriptionID, guid).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("查询RSS记录失败: %w", err)
	}
	return count > 0, nil
}

// MarkProcessed 标记RSS条目为已处理
func (r *Repository) MarkProcessed(ctx context.Context, subscriptionID string, guid ...string) error {
	records := make([]rssRecordSchema, 0, len(guid))
	for _, g := range guid {
		records = append(records, rssRecordSchema{
			SubscriptionID: subscriptionID,
			GUID:           g,
		})
	}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "subscription_id"}, {Name: "guid"}},
		DoNothing: true,
	}).Create(&records).Error
	if err != nil {
		return fmt.Errorf("创建RSS记录失败: %w", err)
	}
	return nil
}

// ListProcessedGUIDs 获取订阅下所有已处理的GUID
func (r *Repository) ListProcessedGUIDs(ctx context.Context, subscriptionID string) ([]string, error) {
	var records []rssRecordSchema
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("查询RSS记录失败: %w", err)
	}

	guids := make([]string, 0, len(records))
	for _, record := range records {
		guids = append(guids, record.GUID)
	}
	return guids, nil
}

// DeleteProcessed 删除RSS条目处理记录
func (r *Repository) DeleteProcessed(ctx context.Context, subscriptionID string, guid ...string) error {
	err := r.db.WithContext(ctx).
		Where("subscription_id = ? AND guid IN ?", subscriptionID, guid).
		Delete(&rssRecordSchema{}).Error
	if err != nil {
		return fmt.Errorf("删除RSS记录失败: %w", err)
	}
	return nil
}

// StopSubscription 停止订阅
func (r *Repository) StopSubscription(ctx context.Context, subscriptionID string) error {
	return r.db.WithContext(ctx).Model(&bangumiSchema{}).Where("subscription_id = ?", subscriptionID).Update("active", false).Error
}
