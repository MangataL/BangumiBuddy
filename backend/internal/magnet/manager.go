package magnet

import (
	"context"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/meta"
	"github.com/MangataL/BangumiBuddy/internal/types"
	"github.com/MangataL/BangumiBuddy/pkg/bangumifile"
	"github.com/MangataL/BangumiBuddy/pkg/errs"
	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/MangataL/BangumiBuddy/pkg/utils"
)

//go:generate mockgen -destination manager_mock.go -source $GOFILE -package $GOPACKAGE

var (
	ErrMagnetTaskNotFound = errors.New("磁力任务不存在")
)

// Repository 下载任务数据访问层接口
type Repository interface {
	SaveTask(ctx context.Context, task Task) error
	GetTask(ctx context.Context, taskID string) (Task, error)
	GetTaskByHash(ctx context.Context, hash string) (Task, error)
	ListTasks(ctx context.Context, req ListTasksReq) ([]Task, int, error)
	DeleteTask(ctx context.Context, taskID string) error
}

// TaskFilter 任务过滤条件
type TaskFilter struct {
	TaskIDs     []string
	TorrentName string
	Page        types.Page
	Order       types.Order
}

type Dependency struct {
	Downloader        downloader.Interface
	TorrentOp         downloader.TorrentOperator
	MetaParser        meta.Parser
	BangumiFileParser bangumifile.Parser
	Repository        Repository
}

func New(dep Dependency) Interface {
	return &Manager{
		downloader: dep.Downloader,
		torrentOp:  dep.TorrentOp,
		metaParser: dep.MetaParser,
		bfParser:   dep.BangumiFileParser,
		repository: dep.Repository,
	}
}

// Manager 种子下载管理器
type Manager struct {
	downloader downloader.Interface
	torrentOp  downloader.TorrentOperator
	metaParser meta.Parser
	bfParser   bangumifile.Parser
	repository Repository
}

// calculatePathDepth 使用filepath包计算文件路径的目录深度
func calculatePathDepth(path string) int {
	cleanPath := filepath.Clean(path)
	if cleanPath == "." || cleanPath == string(filepath.Separator) {
		return 0
	}
	return 1 + calculatePathDepth(filepath.Dir(cleanPath))
}

// getMediaFileMap 获取媒体文件标记map，只有目录深度最浅的媒体文件才标记为true
func getMediaFileMap(fileNames []string) map[string]bool {
	// 媒体文件信息
	type mediaFileInfo struct {
		fileName string
		depth    int
	}

	var mediaFiles []mediaFileInfo
	minDepth := math.MaxInt

	// 找出所有媒体文件并计算它们的目录深度
	for _, fileName := range fileNames {
		if !utils.IsMediaFile(fileName) {
			continue
		}

		depth := calculatePathDepth(fileName)

		mediaFiles = append(mediaFiles, mediaFileInfo{
			fileName: fileName,
			depth:    depth,
		})

		// 记录最小深度
		if depth < minDepth {
			minDepth = depth
		}
	}

	// 创建一个map来标记哪些文件应该设置Media为true
	mediaFileMap := make(map[string]bool)
	for _, mf := range mediaFiles {
		// 只有深度最浅的媒体文件才设置为true
		mediaFileMap[mf.fileName] = (mf.depth == minDepth)
	}

	return mediaFileMap
}

