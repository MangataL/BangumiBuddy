package scrape

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MangataL/BangumiBuddy/internal/meta"
)

type memoryRepo struct {
	mu     sync.Mutex
	nextID uint
	tasks  map[string]MetadataCheckTask
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func setupScrapeTestRepository(t *testing.T) *memoryRepo {
	t.Helper()

	return &memoryRepo{
		tasks: make(map[string]MetadataCheckTask),
	}
}

func cloneTask(task MetadataCheckTask) MetadataCheckTask {
	task.Statuses = append([]ScrapeStatus(nil), task.Statuses...)
	return task
}

func (r *memoryRepo) Add(ctx context.Context, task MetadataCheckTask) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if old, ok := r.tasks[task.FilePath]; ok {
		task.ID = old.ID
	} else {
		r.nextID++
		task.ID = r.nextID
	}
	r.tasks[task.FilePath] = cloneTask(task)
	return nil
}

func (r *memoryRepo) Get(ctx context.Context, id uint) (MetadataCheckTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, task := range r.tasks {
		if task.ID == id {
			return cloneTask(task), nil
		}
	}
	return MetadataCheckTask{}, ErrTaskNotFound
}

func (r *memoryRepo) List(ctx context.Context) ([]MetadataCheckTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	tasks := make([]MetadataCheckTask, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, cloneTask(task))
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return tasks, nil
}

func (r *memoryRepo) Delete(ctx context.Context, filePath string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.tasks, filePath)
	return nil
}

func (r *memoryRepo) UpdateStatuses(ctx context.Context, filePath string, statuses []ScrapeStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[filePath]
	if !ok {
		return nil
	}
	task.Statuses = append([]ScrapeStatus(nil), statuses...)
	r.tasks[filePath] = task
	return nil
}

func (r *memoryRepo) Clean(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks = make(map[string]MetadataCheckTask)
	return nil
}

