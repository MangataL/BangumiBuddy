package web

import (
	"cmp"
	"context"
	"errors"
	"os"
	"slices"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
	"github.com/MangataL/BangumiBuddy/internal/transfer"
	"github.com/MangataL/BangumiBuddy/internal/utils"
	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/samber/lo"
)

type Web struct {
	subscriber      subscriber.Interface
	downloader      downloader.Interface
	torrentOperator downloader.TorrentOperator
	transfer        transfer.Interface
}

type Dependency struct {
	Subscriber      subscriber.Interface
	Downloader      downloader.Interface
	TorrentOperator downloader.TorrentOperator
	Transfer        transfer.Interface
}

func New(dep Dependency) Interface {
	return &Web{
		subscriber:      dep.Subscriber,
		downloader:      dep.Downloader,
		torrentOperator: dep.TorrentOperator,
		transfer:        dep.Transfer,
	}
}

func (w *Web) ListBangumis(ctx context.Context, req subscriber.ListBangumiReq) ([]BangumiBase, error) {
	bangumis, err := w.subscriber.List(ctx, req)
	if err != nil {
		return nil, err
	}

	bangumiMap := make(map[string]BangumiBase)
	for _, bangumi := range bangumis {
		bangumiBase, ok := bangumiMap[bangumi.Name]
		if !ok {
			bangumiBase = BangumiBase{
				BangumiName:   bangumi.Name,
				Season:        bangumi.Season,
				PosterURL:     bangumi.PosterURL,
				BackdropURL:   bangumi.BackdropURL,
				Overview:      bangumi.Overview,
				Genres:        bangumi.Genres,
				AirWeekday:    bangumi.AirWeekday,
				ReleaseGroups: make([]ReleaseGroupSubsription, 0),
				CreatedAt:     bangumi.CreatedAt,
			}
		}
		if bangumi.CreatedAt.Before(bangumiBase.CreatedAt) {
			bangumiBase.CreatedAt = bangumi.CreatedAt
		}
		bangumiBase.ReleaseGroups = append(bangumiBase.ReleaseGroups, ReleaseGroupSubsription{
			ReleaseGroup:    bangumi.ReleaseGroup,
			EpisodeTotalNum: bangumi.EpisodeTotalNum,
			LastAirEpisode:  bangumi.LastAirEpisode,
			SubscriptionID:  bangumi.SubscriptionID,
			Priority:        bangumi.Priority,
			Active:          bangumi.Active,
		})
		bangumiMap[bangumi.Name] = bangumiBase
	}

	bangumiBases := make([]BangumiBase, 0, len(bangumiMap))
	for _, bangumi := range bangumiMap {
		bangumiBases = append(bangumiBases, bangumi)
	}
	slices.SortFunc(bangumiBases, func(b1, b2 BangumiBase) int {
		return cmp.Compare(b2.CreatedAt.UnixNano(), b1.CreatedAt.UnixNano())
	})
	for _, bangumi := range bangumiBases {
		slices.SortStableFunc(bangumi.ReleaseGroups, func(rs1, rs2 ReleaseGroupSubsription) int {
			if rs1.Active && !rs2.Active {
				return -1
			}
			if !rs1.Active && rs2.Active {
				return 1
			}
			if rs1.Priority == rs2.Priority {
				return cmp.Compare(rs1.ReleaseGroup, rs2.ReleaseGroup)
			}
			return cmp.Compare(rs2.Priority, rs1.Priority)
		})
	}
	return bangumiBases, nil
}

func (w *Web) GetBangumiTorrents(ctx context.Context, subscriptionID string) ([]Torrent, error) {
	ts, _, err := w.torrentOperator.List(ctx, downloader.TorrentFilter{
		SubscriptionID: subscriptionID,
		Order: downloader.Order{
			Field: "name",
			Way:   "desc",
		},
	})
	if err != nil {
		return nil, err
	}

	torrents := make([]Torrent, 0, len(ts))
	for _, t := range ts {
		torrent := Torrent{
			Name:         t.Name,
			Hash:         t.Hash,
			Status:       t.Status,
			StatusDetail: t.StatusDetail,
			RSSGUID:      t.RSSGUID,
			CreatedAt:    t.CreatedAt,
		}
		if t.Status != downloader.TorrentStatusTransferred && t.Status != downloader.TorrentStatusTransferredError {
			ds, err := w.downloader.GetDownloadStatuses(ctx, []string{t.Hash})
			if err != nil {
				return nil, err
			}
			if len(ds) == 0 {
				continue
			}
			torrent.DownloadSpeed = ds[0].DownloadSpeed
			torrent.Progress = ds[0].Progress
			torrent.Status = ds[0].Status
			torrent.StatusDetail = ds[0].Error
		}
		torrents = append(torrents, torrent)
	}
	return torrents, nil
}

