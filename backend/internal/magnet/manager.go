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
	"strings"
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/meta"
	"github.com/MangataL/BangumiBuddy/internal/types"
	"github.com/MangataL/BangumiBuddy/pkg/errs"
	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/MangataL/BangumiBuddy/pkg/utils"
)

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
	Downloader downloader.Interface
	TorrentOp  downloader.TorrentOperator
	MetaParser MetaParser
	Repository Repository
}

func New(dep Dependency) Interface {
	return &Manager{
		downloader: dep.Downloader,
		torrentOp:  dep.TorrentOp,
		metaParser: dep.MetaParser,
		repository: dep.Repository,
	}
}

// Manager 种子下载管理器
type Manager struct {
	downloader downloader.Interface
	torrentOp  downloader.TorrentOperator
	metaParser MetaParser
	repository Repository
}

// MetaParser 种子元数据解析器
type MetaParser interface {
	ParseReleaseGroup(ctx context.Context, torrentName string) (string, error)
	ParseMetaByTorrent(ctx context.Context, torrentName string, downloadType downloader.DownloadType) (meta.Meta, error)
	ParseMetaByID(ctx context.Context, tmdbID int, downloadType downloader.DownloadType) (meta.Meta, error)
	ParseFile(ctx context.Context, fileName string) (season, episode int, err error)
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
				log.Debugf(ctx, "初始化任务：等待种子添加完成")
				return false, nil
			}
			if statuses[0].Status != downloader.TorrentStatusDownloadPaused {
				log.Debugf(ctx, "初始化任务：等待种子元数据信息拉取完成")
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
	releaseGroup, _ := m.metaParser.ParseReleaseGroup(ctx, torrentName)

	// 2. 获取种子文件列表
	fileNames, err := m.downloader.GetTorrentFileNames(ctx, task.Torrent.Hash)
	if err != nil {
		return Task{}, fmt.Errorf("获取种子文件失败: %w", err)
	}

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
			season, episode, err := m.metaParser.ParseFile(ctx, fileName)
			if err != nil {
				log.Errorf(ctx, "解析文件名 %s 获取季数和集数失败: %v", fileName, err)
				continue
			}
			file.Season = season
			file.Episode = episode
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
		meta, err = m.metaParser.ParseMetaByID(ctx, task.Meta.TMDBID, task.DownloadType)
	} else {
		meta, err = m.metaParser.ParseMetaByTorrent(ctx, torrentName, task.DownloadType)
	}
	if err != nil {
		return Task{}, fmt.Errorf("解析种子元数据失败: %w", err)
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

// AddSubtitles implements Interface.
func (m *Manager) AddSubtitles(ctx context.Context, req AddSubtitlesReq) (int, error) {
	// 1. 获取任务信息
	task, err := m.repository.GetTask(ctx, req.TaskID)
	if err != nil {
		return 0, fmt.Errorf("获取任务信息失败: %w", err)
	}
	if task.Status != TaskStatusInitSuccess {
		return 0, fmt.Errorf("请先解析并确认下载任务后再进行字幕转移")
	}

	// 修正DstDir
	torrent, err := m.torrentOp.Get(ctx, task.Torrent.Hash)
	if err != nil {
		return 0, fmt.Errorf("获取种子信息失败: %w", err)
	}
	dstDirs := append([]string{torrent.Path}, strings.Split(req.DstDir, "/")...)
	path := filepath.Join(dstDirs...)

	// 2. 根据DstDir找到对应的媒体文件
	mediaFiles, err := m.findMediaFilesByDir(task, req.DstDir)
	if err != nil {
		return 0, fmt.Errorf("查找媒体文件失败: %w", err)
	}

	// 3. 获取字幕文件列表
	subtitleFiles, err := m.getSubtitleFiles(req.SubtitleDir)
	if err != nil {
		return 0, fmt.Errorf("获取字幕文件失败: %w", err)
	}

	// 4. 根据下载类型进行不同处理
	if task.DownloadType == downloader.DownloadTypeMovie {
		return m.processMovieSubtitles(subtitleFiles, mediaFiles, path)
	}
	return m.processTVSubtitles(ctx, subtitleFiles, mediaFiles, req, path)
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
		meta, err := m.metaParser.ParseMetaByID(ctx, req.TMDBID, task.DownloadType)
		if err != nil {
			return fmt.Errorf("解析种子元数据失败: %w", err)
		}
		task.Meta = Meta{
			ChineseName: meta.ChineseName,
			Year:        meta.Year,
			TMDBID:      req.TMDBID,
		}
	}

	task.Meta.ReleaseGroup = req.ReleaseGroup
	task.Status = TaskStatusInitSuccess
	task.Torrent = req.Torrent

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

// findMediaFilesByDir 根据DstDir找到对应的媒体文件组
func (m *Manager) findMediaFilesByDir(task Task, dstDir string) ([]TorrentFile, error) {
	var mediaFiles []TorrentFile

	for _, file := range task.Torrent.Files {
		if !file.Media {
			continue
		}

		// 检查文件是否在指定的目录中
		fileDir := filepath.Dir(file.FileName)
		if fileDir == dstDir || filepath.Clean(fileDir) == filepath.Clean(dstDir) {
			mediaFiles = append(mediaFiles, file)
		}
	}

	if len(mediaFiles) == 0 {
		if dstDir == "" {
			return nil, fmt.Errorf("在根目录中未找到媒体文件")
		}
		return nil, fmt.Errorf("在目录 %s 中未找到媒体文件", dstDir)
	}

	return mediaFiles, nil
}

// getSubtitleFiles 获取字幕文件列表
func (m *Manager) getSubtitleFiles(subtitleDir string) ([]string, error) {
	if subtitleDir == "" {
		return nil, fmt.Errorf("需要提供字幕目录路径")
	}

	files, err := filepath.Glob(filepath.Join(subtitleDir, "*"))
	if err != nil {
		return nil, fmt.Errorf("扫描字幕目录失败: %w", err)
	}

	var subtitleFiles []string
	for _, file := range files {
		if utils.IsSubtitleFile(file) {
			subtitleFiles = append(subtitleFiles, file)
		}
	}

	if len(subtitleFiles) == 0 {
		return nil, fmt.Errorf("在目录 %s 中未找到字幕文件", subtitleDir)
	}

	return subtitleFiles, nil
}

// processMovieSubtitles 处理电影字幕
func (m *Manager) processMovieSubtitles(subtitleFiles []string, mediaFiles []TorrentFile, path string) (int, error) {
	if len(mediaFiles) != 1 {
		return 0, fmt.Errorf("剧场版应该只有一个媒体文件，但找到了 %d 个", len(mediaFiles))
	}

	mediaFile := mediaFiles[0]

	successCount := 0
	for _, subtitleFile := range subtitleFiles {
		// 生成目标文件名
		targetFileName := m.generateSubtitleFileName(mediaFile, subtitleFile)
		targetPath := filepath.Join(path, targetFileName)

		// 移动字幕文件
		if err := m.copySubtitleFile(subtitleFile, targetPath); err != nil {
			log.Warnf(context.Background(), "拷贝字幕 %s 到 %s 失败: %v", subtitleFile, targetPath, err)
			continue
		}
		successCount++
		log.Infof(context.Background(), "拷贝字幕 %s 到 %s", subtitleFile, targetPath)
	}
	return successCount, nil
}

// processTVSubtitles 处理TV字幕
func (m *Manager) processTVSubtitles(ctx context.Context, subtitleFiles []string, mediaFiles []TorrentFile, req AddSubtitlesReq, path string) (int, error) {
	// 为每个字幕文件解析集数信息
	subtitleEpisodes := make(map[int][]string) // episode -> subtitleFiles
	offset := 0
	if req.EpisodeOffset != nil {
		offset = *req.EpisodeOffset
	}

	for _, subtitleFile := range subtitleFiles {
		var episode int
		var err error

		if req.EpisodeLocation != "" {
			// 使用EpisodeLocation解析
			episode, err = utils.ParseEpisodeWithLocation(filepath.Base(subtitleFile), req.EpisodeLocation)
			if err != nil {
				log.Warnf(ctx, "使用EpisodeLocation解析字幕文件 %s 集数失败: %v", subtitleFile, err)
				continue
			}
		} else {
			// 使用ParseFile解析
			_, episode, err = m.metaParser.ParseFile(ctx, filepath.Base(subtitleFile))
			if err != nil {
				log.Warnf(ctx, "解析字幕文件 %s 集数失败: %v", subtitleFile, err)
				continue
			}
		}
		episode += offset
		subtitleEpisodes[episode] = append(subtitleEpisodes[episode], subtitleFile)
	}

	// 遍历媒体文件，匹配字幕文件
	successCount := 0
	for _, mediaFile := range mediaFiles {
		subtitleFiles, exists := subtitleEpisodes[mediaFile.Episode]
		if !exists {
			log.Warnf(ctx, "未找到第 %d 集的字幕文件", mediaFile.Episode)
			continue
		}

		for _, subtitleFile := range subtitleFiles {
			// 生成目标文件名
			targetFileName := m.generateSubtitleFileName(mediaFile, subtitleFile)
			targetPath := filepath.Join(path, targetFileName)

			// 移动字幕文件
			if err := m.copySubtitleFile(subtitleFile, targetPath); err != nil {
				log.Warnf(ctx, "拷贝字幕 %s 到 %s 失败: %v", subtitleFile, targetPath, err)
				continue
			}
			successCount++
			log.Infof(ctx, "拷贝字幕 %s 到 %s", subtitleFile, targetPath)
		}
	}

	return successCount, nil
}

// generateSubtitleFileName 生成字幕文件的目标文件名
func (m *Manager) generateSubtitleFileName(mediaFile TorrentFile, subtitleFilePath string) string {
	// 获取媒体文件的基础名称（不含扩展名）
	mediaBaseName := utils.GetFileBaseName(filepath.Base(mediaFile.FileName))

	// 获取字幕文件的所有扩展名（包括语言、地区等标识）
	subtitleExt := m.getAllExtensions(subtitleFilePath)

	return mediaBaseName + subtitleExt
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

	log.Infof(context.Background(), "成功拷贝字幕文件: %s -> %s", sourcePath, targetPath)
	return nil
}

// getAllExtensions 获取文件的所有扩展名，包括多级扩展名
// 例如: movie.zh.srt -> .zh.srt
//
//	movie.en-US.ass -> .en-US.ass
func (m *Manager) getAllExtensions(filePath string) string {
	baseName := filepath.Base(filePath)

	// 使用 filepath.Ext 循环获取所有扩展名
	var allExts string
	for {
		ext := filepath.Ext(baseName)
		if ext == "" {
			break
		}
		allExts = ext + allExts
		baseName = strings.TrimSuffix(baseName, ext)
	}

	return allExts
}
