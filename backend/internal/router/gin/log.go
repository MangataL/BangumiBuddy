package gin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"slices"
	"strings"

	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/gin-gonic/gin"
)

// GetLogContent 获取日志文件内容
// GET /apis/v1/logs?level=debug
func GetLogContent(c *gin.Context) {
	fileName := log.GetFileName()
	if fileName == "" {
		c.JSON(http.StatusOK, gin.H{"error": "日志文件名不存在"})
		return
	}
	level := c.Query("level")
	// level参数可选，空值时返回全量日志
	logs, err := grepLogs(level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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

// grepLogs 使用 grep 过滤日志文件
func grepLogs(level string) ([]LogEntry, error) {
	var cmd *exec.Cmd
	// 判断level是否为空，为空时返回全量日志
	if level == "" {
		cmd = exec.Command("cat", log.GetFileName())
	} else {
		// 使用通配符 app.log* 匹配主文件和备份文件
		cmd = exec.Command("grep", "-i", `"level":"`+level+`"`, log.GetFileName())
	}
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		// grep 退出码 1 表示未找到匹配行，视为正常情况
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []LogEntry{}, nil
		}
		// 如果文件不存在，返回空结果
		if strings.Contains(err.Error(), "No such file or directory") {
			return []LogEntry{}, nil
		}
		return nil, err
	}

	var logs []LogEntry
	// 按行处理 grep 输出
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // 跳过无效 JSON
		}
		logs = append(logs, entry)
	}

	return logs, nil
}
