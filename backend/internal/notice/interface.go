package notice

import "context"

type Notifier interface {
	NoticeSubscriptionUpdated(ctx context.Context, req NoticeSubscriptionUpdatedReq) error
	NoticeDownloaded(ctx context.Context, req NoticeDownloadedReq) error
	NoticeSubscriptionTransferred(ctx context.Context, req NoticeSubscriptionTransferredReq) error
	NoticeTaskTransferred(ctx context.Context, req NoticeTaskTransferredReq) error
}
