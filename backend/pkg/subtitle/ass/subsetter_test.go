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
