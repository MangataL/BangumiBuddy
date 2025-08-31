package ass

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/image/font/sfnt"

	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
	"github.com/MangataL/BangumiBuddy/pkg/utils"
)

var (
	ErrFontNotFound = errors.New("字体未找到")
)

type FontMetaSet struct {
	mu      sync.Mutex
	fontDir string
	repo    FontMetaRepository
}

type FontMetaRepository interface {
	Clean(ctx context.Context) error
	Save(ctx context.Context, fontMetas []FontMeta) error
	Find(ctx context.Context, req FindFontMetaReq) ([]FontMeta, error)
	Count(ctx context.Context) (int64, error)
}

type FindFontMetaReq struct {
	FullName       string
	PostScriptName string
	FamilyName     string
	Type           FontType
}

func NewFontMetaSet(fontDir string, repo FontMetaRepository) *FontMetaSet {
	return &FontMetaSet{fontDir: fontDir, repo: repo}
}

func (r *FontMetaSet) GetFontMetaSetStats(ctx context.Context) (subtitle.FontMetaSetStats, error) {
	total, err := r.repo.Count(ctx)
	if err != nil {
		return subtitle.FontMetaSetStats{}, err
	}
	var initDone bool
	if total == 0 {
		initDone = false
	} else {
		if ok := r.mu.TryLock(); ok {
			initDone = true
			r.mu.Unlock()
		}
	}
	return subtitle.FontMetaSetStats{
		Total:    total,
		InitDone: initDone,
	}, nil
}

func (r *FontMetaSet) Init(ctx context.Context, useSystemFontsDir bool) error {
	if ok := r.mu.TryLock(); !ok {
		return errors.New("字体库正在初始化中，请稍后再试")
	}
	defer r.mu.Unlock()

	if err := r.repo.Clean(ctx); err != nil {
		return fmt.Errorf("清理已有字体元数据失败: %w", err)
	}

	fontPaths := r.getFontPaths(useSystemFontsDir)
	if len(fontPaths) == 0 {
		return errors.New("未找到字体文件")
	}

	var allFontMetas []FontMeta
	for _, fontPath := range fontPaths {
		fontMetas, err := r.parseFontMeta(fontPath)
		if err != nil {
			log.Warnf(ctx, "解析字体元数据失败 %s: %s", fontPath, err)
			continue
		}

		allFontMetas = append(allFontMetas, fontMetas...)
	}

	if err := r.repo.Save(ctx, allFontMetas); err != nil {
		return fmt.Errorf("保存字体元数据失败: %w", err)
	}
	return nil
}

