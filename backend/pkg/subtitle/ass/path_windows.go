package ass

import (
	"os"
	"path/filepath"
)

func getDefaultFontPaths() []string {
	if windir := os.Getenv("WINDIR"); windir != "" {
		fontDir := filepath.Join(windir, "Fonts")
		if isDir(fontDir) {
			return []string{fontDir}
		}
	}
	return nil
}
