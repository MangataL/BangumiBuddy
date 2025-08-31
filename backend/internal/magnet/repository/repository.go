package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/glebarez/go-sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/MangataL/BangumiBuddy/internal/magnet"
)

var _ magnet.Repository = &Repository{}

// Repository 实现 magnet.Repository 接口
type Repository struct {
	db *gorm.DB
}

// New 创建存储层实例
func New(db *gorm.DB) magnet.Repository {
	db.AutoMigrate(&taskSchema{})
	return &Repository{db: db}
}

// SaveTask 保存下载任务
func (r *Repository) SaveTask(ctx context.Context, task magnet.Task) error {
	model := fromTask(task)

	// 使用 upsert 操作，如果存在则更新，不存在则插入
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "task_id"}},
		UpdateAll: true,
	}).Create(&model)

	if result.Error != nil {
		if err := (&sqlite.Error{}); errors.As(result.Error, &err) {
			if err.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				return errors.New("任务已存在")
			}
		}
		return fmt.Errorf("保存下载任务失败: %w", result.Error)
	}

	return nil
}

// GetTask 获取单个下载任务
func (r *Repository) GetTask(ctx context.Context, taskID string) (magnet.Task, error) {
	var model taskSchema

	result := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		First(&model)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return magnet.Task{}, magnet.ErrMagnetTaskNotFound
		}
		return magnet.Task{}, fmt.Errorf("获取下载任务失败: %w", result.Error)
	}

	return toTask(model)
}

// GetTaskByHash 获取单个下载任务

func (r *Repository) GetTaskByHash(ctx context.Context, hash string) (magnet.Task, error) {
	var model taskSchema

	result := r.db.WithContext(ctx).
		Where("torrent_hash = ?", hash).
		First(&model)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return magnet.Task{}, magnet.ErrMagnetTaskNotFound
		}
		return magnet.Task{}, fmt.Errorf("获取下载任务失败: %w", result.Error)
	}

	return toTask(model)
}

// ListTasks 列出下载任务
func (r *Repository) ListTasks(ctx context.Context, req magnet.ListTasksReq) ([]magnet.Task, int, error) {
	var models []taskSchema
	var total int64

	db := r.db.WithContext(ctx).Model(&taskSchema{})

	// 应用过滤条件
	if len(req.TaskIDs) > 0 {
		db = db.Where("task_id IN ?", req.TaskIDs)
	}
	if req.TorrentName != "" {
		db = db.Where("torrent_name LIKE ?", "%"+req.TorrentName+"%")
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计磁力任务数量失败: %w", err)
	}

	db = req.Page.Apply(db)
	db = req.Order.Apply(db)

	if err := db.Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("获取磁力任务列表失败: %w", err)
	}

	// 转换为业务模型
	tasks := make([]magnet.Task, 0, len(models))
	for _, model := range models {
		task, err := toTask(model)
		if err != nil {
			return nil, 0, fmt.Errorf("转换磁力任务模型失败: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, int(total), nil
}

// DeleteTask 删除下载任务
func (r *Repository) DeleteTask(ctx context.Context, taskID string) error {
	result := r.db.WithContext(ctx).
		Where("task_id = ?", taskID).
		Delete(&taskSchema{})

	if result.Error != nil {
		return fmt.Errorf("删除磁力任务失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("磁力任务不存在")
	}

	return nil
}
