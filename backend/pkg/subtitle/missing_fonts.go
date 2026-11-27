package subtitle

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var ErrFontNotFound = errors.New("字体未找到")

type FontStyle struct {
	FontName   string
	BoldWeight int
	Italic     bool
}

var weightNames = map[int]string{
	100: "极细", 200: "特细", 300: "细", 400: "", 500: "中等",
	600: "半粗", 700: "粗", 800: "特粗", 900: "超粗",
}

func (f FontStyle) styleLabel() string {
	weight, known := weightNames[f.BoldWeight]
	if !known {
		label := fmt.Sprintf("字重 %d", f.BoldWeight)
		if f.Italic {
			label += "，斜体"
		}
		return label
	}
	if f.Italic {
		return weight + "斜体"
	}
	if weight == "" {
		return ""
	}
	return weight + "体"
}

// MissingFonts lists the subtitles using each unavailable face.
type MissingFonts map[FontStyle][]string

func (m MissingFonts) Error() string {
	lines := make([]string, 0, len(m))
	for font, files := range m {
		line := font.FontName
		if style := font.styleLabel(); style != "" {
			line += "（" + style + "）"
		}
		if len(files) > 0 {
			files = append([]string(nil), files...)
			sort.Strings(files)
			line += "：" + strings.Join(files, "、")
		}
		lines = append(lines, line)
	}
	sort.Strings(lines)
	return fmt.Sprintf("缺少 %d 款字体：\n%s", len(m), strings.Join(lines, "\n"))
}

func (m MissingFonts) Unwrap() error { return ErrFontNotFound }
