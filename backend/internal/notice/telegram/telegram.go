package telegram

import (
	"context"
	"fmt"
	"sync"

	"github.com/MangataL/BangumiBuddy/internal/notice"
	"github.com/MangataL/BangumiBuddy/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// notifier 实现notice.Notifier接口，通过Telegram机器人发送通知
type notifier struct {
	bot *tgbotapi.BotAPI
	mu  sync.Mutex
	cfg Config
}

type Config struct {
	Token  string `mapstructure:"token" json:"token"`
	ChatID int64  `mapstructure:"chat_id" json:"chatID"`
}

// NewTelegramNotifier 创建新的TelegramNotifier实例
func NewTelegramNotifier(cfg Config) notice.Notifier {
	return &notifier{
		cfg: cfg,
	}
}

func (t *notifier) init() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.bot != nil {
		return nil
	}

	bot, err := tgbotapi.NewBotAPI(t.cfg.Token)
	if err != nil {
		return fmt.Errorf("初始化Telegram机器人失败: %w", err)
	}

	t.bot = bot
	t.bot.Debug = true
	return nil
}

// NoticeSubscriptionUpdated 实现Notifier接口，通知订阅更新状态
func (t *notifier) NoticeSubscriptionUpdated(ctx context.Context, req notice.NoticeSubscriptionUpdatedReq) error {
	if err := t.init(); err != nil {
		return err
	}

	var text, emoji string

	// 构建基本信息部分
	baseInfo := fmt.Sprintf("番剧订阅更新通知\n\n"+
		"📺 番剧: %s\n"+
		"🔢 季度: 第%d季\n"+
		"👥 字幕组: %s\n"+
		"🔗 RSS订阅项: %s\n",
		req.BangumiName, req.Season, req.ReleaseGroup, req.RSSGUID)

	// 处理成功/失败状态
	if req.Error != nil {
		emoji = "❌"
		text = fmt.Sprintf("%s\n%s 下载失败\n⚠️ 错误详情: %s", baseInfo, emoji, req.Error.Error())
	} else {
		emoji = "⏬"
		text = fmt.Sprintf("%s\n%s 开始下载...", baseInfo, emoji)
	}

	// 创建消息
	msg := tgbotapi.NewMessage(t.cfg.ChatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	// 添加海报图片
	if req.Poster != "" {
		photoMsg := tgbotapi.NewPhoto(t.cfg.ChatID, tgbotapi.FileURL(req.Poster))
		photoMsg.Caption = text

		_, err := t.bot.Send(photoMsg)
		return err
	}
	// 发送主消息
	_, err := t.bot.Send(msg)
	return err
}

// NoticeDownloaded 实现Notifier接口，通知资源下载状态
func (t *notifier) NoticeDownloaded(ctx context.Context, req notice.NoticeDownloadedReq) error {
	if err := t.init(); err != nil {
		return err
	}

	var text, emoji string

	// 构建基本信息部分
	baseInfo := fmt.Sprintf("*番剧下载通知*\n\n"+
		"🔗 *RSS订阅项*: `%s`\n"+
		"📁 *文件名*: `%s`\n",
		req.RSSGUID, req.TorrentName)

	// 添加下载状态和详情
	if req.Failed {
		emoji = "❌"
		text = fmt.Sprintf("%s\n%s *下载失败*\n⚠️ *错误详情*: %s", baseInfo, emoji, req.FailDetail)
	} else {
		emoji = "✅"
		sizeInfo := utils.FormatFileSize(req.Size)
		costInfo := utils.FormatDuration(req.Cost)
		speedInfo := utils.CalculateAverageSpeed(req.Size, req.Cost)

		text = fmt.Sprintf("%s\n"+
			"%s *下载成功*\n"+
			"📊 *文件大小*: %s\n"+
			"⏱️ *耗时*: %s\n"+
			"🚀 *平均速度*: %s",
			baseInfo, emoji, sizeInfo, costInfo, speedInfo)
	}

	// 创建消息
	msg := tgbotapi.NewMessage(t.cfg.ChatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	// 发送消息
	_, err := t.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("发送下载通知失败: %w", err)
	}

	return nil
}

// NoticeTransferred 实现Notifier接口，通知资源转移状态
func (t *notifier) NoticeTransferred(ctx context.Context, req notice.NoticeTransferredReq) error {
	if err := t.init(); err != nil {
		return err
	}

	var text, emoji string

	// 构建基本信息部分
	baseInfo := fmt.Sprintf("番剧转移媒体库通知\n\n"+
		"📺 番剧: %s\n"+
		"🔢 季度: 第%d季\n"+
		"👥 字幕组: %s\n"+
		"📁 文件名: %s\n"+
		"🔗 RSS订阅项: %s\n",
		req.BangumiName, req.Season, req.ReleaseGroup, req.FileName, req.RSSGUID)

	// 添加转移状态和详情
	if !req.Transferred || req.Error != nil {
		emoji = "❌"
		errorMsg := "未知错误"
		if req.Error != nil {
			errorMsg = req.Error.Error()
		}
		text = fmt.Sprintf("%s\n%s 转移媒体库失败\n⚠️ 错误详情: %s", baseInfo, emoji, errorMsg)
	} else {
		emoji = "✅"
		text = fmt.Sprintf("%s\n"+
			"%s 转移媒体库成功\n"+
			"🗂️ 番剧媒体库信息: %s",
			baseInfo, emoji, req.MediaFilePath)
	}

	// 创建消息
	msg := tgbotapi.NewMessage(t.cfg.ChatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown

	// 添加海报图片
	if req.Poster != "" {
		photoMsg := tgbotapi.NewPhoto(t.cfg.ChatID, tgbotapi.FileURL(req.Poster))
		photoMsg.Caption = text

		_, err := t.bot.Send(photoMsg)
		return err
	}
	// 发送主消息
	_, err := t.bot.Send(msg)
	return err
}