func (w *Web) DeleteTorrent(ctx context.Context, req DeleteTorrentReq) error {
	torrent, err := w.torrentOperator.Get(ctx, req.Hash)
	if err != nil {
		return err
	}
	for _, name := range torrent.FileNames {
		filePath := torrent.Path + name
		transferFile, err := w.transfer.GetTransferFile(ctx, filePath)
		if err != nil {
			if errors.Is(errors.Unwrap(err), transfer.ErrFileTransferredNotFound) {
				continue
			}
			return err
		}
		if err := w.transfer.DeleteTransferFile(ctx, transferFile); err != nil {
			return err
		}
	}
	if req.DeleteOriginFiles {
		if err := w.downloader.DeleteTorrent(ctx, req.Hash); err != nil {
			log.Warnf(ctx, "通过下载器删除种子失败，尝试手动清理: %v", err)
			for _, name := range torrent.FileNames {
				filePath := torrent.Path + name
				if err := os.Remove(filePath); err != nil {
					return err
				}
			}
			return w.torrentOperator.Delete(ctx, req.Hash)
		}
	}
	return nil
}

func (w *Web) DeleteSubscription(ctx context.Context, subscriptionID string, deleteFiles bool) error {
	if deleteFiles {
		files, _, err := w.torrentOperator.List(ctx, downloader.TorrentFilter{
			SubscriptionID: subscriptionID,
		})
		if err != nil {
			return err
		}
		for _, file := range files {
			if err := w.DeleteTorrent(ctx, DeleteTorrentReq{
				Hash:              file.Hash,
				DeleteOriginFiles: true,
			}); err != nil {
				return err
			}
		}
	}
	if err := w.transfer.DeleteTransferCache(ctx, transfer.DeleteFileTransferredReq{
		SubscriptionID: subscriptionID,
	}); err != nil {
		return err
	}
	if err := w.subscriber.DeleteSubscription(ctx, subscriptionID); err != nil {
		return err
	}
	return nil
}

func (w *Web) GetTorrentFiles(ctx context.Context, hash string) ([]File, error) {
	torrent, err := w.torrentOperator.Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	tfs := make([]File, 0, len(torrent.FileNames))
	for _, fileName := range torrent.FileNames {
		filePath := torrent.Path + fileName
		if utils.IsMediaFile(filePath) {
			file := File{FileName: filePath}
			linkFile, err := w.transfer.GetTransferFile(ctx, filePath)
			if err != nil {
				if errors.Is(errors.Unwrap(err), transfer.ErrFileTransferredNotFound) {
					tfs = append(tfs, file)
					continue
				}
				return nil, err
			}
			file.LinkName = linkFile
			tfs = append(tfs, file)
		}
	}
	return tfs, nil
}

func (w *Web) ListRecentUpdatedTorrents(ctx context.Context, req ListRecentUpdatedTorrentsReq) (ListRecentUpdatedTorrentsResp, error) {
	torrents, total, err := w.torrentOperator.List(ctx, downloader.TorrentFilter{
		Page: downloader.Page{
			Num:  req.Page,
			Size: req.PageSize,
		},
		Order: downloader.Order{
			Field: "created_at",
			Way:   "desc",
		},
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	})
	if err != nil {
		return ListRecentUpdatedTorrentsResp{}, err
	}
	subscriptionIDSet := make(map[string]struct{})
	for _, t := range torrents {
		subscriptionIDSet[t.SubscriptionID] = struct{}{}
	}
	subscriptions, err := w.subscriber.List(ctx, subscriber.ListBangumiReq{
		SubscriptionIDs: lo.Keys(subscriptionIDSet),
	})
	if err != nil {
		return ListRecentUpdatedTorrentsResp{}, err
	}
	subscriptionMap := make(map[string]subscriber.Bangumi)
	for _, s := range subscriptions {
		subscriptionMap[s.SubscriptionID] = s
	}
	torrentsResp := make([]RecentUpdatedTorrent, 0, len(torrents))
	for _, t := range torrents {
		torrentsResp = append(torrentsResp, RecentUpdatedTorrent{
			PosterURL:    subscriptionMap[t.SubscriptionID].PosterURL,
			BangumiName:  subscriptionMap[t.SubscriptionID].Name,
			Season:       subscriptionMap[t.SubscriptionID].Season,
			CreatedAt:    t.CreatedAt,
			RSSItem:      t.RSSGUID,
			Status:       t.Status,
			StatusDetail: t.StatusDetail,
		})
	}
	return ListRecentUpdatedTorrentsResp{
		Total:    total,
		Torrents: torrentsResp,
	}, nil
}
