package ass

import (
	"os"
	"path/filepath"
)

func getDefaultFontPaths() []string {
	var paths []string
	macPaths := []string{
		"/Library/Fonts",
		"/System/Library/Fonts",
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, "Library", "Fonts"))
	}
	for _, p := range macPaths {
		if isDir(p) {
			paths = append(paths, p)
		}
	}
	return paths
}