func writeTestNFO(t *testing.T, path, title, plot string, season, episode int, posterPath, artMode string) {
	t.Helper()

	artXML := fmt.Sprintf("  <art>\n    <poster>%s</poster>\n  </art>\n", posterPath)
	switch artMode {
	case "noArt":
		artXML = ""
	case "artNoPoster":
		artXML = "  <art></art>\n"
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<episodedetails>
  <title>%s</title>
  <plot>%s</plot>
  <season>%d</season>
  <episode>%d</episode>
%s
</episodedetails>`, title, plot, season, episode, artXML)

	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}

func newImageOKHTTPClient() *http.Client {
	return &http.Client{
		Timeout: time.Second,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("fake-image")),
				Header:     make(http.Header),
			}, nil
		}),
	}
}

func TestFileExists(t *testing.T) {
	assert.True(t, fileExists("scraper.go"))
	assert.False(t, fileExists("nonexistent_file.txt"))
}

func Test_parseNFO(t *testing.T) {
	scraper := &Scraper{}
	plot := `真昼遇见了神秘的女性菊并与她逐渐打好关系，然而更却看穿了她的真面目。菊虽然很犹豫，还是跟真昼说明了真实情况……后来，更听荠说了菊的过去，便有了更多疑虑。`

	nfoData, err := scraper.parseNFO("./testdata/彻夜之歌 S02E02.nfo")

	require.NoError(t, err)
	assert.Equal(t, "好想见到你", nfoData.title)
	assert.Equal(t, 2, nfoData.season)
	assert.Equal(t, 2, nfoData.episode)
	assert.Equal(t, "/video/bangumi/彻夜之歌/Season 2/彻夜之歌 S02E02-thumb.jpg", nfoData.posterPath)
	assert.Equal(t, plot, nfoData.plot)
}

func Test_parseNFO_NoArtOrPosterField(t *testing.T) {
	scraper := &Scraper{}

	tests := []struct {
		name    string
		artMode string
	}{
		{name: "没有art字段", artMode: "noArt"},
		{name: "art中没有poster字段", artMode: "artNoPoster"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nfoPath := filepath.Join(t.TempDir(), "ep.nfo")
			writeTestNFO(t, nfoPath, "标题", "剧情", 1, 1, "poster.jpg", tt.artMode)

			nfoData, err := scraper.parseNFO(nfoPath)
			require.NoError(t, err)
			assert.Equal(t, "", nfoData.posterPath)
		})
	}
}

func TestPosterExtFromStillPath(t *testing.T) {
	tests := []struct {
		name      string
		stillPath string
		want      string
	}{
		{
			name:      "普通图片链接",
			stillPath: "https://image.test/episode.webp",
			want:      ".webp",
		},
		{
			name:      "带查询参数的图片链接",
			stillPath: "https://image.test/episode.jpeg?token=1",
			want:      ".jpeg",
		},
		{
			name:      "本地路径",
			stillPath: "/tmp/episode.png",
			want:      ".png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, posterExtFromStillPath(tt.stillPath))
		})
	}
}

func TestScraper_AddAndListTasks(t *testing.T) {
	ctx := context.Background()
	repo := setupScrapeTestRepository(t)

	scraper := &Scraper{
		config: Config{Enable: true},
		repo:   repo,
	}

	req := AddMetadataFillTaskReq{
		TMDBID:      1,
		FilePath:    "/tmp/episode.mkv",
		BangumiName: "测试番剧",
		PosterURL:   "https://example.com/poster.jpg",
		Season:      1,
		Episode:     1,
	}
	require.NoError(t, scraper.AddMetadataFillTask(ctx, req))

	tasks, err := scraper.ListTasks(ctx)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, []ScrapeStatus{ScrapeStatusPending}, tasks[0].Statuses)

	disabled := &Scraper{
		config: Config{Enable: false},
		repo:   repo,
	}
	require.NoError(t, disabled.AddMetadataFillTask(ctx, AddMetadataFillTaskReq{
		TMDBID:   2,
		FilePath: "/tmp/episode2.mkv",
	}))

	tasks, err = scraper.ListTasks(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
}

func TestScraper_nfoNeedUpdate(t *testing.T) {
	scraper := &Scraper{}

	tests := []struct {
		name                string
		nfo                 *nfoData
		details             meta.EpisodeDetails
		imageAlreadyChecked bool
		want                bool
	}{
		{
			name: "标题无效时需要更新",
			nfo: &nfoData{
				title: "第 1 集",
				plot:  "剧情",
			},
			details:             meta.EpisodeDetails{Overview: "剧情"},
			imageAlreadyChecked: true,
			want:                true,
		},
		{
			name: "nfo剧情为空时需要更新",
			nfo: &nfoData{
				title: "标题",
				plot:  "",
			},
			details:             meta.EpisodeDetails{Overview: "剧情"},
			imageAlreadyChecked: true,
			want:                true,
		},
		{
			name: "详情剧情为空时需要更新",
			nfo: &nfoData{
				title: "标题",
				plot:  "剧情",
			},
			details:             meta.EpisodeDetails{Overview: ""},
			imageAlreadyChecked: true,
			want:                true,
		},
		{
			name: "剧情不一致时需要更新",
			nfo: &nfoData{
				title: "标题",
				plot:  "旧剧情",
			},
			details:             meta.EpisodeDetails{Overview: "新剧情"},
			imageAlreadyChecked: true,
			want:                true,
		},
		{
			name: "图片未检查且详情有图片时需要更新",
			nfo: &nfoData{
				title: "标题",
				plot:  "剧情",
			},
			details:             meta.EpisodeDetails{Overview: "剧情", StillPath: "https://image.test/episode.jpg"},
			imageAlreadyChecked: false,
			want:                true,
		},
		{
			name: "全部匹配且图片已检查时不需要更新",
			nfo: &nfoData{
				title: "标题",
				plot:  "剧情",
			},
			details:             meta.EpisodeDetails{Overview: "剧情", StillPath: "https://image.test/episode.jpg"},
			imageAlreadyChecked: true,
			want:                false,
		},
		{
			name: "全部匹配且详情无图片时不需要更新",
			nfo: &nfoData{
				title: "标题",
				plot:  "剧情",
			},
			details:             meta.EpisodeDetails{Overview: "剧情"},
			imageAlreadyChecked: false,
			want:                false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, scraper.nfoNeedUpdate(tt.nfo, tt.details, tt.imageAlreadyChecked))
		})
	}
}

func TestScraper_nfoStatus(t *testing.T) {
	scraper := &Scraper{}

	tests := []struct {
		name    string
		nfo     *nfoData
		details meta.EpisodeDetails
		want    []ScrapeStatus
	}{
		{
			name: "全部正常时无缺失状态",
			nfo: &nfoData{
				posterPath: "/tmp/poster.jpg",
			},
			details: meta.EpisodeDetails{
				Name:      "标题",
				Overview:  "剧情",
				StillPath: "https://image.test/episode.jpg",
			},
			want: nil,
		},
		{
			name: "标题无效时标记missingTitle",
			nfo: &nfoData{
				posterPath: "/tmp/poster.jpg",
			},
			details: meta.EpisodeDetails{
				Name:      "第 1 集",
				Overview:  "剧情",
				StillPath: "https://image.test/episode.jpg",
			},
			want: []ScrapeStatus{ScrapeStatusMissingTitle},
		},
		{
			name: "剧情为空时标记missingPlot",
			nfo: &nfoData{
				posterPath: "/tmp/poster.jpg",
			},
			details: meta.EpisodeDetails{
				Name:      "标题",
				Overview:  "",
				StillPath: "https://image.test/episode.jpg",
			},
			want: []ScrapeStatus{ScrapeStatusMissingPlot},
		},
		{
			name: "详情图片为空时标记missingImage",
			nfo: &nfoData{
				posterPath: "/tmp/poster.jpg",
			},
			details: meta.EpisodeDetails{
				Name:      "标题",
				Overview:  "剧情",
				StillPath: "",
			},
			want: []ScrapeStatus{ScrapeStatusMissingImage},
		},
		{
			name: "nfo海报为空时标记missingImage",
			nfo: &nfoData{
				posterPath: "",
			},
			details: meta.EpisodeDetails{
				Name:      "标题",
				Overview:  "剧情",
				StillPath: "https://image.test/episode.jpg",
			},
			want: []ScrapeStatus{ScrapeStatusMissingImage},
		},
		{
			name: "多项缺失时按顺序返回状态",
			nfo: &nfoData{
				posterPath: "",
			},
			details: meta.EpisodeDetails{
				Name:      "第 1 集",
				Overview:  "",
				StillPath: "",
			},
			want: []ScrapeStatus{
				ScrapeStatusMissingTitle,
				ScrapeStatusMissingPlot,
				ScrapeStatusMissingImage,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scraper.nfoStatus(tt.nfo, tt.details)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestScraper_ProcessTask_RecordAndReadResultByPath(t *testing.T) {
	type nfoFixture struct {
		title      string
		plot       string
		season     int
		episode    int
		posterPath string
		artMode    string
	}

	tests := []struct {
		name                  string
		initialStatuses       []ScrapeStatus
		createMediaFile       bool
		createNFOFile         bool
		nfo                   nfoFixture
		details               meta.EpisodeDetails
		expectParserCall      int
		expectTaskCount       int
		expectStatuses        []ScrapeStatus
		expectGeneratedPoster bool
	}{
		{
			name:             "媒体文件不存在时删除任务",
			initialStatuses:  []ScrapeStatus{ScrapeStatusPending},
			createMediaFile:  false,
			createNFOFile:    false,
			expectParserCall: 0,
			expectTaskCount:  0,
		},
		{
			name:             "NFO文件不存在时保持待处理状态",
			initialStatuses:  []ScrapeStatus{ScrapeStatusPending},
			createMediaFile:  true,
			createNFOFile:    false,
			expectParserCall: 0,
			expectTaskCount:  1,
			expectStatuses:   []ScrapeStatus{ScrapeStatusPending},
		},
		{
			name:            "缺少标题时录入missingTitle",
			initialStatuses: []ScrapeStatus{ScrapeStatusPending},
			createMediaFile: true,
			createNFOFile:   true,
			nfo: nfoFixture{
				title:      "第 1 集",
				plot:       "旧剧情",
				season:     1,
				episode:    1,
				posterPath: "poster.jpg",
			},
			details: meta.EpisodeDetails{
				Name:      "第 1 集",
				Overview:  "新剧情",
				StillPath: "https://image.test/episode.jpg",
			},
			expectParserCall: 1,
			expectTaskCount:  1,
			expectStatuses:   []ScrapeStatus{ScrapeStatusMissingTitle},
		},
		{
			name:            "缺少剧情时录入missingPlot",
			initialStatuses: []ScrapeStatus{ScrapeStatusPending},
			createMediaFile: true,
			createNFOFile:   true,
			nfo: nfoFixture{
				title:      "旧标题",
				plot:       "旧剧情",
				season:     1,
				episode:    1,
				posterPath: "poster.jpg",
			},
			details: meta.EpisodeDetails{
				Name:      "新标题",
				Overview:  "",
				StillPath: "https://image.test/episode.jpg",
			},
			expectParserCall: 1,
			expectTaskCount:  1,
			expectStatuses:   []ScrapeStatus{ScrapeStatusMissingPlot},
		},
		{
			name:            "缺少海报时自动补充poster路径",
			initialStatuses: []ScrapeStatus{ScrapeStatusPending},
			createMediaFile: true,
			createNFOFile:   true,
			nfo: nfoFixture{
				title:      "标题",
				plot:       "剧情",
				season:     1,
				episode:    1,
				posterPath: "",
			},
			details: meta.EpisodeDetails{
				Name:      "标题",
				Overview:  "剧情",
				StillPath: "https://image.test/episode.webp",
			},
			expectParserCall:      1,
			expectTaskCount:       0,
			expectGeneratedPoster: true,
		},
		{
			name:            "缺少海报时自动补充poster路径(nfo没有art字段)",
			initialStatuses: []ScrapeStatus{ScrapeStatusPending},
			createMediaFile: true,
			createNFOFile:   true,
			nfo: nfoFixture{
				title:      "标题",
				plot:       "剧情",
				season:     1,
				episode:    1,
				posterPath: "poster.jpg",
				artMode:    "noArt",
			},
			details: meta.EpisodeDetails{
				Name:      "标题",
				Overview:  "剧情",
				StillPath: "https://image.test/episode.png",
			},
			expectParserCall:      1,
			expectTaskCount:       0,
			expectGeneratedPoster: true,
		},
		{
			name:            "缺少海报时自动补充poster路径(art没有poster字段)",
			initialStatuses: []ScrapeStatus{ScrapeStatusPending},
			createMediaFile: true,
			createNFOFile:   true,
			nfo: nfoFixture{
				title:      "标题",
				plot:       "剧情",
				season:     1,
				episode:    1,
				posterPath: "poster.jpg",
				artMode:    "artNoPoster",
			},
			details: meta.EpisodeDetails{
				Name:      "标题",
				Overview:  "剧情",
				StillPath: "https://image.test/episode.jpeg?token=1",
			},
			expectParserCall:      1,
			expectTaskCount:       0,
			expectGeneratedPoster: true,
		},
		{
			name:            "全部都更新时删除任务",
			initialStatuses: []ScrapeStatus{ScrapeStatusPending},
			createMediaFile: true,
			createNFOFile:   true,
			nfo: nfoFixture{
				title:      "第 1 集",
				plot:       "旧剧情",
				season:     1,
				episode:    1,
				posterPath: "poster.jpg",
			},
			details: meta.EpisodeDetails{
				Name:      "新标题",
				Overview:  "新剧情",
				StillPath: "https://image.test/episode.jpg",
			},
			expectParserCall: 1,
			expectTaskCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			repo := setupScrapeTestRepository(t)
			ctrl := gomock.NewController(t)
			parser := meta.NewMockParser(ctrl)
			if tt.expectParserCall > 0 {
				parser.EXPECT().
					GetEpisodeDetails(gomock.Any(), 95231, tt.nfo.season, tt.nfo.episode).
					Return(tt.details, nil).
					Times(tt.expectParserCall)
			}

			scraper := &Scraper{
				config:     Config{Enable: true},
				repo:       repo,
				metaParser: parser,
				client:     newImageOKHTTPClient(),
			}

			tempDir := t.TempDir()
			filePath := filepath.Join(tempDir, "episode.mkv")
			nfoPath := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".nfo"

			if tt.createMediaFile {
				err := os.WriteFile(filePath, []byte("media"), 0644)
				require.NoError(t, err)
			}

			if tt.createNFOFile {
				posterPath := tt.nfo.posterPath
				if posterPath == "poster.jpg" {
					posterPath = filepath.Join(tempDir, "poster.jpg")
				}

				writeTestNFO(
					t,
					nfoPath,
					tt.nfo.title,
					tt.nfo.plot,
					tt.nfo.season,
					tt.nfo.episode,
					posterPath,
					tt.nfo.artMode,
				)
			}

			err := repo.Add(ctx, MetadataCheckTask{
				TMDBID:      95231,
				FilePath:    filePath,
				BangumiName: "测试番剧",
				PosterURL:   "https://image.test/poster.jpg",
				Season:      1,
				Episode:     1,
				Statuses:    tt.initialStatuses,
			})
			require.NoError(t, err)

			tasks, err := repo.List(ctx)
			require.NoError(t, err)
			require.Len(t, tasks, 1)

			require.NoError(t, scraper.processTask(ctx, tasks[0]))

			tasks, err = repo.List(ctx)
			require.NoError(t, err)
			require.Len(t, tasks, tt.expectTaskCount)
			if tt.expectTaskCount == 1 {
				assert.Equal(t, tt.expectStatuses, tasks[0].Statuses)
			}
			if tt.expectGeneratedPoster {
				nfoData, err := scraper.parseNFO(nfoPath)
				require.NoError(t, err)

				expectedPosterPath := strings.TrimSuffix(nfoPath, filepath.Ext(nfoPath)) + "-thumb" + posterExtFromStillPath(tt.details.StillPath)
				assert.Equal(t, expectedPosterPath, nfoData.posterPath)

				imageData, err := os.ReadFile(expectedPosterPath)
				require.NoError(t, err)
				assert.Equal(t, "fake-image", string(imageData))

				nfoContent, err := os.ReadFile(nfoPath)
				require.NoError(t, err)
				assert.Contains(t, string(nfoContent), fmt.Sprintf("  <art>\n    <poster>%s</poster>\n  </art>", expectedPosterPath))
			}
		})
	}
}
