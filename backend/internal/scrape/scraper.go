package scrape

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/beevik/etree"
	"github.com/pkg/errors"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/meta"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

var _ Interface = (*Scraper)(nil)

type Dependency struct {
	Config
	Repository Repository
	MetaParser meta.Parser
}

type Repository interface {
	Add(ctx context.Context, task MetadataCheckTask) error
	List(ctx context.Context) ([]MetadataCheckTask, error)
	Delete(ctx context.Context, filePath string) error
	UpdateImageChecked(ctx context.Context, filePath string) error
	Clean(ctx context.Context) error
}

type Scraper struct {
	mu         sync.Mutex
	config     Config
	repo       Repository
	metaParser meta.Parser
	ticker     *time.Ticker
	stop       func()
	client     *http.Client
}

func NewScraper(dep Dependency) *Scraper {
	ctx, cancel := context.WithCancel(context.Background())

	scraper := &Scraper{
		config:     dep.Config,
		repo:       dep.Repository,
		metaParser: dep.MetaParser,
		stop:       cancel,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	go scraper.run(ctx)
	return scraper
}

func (s *Scraper) run(ctx context.Context) {
	s.ticker = time.NewTicker(time.Duration(s.config.CheckInterval) * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.ticker.C:
			s.checkMetadata(ctx)
		}
	}
}

func (s *Scraper) AddMetadataFillTask(ctx context.Context, req AddMetadataFillTaskReq) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.config.Enable {
		return nil
	}

	// 仅处理 TV 类型
	if req.DownloadType != downloader.DownloadTypeTV {
		log.Infof(ctx, "跳过非TV类型的元数据填充任务: %s", req.FilePath)
		return nil
	}

	task := MetadataCheckTask{
		TMDBID:       req.TMDBID,
		FilePath:     req.FilePath,
		DownloadType: req.DownloadType,
	}

	if err := s.repo.Add(ctx, task); err != nil {
		return errors.WithMessage(err, "添加元数据填充任务失败")
	}

	return nil
}