// AddTask implements Interface.
func (m *Manager) AddTask(ctx context.Context, req AddTaskReq) (Task, error) {
	hash, err := extractHash(req.MagnetLink)
	if err != nil {
		return Task{}, fmt.Errorf("提取磁力链接哈希失败: %w", err)
	}
	log.Debugf(ctx, "提取磁力链接哈希成功: %s", hash)

	if _, err := m.repository.GetTaskByHash(ctx, hash); err != ErrMagnetTaskNotFound {
		if err == nil {
			return Task{}, errs.NewBadRequest("磁力任务已存在，请勿重复添加")
		}
		return Task{}, fmt.Errorf("获取下载任务失败: %w", err)
	}

	// 初始化任务信息，并保存
	task := Task{
		TaskID:       uuid.New().String(),
		MagnetLink:   req.MagnetLink,
		DownloadType: req.Type,
		CreatedAt:    time.Now(),
		Torrent: Torrent{
			Hash: hash,
		},
		Status: TaskStatusWaitingForParsing,
		Meta:   Meta{},
	}
	if err := m.repository.SaveTask(ctx, task); err != nil {
		return Task{}, fmt.Errorf("保存下载任务失败: %w", err)
	}

	// 调用下载器添加下载任务，设置不立即开始
	downloadReq := downloader.DownloadReq{
		TorrentLink:  req.MagnetLink,
		Hash:         hash,
		TaskID:       task.TaskID,
		DownloadType: req.Type,
		NotStart:     true, // 不立即开始下载
	}

	if err := m.downloader.Download(ctx, downloadReq); err != nil {
		m.deleteTask(ctx, task.TaskID)
		if errors.Is(err, context.DeadlineExceeded) {
			return Task{}, fmt.Errorf("种子元数据下载较慢，请等待下载器中种子元数据下载完成后再重试")
		}
		return Task{}, fmt.Errorf("添加下载任务失败: %w", err)
	}

	task, err = m.initTask(ctx, task)
	if err != nil {
		return Task{}, fmt.Errorf("解析种子元数据失败: %w", err)
	}
	return task, nil
}

func extractHash(magnetLink string) (string, error) {
	u, err := url.Parse(magnetLink)
	if err != nil {
		return "", fmt.Errorf("解析磁力链接失败: %w", err)
	}
	if u.Scheme != "magnet" {
		return "", fmt.Errorf("传入的链接不是磁力链接: %s", magnetLink)
	}
	params := u.Query()
	xt := params.Get("xt")
	if xt == "" {
		return "", fmt.Errorf("磁力链接中没有哈希值(xt)")
	}
	return infoHash(xt)
}

func infoHash(xt string) (string, error) {
	// 检查是否以 urn:btih: 开头
	const prefix = "urn:btih:"
	if !strings.HasPrefix(strings.ToLower(xt), prefix) {
		return "", errors.New("只支持btih格式的磁力链接")
	}

	// 提取哈希值部分
	hash := xt[len(prefix):]
	if hash == "" {
		return "", errors.New("磁力链接中没有哈希值")
	}

	// 根据长度判断是十六进制还是Base32编码
	switch len(hash) {
	case 40:
		// 40位，应该是十六进制格式
		hash = strings.ToLower(hash)
		// 验证是否为有效的十六进制
		if _, err := hex.DecodeString(hash); err != nil {
			return "", fmt.Errorf("无效的十六进制哈希值: %w", err)
		}
		return hash, nil

	case 32:
		// 32位，应该是Base32编码格式
		hash = strings.ToUpper(hash)
		// Base32解码
		decoded, err := base32.StdEncoding.DecodeString(hash)
		if err != nil {
			return "", fmt.Errorf("Base32解码失败: %w", err)
		}
		// 转换为十六进制
		return strings.ToLower(hex.EncodeToString(decoded)), nil

	default:
		return "", fmt.Errorf("哈希值长度错误: 期望40(hex)或32(base32)，实际为%d", len(hash))
	}
}

func (m *Manager) deleteTask(ctx context.Context, taskID string) {
	if err := m.DeleteTask(ctx, taskID); err != nil {
		log.Errorf(ctx, "删除下载任务失败: %v", err)
	}
}

