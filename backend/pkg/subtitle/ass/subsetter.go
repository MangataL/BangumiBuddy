package ass

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"golang.org/x/image/font/sfnt"
	"golang.org/x/sync/errgroup"

	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
)

func NewSubsetter(fontMetaSet *FontMetaSet, config FontSubsetterConfig) *Subsetter {
	return &Subsetter{
		parser:       NewParser(),
		fontMetaSet:  fontMetaSet,
		config:       config,
		fontEmbedder: NewFontEmbedder(),
	}
}

type FontSubsetterConfig struct {
	UseOTF            bool `mapstructure:"use_otf" json:"useOTF"`
	UseSimilarFont    bool `mapstructure:"use_similar_font" json:"useSimilarFont"`
	UseSystemFontsDir bool `mapstructure:"use_system_fonts_dir" json:"useSystemFontsDir"`
	CoverExistSubFont bool `mapstructure:"cover_exist_sub_font" json:"coverExistSubFont"`
	GenerateNewFile   bool `mapstructure:"generate_new_file" json:"generateNewFile"`
	CheckGlyphs       bool `mapstructure:"check_glyphs" json:"checkGlyphs"`
}

type Subsetter struct {
	parser       *Parser
	fontMetaSet  *FontMetaSet
	config       FontSubsetterConfig
	fontEmbedder *FontEmbedder
}

func (s *Subsetter) InitFontMetaSet(ctx context.Context) error {
	return s.fontMetaSet.Init(ctx, s.config.UseSystemFontsDir)
}

func (s *Subsetter) SubsetFont(ctx context.Context, filePath string) (string, error) {

	var parseFontOpts []ParseFontSetOption
	if s.config.CoverExistSubFont {
		parseFontOpts = append(parseFontOpts, IgnoreFontExist())
	}
	fontSet, noNeedSubset, err := s.parser.ParseFontSet(ctx, filePath, parseFontOpts...)
	if err != nil {
		return "", fmt.Errorf("解析字体集合失败: %w", err)
	}
	if noNeedSubset {
		log.Infof(ctx, "文件 %s 已经子集化，跳过处理", filePath)
		return filePath, nil
	}
	subsetFonts, err := s.makeSubsetFonts(ctx, fontSet)
	if err != nil {
		return "", fmt.Errorf("获取字体库信息失败: %w", err)
	}

	// 在子集化之前检查字形
	if s.config.CheckGlyphs {
		if err := s.checkGlyphs(subsetFonts); err != nil {
			return "", fmt.Errorf("字形检查失败: %w", err)
		}
	}

	// 并发执行子集化
	fonts, err := s.subsetFontsConcurrently(ctx, subsetFonts)
	if err != nil {
		return "", fmt.Errorf("字体子集化失败: %w", err)
	}

	// 嵌入字体到字幕文件
	outputPath, err := s.fontEmbedder.EmbedFonts(ctx, filePath, fonts, s.config.GenerateNewFile)
	if err != nil {
		return "", fmt.Errorf("嵌入字体失败: %w", err)
	}

	log.Infof(ctx, "字体子集化完成，输出文件: %s", outputPath)
	return outputPath, nil
}

type SubsetFont struct {
	Font       Font
	CodePoints CodePoints
	Location   FontLocation
	FontData   []byte // 原始字体文件数据（用于传递给CGO）
}

func (s *Subsetter) makeSubsetFonts(ctx context.Context, fontSet FontSet) ([]SubsetFont, error) {
	var subsetFonts []SubsetFont
	for font, codePoints := range fontSet {
		subsetFont, err := s.makeSubsetFont(ctx, font, codePoints.Copy())
		if err != nil {
			return nil, err
		}
		subsetFonts = append(subsetFonts, subsetFont)
	}
	return subsetFonts, nil
}

func (s *Subsetter) makeSubsetFont(ctx context.Context, font Font, codePoints CodePoints) (SubsetFont, error) {
	req := FindFontReq{
		Font:           font,
		UseOTF:         s.config.UseOTF,
		UseSimilarFont: s.config.UseSimilarFont,
	}
	fontMeta, err := s.fontMetaSet.FindFontMeta(ctx, req)
	if err != nil {
		if errors.Is(err, ErrFontNotFound) {
			return SubsetFont{}, fmt.Errorf("字体 %s 未找到: %w", font.FontName, err)
		}
		return SubsetFont{}, err
	}
	subsetFont := SubsetFont{
		Font:       font,
		CodePoints: codePoints,
		Location:   fontMeta.Location,
	}
	return subsetFont, nil
}

