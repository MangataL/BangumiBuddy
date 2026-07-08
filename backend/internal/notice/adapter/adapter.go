package adapter

import (
	"context"
	"errors"
	"sync"

	"github.com/MangataL/BangumiBuddy/internal/network"
	"github.com/MangataL/BangumiBuddy/internal/notice"
	"github.com/MangataL/BangumiBuddy/internal/notice/bark"
	"github.com/MangataL/BangumiBuddy/internal/notice/email"
	"github.com/MangataL/BangumiBuddy/internal/notice/telegram"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

func NewAdapter(config Config, provider network.HTTPClientProvider) *Adapter {
	adapter := &Adapter{
		network: provider,
	}
	if err := adapter.Reload(&config); err != nil {
		log.Errorf(context.Background(), "初始化消息通知器失败: %v", err)
		adapter.notifier = &notice.Empty{}
	}
	return adapter
}

type Adapter struct {
	mu       sync.RWMutex
	notifier notice.Notifier
	config   Config
	network  network.HTTPClientProvider
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
	notifier := notice.Notifier(&notice.Empty{})
	switch cfg.Type {
	case "telegram":
		notifier = telegram.NewTelegramNotifier(cfg.Telegram, a.network)
	case "email":
		notifier = email.NewEmailNotifier(cfg.Email)
	case "bark":
		notifier = bark.NewBarkNotifier(cfg.Bark)
	}
	a.mu.Lock()
	a.config = *cfg
	a.notifier = notifier
	a.mu.Unlock()
	return nil
}

// NoticeDownloaded implements notice.Notifier.
func (a *Adapter) NoticeDownloaded(ctx context.Context, req notice.NoticeDownloadedReq) error {
	config, notifier := a.snapshot()
	if !config.Enabled {
		return nil
	}
	if req.Failed && (config.NoticePoints.Error == nil || !*config.NoticePoints.Error) {
		return nil
	}
	if !req.Failed && (config.NoticePoints.Downloaded == nil || !*config.NoticePoints.Downloaded) {
		return nil
	}
	return notifier.NoticeDownloaded(ctx, req)
}

// NoticeSubscriptionUpdated implements notice.Notifier.
func (a *Adapter) NoticeSubscriptionUpdated(ctx context.Context, req notice.NoticeSubscriptionUpdatedReq) error {
	config, notifier := a.snapshot()
	if !config.Enabled {
		return nil
	}
	if req.Error != nil && (config.NoticePoints.Error == nil || !*config.NoticePoints.Error) {
		return nil
	}
	if req.Error == nil && (config.NoticePoints.SubscriptionUpdated == nil || !*config.NoticePoints.SubscriptionUpdated) {
		return nil
	}
	return notifier.NoticeSubscriptionUpdated(ctx, req)
}

// NoticeSubscriptionTransferred implements notice.Notifier.
func (a *Adapter) NoticeSubscriptionTransferred(ctx context.Context, req notice.NoticeSubscriptionTransferredReq) error {
	config, notifier := a.snapshot()
	if !config.Enabled {
		return nil
	}
	if req.Error != nil && (config.NoticePoints.Error == nil || !*config.NoticePoints.Error) {
		return nil
	}
	if req.Error == nil && (config.NoticePoints.Transferred == nil || !*config.NoticePoints.Transferred) {
		return nil
	}
	return notifier.NoticeSubscriptionTransferred(ctx, req)
}

// NoticeTaskTransferred implements notice.Notifier.
func (a *Adapter) NoticeTaskTransferred(ctx context.Context, req notice.NoticeTaskTransferredReq) error {
	config, notifier := a.snapshot()
	if !config.Enabled {
		return nil
	}
	if req.Error != nil && (config.NoticePoints.Error == nil || !*config.NoticePoints.Error) {
		return nil
	}
	if req.Error == nil && (config.NoticePoints.Transferred == nil || !*config.NoticePoints.Transferred) {
		return nil
	}
	return notifier.NoticeTaskTransferred(ctx, req)
}

func (a *Adapter) snapshot() (Config, notice.Notifier) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.notifier == nil {
		return a.config, &notice.Empty{}
	}
	return a.config, a.notifier
}