func (m *Manager) initTask(ctx context.Context, task Task) (Task, error) {
	var status downloader.DownloadStatus
	// 解析元数据
	if err := wait.PollUntilContextTimeout(ctx, 500*time.Millisecond, 10*time.Second, true,
		func(ctx context.Context) (bool, error) {
			statuses, err := m.downloader.GetDownloadStatuses(ctx, []string{task.Torrent.Hash})
			if err != nil {
				log.Errorf(ctx, "初始化任务：获取种子状态失败: %v", err)
				return false, nil
			}
			if len(statuses) == 0 {
				return false, nil
			}
			if statuses[0].Status != downloader.TorrentStatusDownloadPaused {
				return false, nil
			}
			status = statuses[0]
			return true, nil
		}); err != nil {
		return Task{}, fmt.Errorf("获取种子状态失败: %w", err)
	}
	torrentName := status.Name
	log.Debugf(ctx, "解析出磁力链接 %s 的种子名称: %s", task.MagnetLink, torrentName)
	task.Torrent.Name = torrentName
	task.Torrent.Size = status.Size
	bangumiFile, _ := m.bfParser.Parse(ctx, torrentName)
	releaseGroup := bangumiFile.ReleaseGroup

	// 2. 获取种子文件列表
	fileNames, err := m.downloader.GetTorrentFileNames(ctx, task.Torrent.Hash)
	if err != nil {
		return Task{}, fmt.Errorf("获取种子文件失败: %w", err)
	}
	sort.Strings(fileNames)

	// 处理种子文件，设置TorrentFile属性
	var torrentFiles []TorrentFile

	// 获取媒体文件标记map
	mediaFileMap := getMediaFileMap(fileNames)

	// 处理所有文件
	for _, fileName := range fileNames {
		file := TorrentFile{
			FileName: fileName,
			Media:    mediaFileMap[fileName],
			Download: true,
		}
		if !file.Media {
			torrentFiles = append(torrentFiles, file)
			continue
		}

		if task.DownloadType == downloader.DownloadTypeTV {
			// 解析季数和集数
			bf, err := m.bfParser.Parse(ctx, fileName)
			if err != nil {
				log.Errorf(ctx, "解析文件名 %s 获取季数和集数失败: %v", fileName, err)
			} else {
				file.Season = bf.Season
				file.Episode = bf.Episode
			}
		}

		torrentFiles = append(torrentFiles, file)
	}
	task.Torrent.Files = torrentFiles
	task.Meta.ReleaseGroup = releaseGroup

	// 先保存一次，因为解析元数据比较容易出错，先把能保存的数据保存了
	if err := m.repository.SaveTask(ctx, task); err != nil {
		return Task{}, fmt.Errorf("保存下载任务失败: %w", err)
	}

	var (
		meta meta.Meta
	)
	if task.Meta.TMDBID != 0 {
		meta, err = m.parseMetaByID(ctx, task.Meta.TMDBID, task.DownloadType)
	} else {
		meta, err = m.parseMetaByTorrent(ctx, torrentName, task.DownloadType)
	}
	if err != nil {
		return Task{}, fmt.Errorf("解析种子元数据失败，请手动解析: %w", err)
	}
	task.Meta = Meta{
		ChineseName:  meta.ChineseName,
		Year:         meta.Year,
		TMDBID:       meta.TMDBID,
		ReleaseGroup: releaseGroup,
	}
	task.Status = TaskStatusWaitingForConfirmation

	if err := m.repository.SaveTask(ctx, task); err != nil {
		return Task{}, fmt.Errorf("保存下载任务失败: %w", err)
	}

	return task, nil
}

func (m *Manager) parseMetaByID(ctx context.Context, tmdbID int, downloadType downloader.DownloadType) (meta.Meta, error) {
	if downloadType == downloader.DownloadTypeTV {
		return m.metaParser.ParseTV(ctx, tmdbID)
	}
	return m.metaParser.ParseMovie(ctx, tmdbID)
}

func (m *Manager) parseMetaByTorrent(ctx context.Context, torrentName string, downloadType downloader.DownloadType) (meta.Meta, error) {
	bangumiFile, err := m.bfParser.Parse(ctx, torrentName, bangumifile.IgnoreValidateEpisode())
	if err != nil {
		return meta.Meta{}, fmt.Errorf("解析种子 %s 获取标题失败: %v", torrentName, err)
	}
	searchKey := bangumiFile.AnimeTitle
	if searchKey == "" {
		return meta.Meta{}, errors.New("无法从种子名称中解析出番剧名称，请手动解析")
	}
	if downloadType == downloader.DownloadTypeTV {
		return m.metaParser.SearchTV(ctx, searchKey)
	}
	return m.metaParser.SearchMovie(ctx, searchKey)
}

// InitTask 解析下载任务
func (m *Manager) InitTask(ctx context.Context, taskID string, tmdbID int) (Task, error) {
	task, err := m.repository.GetTask(ctx, taskID)
	if err != nil {
		return Task{}, fmt.Errorf("获取下载任务失败: %w", err)
	}
	if tmdbID != 0 {
		task.Meta.TMDBID = tmdbID
	}
	return m.initTask(ctx, task)
}

