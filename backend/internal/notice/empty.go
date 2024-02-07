package notice

import (
	"context"
	"errors"
)

var _ Notifier = (*Empty)(nil)

type Empty struct{}

var ErrNofierNotSet = errors.New("消息通知渠道未设置，如已设置请检查通知配置是否正确")

// NoticeDownloaded implements Notifier.
func (e *Empty) NoticeDownloaded(ctx context.Context, req NoticeDownloadedReq) error {
	return ErrNofierNotSet
}

// NoticeSubscriptionUpdated implements Notifier.
func (e *Empty) NoticeSubscriptionUpdated(ctx context.Context, req NoticeSubscriptionUpdatedReq) error {
	return ErrNofierNotSet
}

// NoticeTransferred implements Notifier.
func (e *Empty) NoticeTransferred(ctx context.Context, req NoticeTransferredReq) error {
	return ErrNofierNotSet
}
