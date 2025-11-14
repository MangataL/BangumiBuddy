package gin

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/icza/backscanner"
)

// GetLogContent 获取日志文件内容
// GET /apis/v1/logs?level=debug&keyword=error
func GetLogContent(c *gin.Context) {
	fileName := log.GetFileName()
	if fileName == "" {
		c.JSON(http.StatusOK, gin.H{"error": "日志文件名不存在"})
		return
	}
	const (
		defaultLimit = 50
		maxLimit     = 200
	)
	level := c.Query("level")
	keyword := c.Query("keyword")
	limit := defaultLimit
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			if parsedLimit > maxLimit {
				parsedLimit = maxLimit
			}
			limit = parsedLimit
		}
	}
	offset := 0
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}
	// level和keyword参数可选
	logs, err := grepLogs(level, keyword, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 按时间戳倒序排序
	slices.SortFunc(logs, func(a, b LogEntry) int {
		return int(b.Ts - a.Ts)
	})

	c.JSON(http.StatusOK, logs)
}

// LogEntry 表示日志文件的单行 JSON 结构
type LogEntry struct {
	Level   string  `json:"level"`
	Ts      float64 `json:"ts"`
	Message string  `json:"msg"`
}

// grepLogs 借助 backscanner 倒序遍历，只保留 limit 大小的结果集
func grepLogs(level, keyword string, limit, offset int) ([]LogEntry, error) {
	if limit <= 0 {
		return []LogEntry{}, nil
	}
	fileName := log.GetFileName()
	f, err := os.Open(fileName)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []LogEntry{}, nil
		}
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() == 0 {
		return []LogEntry{}, nil
	}

	scanner := backscanner.New(f, int(info.Size()))
	results := make([]LogEntry, 0, limit)
	keywordLower := strings.ToLower(keyword)
	levelPattern := `"level":"` + strings.ToLower(level) + `"`

	needCaseInsensitive := level != "" || keyword != ""
	skipped := 0

	for len(results) < limit {
		line, _, err := scanner.Line()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if errors.Is(err, backscanner.ErrLongLine) {
				// 行内容过长，跳过这一条
				continue
			}
			return nil, err
		}
		compareLine := line
		if needCaseInsensitive {
			compareLine = strings.ToLower(line)
			if keyword != "" {
				if !strings.Contains(compareLine, keywordLower) {
					continue
				}
			}
			if level != "" {
				if !strings.Contains(compareLine, levelPattern) {
					continue
				}
			}
		}

		if skipped < offset {
			skipped++
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		results = append(results, entry)
	}

	return results, nil
}
