package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/MangataL/BangumiBuddy/internal/scrape"
	"github.com/MangataL/BangumiBuddy/pkg/utils"
)

var _ scrape.Repository = &Repository{}

// Repository 实现 scrape.Repository 接口
type Repository struct {
	db *gorm.DB
}

// New 创建存储层实例
func New(db *gorm.DB) *Repository {
	utils.DropUnusedColumns(db, &metadataCheckSchema{})
	_ = db.AutoMigrate(&metadataCheckSchema{})
	return &Repository{db: db}
}

// Add 添加元数据检查任务
func (r *Repository) Add(ctx context.Context, task scrape.MetadataCheckTask) error {
	model := fromTask(task)

	// 使用 upsert 操作，如果文件路径存在则更新，不存在则插入
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "file_path"}},
		UpdateAll: true,
	}).Create(&model)

	if result.Error != nil {
		return fmt.Errorf("添加元数据检查任务失败: %w", result.Error)
	}

	return nil
}

// Get 根据ID获取任务
func (r *Repository) Get(ctx context.Context, id uint) (scrape.MetadataCheckTask, error) {
	var model metadataCheckSchema
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return scrape.MetadataCheckTask{}, scrape.ErrTaskNotFound
		}
		return scrape.MetadataCheckTask{}, fmt.Errorf("获取任务失败: %w", err)
	}
	return toTask(model), nil
}

// List 列出所有待处理任务
func (r *Repository) List(ctx context.Context) ([]scrape.MetadataCheckTask, error) {
	var models []metadataCheckSchema

	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %w", err)
	}

	tasks := make([]scrape.MetadataCheckTask, 0, len(models))
	for _, model := range models {
		tasks = append(tasks, toTask(model))
	}

	return tasks, nil
}

// Delete 根据文件路径删除任务
func (r *Repository) Delete(ctx context.Context, filePath string) error {
	result := r.db.WithContext(ctx).Where("file_path = ?", filePath).Delete(&metadataCheckSchema{})
	if result.Error != nil {
		return fmt.Errorf("删除任务失败: %w", result.Error)
	}
	return nil
}

// UpdateStatuses 更新任务刮削状态
func (r *Repository) UpdateStatuses(ctx context.Context, filePath string, statuses []scrape.ScrapeStatus) error {
	statusesJSON, err := json.Marshal(statuses)
	if err != nil {
		return fmt.Errorf("序列化状态失败: %w", err)
	}
	result := r.db.WithContext(ctx).Model(&metadataCheckSchema{}).
		Where("file_path = ?", filePath).
		Update("statuses", string(statusesJSON))
	if result.Error != nil {
		return fmt.Errorf("更新刮削状态失败: %w", result.Error)
	}
	return nil
}

// Clean 清空所有任务
func (r *Repository) Clean(ctx context.Context) error {
	if err := r.db.Session(&gorm.Session{
		AllowGlobalUpdate: true,
		Context:           ctx,
	}).Delete(&metadataCheckSchema{}).Error; err != nil {
		return fmt.Errorf("清空任务失败: %w", err)
	}

	r.db.WithContext(ctx).Exec("DELETE FROM sqlite_sequence WHERE name='metadata_checks'")

	return nil
}
