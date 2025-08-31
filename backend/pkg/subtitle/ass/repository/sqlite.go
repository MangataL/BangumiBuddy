package repository

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass"
)

func New(db *gorm.DB) ass.FontMetaRepository {
	// 自动迁移表结构
	db.AutoMigrate(&fontMetaSchema{})
	return &fontMetaRepository{db: db}
}

type fontMetaRepository struct {
	db *gorm.DB
}

// Clean implements FontMetaRepository.
func (f *fontMetaRepository) Clean(ctx context.Context) error {
	// 使用ORM方式删除所有记录
	if err := f.db.Session(&gorm.Session{
		AllowGlobalUpdate: true,
		Context:           ctx,
	}).Delete(&fontMetaSchema{}).Error; err != nil {
		return fmt.Errorf("清理字体元数据失败: %w", err)
	}

	// 重置SQLite自增值（SQLite特有，无法通过ORM实现）
	if err := f.db.WithContext(ctx).Exec("DELETE FROM sqlite_sequence WHERE name='fonts_meta'").Error; err != nil {
		// 如果sqlite_sequence表不存在或者没有记录，这是正常的，不算错误
		// 忽略这个错误
		log.Warnf(ctx, "重置SQLite自增值失败: %v", err)
	}

	return nil
}

// Save implements FontMetaRepository.
func (f *fontMetaRepository) Save(ctx context.Context, fontMetas []ass.FontMeta) error {
	if len(fontMetas) == 0 {
		return nil
	}

	// 转换为数据库模型
	models := make([]fontMetaSchema, 0, len(fontMetas))
	for _, fontMeta := range fontMetas {
		models = append(models, fromFontMeta(fontMeta))
	}

	// 使用Clauses进行批量Upsert操作，避免重复插入相同fullName的记录
	// SQLite使用ON CONFLICT DO NOTHING来处理唯一约束冲突
	return errors.WithMessage(
		f.db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "full_name"}},
			DoNothing: true,
		}).CreateInBatches(&models, 200).Error,
		"保存字体元数据失败",
	)
}

// Find implements FontMetaRepository.
func (f *fontMetaRepository) Find(ctx context.Context, req ass.FindFontMetaReq) ([]ass.FontMeta, error) {
	var models []fontMetaSchema
	query := f.db.WithContext(ctx)

	// 构建查询条件
	if req.FullName != "" {
		query = query.Where("full_name = ?", req.FullName)
	}
	if req.PostScriptName != "" {
		query = query.Where("post_script_name = ?", req.PostScriptName)
	}
	if req.FamilyName != "" {
		query = query.Where("family_name = ?", req.FamilyName)
	}
	if req.Type != "" {
		query = query.Where("type = ?", string(req.Type))
	}

	// 执行查询
	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("查询字体元数据失败: %w", err)
	}

	// 转换为业务模型
	fontMetas := make([]ass.FontMeta, 0, len(models))
	for _, model := range models {
		fontMetas = append(fontMetas, toFontMeta(model))
	}

	return fontMetas, nil
}

func (f *fontMetaRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := f.db.WithContext(ctx).Model(&fontMetaSchema{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("查询字体元数据数量失败: %w", err)
	}
	return count, nil
}
