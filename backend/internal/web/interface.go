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
	ListMagnetTasks(ctx context.Context, req ListMagnetTasksReq) (ListMagnetTasksResp, error)
	GetMagnetTask(ctx context.Context, taskID string) (MagnetTask, error)
	DeleteMagnetTask(ctx context.Context, taskID string, deleteFiles bool) error
	ListDirs(ctx context.Context, path string) (ListDirsResp, error)
}
