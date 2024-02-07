package bark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/MangataL/BangumiBuddy/internal/notice"
	"github.com/MangataL/BangumiBuddy/internal/utils"
)

const (
	groupName = "BangumiBuddy"
	icon      = "https://raw.githubusercontent.com/MangataL/BangumiBuddy/main/webui/public/logo.png"
)

// notifier 实现notice.Notifier接口，通过Bark发送通知
type notifier struct {
	mu  sync.Mutex
	cfg Config
}

// Config Bark通知配置
type Config struct {
	ServerPath   string  `mapstructure:"server_path" json:"serverPath"`                     // Bark服务器地址
	Sound        *string `mapstructure:"sound" json:"sound"`                                // 通知铃声
	Interruption *string `mapstructure:"interruption" json:"interruption" default:"active"` // 中断级别
	AutoSave     *bool   `mapstructure:"auto_save" json:"autoSave" default:"true"`          // 自动保存
}

// NewBarkNotifier 创建新的Bark通知器实例
func NewBarkNotifier(cfg Config) notice.Notifier {
	return &notifier{
		cfg: cfg,
	}
}

// 发送Bark通知
func (n *notifier) sendNotification(title, body string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 构建POST请求体
	requestData := map[string]interface{}{
		"title": title,
		"body":  body,
		"icon":  icon,
		"group": groupName,
	}

	// 设置通知铃声
	if n.cfg.Sound != nil {
		requestData["sound"] = *n.cfg.Sound
	}

	// 设置中断级别
	if n.cfg.Interruption != nil {
		requestData["level"] = *n.cfg.Interruption
	}

	// 设置自动保存
	if n.cfg.AutoSave != nil && *n.cfg.AutoSave {
		requestData["isArchive"] = 1
	}

	// 序列化为JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("序列化Bark请求数据失败: %w", err)
	}

	// 构建API URL
	apiURL := n.cfg.ServerPath

	// 创建POST请求
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建Bark POST请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送HTTP请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("发送Bark通知失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("发送Bark通知失败: %d, %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// NoticeSubscriptionUpdated 实现Notifier接口，通知订阅更新状态
func (n *notifier) NoticeSubscriptionUpdated(ctx context.Context, req notice.NoticeSubscriptionUpdatedReq) error {
	var title, body string

	title = fmt.Sprintf("番剧订阅更新：%s", req.BangumiName)
	baseBody := fmt.Sprintf("季度: 第%d季\n字幕组: %s\nRSS订阅项: %s",
		req.Season, req.ReleaseGroup, req.RSSGUID)
	if req.Error != nil {
		body = fmt.Sprintf("%s\n下载失败: %s", baseBody, req.Error.Error())
	} else {
		body = fmt.Sprintf("%s\n开始下载...", baseBody)
	}

	return n.sendNotification(title, body)
}

// NoticeDownloaded 实现Notifier接口，通知资源下载状态
func (n *notifier) NoticeDownloaded(ctx context.Context, req notice.NoticeDownloadedReq) error {
	var title, body string

	if req.Failed {
		title = "番剧下载失败"
		body = fmt.Sprintf("文件名: %s\n错误: %s", req.TorrentName, req.FailDetail)
	} else {
		title = "番剧下载完成"
		body = fmt.Sprintf("文件名: %s\n大小: %s\n耗时: %s\n平均速度: %s",
			req.TorrentName,
			utils.FormatFileSize(req.Size),
			utils.FormatDuration(req.Cost),
			utils.CalculateAverageSpeed(req.Size, req.Cost))
	}

	return n.sendNotification(title, body)
}

// NoticeTransferred 实现Notifier接口，通知资源转移状态
func (n *notifier) NoticeTransferred(ctx context.Context, req notice.NoticeTransferredReq) error {
	var title, body string

	baseBody := fmt.Sprintf("季度: 第%d季\n字幕组: %s\n文件名: %s\nRSS订阅项: %s",
		req.Season, req.ReleaseGroup, req.FileName, req.RSSGUID)

	if !req.Transferred || req.Error != nil {
		title = fmt.Sprintf("番剧转移失败：%s", req.BangumiName)

		errorMsg := "未知错误"
		if req.Error != nil {
			errorMsg = req.Error.Error()
		}

		body = fmt.Sprintf("%s\n文件: %s\n错误: %s",
			baseBody, req.FileName, errorMsg)
	} else {
		title = fmt.Sprintf("番剧转移成功：%s", req.BangumiName)
		body = fmt.Sprintf("%s\n媒体库信息: %s",
			baseBody, req.MediaFilePath)
	}

	return n.sendNotification(title, body)
}