// DeleteTask implements Interface.
func (m *Manager) DeleteTask(ctx context.Context, taskID string) error {
	if taskID == "" {
		return fmt.Errorf("要删除的任务ID不能为空")
	}
	if err := m.repository.DeleteTask(ctx, taskID); err != nil {
		return fmt.Errorf("删除下载任务失败: %w", err)
	}
	return nil
}

func (m *Manager) PreviewAddSubtitles(ctx context.Context, req PreviewAddSubtitlesReq) (PreviewAddSubtitlesResp, error) {
	// 1. 获取任务信息
	task, err := m.repository.GetTask(ctx, req.TaskID)
	if err != nil {
		return PreviewAddSubtitlesResp{}, fmt.Errorf("获取任务信息失败: %w", err)
	}
	if task.Status != TaskStatusInitSuccess {
		return PreviewAddSubtitlesResp{}, fmt.Errorf("请先解析并确认下载任务后再进行字幕转移")
	}

	// 2. 获取种子信息，确定 torrent 根目录（用于拼出目标绝对路径）
	torrent, err := m.torrentOp.Get(ctx, task.Torrent.Hash)
	if err != nil {
		return PreviewAddSubtitlesResp{}, fmt.Errorf("获取种子信息失败: %w", err)
	}

	// 3. 获取字幕文件列表（支持文件/目录）
	subtitleFiles, err := m.getSubtitleFilesFromPath(req.SubtitlePath)
	if err != nil {
		return PreviewAddSubtitlesResp{}, fmt.Errorf("获取字幕文件失败: %w", err)
	}

	// 4. 解析 DstPath（支持文件/目录），并找出候选媒体文件
	mediaFiles, err := m.resolveDstMediaCandidates(task, req.DstPath, req.Season)
	if err != nil {
		return PreviewAddSubtitlesResp{}, fmt.Errorf("解析目标路径失败: %w", err)
	}

	// 5. 逐个字幕匹配并生成预览结果
	resp := PreviewAddSubtitlesResp{
		SubtileFiles: make(map[string]AddSubtitlesResult, len(subtitleFiles)),
	}

	offset := 0
	if req.EpisodeOffset != nil {
		offset = *req.EpisodeOffset
	}

	for _, subtitleFile := range subtitleFiles {
		res := m.previewMatchOneSubtitle(
			ctx,
			mediaFiles,
			subtitleFile,
			task.DownloadType,
			req.EpisodeLocation,
			offset,
			torrent.Path,
			req.ExtensionLevel,
		)
		resp.SubtileFiles[subtitleFile] = res
	}

	return resp, nil
}

// AddSubtitles implements Interface.
func (m *Manager) AddSubtitles(ctx context.Context, req AddSubtitlesReq) AddSubtitlesResp {
	resp := AddSubtitlesResp{
		FailedDetails: make(map[string]string),
	}

	for src, dst := range req.SubtitleFiles {
		var err error
		if req.PreserveOriginal {
			err = m.copySubtitleFile(src, dst)
		} else {
			err = m.moveSubtitleFile(ctx, src, dst)

		}

		if err != nil {
			resp.FailedDetails[src] = err.Error()
		} else {
			resp.SuccessCount++
		}
	}

	return resp
}

// copySubtitleFile 拷贝字幕文件到目标位置
func (m *Manager) copySubtitleFile(sourcePath, targetPath string) error {
	// 确保目标目录存在
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 打开源文件
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer sourceFile.Close()

	// 创建目标文件
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer targetFile.Close()

	// 拷贝文件内容
	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return fmt.Errorf("拷贝文件内容失败: %w", err)
	}

	return nil
}

func (m *Manager) moveSubtitleFile(ctx context.Context, src, dst string) error {
	// 确保目标目录存在
	targetDir := filepath.Dir(dst)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
	// 尝试直接移动
	if err := os.Rename(src, dst); err != nil {
		// 如果移动失败（可能是跨分区），尝试先拷贝再删除
		if err = m.copySubtitleFile(src, dst); err != nil {
			return err
		}
		if err := os.Remove(src); err != nil {
			log.Errorf(ctx, "拷贝成功但删除源文件失败: %v", err)
			return nil // 删除失败不视为整体失败，因为文件已经到达目标位置
		}
	}
	return nil
}

// GetTask implements Interface.
func (m *Manager) GetTask(ctx context.Context, taskID string) (Task, error) {
	return m.repository.GetTask(ctx, taskID)
}

