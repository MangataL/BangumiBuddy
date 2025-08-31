package ass

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// 基础的 ASS 文件模板，除了 Dialogue 之外的内容可以复用
const assBaseTemplate = `[Script Info]
Title: Test Subtitle
ScriptType: v4.00+

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,0,0,0,0,100,100,0,0,1,2,0,2,10,10,10,1
Style: Bold,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,-1,0,0,0,100,100,0,0,1,2,0,2,10,10,10,1
Style: Italic,Arial,20,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,0,-1,0,0,100,100,0,0,1,2,0,2,10,10,10,1
Style: CustomFont,SimHei,20,&H00FFFFFF,&H000000FF,&H00000000,&H00000000,0,0,0,0,100,100,0,0,1,2,0,2,10,10,10,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
%s
`

func TestParseFontSet(t *testing.T) {
	tests := []struct {
		name        string
		dialogue    string
		expectedSet FontSet
	}{
		{
			name:     "无覆写",
			dialogue: `Dialogue: 0,0:00:00.00,0:00:05.00,Default,,0,0,0,,\hHello \nWorld\N`,
			expectedSet: FontSet{
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: false}: buildCodePoint("Hello World"),
			},
		},
		{
			name:     "有粗体覆写",
			dialogue: `Dialogue: 0,0:00:00.00,0:00:05.00,Default,,0,0,0,,Normal {\b1}Bold Text Normal`,
			expectedSet: FontSet{
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: false}: buildCodePoint("Normal "),
				Font{FontName: "Arial", BoldWeight: WeightBold, Italic: false}:   buildCodePoint("Bold Text Normal"),
			},
		},
		{
			name:     "有斜体覆写",
			dialogue: `Dialogue: 0,0:00:00.00,0:00:05.00,Default,,0,0,0,,Normal {\i1}Italic Text{\i} Normal`,
			expectedSet: FontSet{
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: false}: buildCodePoint("Normal "),
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: true}:  buildCodePoint("Italic Text"),
			},
		},
		{
			name:     "有字体名称覆写",
			dialogue: `Dialogue: 0,0:00:00.00,0:00:05.00,Default,,0,0,0,,Normal {\fnSimHei}Chinese Text{\fn} Normal`,
			expectedSet: FontSet{
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: false}:  buildCodePoint("Normal "),
				Font{FontName: "SimHei", BoldWeight: WeightNormal, Italic: false}: buildCodePoint("Chinese Text"),
			},
		},
		{
			name:     "多个关心属性的覆写",
			dialogue: `Dialogue: 0,0:00:00.00,0:00:05.00,Default,,0,0,0,,Normal {\b1\i1\fnSimHei}Bold Italic Custom{\b\i\fn} Normal`,
			expectedSet: FontSet{
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: false}: buildCodePoint("Normal "),
				Font{FontName: "SimHei", BoldWeight: WeightBold, Italic: true}:   buildCodePoint("Bold Italic Custom"),
			},
		},
		{
			name:     "覆写后又覆写",
			dialogue: `Dialogue: 0,0:00:00.00,0:00:05.00,Default,,0,0,0,,Normal {\b-1}Bold{\b0}Extra{\b700}Bold{\b} Normal`,
			expectedSet: FontSet{
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: false}: buildCodePoint("Extra Normal "),
				Font{FontName: "Arial", BoldWeight: WeightBold, Italic: false}:   buildCodePoint("Bold"),
			},
		},
		{
			name:     "覆写后撤销",
			dialogue: `Dialogue: 0,0:00:00.00,0:00:05.00,Default,,0,0,0,,Normal {\b1}Bold{\b}Back to Normal`,
			expectedSet: FontSet{
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: false}: buildCodePoint("Normal Back to"),
				Font{FontName: "Arial", BoldWeight: WeightBold, Italic: false}:   buildCodePoint("Bold"),
			},
		},
		{
			name:     "复杂的混合覆写场景",
			dialogue: `Dialogue: 0,0:00:00.00,0:00:05.00,Default,,0,0,0,,A{\b1}B{\i1}C{\b}D{\i}E`,
			expectedSet: FontSet{
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: false}: buildCodePoint("AE"),
				Font{FontName: "Arial", BoldWeight: WeightBold, Italic: false}:   buildCodePoint("B"),
				Font{FontName: "Arial", BoldWeight: WeightBold, Italic: true}:    buildCodePoint("C"),
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: true}:  buildCodePoint("D"),
			},
		},
		{
			name:     "使用预设的带粗体样式",
			dialogue: `Dialogue: 0,0:00:00.00,0:00:05.00,Bold,,0,0,0,,{\b0}Bold {\b}Style Text`,
			expectedSet: FontSet{
				Font{FontName: "Arial", BoldWeight: WeightBold, Italic: false}:   buildCodePoint("Style Text"),
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: false}: buildCodePoint("Bold "),
			},
		},
		{
			name:     "使用预设的带斜体样式",
			dialogue: `Dialogue: 0,0:00:00.00,0:00:05.00,Italic,,0,0,0,,{\i0}Italic{\i} Style Text`,
			expectedSet: FontSet{
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: false}: buildCodePoint("Italic"),
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: true}:  buildCodePoint(" Style Text"),
			},
		},
		{
			name:     "多次字体名称覆写",
			dialogue: `Dialogue: 0,0:00:00.00,0:00:05.00,Default,,0,0,0,,A{\fnSimHei}B{\fnTimes New Roman}C{\fn}D`,
			expectedSet: FontSet{
				Font{FontName: "Arial", BoldWeight: WeightNormal, Italic: false}:           buildCodePoint("AD"),
				Font{FontName: "SimHei", BoldWeight: WeightNormal, Italic: false}:          buildCodePoint("B"),
				Font{FontName: "Times New Roman", BoldWeight: WeightNormal, Italic: false}: buildCodePoint("C"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建临时 ASS 文件
			content := fmt.Sprintf(assBaseTemplate, tt.dialogue)
			tmpFile := createTempASSFile(t, content)
			defer os.Remove(tmpFile)

			// 解析字体集
			parser := NewParser()
			fontSet, needTransfer, err := parser.ParseFontSet(context.Background(), tmpFile)
			if err != nil {
				t.Fatalf("ParseFontSet failed: %v", err)
			}

			if needTransfer {
				t.Fatal("Expected needTransfer to be false")
			}

			// 验证 FontSet
			compareFontSets(t, fontSet, tt.expectedSet)
		})
	}
}

