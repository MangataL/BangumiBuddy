package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass/freetype"
)

func New(db *gorm.DB) (ass.FontMetaRepository, error) {
	// 自动迁移表结构
	if err := db.AutoMigrate(&fontMetaSchema{}, &fontNameSchema{}); err != nil {
		return nil, err
	}
	return &fontMetaRepository{db: db}, nil
}

type fontMetaRepository struct {
	db *gorm.DB
}

func (f *fontMetaRepository) Replace(ctx context.Context, fonts []ass.FontMeta) error {
	return f.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM font_names").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM fonts_meta").Error; err != nil {
			return err
		}
		var aliases []fontNameSchema
		for _, font := range fonts {
			face := fromFontMeta(font)
			result := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "full_name"}}, DoNothing: true}).Create(&face)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				continue
			}
			for _, name := range font.Aliases {
				aliases = append(aliases, fontNameSchema{FontID: face.ID, Kind: name.Kind, Name: ass.FontNameKey(name.Value)})
			}
		}
		if len(aliases) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(aliases, 200).Error
	})
}

// Find implements FontMetaRepository.
func (f *fontMetaRepository) Find(ctx context.Context, req ass.FindFontMetaReq) ([]ass.FontMeta, error) {
	var models []fontMetaSchema
	query := f.db.WithContext(ctx)
	byName := func(columns string, kind uint16, name string) {
		query = query.Where("("+columns+" OR id IN (SELECT font_id FROM font_names WHERE kind = ? AND name = ?))",
			name, name, kind, ass.FontNameKey(name))
	}
	if req.FullName != "" {
		byName("full_name = ? OR chinese_full_name = ?", freetype.NameIDFullName, req.FullName)
	}
	if req.PostScriptName != "" {
		byName("post_script_name = ? OR chinese_post_script_name = ?", freetype.NameIDPostScript, req.PostScriptName)
	}
	if req.FamilyName != "" {
		byName("family_name = ? OR chinese_family_name = ?", freetype.NameIDFamily, req.FamilyName)
	}
	if req.Type != "" {
		query = query.Where("type = ?", string(req.Type))
	}
	query = req.Page.Apply(query)

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

func (f *fontMetaRepository) Stats(ctx context.Context) (subtitle.FontMetaSetStats, error) {
	var stats subtitle.FontMetaSetStats
	err := f.db.WithContext(ctx).Raw("SELECT count(*) AS total, count(*) > 0 AND NOT EXISTS (SELECT 1 FROM font_names) AS needs_rebuild FROM fonts_meta").Scan(&stats).Error
	return stats, err
}
