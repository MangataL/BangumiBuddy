package subscriber

import (
	"context"
	"time"

	"github.com/MangataL/BangumiBuddy/pkg/errs"
)

//go:generate mockgen -destination interface_mock.go -source $GOFILE -package $GOPACKAGE

var ErrSubscriberNotFound = errs.NewNotFound("番剧未找到")

// Interface 服务层
type Interface interface {
	// ParseRSS 解析RSS链接，获取番剧的基本信息，供用户确认
	ParseRSS(ctx context.Context, rssLink string) (ParseRSSRsp, error)
	// Subscribe 订阅番剧
	Subscribe(ctx context.Context, req SubscribeReq) (Bangumi, error)
	// Get 获取番剧信息
	Get(ctx context.Context, id string) (Bangumi, error)
	// List 列出所有订阅
	List(ctx context.Context, req ListBangumiReq) ([]Bangumi, error)
	// UpdateSubscription 更新订阅
	UpdateSubscription(ctx context.Context, req UpdateSubscribeReq) error
	// DeleteSubscription 删除订阅
	DeleteSubscription(ctx context.Context, id string) error
	// UpdateLastAirEpisode 更新番剧的最新集数
	UpdateLastAirEpisode(ctx context.Context, subscriptionID string, episode int) error
	// GetRSSMatch 获取RSS匹配
	GetRSSMatch(ctx context.Context, subscriptionID string) ([]RSSMatch, error)
	// HandleBangumiSubscription 运行订阅下载任务
	HandleBangumiSubscription(ctx context.Context, subscriptionID string) error
	// MarkRSSRecord 标记RSS记录
	MarkRSSRecord(ctx context.Context, req MarkRSSRecordReq) error
	// GetSubscriptionCalendar 获取订阅日历
	GetSubscriptionCalendar(ctx context.Context) (map[time.Weekday][]CalendarItem, error)
	// StopSubscription 停止订阅
	StopSubscription(ctx context.Context, id string) error
	// AutoStopSubscription 是否自动停止订阅
	AutoStopSubscription(ctx context.Context, id string) bool
}