// checkGlyphs 检查所有字体的字形
func (s *Subsetter) checkGlyphs(subsetFonts []SubsetFont) error {
	for i := range subsetFonts {
		subsetFont := &subsetFonts[i]

		// 读取字体文件数据
		fontData, err := os.ReadFile(subsetFont.Location.Path)
		if err != nil {
			return fmt.Errorf("读取字体文件失败[%s]: %w", subsetFont.Font.FontName, err)
		}

		// 保存字体数据供后续子集化使用
		subsetFont.FontData = fontData

		// 使用 sfnt 包检查字形
		missingGlyphs, err := s.checkGlyphsWithSfnt(fontData, subsetFont.Location.Index, subsetFont.CodePoints)
		if err != nil {
			return fmt.Errorf("检查字体字形失败[%s]: %w", subsetFont.Font.FontName, err)
		}

		// 如果有缺失的字形，返回错误
		if len(missingGlyphs) > 0 {
			return fmt.Errorf("字体 %s 缺失 %d 个字形: %s",
				subsetFont.Font.FontName,
				len(missingGlyphs),
				formatMissingGlyphs(missingGlyphs),
			)
		}
	}

	return nil
}

// checkGlyphsWithSfnt 使用 sfnt 包检查字体是否包含指定的码点
func (s *Subsetter) checkGlyphsWithSfnt(fontData []byte, fontIndex int, codePoints CodePoints) ([]rune, error) {
	// 解析字体
	font, err := sfnt.Parse(fontData)
	if err != nil {
		return nil, fmt.Errorf("解析字体失败: %w", err)
	}

	// 获取字体缓冲区
	buf := &sfnt.Buffer{}
	var missingGlyphs []rune

	// 检查每个码点
	for cp := range codePoints {
		// 尝试获取该码点的字形索引
		glyphIndex, err := font.GlyphIndex(buf, cp)
		if err != nil {
			return nil, fmt.Errorf("获取字形索引失败: %w", err)
		}

		// glyphIndex 为 0 表示该字符没有对应的字形（.notdef）
		if glyphIndex == 0 {
			missingGlyphs = append(missingGlyphs, cp)
		}
	}

	return missingGlyphs, nil
}

// subsetFontsConcurrently 并发执行字体子集化
func (s *Subsetter) subsetFontsConcurrently(ctx context.Context, subsetFonts []SubsetFont) (map[Font][]byte, error) {
	g, ctx := errgroup.WithContext(ctx)

	var (
		fonts = make(map[Font][]byte)
		mu    sync.Mutex
	)

	for i := range subsetFonts {
		subsetFont := subsetFonts[i] // 使用指针

		g.Go(func() error {
			// 检查上下文是否已取消
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			// 执行子集化
			subsetData, err := s.subsetSingleFont(ctx, subsetFont)
			if err != nil {
				return fmt.Errorf("子集化字体失败[%s]: %w", subsetFont.Font.FontName, err)
			}

			mu.Lock()
			fonts[subsetFont.Font] = subsetData
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return fonts, nil
}

// subsetSingleFont 执行单个字体的子集化
func (s *Subsetter) subsetSingleFont(ctx context.Context, subsetFont SubsetFont) ([]byte, error) {
	// 将CodePoints转换为rune切片
	codePoints := make([]rune, 0, len(subsetFont.CodePoints))
	for cp := range subsetFont.CodePoints {
		codePoints = append(codePoints, cp)
	}

	// 如果字体数据已经加载（CheckGlyphs时加载），使用已加载的数据
	// 否则读取字体文件
	var fontData []byte
	var err error
	if len(subsetFont.FontData) > 0 {
		fontData = subsetFont.FontData
	} else {
		fontData, err = os.ReadFile(subsetFont.Location.Path)
		if err != nil {
			return nil, fmt.Errorf("读取字体文件失败: %w", err)
		}
	}

	// 调用harfbuzz进行子集化
	subfontData, err := CreateSubfontData(
		fontData,
		subsetFont.Location.Index,
		codePoints,
	)
	if err != nil {
		return nil, err
	}

	return subfontData, nil
}

// formatMissingGlyphs 格式化缺失的字形，用于日志输出
func formatMissingGlyphs(glyphs []rune) string {
	if len(glyphs) == 0 {
		return ""
	}

	// 最多显示20个字符
	maxDisplay := 20
	if len(glyphs) > maxDisplay {
		return fmt.Sprintf("%s... (共%d个)", string(glyphs[:maxDisplay]), len(glyphs))
	}
	return string(glyphs)
}

func (s *Subsetter) GetFontMetaSetStats(ctx context.Context) (subtitle.FontMetaSetStats, error) {
	return s.fontMetaSet.GetFontMetaSetStats(ctx)
}

func (s *Subsetter) Reload(config interface{}) error {
	cfg, ok := config.(*FontSubsetterConfig)
	if !ok {
		return errors.New("配置类型错误")
	}
	s.config = *cfg
	return nil
}

func (s *Subsetter) UsingTempFontDir(ctx context.Context, fontDir string) (subtitle.Subsetter, error) {
	fms, err := s.fontMetaSet.UsingTempFontDir(ctx, fontDir)
	if err != nil {
		return nil, err
	}
	return NewSubsetter(fms, s.config), nil
}
