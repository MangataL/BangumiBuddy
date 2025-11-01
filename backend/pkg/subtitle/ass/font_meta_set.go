package ass

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass/freetype"
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
	FullName              string
	PostScriptName        string
	FamilyName            string
	ChineseFullName       string
	ChinesePostScriptName string
	ChineseFamilyName     string
	Type                  FontType
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
	// 初始化 FreeType 库
	lib, err := freetype.NewLibrary()
	if err != nil {
		return nil, fmt.Errorf("初始化FreeType库失败: %w", err)
	}
	defer lib.Done()

	// 获取字体类型
	fontType := getFontType(fontPath)

	// 先尝试加载索引0的字体，检查是否是字体集合
	face0, err := lib.NewFaceFromFile(fontPath, 0)
	if err != nil {
		return nil, fmt.Errorf("加载字体失败: %w", err)
	}

	numFaces := face0.NumFaces()
	face0.Done()

	// 如果是字体集合，遍历所有字体
	if numFaces > 1 {
		fontMetas := make([]FontMeta, 0, numFaces)
		for i := 0; i < numFaces; i++ {
			face, err := lib.NewFaceFromFile(fontPath, i)
			if err != nil {
				log.Warnf(context.Background(), "加载字体集合 %s 第 %d 个字体失败: %s", fontPath, i, err)
				continue
			}

			meta, err := r.extractFontMeta(face, fontPath, i, fontType)
			face.Done()
			if err != nil {
				log.Warnf(context.Background(), "提取字体集合 %s 第 %d 个字体元数据失败: %s", fontPath, i, err)
				continue
			}
			fontMetas = append(fontMetas, meta)
		}
		return fontMetas, nil
	}

	// 单个字体
	face, err := lib.NewFaceFromFile(fontPath, 0)
	if err != nil {
		return nil, fmt.Errorf("加载字体失败: %w", err)
	}
	defer face.Done()

	meta, err := r.extractFontMeta(face, fontPath, 0, fontType)
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

// extractFontMeta 从 FreeType Face 提取字体元数据
func (r *FontMetaSet) extractFontMeta(face *freetype.Face, fontPath string, index int, fontType FontType) (FontMeta, error) {
	var meta FontMeta

	// 设置位置信息
	meta.Location = FontLocation{
		Path:  fontPath,
		Index: index,
	}
	meta.Type = fontType

	// 提取字体名称（包括英文和中文）
	names := freetype.ExtractFontNames(face)
	meta.FamilyName = names.FamilyName
	meta.FullName = names.FullName
	meta.PostScriptName = names.PostScriptName
	meta.ChineseFamilyName = names.ChineseFamilyName
	meta.ChineseFullName = names.ChineseFullName
	meta.ChinesePostScriptName = names.ChinesePostScriptName

	// 提取字重和倾斜度
	weight, italic, err := face.GetOS2Table()
	if err != nil {
		// 如果无法获取OS/2表，使用默认值
		log.Warnf(context.Background(), "无法获取字体 %s 的OS/2表: %s，使用默认值", fontPath, err)
		styleFlags := face.StyleFlags()
		meta.BoldWeight = WeightNormal
		if styleFlags&freetype.StyleFlagBold != 0 {
			meta.BoldWeight = WeightBold
		}
		meta.Italic = styleFlags&freetype.StyleFlagItalic != 0
	} else {
		meta.BoldWeight = weight
		meta.Italic = italic
	}

	return meta, nil
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

// ParseFontMetaTest 测试用方法，直接解析字体元数据（不使用repository）
func (r *FontMetaSet) ParseFontMetaTest(fontPath string) ([]FontMeta, error) {
	return r.parseFontMeta(fontPath)
}
