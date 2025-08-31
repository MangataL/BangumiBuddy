package ass

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
)

// FontEmbedder 字体嵌入器，负责将子集化的字体嵌入到字幕文件中
type FontEmbedder struct{}

// NewFontEmbedder 创建字体嵌入器
func NewFontEmbedder() *FontEmbedder {
	return &FontEmbedder{}
}

// EmbedFonts 将子集化的字体嵌入到字幕文件中
func (fe *FontEmbedder) EmbedFonts(
	ctx context.Context,
	filePath string,
	fonts map[Font][]byte,
	newFile bool) (string, error) {
	if len(fonts) == 0 {
		return filePath, nil
	}

	// 读取原始字幕文件
	originalContent, err := fe.readOriginalFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取原始字幕文件失败: %w", err)
	}

	ext := filepath.Ext(filePath)
	// 生成嵌入的字体内容
	fontEmbedContent, err := fe.generateFontEmbedContent(fonts, ext)
	if err != nil {
		return "", fmt.Errorf("生成字体嵌入内容失败: %w", err)
	}

	// 合并内容
	finalContent := fe.mergeContent(originalContent, fontEmbedContent)

	// 写入文件
	outputPath := fe.getOutputPath(filePath, newFile)
	if err := fe.writeToFile(outputPath, finalContent); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}
	return outputPath, nil
}

// readOriginalFile 读取原始字幕文件
func (fe *FontEmbedder) readOriginalFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := newBufferScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// generateFontEmbedContent 生成字体嵌入内容
func (fe *FontEmbedder) generateFontEmbedContent(fonts map[Font][]byte, ext string) ([]string, error) {
	var content []string

	// 添加 [Fonts] 节标题
	content = append(content, "[Fonts]")

	names := make([]Font, 0, len(fonts))
	for font := range fonts {
		names = append(names, font)
	}
	sort.Slice(names, func(i, j int) bool {
		return names[i].FontName < names[j].FontName
	})

	for _, font := range names {
		data := fonts[font]
		// 生成字体文件名
		fontFileName := fe.generateFontFileName(font, ext)

		// 添加字体声明行
		fontDeclareLine := fmt.Sprintf("fontname: %s", fontFileName)
		content = append(content, fontDeclareLine)

		// 将字体数据转换为uuencoding格式
		uuencodedData, err := fe.uuencodeFontData(data)
		if err != nil {
			return nil, fmt.Errorf("编码字体数据失败[%s]: %w", font.FontName, err)
		}

		// 添加编码后的字体数据
		content = append(content, uuencodedData...)
	}

	return content, nil
}

// generateFontFileName 生成SSA格式的字体文件名
// 格式：原字体名_样式标识_编码.ttf
// 例如：chaucer_B0.ttf, comic_0.ttf
func (fe *FontEmbedder) generateFontFileName(font Font, ext string) string {
	// 生成样式标识
	var styleSuffix string
	if font.BoldWeight >= WeightBold {
		styleSuffix += "B"
	}
	if font.Italic {
		styleSuffix += "I"
	}

	// 编码标识（通常为0，表示默认编码）
	encodingSuffix := "0"

	// 组合文件名
	fileName := fmt.Sprintf("%s_%s%s%s", font.FontName, styleSuffix, encodingSuffix, ext)

	return fileName
}

// uuencodeFontData 将字体数据转换为uuencoding格式
func (fe *FontEmbedder) uuencodeFontData(data []byte) ([]string, error) {
	var buf strings.Builder
	var err error
	size := len(data)
	written := 0

	for pos := 0; pos < size; pos += 3 {
		src := [3]byte{0, 0, 0}
		n := copy(src[:], data[pos:min(pos+3, size)])

		dst := [4]byte{
			src[0] >> 2,
			((src[0]&0x3)<<4 | (src[1]&0xF0)>>4),
			((src[1]&0xF)<<2 | (src[2]&0xC0)>>6),
			src[2] & 0x3F,
		}

		for i := 0; i < min(n+1, 4); i++ {
			b := dst[i] + 33
			if err = buf.WriteByte(b); err != nil {
				return nil, fmt.Errorf("write error when UUencoding: %w", err)
			}
			written++
			if written == 80 && pos+3 < size {
				if err = buf.WriteByte('\n'); err != nil {
					return nil, fmt.Errorf("write error when UUencoding: %w", err)
				}
				written = 0
			}
		}
	}

	// 将结果按行分割
	result := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	return result, nil
}

// mergeContent 合并原始内容和字体嵌入内容
func (fe *FontEmbedder) mergeContent(originalLines []string, fontEmbedLines []string) []string {
	var result []string
	var fontsSectionFound bool
	var inFontsSection bool

	for _, line := range originalLines {
		trimmedLine := strings.TrimSpace(line)

		// 检查是否遇到 [Fonts] 节
		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			sectionName := strings.ToLower(strings.Trim(trimmedLine, "[]"))
			if sectionName == "fonts" {
				fontsSectionFound = true
				inFontsSection = true
				// 添加 [Fonts] 节标题和嵌入的字体内容
				result = append(result, fontEmbedLines...)
				continue
			} else {
				inFontsSection = false
			}
		}

		// 如果在 [Fonts] 节中，跳过原始内容
		if inFontsSection {
			continue
		}

		result = append(result, line)
	}

	// 如果没有找到 [Fonts] 节，在文件末尾添加
	if !fontsSectionFound {
		result = append(result, "")
		result = append(result, fontEmbedLines...)
	}

	return result
}

// getOutputPath 获取输出文件路径
func (fe *FontEmbedder) getOutputPath(originalPath string, newFile bool) string {
	if newFile {
		dir := filepath.Dir(originalPath)
		baseName := filepath.Base(originalPath)
		// 找到第一个点的位置，在文件名和第一个扩展名之间插入.subset
		firstDot := strings.Index(baseName, ".")
		if firstDot == -1 {
			// 没有扩展名，直接添加.subset
			return filepath.Join(dir, fmt.Sprintf("%s%s", baseName, subtitle.SubtitleSubsetExt))
		}
		name := baseName[:firstDot]
		extensions := baseName[firstDot:]
		return filepath.Join(dir, fmt.Sprintf("%s%s%s", name, subtitle.SubtitleSubsetExt, extensions))
	}
	// 覆盖原文件
	return originalPath
}

// writeToFile 写入文件
func (fe *FontEmbedder) writeToFile(filePath string, content []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range content {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