// ListTasks implements Interface.
func (m *Manager) ListTasks(ctx context.Context, req ListTasksReq) ([]Task, int, error) {
	return m.repository.ListTasks(ctx, req)
}

// UpdateTask implements Interface.
func (m *Manager) UpdateTask(ctx context.Context, req UpdateTaskReq) error {
	task, err := m.repository.GetTask(ctx, req.TaskID)
	if err != nil {
		return fmt.Errorf("获取下载任务失败: %w", err)
	}

	if req.TMDBID != task.Meta.TMDBID {
		meta, err := m.parseMetaByID(ctx, req.TMDBID, task.DownloadType)
		if err != nil {
			return fmt.Errorf("解析种子元数据失败: %w", err)
		}
		task.Meta = Meta{
			ChineseName: meta.ChineseName,
			Year:        meta.Year,
			TMDBID:      req.TMDBID,
		}
	}

	// 处理文件级别的元数据
	for i := range req.Torrent.Files {
		file := &req.Torrent.Files[i]
		if file.Meta == nil {
			continue
		}
		if file.Meta.TMDBID == 0 || (file.Meta.MediaType != downloader.DownloadTypeMovie && file.Meta.MediaType != downloader.DownloadTypeTV) {
			return errs.NewBadRequest("文件级别元数据不能为空")
		}
		// 获取文件原有元数据（如果有）
		var oldTMDBID int
		for _, oldFile := range task.Torrent.Files {
			if oldFile.FileName == file.FileName && oldFile.Meta != nil {
				oldTMDBID = oldFile.Meta.TMDBID
				break
			}
		}
		// 如果 TMDBID 变化了或之前没有元数据，重新获取
		if file.Meta.TMDBID != oldTMDBID {
			meta, err := m.parseMetaByID(ctx, file.Meta.TMDBID, file.Meta.MediaType)
			if err != nil {
				return fmt.Errorf("解析文件 %s 元数据失败: %w", file.FileName, err)
			}
			file.Meta = &TorrentFileMeta{
				MediaType:   file.Meta.MediaType,
				ChineseName: meta.ChineseName,
				Year:        meta.Year,
				TMDBID:      file.Meta.TMDBID,
			}
		}
	}

	torrentSize, err := m.dealTorrentDownload(ctx, task.Torrent.Hash, req.Torrent.Files, task.Torrent.Size)
	if err != nil {
		return err
	}
	task.Torrent = req.Torrent
	if torrentSize != 0 {
		task.Torrent.Size = torrentSize
	}
	task.Meta.ReleaseGroup = req.ReleaseGroup
	task.Status = TaskStatusInitSuccess

	if err := m.repository.SaveTask(ctx, task); err != nil {
		return fmt.Errorf("保存下载任务失败: %w", err)
	}

	if req.ContinueDownload != nil && *req.ContinueDownload {
		if err := m.downloader.ContinueDownload(ctx, task.Torrent.Hash); err != nil {
			return fmt.Errorf("继续下载任务失败: %w", err)
		}
	}

	return nil
}

func (m *Manager) dealTorrentDownload(ctx context.Context, hash string, files []TorrentFile, curSize int64) (int64, error) {
	torrent, err := m.torrentOp.Get(ctx, hash)
	if err != nil {
		return 0, fmt.Errorf("获取种子信息失败: %w", err)
	}
	if !torrent.Status.IsDownloading() {
		return 0, nil
	}
	if len(files) == 0 {
		return 0, errs.NewBadRequest("未选择任何文件，无法开始下载")
	}
	fileSelections := make([]downloader.TorrentFileSelection, 0, len(files))
	selectedCount := 0
	for _, file := range files {
		if file.Download {
			selectedCount++
		}
		fileSelections = append(fileSelections, downloader.TorrentFileSelection{
			FileName: file.FileName,
			Download: file.Download,
		})
	}
	if selectedCount == 0 {
		return 0, errs.NewBadRequest("未选择任何文件，无法开始下载")
	}

	if err := m.downloader.SetTorrentFilePriorities(ctx, hash, fileSelections); err != nil {
		return 0, fmt.Errorf("设置种子文件下载选择失败: %w", err)
	}
	var size int64
	wait.PollUntilContextTimeout(ctx, 100*time.Millisecond, 5*time.Second, false, func(ctx context.Context) (bool, error) {
		statuses, err := m.downloader.GetDownloadStatuses(ctx, []string{hash})
		if err != nil {
			return false, err
		}
		if len(statuses) == 0 {
			return false, nil
		}
		size = statuses[0].Size
		if size != curSize {
			return true, nil
		}
		return false, nil
	})
	return size, nil
}

