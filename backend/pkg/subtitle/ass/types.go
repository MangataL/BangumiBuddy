package ass

import (
	"golang.org/x/text/cases"

	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass/freetype"
)

type ParseFontSetOption func(*parseFontSetOptions)

type parseFontSetOptions struct {
	IgnoreFontExist bool
}

func IgnoreFontExist() ParseFontSetOption {
	return func(o *parseFontSetOptions) {
		o.IgnoreFontExist = true
	}
}

type FontSet map[Font]CodePoints

type Font = subtitle.FontStyle

type CodePoints map[rune]struct{}

func (c CodePoints) Copy() CodePoints {
	copy := make(CodePoints)
	for k, v := range c {
		copy[k] = v
	}
	return copy
}

type Dialogue struct {
	Style string
	Text  string
}

const (
	WeightNormal = 400
	WeightBold   = 700
)

type FontLocation struct {
	Path  string
	Index int
}

type FontType string

const (
	FontTypeTTF FontType = "ttf"
	FontTypeOTF FontType = "otf"
)

// FontNameKey 供临时库和 SQLite 按相同的 Unicode 大小写规则查询名称。
func FontNameKey(name string) string { return fontNameFold.String(name) }

var fontNameFold = cases.Fold()

type FontMeta struct {
	Aliases               []freetype.FontName
	FamilyName            string
	FullName              string
	PostScriptName        string
	ChineseFamilyName     string // 中文字体家族名（优先简体中文，其次繁体中文）
	ChineseFullName       string // 中文完整字体名
	ChinesePostScriptName string // 中文PostScript名称
	Location              FontLocation
	BoldWeight            int  // 字重
	Italic                bool // 是否斜体
	Type                  FontType
}

type FindFontReq struct {
	Font           Font
	UseOTF         bool
	UseSimilarFont bool
}
