package ass

import (
	"os"
	"testing"

	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass/freetype"
)

func TestCreateSubfontData_PreserveChineseNamesAndKeepBasicGlyphs(t *testing.T) {
	fontPath := "./testdata/方正隶变_GBK.ttf"
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		t.Fatalf("read font failed: %v", err)
	}

	lib, err := freetype.NewLibrary()
	if err != nil {
		t.Fatalf("init freetype failed: %v", err)
	}
	defer lib.Done()

	origFace, err := lib.NewFaceFromMemory(fontData, 0)
	if err != nil {
		t.Fatalf("load original font failed: %v", err)
	}
	origNames := freetype.ExtractFontNames(origFace)
	origFace.Done()

	candidates := []rune{'你', '我', '的', '中', '汉', '天', '地', '一'}
	var sample rune
	for _, r := range candidates {
		face, err := lib.NewFaceFromMemory(fontData, 0)
		if err != nil {
			t.Fatalf("load original font failed: %v", err)
		}
		if face.GetCharIndex(r) != 0 {
			sample = r
			face.Done()
			break
		}
		face.Done()
	}
	if sample == 0 {
		t.Skip("no candidate glyph found in test font")
	}

	codePoints := appendAdditionalCodePoints([]rune{sample})
	subsetData, err := CreateSubfontData(fontData, 0, codePoints)
	if err != nil {
		t.Fatalf("CreateSubfontData failed: %v", err)
	}

	subsetFace, err := lib.NewFaceFromMemory(subsetData, 0)
	if err != nil {
		t.Fatalf("load subset font failed: %v", err)
	}
	defer subsetFace.Done()

	// 1) 额外码点：确保空格等基础字符仍在子集中
	if subsetFace.GetCharIndex(' ') == 0 {
		t.Fatalf("subset font missing space glyph")
	}

	// name table：如果原字体有中文名，子集化后也应保留（避免丢中文名导致匹配失败）。
	if origNames.ChineseFamilyName != "" {
		subsetNames := freetype.ExtractFontNames(subsetFace)
		if subsetNames.ChineseFamilyName == "" {
			t.Fatalf("subset font lost Chinese family name (orig=%q)", origNames.ChineseFamilyName)
		}
	}
}
