package qbittorrent

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/autobrr/go-qbittorrent"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

// 确保QBittorrent实现了Downloader接口
var _ downloader.Downloader = (*QBittorrent)(nil)

const (
	// 标签名称
	tag = "BangumiBuddy"
)

// Config qBittorrent配置接口
type Config struct {
	Host     string `mapstructure:"host" json:"host"`
	Username string `mapstructure:"username" json:"username"`
	Password string `mapstructure:"password" json:"password"`
}

// QBittorrent 实现Downloader接口
type QBittorrent struct {
	client *qbittorrent.Client
	login  bool
}

// NewQBittorrent 创建一个新的QBittorrent实例
func NewQBittorrent(config Config) *QBittorrent {
	// 创建qBittorrent客户端
	client := qbittorrent.NewClient(qbittorrent.Config{
		Host:     config.Host,
		Username: config.Username,
		Password: config.Password,
	})
	return &QBittorrent{
		client: client,
	}
}

func (q *QBittorrent) init() error {
	if q.login {
		return nil
	}
	if err := q.client.Login(); err != nil {
		return fmt.Errorf("qbittorrent 登录失败，请检查配置的账号密码是否正确: %w", err)
	}
	tags, err := q.client.GetTags()
	if err != nil {
		return fmt.Errorf("获取标签失败: %w", err)
	}
	if !slices.Contains(tags, tag) {
		if err := q.client.CreateTags([]string{tag}); err != nil {
			return fmt.Errorf("创建标签失败: %w", err)
		}
	}
	return nil
}

// GetTorrentContents implements downloader.Downloader.
func (q *QBittorrent) GetTorrentFileNames(ctx context.Context, hash string) ([]string, error) {
	if err := q.init(); err != nil {
		return nil, err
	}
	files, err := q.client.GetFilesInformationCtx(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("获取种子文件列表失败: %w", err)
	}

	fileNames := make([]string, 0, len(*files))
	for _, file := range *files {
		fileNames = append(fileNames, file.Name)
	}

	return fileNames, nil
}

// GetDownloadStatuses 实现Downloader接口的GetDownloadStatuses方法
func (q *QBittorrent) GetDownloadStatuses(ctx context.Context, hashes []string) ([]downloader.DownloadStatus, error) {
	if err := q.init(); err != nil {
		return nil, err
	}
	// 获取种子列表
	options := qbittorrent.TorrentFilterOptions{}
	if len(hashes) > 0 {
		options.Hashes = hashes
	}

	torrents, err := q.client.GetTorrentsCtx(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("获取种子列表失败: %w", err)
	}

	return q.convertTorrentsToDownloadStatuses(ctx, torrents), nil
}

// convertTorrentsToDownloadStatuses 将qbittorrent的种子信息转换为DownloadStatus
func (q *QBittorrent) convertTorrentsToDownloadStatuses(ctx context.Context, torrents []qbittorrent.Torrent) []downloader.DownloadStatus {
	// 转换为DownloadStatus
	statuses := make([]downloader.DownloadStatus, 0, len(torrents))
	for _, t := range torrents {
		status := downloader.DownloadStatus{
			Hash:          t.Hash,
			Name:          t.Name,
			Progress:      t.Progress,
			DownloadSpeed: t.DlSpeed,
			Size:          t.TotalSize,
		}

		switch {
		case t.State == qbittorrent.TorrentStateError || t.State == qbittorrent.TorrentStateMissingFiles:
			// 如果状态是error，获取错误信息
			status.Status = downloader.TorrentStatusDownloadError
			props, err := q.client.GetTorrentPropertiesCtx(ctx, t.Hash)
			if err == nil {
				status.Error = props.Comment
			}
		case t.State == qbittorrent.TorrentStateStoppedDl:
			status.Status = downloader.TorrentStatusDownloadPaused
		case t.AmountLeft == 0 && t.State != qbittorrent.TorrentStateMoving:
			// 如果amount_left为0并且种子未处于移动中，则表示下载已完成
			status.Status = downloader.TorrentStatusDownloaded
			status.Cost = time.Duration(t.CompletionOn-t.AddedOn) * time.Second
		default:
			// 其他情况为下载中
			status.Status = downloader.TorrentStatusDownloading
		}
		statuses = append(statuses, status)
	}

	return statuses
}

func (q *QBittorrent) AddTorrent(ctx context.Context, torrentLink, savePath, stopCondition string) error {
	if err := q.init(); err != nil {
		return err
	}
	options := &qbittorrent.TorrentAddOptions{
		SavePath: savePath,
		Tags:     tag,
	}
	opts := options.Prepare()

	if stopCondition != "" {
		opts["stopCondition"] = stopCondition
	}
	return q.client.AddTorrentFromUrlCtx(ctx, torrentLink, opts)
}

func (q *QBittorrent) SetLocation(ctx context.Context, hash, savePath string) error {
	if err := q.init(); err != nil {
		return err
	}
	if err := q.client.AddTagsCtx(ctx, []string{hash}, tag); err != nil {
		return fmt.Errorf("添加标签失败: %w", err)
	}
	return q.client.SetLocationCtx(ctx, []string{hash}, savePath)
}

func (q *QBittorrent) GetTorrentName(ctx context.Context, hash string) (string, error) {
	if err := q.init(); err != nil {
		return "", err
	}
	var name string
	if err := wait.PollUntilContextTimeout(ctx, 500*time.Millisecond, 10*time.Second, false, func(ctx context.Context) (bool, error) {
		props, err := q.client.GetTorrentPropertiesCtx(ctx, hash)
		if err != nil {
			log.Errorf(ctx, "获取种子属性信息失败: %s", err)
			return false, nil
		}
		if props.Name != hash {
			name = props.Name
			return true, nil
		}
		return false, nil
	}); err != nil {
		return "", err
	}
	return name, nil
}

func (q *QBittorrent) ListTorrentsStatus(ctx context.Context) ([]downloader.DownloadStatus, error) {
	if err := q.init(); err != nil {
		return nil, err
	}
	// 获取所有带有BangumiBuddy标签的种子
	options := qbittorrent.TorrentFilterOptions{
		Filter: qbittorrent.TorrentFilterAll,
		Tag:    tag,
	}

	// 获取所有带有BangumiBuddy标签的种子
	torrents, err := q.client.GetTorrentsCtx(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("获取种子列表失败: %w", err)
	}

	return q.convertTorrentsToDownloadStatuses(ctx, torrents), nil
}

// DeleteTorrent 删除种子文件
func (q *QBittorrent) DeleteTorrent(ctx context.Context, hash string) error {
	if err := q.init(); err != nil {
		return err
	}
	if err := q.client.DeleteTorrentsCtx(ctx, []string{hash}, true); err != nil {
		return fmt.Errorf("删除种子文件失败: %w", err)
	}
	return nil
}

func (q *QBittorrent) ContinueDownload(ctx context.Context, hash string) error {
	if err := q.init(); err != nil {
		return err
	}
	return q.client.ResumeCtx(ctx, []string{hash})
}
