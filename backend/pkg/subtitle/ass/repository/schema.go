package repository

import (
	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass"
)

// fontMetaSchema 字体元数据数据库模型
type fontMetaSchema struct {
	ID                    uint   `gorm:"type:int;primaryKey;autoIncrement"`
	FamilyName            string `gorm:"type:varchar(255);not null;index"`
	FullName              string `gorm:"column:full_name;type:varchar(255);not null;uniqueIndex:idx_full_name"`
	PostScriptName        string `gorm:"type:varchar(255);not null;index"`
	ChineseFamilyName     string `gorm:"type:varchar(255);not null;default:'';index"`
	ChineseFullName       string `gorm:"type:varchar(255);not null;default:'';index"`
	ChinesePostScriptName string `gorm:"type:varchar(255);not null;default:''"`
	Path                  string `gorm:"type:varchar(512);not null;index"`
	Index                 int    `gorm:"type:int;not null;default:0"`
	BoldWeight            int    `gorm:"type:int;not null;default:400"`
	Italic                bool   `gorm:"type:boolean;not null;default:false"`
	Type                  string `gorm:"type:varchar(16);not null"`
}

// TableName 设置表名
func (fontMetaSchema) TableName() string {
	return "fonts_meta"
}

// fromFontMeta 将业务模型转换为数据库模型
func fromFontMeta(fontMeta ass.FontMeta) fontMetaSchema {
	return fontMetaSchema{
		FamilyName:            fontMeta.FamilyName,
		FullName:              fontMeta.FullName,
		PostScriptName:        fontMeta.PostScriptName,
		ChineseFamilyName:     fontMeta.ChineseFamilyName,
		ChineseFullName:       fontMeta.ChineseFullName,
		ChinesePostScriptName: fontMeta.ChinesePostScriptName,
		Path:                  fontMeta.Location.Path,
		Index:                 fontMeta.Location.Index,
		BoldWeight:            fontMeta.BoldWeight,
		Italic:                fontMeta.Italic,
		Type:                  string(fontMeta.Type),
	}
}

// toFontMeta 将数据库模型转换为业务模型
func toFontMeta(model fontMetaSchema) ass.FontMeta {
	return ass.FontMeta{
		FamilyName:            model.FamilyName,
		FullName:              model.FullName,
		PostScriptName:        model.PostScriptName,
		ChineseFamilyName:     model.ChineseFamilyName,
		ChineseFullName:       model.ChineseFullName,
		ChinesePostScriptName: model.ChinesePostScriptName,
		Location: ass.FontLocation{
			Path:  model.Path,
			Index: model.Index,
		},
		BoldWeight: model.BoldWeight,
		Italic:     model.Italic,
		Type:       ass.FontType(model.Type),
	}
}
