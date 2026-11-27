package ass

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
)

func TestPreflightIncludesMovedFonts(t *testing.T) {
	dir := t.TempDir()
	repo := NewMemoryRepository()
	var fonts []FontMeta
	for _, name := range []string{"Moved A", "Moved B"} {
		path := filepath.Join(dir, name+".ttf")
		require.NoError(t, os.WriteFile(path, nil, 0600))
		fonts = append(fonts, FontMeta{FullName: name, Type: FontTypeTTF, Location: FontLocation{Path: path}})
	}
	require.NoError(t, repo.Replace(t.Context(), fonts))
	for _, font := range fonts {
		require.NoError(t, os.Remove(font.Location.Path))
	}
	s := NewSubsetter(NewFontMetaSet(dir, repo), FontSubsetterConfig{})
	file := filepath.Join(dir, "example.ass")
	original := "[V4+ Styles]\nFormat: Name, Fontname, Bold, Italic\nStyle: Default,Moved A,0,0\n[Events]\nFormat: Style,Text\nDialogue: Default,A{\\fnMoved B}B{\\fnNever Installed}C"
	require.NoError(t, os.WriteFile(file, []byte(original), 0600))
	write, err := s.PrepareSubset(t.Context(), file)
	require.Nil(t, write)
	var missing subtitle.MissingFonts
	require.True(t, errors.As(err, &missing))
	require.Len(t, missing, 3)
	for _, name := range []string{"Moved A", "Moved B", "Never Installed"} {
		require.Contains(t, missing, Font{FontName: name, BoldWeight: WeightNormal})
	}
	data, err := os.ReadFile(file)
	require.NoError(t, err)
	require.Equal(t, original, string(data))
}

func TestGlyphCheckKeepsFontBytesOutOfPreparedSubtitles(t *testing.T) {
	file := filepath.Join(t.TempDir(), "01.ass")
	require.NoError(t, os.WriteFile(file, []byte("[Events]\n"), 0600))
	fonts := []SubsetFont{{Location: FontLocation{Path: "testdata/multilingual.ttf"}, CodePoints: CodePoints{'A': {}}}}
	s := NewSubsetter(nil, FontSubsetterConfig{CheckGlyphs: true, GenerateNewFile: true})
	subsetPath := filepath.Join(filepath.Dir(file), "01.subset.ass")
	output, err := s.embedSubset(t.Context(), file, subsetPath, fonts)
	require.NoError(t, err)
	data, err := os.ReadFile(output)
	require.NoError(t, err)
	require.Contains(t, string(data), "[Fonts]")
	require.Nil(t, fonts[0].FontData)

	fonts[0].CodePoints['界'] = struct{}{}
	_, err = s.embedSubset(t.Context(), file, subsetPath, fonts)
	require.ErrorContains(t, err, "缺失")
	require.Nil(t, fonts[0].FontData)
}