func (r *FontMetaSet) getFontPaths(useSystemFontsDir bool) []string {
	dirs := []string{r.fontDir}
	if useSystemFontsDir {
		dirs = append(dirs, getDefaultFontPaths()...)
	}
	var fontPaths []string
	for _, dir := range dirs {
		filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // 忽略错误
			}
			if !d.IsDir() && utils.IsFontFile(d.Name()) {
				if absPath, err := filepath.Abs(path); err == nil {
					fontPaths = append(fontPaths, absPath)
				}
				// 如果获取绝对路径失败，跳过该文件（通常不会发生）
			}
			return nil
		})
	}
	return fontPaths
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func (r *FontMetaSet) parseFontMeta(fontPath string) ([]FontMeta, error) {
	file, err := os.Open(fontPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 读取文件内容到内存
	fontData, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// 获取字体类型
	fontType := getFontType(fontPath)

	// 尝试解析为字体集合 (TTC/OTC)
	collection, err := sfnt.ParseCollection(fontData)
	if err == nil {
		// 是字体集合文件，遍历每个字体
		numFonts := collection.NumFonts()
		fontMetas := make([]FontMeta, 0, numFonts)

		for i := 0; i < numFonts; i++ {
			font, err := collection.Font(i)
			if err != nil {
				continue
			}

			meta, err := r.extractFontMeta(font, fontPath, i, fontType, fontData)
			if err != nil {
				continue
			}
			fontMetas = append(fontMetas, meta)
		}

		return fontMetas, nil
	}

	// 不是字体集合，尝试解析为单个字体
	font, err := sfnt.Parse(fontData)
	if err != nil {
		return nil, err
	}

	meta, err := r.extractFontMeta(font, fontPath, 0, fontType, fontData)
	if err != nil {
		return nil, err
	}

	return []FontMeta{meta}, nil
}

// getFontType 根据文件扩展名获取字体类型
func getFontType(fontPath string) FontType {
	ext := strings.ToLower(filepath.Ext(fontPath))
	if ext == ".otf" || ext == ".otc" {
		return FontTypeOTF
	}
	return FontTypeTTF
}

// extractFontMeta 从 sfnt.Font 提取字体元数据
func (r *FontMetaSet) extractFontMeta(font *sfnt.Font, fontPath string, index int, fontType FontType, fontData []byte) (FontMeta, error) {
	var meta FontMeta

	// 设置位置信息
	meta.Location = FontLocation{
		Path:  fontPath,
		Index: index,
	}
	meta.Type = fontType

	// 提取字体名称信息
	// NameID 定义: https://docs.microsoft.com/en-us/typography/opentype/spec/name
	// 1 = Font Family name
	// 4 = Full font name
	// 6 = PostScript name

	// 获取 Family Name (NameID 1)
	familyName, err := font.Name(nil, sfnt.NameIDFamily)
	if err == nil {
		meta.FamilyName = familyName
	}

	// 获取 Full Name (NameID 4)
	fullName, err := font.Name(nil, sfnt.NameIDFull)
	if err == nil {
		meta.FullName = fullName
	}

	// 获取 PostScript Name (NameID 6)
	postScriptName, err := font.Name(nil, sfnt.NameIDPostScript)
	if err == nil {
		meta.PostScriptName = postScriptName
	}

	// 提取字重和倾斜度
	// 检查是否是字体集合文件（TTC/OTC）
	if len(fontData) >= 4 && string(fontData[0:4]) == "ttcf" {
		// 是 TTC/OTC 文件，需要特殊处理
		// TTC 文件中的表偏移量是相对于整个文件的，不能简单截取
		weight, slant := extractWeightAndSlantFromTTC(fontData, index)
		meta.BoldWeight = int(weight)
		meta.Italic = slant
	} else {
		// 单字体文件，直接解析
		weight, slant := extractWeightAndSlant(fontData, 0)
		meta.BoldWeight = int(weight)
		meta.Italic = slant
	}

	return meta, nil
}

// extractWeightAndSlantFromTTC 从 TTC 文件中提取指定字体的字重和倾斜度
func extractWeightAndSlantFromTTC(fontData []byte, fontIndex int) (weight uint, slant bool) {
	// 获取字体在 TTC 中的偏移量
	if len(fontData) < 16+fontIndex*4 {
		return WeightNormal, false
	}

	offsetPos := 12 + fontIndex*4
	fontOffset := int(fontData[offsetPos])<<24 | int(fontData[offsetPos+1])<<16 |
		int(fontData[offsetPos+2])<<8 | int(fontData[offsetPos+3])

	if fontOffset >= len(fontData) {
		return WeightNormal, false
	}

	return extractWeightAndSlant(fontData, fontOffset)
}

// extractWeightAndSlant 从字体数据中提取字重和倾斜度
func extractWeightAndSlant(fontData []byte, fontOffset int) (weight uint, slant bool) {
	// 默认值
	weight = WeightNormal
	slant = false

	// 查找 OS/2 表
	os2Table := findTableAtOffset(fontData, "OS/2", fontOffset)
	if len(os2Table) < 78 {
		return
	}

	// usWeightClass 位于 OS/2 表偏移量 4-5 (uint16, big-endian)
	weight = uint(os2Table[4])<<8 | uint(os2Table[5])

	// fsSelection 位于 OS/2 表偏移量 62-63 (uint16, big-endian)
	fsSelection := uint(os2Table[62])<<8 | uint(os2Table[63])

	// fsSelection bit 0: ITALIC
	// fsSelection bit 9: OBLIQUE
	// 只要其中一个标志位被设置，就认为是倾斜字体
	slant = (fsSelection&0x01 != 0) || (fsSelection&0x200 != 0)

	return
}

// findTableAtOffset 从指定偏移量开始查找表
// fontOffset: 字体在文件中的起始偏移量（对于单字体文件是 0，对于 TTC 文件是字体的偏移量）
// 返回的表数据中的偏移量是相对于整个文件的
func findTableAtOffset(fontData []byte, tag string, fontOffset int) []byte {
	if fontOffset+12 > len(fontData) {
		return nil
	}

	// 读取表数量 (相对于 fontOffset 的偏移量 4, uint16)
	numTablesOffset := fontOffset + 4
	numTables := int(fontData[numTablesOffset])<<8 | int(fontData[numTablesOffset+1])

	// 表目录从 fontOffset+12 开始
	tableDirectoryOffset := fontOffset + 12
	for i := 0; i < numTables; i++ {
		entryOffset := tableDirectoryOffset + i*16
		if entryOffset+16 > len(fontData) {
			break
		}

		// 表标签 (4 字节)
		tableTag := string(fontData[entryOffset : entryOffset+4])

		// 表偏移量 (uint32, big-endian) - 相对于整个文件
		tableOffset := int(fontData[entryOffset+8])<<24 | int(fontData[entryOffset+9])<<16 |
			int(fontData[entryOffset+10])<<8 | int(fontData[entryOffset+11])

		// 表长度 (uint32, big-endian)
		tableLength := int(fontData[entryOffset+12])<<24 | int(fontData[entryOffset+13])<<16 |
			int(fontData[entryOffset+14])<<8 | int(fontData[entryOffset+15])

		if tableTag == tag {
			if tableOffset+tableLength > len(fontData) {
				return nil
			}
			return fontData[tableOffset : tableOffset+tableLength]
		}
	}

	return nil
}

func (r *FontMetaSet) FindFontMeta(ctx context.Context, req FindFontReq) (FontMeta, error) {
	// 先查找TTF字体
	fontMetaTTF, deviationValueTTF, err := r.findFontMeta(ctx, req.Font, FontTypeTTF, req.UseSimilarFont)
	if err != nil {
		// ttf匹配不到时，如果可以尝试otf，则尝试otf，否则返回错误
		if errors.Is(err, ErrFontNotFound) && req.UseOTF {
			fontMeta, _, err := r.findFontMeta(ctx, req.Font, FontTypeOTF, req.UseSimilarFont)
			return fontMeta, err
		}
		return FontMeta{}, err
	}

	// 查找OTF字体
	fontMetaOTF, deviationValueOTF, err := r.findFontMeta(ctx, req.Font, FontTypeOTF, req.UseSimilarFont)
	if err != nil {
		// otf查找不到时返回最相似的TTF字体
		if errors.Is(err, ErrFontNotFound) {
			return fontMetaTTF, nil
		}
		return FontMeta{}, err
	}

	// 两者都查找成功，选择一个偏离值最小的字体
	if deviationValueTTF <= deviationValueOTF {
		return fontMetaTTF, nil
	}
	return fontMetaOTF, nil
}

func (r *FontMetaSet) findFontMeta(ctx context.Context, font Font, fontType FontType, useSimilarFont bool) (FontMeta, int, error) {
	var deviationValue = math.MaxInt
	// 先用FullName查找
	fontMetas, err := r.repo.Find(ctx, FindFontMetaReq{
		FullName: font.FontName,
		Type:     fontType,
	})
	if err != nil {
		return FontMeta{}, 0, err
	}
	if len(fontMetas) != 0 {
		return fontMetas[0], 0, nil
	}

	// 再用PostScriptName查找
	fontMetas, err = r.repo.Find(ctx, FindFontMetaReq{
		PostScriptName: font.FontName,
		Type:           fontType,
	})
	if err != nil {
		return FontMeta{}, 0, err
	}

	if len(fontMetas) != 0 {
		return fontMetas[0], 0, nil
	}

	if !useSimilarFont {
		return FontMeta{}, 0, ErrFontNotFound
	}

	// 查找同Family字体，寻找最相似的字体
	fontMetas, err = r.repo.Find(ctx, FindFontMetaReq{
		FamilyName: font.FontName,
		Type:       fontType,
	})
	if err != nil {
		return FontMeta{}, 0, err
	}
	var (
		bestMatchFont FontMeta
	)
	for _, fontMeta := range fontMetas {
		if fontMeta.Italic != font.Italic {
			continue
		}
		if fontMeta.BoldWeight == font.BoldWeight {
			return fontMeta, 0, nil
		}
		deviation := int(math.Abs(float64(fontMeta.BoldWeight - font.BoldWeight)))
		if deviation < deviationValue {
			deviationValue = deviation
			bestMatchFont = fontMeta
		}
	}
	if deviationValue == math.MaxInt {
		return FontMeta{}, 0, ErrFontNotFound
	}

	return bestMatchFont, deviationValue, nil
}

func (r *FontMetaSet) UsingTempFontDir(ctx context.Context, fontDir string) (*FontMetaSet, error) {
	repo := NewTempFontMetaRepository(NewMemoryRepository(), r.repo)
	fms := &FontMetaSet{
		fontDir: fontDir,
		repo:    repo,
	}
	if err := fms.Init(ctx, false); err != nil {
		return nil, err
	}
	return fms, nil
}
