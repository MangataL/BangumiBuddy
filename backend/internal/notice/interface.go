package notice

import "context"

type Notifier interface {
	NoticeSubscriptionUpdated(ctx context.Context, req NoticeSubscriptionUpdatedReq) error
	NoticeDownloaded(ctx context.Context, req NoticeDownloadedReq) error
	NoticeTransferred(ctx context.Context, req NoticeTransferredReq) error
}
