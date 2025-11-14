package web

import (
	"cmp"
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"

	"github.com/samber/lo"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/magnet"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
	"github.com/MangataL/BangumiBuddy/internal/transfer"
	"github.com/MangataL/BangumiBuddy/internal/types"
	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/MangataL/BangumiBuddy/pkg/utils"
)

type Web struct {
	subscriber      subscriber.Interface
	downloader      downloader.Interface
	torrentOperator downloader.TorrentOperator
	transfer        transfer.Interface
	magnet          magnet.Interface
}

type Dependency struct {
	Subscriber      subscriber.Interface
	Downloader      downloader.Interface
	TorrentOperator downloader.TorrentOperator
	Transfer        transfer.Interface
	Magnet          magnet.Interface
}

func New(dep Dependency) Interface {
	return &Web{
		subscriber:      dep.Subscriber,
		downloader:      dep.Downloader,
		torrentOperator: dep.TorrentOperator,
		transfer:        dep.Transfer,
		magnet:          dep.Magnet,
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
		if bangumi.CreatedAt.After(bangumiBase.CreatedAt) {
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
	slices.SortStableFunc(bangumiBases, func(b1, b2 BangumiBase) int {
		b1Active := getActive(b1.ReleaseGroups)
		b2Active := getActive(b2.ReleaseGroups)
		if b1Active && !b2Active {
			return -1
		}
		if !b1Active && b2Active {
			return 1
		}
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

func getActive(releaseGroups []ReleaseGroupSubsription) bool {
	for _, rs := range releaseGroups {
		if rs.Active {
			return true
		}
	}
	return false
}

func (w *Web) GetBangumiTorrents(ctx context.Context, subscriptionID string) ([]Torrent, error) {
	ts, _, err := w.torrentOperator.List(ctx, downloader.TorrentFilter{
		SubscriptionID: subscriptionID,
		Order: types.Order{
			Field: "name",
			Way:   "desc",
		},
	})
	if err != nil {
		return nil, err
	}
	bangumi, err := w.getBangumi(ctx, subscriptionID)
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

		torrent.Collection = w.isCollection(ctx, t.Name)
		if !torrent.Collection {
			episode, _ := w.transfer.ParseEpisode(ctx, t.Name, bangumi.EpisodeLocation)
			torrent.Episode = episode
			torrent.Season = bangumi.Season
		}

		torrents = append(torrents, torrent)
	}
	return torrents, nil
}

var collectionRegex = regexp.MustCompile(`\d+-\d+`)

func (w *Web) isCollection(ctx context.Context, torrentName string) bool {
	if utils.IsMediaFile(torrentName) {
		return false
	}
	return collectionRegex.MatchString(torrentName)
}

func (w *Web) DeleteTorrent(ctx context.Context, req DeleteTorrentReq) error {
	torrent, err := w.torrentOperator.Get(ctx, req.Hash)
	if err != nil {
		return err
	}
	for _, name := range torrent.FileNames {
		filePath := filepath.Join(torrent.Path, name)
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
				filePath := filepath.Join(torrent.Path, name)
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
	} else {
		if err := w.torrentOperator.DeleteBySubscriptionID(ctx, subscriptionID); err != nil {
			return err
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

func (w *Web) GetBangumiTorrentFiles(ctx context.Context, hash string) ([]File, error) {
	torrent, err := w.torrentOperator.Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	bangumi, err := w.getBangumi(ctx, torrent.SubscriptionID)
	if err != nil {
		return nil, err
	}
	tfs := make([]File, 0, len(torrent.FileNames))
	for _, fileName := range torrent.FileNames {
		filePath := filepath.Join(torrent.Path, fileName)
		if utils.IsMediaFile(filePath) {
			file := File{FileName: filePath}
			linkFile, err := w.transfer.GetTransferFile(ctx, filePath)
			episode, _ := w.transfer.ParseEpisode(ctx, fileName, bangumi.EpisodeLocation)
			file.Episode = episode
			file.Season = bangumi.Season
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

func (w *Web) getBangumi(ctx context.Context, subscriptionID string) (subscriber.Bangumi, error) {
	if subscriptionID != "" {
		subscription, err := w.subscriber.Get(ctx, subscriptionID)
		if err != nil {
			return subscriber.Bangumi{}, err
		}
		return subscription, nil
	}
	return subscriber.Bangumi{}, subscriber.ErrSubscriberNotFound
}

func (w *Web) ListRecentUpdatedTorrents(ctx context.Context, req ListRecentUpdatedTorrentsReq) (ListRecentUpdatedTorrentsResp, error) {
	torrents, total, err := w.torrentOperator.List(ctx, downloader.TorrentFilter{
		Page: types.Page{
			Num:  req.Page,
			Size: req.PageSize,
		},
		Order: types.Order{
			Field: "created_at",
			Way:   "desc",
		},
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		MagnetTask: lo.ToPtr(false),
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

// ListMagnetTasks implements Interface.
func (w *Web) ListMagnetTasks(ctx context.Context, req ListMagnetTasksReq) (ListMagnetTasksResp, error) {
	tasks, total, err := w.magnet.ListTasks(ctx, magnet.ListTasksReq{
		Page: types.Page{
			Num:  req.Page,
			Size: req.PageSize,
		},
		Order: types.Order{
			Field: "id",
			Way:   "desc",
		},
	})
	if err != nil {
		return ListMagnetTasksResp{}, err
	}

	downloadTasks := make([]MagnetTask, 0, len(tasks))
	for _, task := range tasks {
		dt, err := w.getMagnetTask(ctx, task)
		if err != nil {
			return ListMagnetTasksResp{}, err
		}
		downloadTasks = append(downloadTasks, dt)
	}
	return ListMagnetTasksResp{
		Total: total,
		Tasks: downloadTasks,
	}, nil
}

func (w *Web) getMagnetTask(ctx context.Context, task magnet.Task) (MagnetTask, error) {
	dt := MagnetTask{
		Task: task,
	}

	// 如果任务状态不是 init success，说明还在等待解析或确认，不需要查询下载状态
	if task.Status != magnet.TaskStatusInitSuccess {
		return dt, nil
	}

	// 查询种子状态
	torrent, err := w.torrentOperator.Get(ctx, task.Torrent.Hash)
	if err != nil {
		log.Warnf(ctx, "查询种子状态失败: %v, taskID: %s, hash: %s", err, task.TaskID, task.Torrent.Hash)
		return dt, nil
	}

	dt.DownloadStatus = torrent.Status
	dt.DownloadStatusDetail = torrent.StatusDetail

	// 如果种子还在下载中，查询下载进度和大小
	if torrent.Status != downloader.TorrentStatusTransferred && torrent.Status != downloader.TorrentStatusTransferredError {
		ds, err := w.downloader.GetDownloadStatuses(ctx, []string{torrent.Hash})
		if err != nil {
			return MagnetTask{}, err
		}
		if len(ds) == 0 {
			return dt, nil
		}
		dt.DownloadSpeed = ds[0].DownloadSpeed
		dt.Progress = ds[0].Progress
		dt.DownloadStatus = ds[0].Status
		dt.DownloadStatusDetail = ds[0].Error
	}
	return dt, nil
}

func (w *Web) GetMagnetTask(ctx context.Context, taskID string) (MagnetTask, error) {
	task, err := w.magnet.GetTask(ctx, taskID)
	if err != nil {
		return MagnetTask{}, err
	}
	mt, err := w.getMagnetTask(ctx, task)
	if err != nil {
		return MagnetTask{}, err
	}
	return w.fillLinkFile(ctx, mt)
}

func (w *Web) fillLinkFile(ctx context.Context, mt MagnetTask) (MagnetTask, error) {
	torrent, err := w.torrentOperator.Get(ctx, mt.Torrent.Hash)
	if err != nil {
		return MagnetTask{}, err
	}
	for i, file := range mt.Torrent.Files {
		if file.Media && file.Download {
			linkFile, err := w.transfer.GetTransferFile(ctx, filepath.Join(torrent.Path, file.FileName))
			if err != nil {
				if errors.Is(errors.Unwrap(err), transfer.ErrFileTransferredNotFound) {
					continue
				}
				return MagnetTask{}, err
			}
			mt.Torrent.Files[i].LinkFile = linkFile
		}
	}
	return mt, nil
}

func (w *Web) DeleteMagnetTask(ctx context.Context, taskID string, deleteFiles bool) error {
	if deleteFiles {
		files, _, err := w.magnet.ListTasks(ctx, magnet.ListTasksReq{
			TaskIDs: []string{taskID},
		})
		if err != nil {
			return err
		}
		for _, file := range files {
			if err := w.DeleteTorrent(ctx, DeleteTorrentReq{
				Hash:              file.Torrent.Hash,
				DeleteOriginFiles: true,
			}); err != nil {
				return err
			}
		}
	} else {
		if err := w.torrentOperator.DeleteByTaskID(ctx, taskID); err != nil {
			return err
		}
	}
	return w.magnet.DeleteTask(ctx, taskID)
}

func (w *Web) ListDirs(ctx context.Context, path string) (ListDirsResp, error) {
	if path == "" {
		return ListDirsResp{}, errors.New("请传入有效的路径")
	}

	path = filepath.Clean(path)
	entries, err := os.ReadDir(path)
	if err != nil {
		return ListDirsResp{}, err
	}

	paths := make([]FileDir, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			dirPath := filepath.Join(path, entry.Name())
			hasDir, subtitleCount := w.dirStats(dirPath)
			paths = append(paths, FileDir{
				Path:          dirPath,
				HasDir:        hasDir,
				SubtitleCount: subtitleCount,
			})
		}
	}

	return ListDirsResp{
		Dirs:          paths,
		FilePathSplit: string(filepath.Separator),
		FileRoots:     w.getFileRoots(),
	}, nil
}

func (w *Web) getFileRoots() []string {
	if runtime.GOOS == "windows" {
		return w.getWindowsDrives()
	}
	return []string{"/"}
}

func (w *Web) getWindowsDrives() []string {
	drives := make([]string, 0, 26)
	for i := 'A'; i <= 'Z'; i++ {
		drive := string(i) + ":\\"
		if _, err := os.Stat(drive); err == nil {
			drives = append(drives, drive)
		}
	}
	return drives
}

// dirStats 返回目录下是否包含子目录以及字幕文件数量
func (w *Web) dirStats(dirPath string) (bool, int) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false, 0
	}

	hasDir := false
	subtitleCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			hasDir = true
		}
		if utils.IsSubtitleFile(entry.Name()) {
			subtitleCount++
		}
	}

	return hasDir, subtitleCount
}