// buildCodePoint 从字符串构建 CodePoint
func buildCodePoint(text string) CodePoints {
	cp := make(CodePoints)
	for _, char := range text {
		cp[char] = struct{}{}
	}
	return cp
}

// compareFontSets 比较两个 FontSet 是否相等
func compareFontSets(t *testing.T, actual, expected FontSet) {
	t.Helper()

	// 检查 FontSet 的大小
	if len(actual) != len(expected) {
		t.Errorf("FontSet size mismatch: got %d fonts, want %d fonts", len(actual), len(expected))
		t.Logf("Actual FontSet:")
		for font := range actual {
			t.Logf("  %+v", font)
		}
		t.Logf("Expected FontSet:")
		for font := range expected {
			t.Logf("  %+v", font)
		}
		return
	}

	// 检查每个期望的字体是否存在
	for expectedFont, expectedCodePoint := range expected {
		actualCodePoint, exists := actual[expectedFont]
		if !exists {
			t.Errorf("Expected font %+v not found in actual FontSet", expectedFont)
			continue
		}

		// 比较 CodePoint
		compareCodePoints(t, actualCodePoint, expectedCodePoint, expectedFont)
	}

	// 检查是否有额外的字体
	for actualFont := range actual {
		if _, exists := expected[actualFont]; !exists {
			t.Errorf("Unexpected font %+v found in actual FontSet", actualFont)
		}
	}
}

// compareCodePoints 比较两个 CodePoint 是否相等
func compareCodePoints(t *testing.T, actual, expected CodePoints, font Font) {
	t.Helper()

	// 检查 CodePoint 的大小
	if len(actual) != len(expected) {
		t.Errorf("CodePoint size mismatch for font %+v: got %d chars, want %d chars", font, len(actual), len(expected))

		// 打印实际的字符
		actualChars := make([]rune, 0, len(actual))
		for char := range actual {
			actualChars = append(actualChars, char)
		}
		t.Logf("  Actual chars: %q", string(actualChars))

		// 打印期望的字符
		expectedChars := make([]rune, 0, len(expected))
		for char := range expected {
			expectedChars = append(expectedChars, char)
		}
		t.Logf("  Expected chars: %q", string(expectedChars))
	}

	// 检查每个期望的字符是否存在
	for expectedChar := range expected {
		if _, exists := actual[expectedChar]; !exists {
			t.Errorf("Expected character '%c' (U+%04X) not found in CodePoint for font %+v", expectedChar, expectedChar, font)
		}
	}

	// 检查是否有额外的字符
	for actualChar := range actual {
		if _, exists := expected[actualChar]; !exists {
			t.Errorf("Unexpected character '%c' (U+%04X) found in CodePoint for font %+v", actualChar, actualChar, font)
		}
	}
}

// createTempASSFile 创建临时 ASS 文件
func createTempASSFile(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.ass")

	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	return tmpFile
}