// getSubtitleFiles 获取字幕文件列表
func (m *Manager) getSubtitleFiles(subtitleDir string) ([]string, error) {
	if subtitleDir == "" {
		return nil, fmt.Errorf("需要提供字幕目录路径")
	}

	entries, err := os.ReadDir(subtitleDir)
	if err != nil {
		return nil, fmt.Errorf("读取字幕目录失败: %w", err)
	}

	var subtitleFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fp := filepath.Join(subtitleDir, entry.Name())
		if utils.IsSubtitleFile(fp) {
			subtitleFiles = append(subtitleFiles, fp)
		}
	}

	if len(subtitleFiles) == 0 {
		return nil, fmt.Errorf("在目录 %s 中未找到字幕文件", subtitleDir)
	}

	return subtitleFiles, nil
}

// getSubtitleFilesFromPath 获取字幕文件列表，支持目录或单文件路径
func (m *Manager) getSubtitleFilesFromPath(subtitlePath string) ([]string, error) {
	if subtitlePath == "" {
		return nil, fmt.Errorf("需要提供字幕路径")
	}

	info, err := os.Stat(subtitlePath)
	if err != nil {
		return nil, fmt.Errorf("读取字幕路径失败: %w", err)
	}

	if info.IsDir() {
		return m.getSubtitleFiles(subtitlePath)
	}

	if !utils.IsSubtitleFile(subtitlePath) {
		return nil, fmt.Errorf("不是有效的字幕文件: %s", subtitlePath)
	}
	return []string{subtitlePath}, nil
}

// resolveDstMediaCandidates 解析 DstPath（torrent 根目录内相对路径），支持文件/目录；返回候选媒体文件、目标目录绝对路径
func (m *Manager) resolveDstMediaCandidates(task Task, dstPath string, season *int) ([]TorrentFile, error) {
	cleanDst := filepath.Clean(dstPath)
	var mediaFiles []TorrentFile

	// 文件级：DstPath 精确匹配到某个媒体文件
	for _, f := range task.Torrent.Files {
		if !f.Media {
			continue
		}
		if filepath.Clean(f.FileName) == cleanDst {
			// dstPath是一个精确精确的文件名，提前返回
			return []TorrentFile{f}, nil
		}

		fileDir := filepath.Dir(f.FileName)
		if fileDir == cleanDst || filepath.Clean(fileDir) == filepath.Clean(cleanDst) {
			mediaFiles = append(mediaFiles, f)
		}
	}
	if season != nil && task.DownloadType == downloader.DownloadTypeTV {
		mediaFiles = lo.Filter(mediaFiles, func(mf TorrentFile, _ int) bool {
			return mf.Season == *season
		})
	}
	if len(mediaFiles) == 0 {
		if dstPath == "" {
			return nil, fmt.Errorf("在根目录中未找到媒体文件")
		}
		return nil, fmt.Errorf("在目录 %s 中未找到媒体文件", dstPath)
	}

	return mediaFiles, nil
}

