package ass

import (
	"os"
	"path/filepath"
)

// 根据 https://wiki.archlinux.org/title/Fonts#Manual_installation 规范获取默认字体路径
func getDefaultFontPaths() []string {
	var paths []string
	linuxPaths := []string{
		"/usr/share/fonts",
		"/usr/local/share/fonts",
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".local", "share", "fonts"))
	}
	for _, p := range linuxPaths {
		if isDir(p) {
			paths = append(paths, p)
		}
	}

	return paths
}