func (s *Scraper) checkMetadata(ctx context.Context) {
	if !s.Enable() {
		return
	}

	tasks, err := s.repo.List(ctx)
	if err != nil {
		log.Errorf(ctx, "获取元数据填充任务列表失败: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	log.Infof(ctx, "开始巡检元数据，共 %d 个任务", len(tasks))

	for _, task := range tasks {
		if err := s.processTask(ctx, task); err != nil {
			log.Errorf(ctx, "处理元数据填充任务失败(FilePath=%s): %v", task.FilePath, err)
		}
	}
}

func (s *Scraper) processTask(ctx context.Context, task MetadataCheckTask) error {
	if !fileExists(task.FilePath) {
		log.Infof(ctx, "文件 %s 已不存在，不再处理元数据", task.FilePath)
		_ = s.repo.Delete(ctx, task.FilePath)
		return nil
	}
	// 1. 推导 NFO 文件路径
	nfoPath := strings.TrimSuffix(task.FilePath, filepath.Ext(task.FilePath)) + ".nfo"

	// 2. 检查 NFO 文件是否存在
	if !fileExists(nfoPath) {
		log.Debugf(ctx, "NFO 文件不存在，等待下次轮询: %s", nfoPath)
		return nil // 等待下次轮询
	}

	// 3. 解析 NFO 文件
	nfoData, err := s.parseNFO(nfoPath)
	if err != nil {
		return errors.WithMessage(err, "解析 NFO 文件失败")
	}

	if nfoData.season == 0 || nfoData.episode == 0 {
		log.Debugf(ctx, "NFO 文件中没有季数或集数，等待下次轮询: %s", nfoPath)
		return nil // 等待下次轮询
	}

	// 4. 从 TMDB 获取正确的元数据
	episodeDetails, err := s.metaParser.GetEpisodeDetails(ctx, task.TMDBID, nfoData.season, nfoData.episode)
	if err != nil {
		return errors.WithMessage(err, "获取 TMDB 单集元数据失败")
	}

	// 5. 如果需要更新，则更新 NFO
	if s.nfoNeedUpdate(nfoData, episodeDetails, task) {
		allUpdated, err := s.updateNFO(ctx, nfoPath, nfoData, episodeDetails, task)
		if err != nil {
			return errors.WithMessage(err, "更新 NFO 文件失败")
		}
		if !allUpdated {
			return nil
		}
	}

	// 所有元数据都正确，立即删除任务
	if err := s.repo.Delete(ctx, task.FilePath); err != nil {
		log.Warnf(ctx, "删除任务失败: %v", err)
	}

	return nil
}

type nfoData struct {
	doc        *etree.Document
	title      string
	plot       string
	posterPath string
	season     int
	episode    int
}

// parseNFO 解析 NFO 文件，提取关键信息
func (s *Scraper) parseNFO(path string) (*nfoData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.WithMessage(err, "读取 NFO 文件失败")
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, errors.WithMessage(err, "解析 XML 失败")
	}

	root := doc.SelectElement("episodedetails")
	if root == nil {
		return nil, errors.New("NFO 文件格式错误：缺少 episodedetails 根节点")
	}

	var title, plot, posterPath string
	var season, episode int

	if el := root.SelectElement("title"); el != nil {
		title = el.Text()
	}
	if el := root.SelectElement("plot"); el != nil {
		plot = el.Text()
	}
	if el := root.SelectElement("season"); el != nil {
		fmt.Sscanf(el.Text(), "%d", &season)
	}
	if el := root.SelectElement("episode"); el != nil {
		fmt.Sscanf(el.Text(), "%d", &episode)
	}
	if el := root.SelectElement("art"); el != nil {
		if posterEl := el.SelectElement("poster"); posterEl != nil {
			posterPath = posterEl.Text()
		}
	}

	return &nfoData{
		doc:        doc,
		title:      title,
		plot:       plot,
		posterPath: posterPath,
		season:     season,
		episode:    episode,
	}, nil
}

// nfoNeedUpdate 检查元数据是否需要更新
func (s *Scraper) nfoNeedUpdate(nfo *nfoData, details meta.EpisodeDetails, task MetadataCheckTask) bool {
	needUpdate := false
	if episodeNameInvalid(nfo.title) {
		needUpdate = true
	}
	if nfo.plot == "" {
		needUpdate = true
	}
	if details.StillPath != "" && !task.ImageChecked {
		needUpdate = true
	}

	return needUpdate
}

func episodeNameInvalid(name string) bool {
	return name == "" || isInvalidTitle(name)
}

// isInvalidTitle 判断标题是否无效（包含"第 x 集"格式）
func isInvalidTitle(title string) bool {
	// 匹配 "第 x 集" 格式，空格可选（\s* 表示零个或多个空格）
	re := regexp.MustCompile(`第\s*\d+\s*集`)
	return re.MatchString(title)
}

// updateNFO 更新 NFO 文件
func (s *Scraper) updateNFO(ctx context.Context, path string, nfo *nfoData, details meta.EpisodeDetails, task MetadataCheckTask) (allUpdated bool, err error) {
	nfoUpdated := false
	var (
		titleUpdated bool
		plotUpdated  bool
		imageUpdated bool
	)

	root := nfo.doc.SelectElement("episodedetails")
	if root == nil {
		return false, errors.New("NFO 格式错误：缺少 episodedetails")
	}

	// 更新 title
	if !episodeNameInvalid(details.Name) {
		if el := root.SelectElement("title"); el != nil {
			el.SetText(details.Name)
			nfoUpdated = true
			titleUpdated = true
		}
	}

	// 更新 plot
	if details.Overview != "" {
		if el := root.SelectElement("plot"); el != nil {
			el.SetText(details.Overview)
			nfoUpdated = true
			plotUpdated = true
		}
	}

	// 更新图片（仅在未检查过时执行）
	if !task.ImageChecked && details.StillPath != "" && nfo.posterPath != "" {
		// 下载图片并比对
		if err := s.checkAndReplaceImage(ctx, nfo.posterPath, details.StillPath); err != nil {
			log.Warnf(ctx, "更新图片失败: %v", err)
			return false, errors.WithMessage(err, "更新图片失败")
		}

		imageUpdated = true
		if err := s.repo.UpdateImageChecked(ctx, task.FilePath); err != nil {
			log.Warnf(ctx, "更新图片检查状态失败: %v", err)
		}
	}

	// 写回 NFO 文件
	if nfoUpdated {
		if err := nfo.doc.WriteToFile(path); err != nil {
			return false, errors.WithMessage(err, "写入 NFO 文件失败")
		}
	}
	if titleUpdated && plotUpdated && (imageUpdated || task.ImageChecked) {
		return true, nil
	}

	return false, nil
}

// checkAndReplaceImage 下载图片、比对 MD5 并在需要时替换（一次性操作）
// 返回是否进行了替换
func (s *Scraper) checkAndReplaceImage(ctx context.Context, currentPosterPath, posterPath string) error {
	// 下载 TMDB 图片（只下载一次）
	newImageData, err := s.downloadImage(ctx, posterPath)
	if err != nil {
		return errors.WithMessage(err, "下载 TMDB 图片失败")
	}

	// 计算新图片的 MD5
	newMD5 := md5.Sum(newImageData)

	// 读取当前图片并计算 MD5
	var currentMD5 [16]byte
	needReplace := true
	if fileExists(currentPosterPath) {
		currentImageData, err := os.ReadFile(currentPosterPath)
		if err == nil {
			currentMD5 = md5.Sum(currentImageData)
			needReplace = (newMD5 != currentMD5)
		}
	}

	// 如果需要替换
	if needReplace {
		// 写入新文件
		return os.WriteFile(currentPosterPath, newImageData, 0644)
	}

	return nil
}

// downloadImage 下载 TMDB 图片
func (s *Scraper) downloadImage(ctx context.Context, tmdbImagePath string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tmdbImagePath, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载图片失败，状态码: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// fileExists 检查文件是否存在
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *Scraper) Close() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.stop()
}

func (s *Scraper) Enable() bool {
	return s.config.Enable
}

func (s *Scraper) Reload(config interface{}) error {
	cfg, ok := config.(*Config)
	if !ok {
		return errors.New("配置类型错误")
	}
	if s.ticker != nil {
		s.ticker.Reset(time.Duration(cfg.CheckInterval) * time.Hour)
	}
	s.config = *cfg
	if !s.config.Enable {
		if err := s.repo.Clean(context.Background()); err != nil {
			log.Warnf(context.Background(), "清空任务失败: %v", err)
		}
	}
	return nil
}
