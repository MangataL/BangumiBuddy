package gin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"slices"
	"strings"

	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/gin-gonic/gin"
)

// GetLogContent 获取日志文件内容
// GET /apis/v1/logs?level=debug&keyword=error
func GetLogContent(c *gin.Context) {
	fileName := log.GetFileName()
	if fileName == "" {
		c.JSON(http.StatusOK, gin.H{"error": "日志文件名不存在"})
		return
	}
	level := c.Query("level")
	keyword := c.Query("keyword")
	// level和keyword参数可选
	logs, err := grepLogs(level, keyword)
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

// grepLogs 使用 grep 过滤日志文件
func grepLogs(level, keyword string) ([]LogEntry, error) {
	var cmd *exec.Cmd

	// 根据过滤条件构建命令
	if level == "" && keyword == "" {
		// 无过滤条件，直接返回最新50条
		cmdStr := fmt.Sprintf("tail -n 50 %s", log.GetFileName())
		cmd = exec.Command("sh", "-c", cmdStr)
	} else if level != "" && keyword == "" {
		// 先按级别过滤，再获取最新50条
		cmdStr := fmt.Sprintf("grep -i '\"level\":\"%s\"' %s | tail -n 50", level, log.GetFileName())
		cmd = exec.Command("sh", "-c", cmdStr)
	} else if level == "" && keyword != "" {
		// 先按关键字过滤，再获取最新50条
		cmdStr := fmt.Sprintf("grep -i \"%s\" %s | tail -n 50", keyword, log.GetFileName())
		cmd = exec.Command("sh", "-c", cmdStr)
	} else {
		// 同时按级别和关键字过滤，再获取最新50条
		cmdStr := fmt.Sprintf("grep -i '\"level\":\"%s\"' %s | grep -i \"%s\" | tail -n 50", level, log.GetFileName(), keyword)
		cmd = exec.Command("sh", "-c", cmdStr)
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
	if len(logs) == 0 {
		return []LogEntry{}, nil
	}
	return logs, nil
}
