package web

import (
	"context"

	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

type Interface interface {
	ListBangumis(ctx context.Context, req subscriber.ListBangumiReq) ([]BangumiBase, error)
	GetBangumiTorrents(ctx context.Context, subscriptionID string) ([]Torrent, error)
	DeleteTorrent(ctx context.Context, req DeleteTorrentReq) error
	DeleteSubscription(ctx context.Context, subscriptionID string, deleteFiles bool) error
	GetTorrentFiles(ctx context.Context, hash string) ([]File, error)
	ListRecentUpdatedTorrents(ctx context.Context, req ListRecentUpdatedTorrentsReq) (ListRecentUpdatedTorrentsResp, error)
}