func (m *Manager) previewMatchOneSubtitle(
	ctx context.Context,
	mediaFiles []TorrentFile,
	subtitleFile string,
	downloadType downloader.DownloadType,
	episodeLocation string,
	offset int,
	root string,
	extensionLevel *int,
) AddSubtitlesResult {
	makeFilePath := func(fileName string) string {
		return filepath.Join(root, filepath.FromSlash(fileName))
	}
	// 先找有没有同名文件
	for _, mf := range mediaFiles {
		subtitlePerfix := utils.GetFileBaseName(mf.FileName) + "."
		if strings.HasPrefix(filepath.Base(subtitleFile), subtitlePerfix) {
			return AddSubtitlesResult{
				SubtitleFile:  subtitleFile,
				NewFileName:   filepath.Base(subtitleFile),
				TargetPath:    filepath.Join(filepath.Dir(makeFilePath(mf.FileName)), filepath.Base(subtitleFile)),
				MediaFileName: filepath.Base(mf.FileName),
			}
		}
	}

	makeAddSubtitlesResult := func(mf, subtitleFile string) AddSubtitlesResult {
		newFileName := m.generateSubtitleFileName(mf, subtitleFile, extensionLevel)
		filePath := filepath.Join(filepath.Dir(makeFilePath(mf)), newFileName)
		return AddSubtitlesResult{
			SubtitleFile:  subtitleFile,
			NewFileName:   newFileName,
			TargetPath:    filePath,
			MediaFileName: filepath.Base(mf),
		}
	}

	if downloadType == downloader.DownloadTypeMovie {
		// 剧场版，非同名情况下只能有一个媒体文件，否则不知道字幕是哪个文件的
		if len(mediaFiles) != 1 {
			return AddSubtitlesResult{
				SubtitleFile: subtitleFile,
				Error:        fmt.Sprintf("movie 模式目标目录包含 %d 个媒体文件，且字幕文件名无法唯一对应媒体文件", len(mediaFiles)),
			}
		}
		return makeAddSubtitlesResult(mediaFiles[0].FileName, subtitleFile)
	}
	bf, err := m.bfParser.Parse(ctx, subtitleFile,
		bangumifile.WithEpisodeLocation(episodeLocation),
		bangumifile.WithEpisodeOffset(offset),
	)
	if err != nil {
		return AddSubtitlesResult{
			SubtitleFile: subtitleFile,
			Error:        err.Error(),
		}
	}

	for _, mf := range mediaFiles {
		if mf.Episode == bf.Episode {
			return makeAddSubtitlesResult(mf.FileName, subtitleFile)
		}
	}

	return AddSubtitlesResult{
		SubtitleFile: subtitleFile,
		Error:        fmt.Sprintf("未找到第 %d 集媒体文件", bf.Episode),
	}
}

// generateSubtitleFileName 生成字幕文件的目标文件名
func (m *Manager) generateSubtitleFileName(mediaFile, subtitleFilePath string, extensionLevel *int) string {
	// 获取媒体文件的基础名称（不含扩展名）
	mediaBaseName := utils.GetFileBaseName(filepath.Base(mediaFile))

	// 获取字幕文件的所有扩展名（包括语言、地区等标识）
	subtitleExt := m.getAllExtensions(subtitleFilePath, extensionLevel)
	return mediaBaseName + subtitleExt
}

// getAllExtensions 获取文件的所有扩展名，包括多级扩展名
// 例如: movie.zh.srt -> .zh.srt
//
//	movie.en-US.ass -> .en-US.ass
func (m *Manager) getAllExtensions(filePath string, extensionLevel *int) string {
	baseName := filepath.Base(filePath)
	level := math.MaxInt
	if extensionLevel != nil {
		level = *extensionLevel
	}

	// 使用 filepath.Ext 循环获取所有扩展名
	var allExts string
	for {
		ext := filepath.Ext(baseName)
		if ext == "" || level == 0 {
			break
		}
		allExts = ext + allExts
		baseName = strings.TrimSuffix(baseName, ext)
		level--
	}

	return allExts
}

func (m *Manager) FindTaskSimilarFiles(ctx context.Context, taskID, filePath string) ([]string, error) {
	if filePath == "" {
		return nil, fmt.Errorf("文件路径不能为空")
	}
	task, err := m.repository.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取下载任务失败: %w", err)
	}
	filePath = filepath.Clean(filePath)
	targetDir := filepath.Dir(filePath)

	var targetTitle string
	bf, _ := m.bfParser.Parse(ctx, filePath, bangumifile.IgnoreValidateEpisode())
	targetTitle = bf.AnimeTitle

	var result []string
	for _, file := range task.Torrent.Files {
		if filepath.Dir(file.FileName) != targetDir {
			continue
		}
		if filepath.Clean(file.FileName) == filePath {
			result = append(result, file.FileName)
			continue
		}
		if !file.Media {
			continue
		}
		bf, _ := m.bfParser.Parse(ctx, file.FileName, bangumifile.IgnoreValidateEpisode())
		if bf.AnimeTitle == targetTitle {
			result = append(result, file.FileName)
		}
	}

	return result, nil
}
