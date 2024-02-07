package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/MangataL/BangumiBuddy/internal/transfer"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"github.com/glebarez/go-sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

func NewTransferFilesRepo(db *gorm.DB) transfer.TransferFilesRepo {
	_ = db.AutoMigrate(&fileTransferredSchema{})
	return &transferFilesRepo{db: db}
}

type transferFilesRepo struct {
	db *gorm.DB
}

func (r *transferFilesRepo) Set(ctx context.Context, fileTransferred transfer.FileTransferred) error {
	schema := fromFileTransferred(fileTransferred)
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "new_file_id"}},
		UpdateAll: true,
	}).Create(&schema).Error; err != nil {
		// 由于gorm不支持sqlite多个on conflict，所以这里调用两次来兼容。
		// 目前不存在性能问题，所以不考虑手动写sql
		if err := (&sqlite.Error{}); errors.As(err, &err) {
			if err.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE {
				return r.db.WithContext(ctx).Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "origin_file"}},
					UpdateAll: true,
				}).Create(&schema).Error
			}
		}
	}
	return nil
}

func (r *transferFilesRepo) Get(ctx context.Context, req transfer.GetFileTransferredReq) (transfer.FileTransferred, error) {
	var model fileTransferredSchema
	stmt := r.db.WithContext(ctx)
	if req.NewFileID != "" {
		stmt = stmt.Where("new_file_id = ?", req.NewFileID)
	}
	if req.OriginFile != "" {
		stmt = stmt.Where("origin_file = ?", req.OriginFile)
	}

	if err := stmt.First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return transfer.FileTransferred{}, transfer.ErrFileTransferredNotFound
		}
		return transfer.FileTransferred{}, err
	}

	return toFileTransferred(model), nil
}

// List implements transfer.TransferFilesRepo.
func (r *transferFilesRepo) List(ctx context.Context, req transfer.ListFileTransferredReq) ([]transfer.FileTransferred, error) {
	var models []fileTransferredSchema
	stmt := r.db.WithContext(ctx)
	if req.BangumiName != "" {
		stmt = stmt.Where("bangumi_name = ?", req.BangumiName)
	}
	if req.Season != 0 {
		stmt = stmt.Where("season = ?", req.Season)
	}
	if err := stmt.Find(&models).Error; err != nil {
		return nil, err
	}

	var fileTransferreds []transfer.FileTransferred
	for _, model := range models {
		fileTransferreds = append(fileTransferreds, toFileTransferred(model))
	}

	return fileTransferreds, nil
}

func (r *transferFilesRepo) Del(ctx context.Context, req transfer.DeleteFileTransferredReq) error {
	if req.NewFileID == "" && req.OriginFile == "" && req.SubscriptionID == "" && req.NewFile == "" {
		return nil
	}

	stmt := r.db.WithContext(ctx)
	if req.NewFileID != "" {
		stmt = stmt.Where("new_file_id = ?", req.NewFileID)
	}
	if req.NewFile != "" {
		stmt = stmt.Where("new_file = ?", req.NewFile)
	}
	if req.OriginFile != "" {
		stmt = stmt.Where("origin_file = ?", req.OriginFile)
	}
	if req.SubscriptionID != "" {
		stmt = stmt.Where("subscription_id = ?", req.SubscriptionID)
	}
	if err := stmt.Delete(&fileTransferredSchema{}).Error; err != nil {
		return fmt.Errorf("删除文件转移记录失败: %w", err)
	}
	return nil
}
