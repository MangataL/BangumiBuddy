package adapter

import (
	"context"
	"errors"

	"github.com/MangataL/BangumiBuddy/internal/notice"
	"github.com/MangataL/BangumiBuddy/internal/notice/bark"
	"github.com/MangataL/BangumiBuddy/internal/notice/email"
	"github.com/MangataL/BangumiBuddy/internal/notice/telegram"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

func NewAdapter(config Config) *Adapter {
	adapter := &Adapter{}
	if err := adapter.Reload(&config); err != nil {
		log.Errorf(context.Background(), "初始化消息通知器失败: %v", err)
		adapter.notifier = &notice.Empty{}
	}
	return adapter
}

type Adapter struct {
	notifier notice.Notifier
	config   Config
}

type Config struct {
	Enabled      bool            `mapstructure:"enabled" json:"enabled"`
	Type         string          `mapstructure:"type" json:"type"`
	Telegram     telegram.Config `mapstructure:"telegram" json:"telegram"`
	Email        email.Config    `mapstructure:"email" json:"email"`
	Bark         bark.Config     `mapstructure:"bark" json:"bark"`
	NoticePoints NoticePoints    `mapstructure:"notice_points" json:"noticePoints"`
}

// NoticePoints 消息通知点
type NoticePoints struct {
	SubscriptionUpdated *bool `mapstructure:"subscription_updated" json:"subscriptionUpdated"`
	Downloaded          *bool `mapstructure:"downloaded" json:"downloaded"`
	Transferred         *bool `mapstructure:"transferred" json:"transferred" default:"true"`
	Error               *bool `mapstructure:"error" json:"error" default:"true"`
}

func (a *Adapter) Reload(config interface{}) error {
	cfg, ok := config.(*Config)
	if !ok {
		return errors.New("配置类型错误")
	}
	a.config = *cfg
	switch cfg.Type {
	case "telegram":
		a.notifier = telegram.NewTelegramNotifier(cfg.Telegram)
	case "email":
		a.notifier = email.NewEmailNotifier(cfg.Email)
	case "bark":
		a.notifier = bark.NewBarkNotifier(cfg.Bark)
	default:
		a.notifier = &notice.Empty{}
	}
	return nil
}

// NoticeDownloaded implements notice.Notifier.
func (a *Adapter) NoticeDownloaded(ctx context.Context, req notice.NoticeDownloadedReq) error {
	if !a.config.Enabled {
		return nil
	}
	if req.Failed && (a.config.NoticePoints.Error == nil || !*a.config.NoticePoints.Error) {
		return nil
	}
	if !req.Failed && (a.config.NoticePoints.Downloaded == nil || !*a.config.NoticePoints.Downloaded) {
		return nil
	}
	return a.notifier.NoticeDownloaded(ctx, req)
}

// NoticeSubscriptionUpdated implements notice.Notifier.
func (a *Adapter) NoticeSubscriptionUpdated(ctx context.Context, req notice.NoticeSubscriptionUpdatedReq) error {
	if !a.config.Enabled {
		return nil
	}
	if req.Error != nil && (a.config.NoticePoints.Error == nil || !*a.config.NoticePoints.Error) {
		return nil
	}
	if req.Error == nil && (a.config.NoticePoints.SubscriptionUpdated == nil || !*a.config.NoticePoints.SubscriptionUpdated) {
		return nil
	}
	return a.notifier.NoticeSubscriptionUpdated(ctx, req)
}

// NoticeTransferred implements notice.Notifier.
func (a *Adapter) NoticeTransferred(ctx context.Context, req notice.NoticeTransferredReq) error {
	if !a.config.Enabled {
		return nil
	}
	if req.Error != nil && (a.config.NoticePoints.Error == nil || !*a.config.NoticePoints.Error) {
		return nil
	}
	if req.Error == nil && (a.config.NoticePoints.Transferred == nil || !*a.config.NoticePoints.Transferred) {
		return nil
	}
	return a.notifier.NoticeTransferred(ctx, req)
}
