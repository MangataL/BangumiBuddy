package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Tnze/go.num/v2/zh"
	"github.com/pkg/errors"
)

var subtitleExtensions = map[string]struct{}{
	".srt": {},
	".ass": {},
	".ssa": {},
	".sub": {},
	".idx": {},
	".vtt": {},
}

var mediaExtensions = map[string]struct{}{
	".mkv": {},
	".mp4": {},
	".avi": {},
	".wmv": {},
}

var fontExtensions = map[string]struct{}{
	".ttf": {},
	".otf": {},
	".ttc": {},
	".otc": {},
}

func IsMediaFile(file string) bool {
	_, ok := mediaExtensions[strings.ToLower(filepath.Ext(file))]
	return ok
}

func IsSubtitleFile(file string) bool {
	_, ok := subtitleExtensions[strings.ToLower(filepath.Ext(file))]
	return ok
}

func IsFontFile(file string) bool {
	_, ok := fontExtensions[strings.ToLower(filepath.Ext(file))]
	return ok
}

func FindSameBaseFiles(filePath string) ([]string, error) {
	dir := filepath.Dir(filePath)
	baseNameWithoutExt := GetFileBaseName(filepath.Base(filePath))
	// 读取目录下的所有文件
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// 存储匹配的文件路径
	var result []string
	for _, file := range files {
		// 跳过目录，只处理文件
		if file.IsDir() {
			continue
		}
		fileName := file.Name()
		if strings.HasPrefix(fileName, baseNameWithoutExt+".") {
			result = append(result, filepath.Join(dir, fileName))
		}
	}
	return result, nil
}

func GetFileBaseName(fp string) string {
	return strings.TrimSuffix(fp, filepath.Ext(fp))
}

func FormatNumber(num int) string {
	if num < 10 {
		return fmt.Sprintf("0%d", num)
	}
	return fmt.Sprintf("%d", num)
}

// formatDuration 格式化时间间隔为人类可读格式
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)

	h := d / time.Hour
	d -= h * time.Hour

	m := d / time.Minute
	d -= m * time.Minute

	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%d小时%d分钟%d秒", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%d分钟%d秒", m, s)
	}
	return fmt.Sprintf("%d秒", s)
}

// formatFileSize 将字节数格式化为人类可读格式
func FormatFileSize(size int64) string {
	const (
		B  int64 = 1
		KB       = B * 1024
		MB       = KB * 1024
		GB       = MB * 1024
		TB       = GB * 1024
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

func CalculateAverageSpeed(size int64, duration time.Duration) string {
	if duration == 0 {
		return "N/A"
	}
	bytesPerSecond := float64(size) / duration.Seconds()
	return FormatFileSize(int64(bytesPerSecond)) + "/s"
}

func ParseEpisodeWithLocation(name string, location string) (int, error) {
	pattern := regexp.QuoteMeta(location)
	pattern = strings.ReplaceAll(pattern, `\{ep\}`, `(\d+|[一二三四五六七八九十百千]+)`)

	reg := regexp.MustCompile(pattern)
	matches := reg.FindStringSubmatch(name)
	if len(matches) < 2 {
		return 0, errors.New("无法从文件名中解析出集数信息")
	}

	epStr := matches[1]

	// 如果是数字直接转换
	if num, err := strconv.Atoi(epStr); err == nil {
		return num, nil
	}

	var ep zh.Uint64
	if _, err := fmt.Sscan(epStr, &ep); err != nil {
		return 0, errors.WithMessage(err, "转换中文集数失败")
	}

	return int(ep), nil
}
