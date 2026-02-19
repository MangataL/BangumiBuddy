package magnet

import (
	"context"
	"testing"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/pkg/bangumifile"
	"github.com/MangataL/BangumiBuddy/pkg/bangumifile/anito"
	"github.com/golang/mock/gomock"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestManager_UpdateTask(t *testing.T) {
	testCases := []struct {
		name    string
		fake    func(t *testing.T) Dependency
		req     UpdateTaskReq
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "未选择任何文件",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-1").Return(Task{
					TaskID:       "task-1",
					DownloadType: downloader.DownloadTypeTV,
					Meta:         Meta{TMDBID: 1},
					Torrent:      Torrent{Hash: "hash-1"},
				}, nil)

				top := downloader.NewMockTorrentOperator(ctrl)
				top.EXPECT().Get(gomock.Any(), "hash-1").Return(downloader.Torrent{
					Status: downloader.TorrentStatusDownloading,
				}, nil)

				return Dependency{
					Repository: repo,
					TorrentOp:  top,
				}
			},
			req: UpdateTaskReq{
				TaskID: "task-1",
				Torrent: Torrent{
					Hash: "hash-1",
					Files: []TorrentFile{
						{FileName: "a.mkv", Download: false},
						{FileName: "b.mkv", Download: false},
					},
				},
				TMDBID: 1,
			},
			wantErr: assert.Error,
		},
		{
			name: "设置文件选择并继续下载",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-2").Return(Task{
					TaskID:       "task-2",
					DownloadType: downloader.DownloadTypeTV,
					Meta:         Meta{TMDBID: 1},
					Torrent:      Torrent{Hash: "hash-2"},
				}, nil)
				repo.EXPECT().SaveTask(gomock.Any(), gomock.Any()).Return(nil)

				top := downloader.NewMockTorrentOperator(ctrl)
				top.EXPECT().Get(gomock.Any(), "hash-2").Return(downloader.Torrent{
					Status: downloader.TorrentStatusDownloading,
				}, nil)

				dl := downloader.NewMockInterface(ctrl)
				dl.EXPECT().SetTorrentFilePriorities(gomock.Any(), "hash-2", gomock.Any()).
					Do(func(ctx context.Context, hash string, files []downloader.TorrentFileSelection) {
						assert.Len(t, files, 2)
					}).
					Return(nil)
				dl.EXPECT().GetDownloadStatuses(gomock.Any(), []string{"hash-2"}).Return([]downloader.DownloadStatus{
					{Size: 123},
				}, nil).AnyTimes()
				dl.EXPECT().ContinueDownload(gomock.Any(), "hash-2").Return(nil)

				return Dependency{
					Repository: repo,
					Downloader: dl,
					TorrentOp:  top,
				}
			},
			req: func() UpdateTaskReq {
				cont := true
				return UpdateTaskReq{
					TaskID: "task-2",
					TMDBID: 1,
					Torrent: Torrent{
						Hash: "hash-2",
						Files: []TorrentFile{
							{FileName: "a.mkv", Download: true},
							{FileName: "b.mkv", Download: false},
						},
					},
					ContinueDownload: &cont,
				}
			}(),
			wantErr: assert.NoError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dep := tc.fake(t)
			m := &Manager{
				repository: dep.Repository,
				downloader: dep.Downloader,
				torrentOp:  dep.TorrentOp,
				metaParser: dep.MetaParser,
			}
			err := m.UpdateTask(context.Background(), tc.req)
			tc.wantErr(t, err)
		})
	}
}

func TestManager_getSubtitleFiles(t *testing.T) {
	m := &Manager{}
	path := "testdata/[VCB-Studio] test"
	subtitleFiles, err := m.getSubtitleFiles(path)
	if err != nil {
		t.Fatalf("getSubtitleFiles failed: %v", err)
	}
	want := []string{
		"testdata/[VCB-Studio] test/a.ass",
	}
	assert.Equal(t, want, subtitleFiles)
}

func TestManager_PreviewAddSubtitles(t *testing.T) {
	testCases := []struct {
		name    string
		fake    func(t *testing.T) Dependency
		req     PreviewAddSubtitlesReq
		want    PreviewAddSubtitlesResp
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "剧场版-匹配成功",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-1").Return(Task{
					TaskID:       "task-1",
					Status:       TaskStatusInitSuccess,
					DownloadType: downloader.DownloadTypeMovie,
					Torrent: Torrent{
						Hash: "movie-hash",
						Files: []TorrentFile{
							{FileName: "movie.mkv", Media: true},
						},
					},
				}, nil)
				top := downloader.NewMockTorrentOperator(ctrl)
				top.EXPECT().Get(gomock.Any(), "movie-hash").Return(downloader.Torrent{
					Path: "/downloads",
				}, nil)

				return Dependency{
					Repository: repo,
					TorrentOp:  top,
				}
			},
			req: PreviewAddSubtitlesReq{
				TaskID:       "task-1",
				SubtitlePath: "testdata/[VCB-Studio] test/a.ass",
				DstPath:      "",
			},
			want: PreviewAddSubtitlesResp{
				SubtileFiles: map[string]AddSubtitlesResult{
					"testdata/[VCB-Studio] test/a.ass": {
						SubtitleFile:  "testdata/[VCB-Studio] test/a.ass",
						NewFileName:   "movie.ass",
						TargetPath:    "/downloads/movie.ass",
						MediaFileName: "movie.mkv",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "TV-按集数匹配成功",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-tv").Return(Task{
					TaskID:       "task-tv",
					Status:       TaskStatusInitSuccess,
					DownloadType: downloader.DownloadTypeTV,
					Torrent: Torrent{
						Hash: "tv-hash",
						Files: []TorrentFile{
							{FileName: "S01E01.mkv", Episode: 1, Media: true},
							{FileName: "S01E02.mkv", Episode: 2, Media: true},
						},
					},
				}, nil)
				top := downloader.NewMockTorrentOperator(ctrl)
				top.EXPECT().Get(gomock.Any(), "tv-hash").Return(downloader.Torrent{
					Path: "/downloads/tv",
				}, nil)
				parser := bangumifile.NewMockParser(ctrl)
				// 模拟字幕文件 a.ass 被解析为第 2 集
				parser.EXPECT().Parse(gomock.Any(), "testdata/[VCB-Studio] test/a.ass", gomock.Any(), gomock.Any()).
					Return(bangumifile.BangumiFile{Episode: 2}, nil)

				return Dependency{
					Repository:        repo,
					TorrentOp:         top,
					BangumiFileParser: parser,
				}
			},
			req: PreviewAddSubtitlesReq{
				TaskID:       "task-tv",
				SubtitlePath: "testdata/[VCB-Studio] test/a.ass",
				DstPath:      "",
			},
			want: PreviewAddSubtitlesResp{
				SubtileFiles: map[string]AddSubtitlesResult{
					"testdata/[VCB-Studio] test/a.ass": {
						SubtitleFile:  "testdata/[VCB-Studio] test/a.ass",
						NewFileName:   "S01E02.ass",
						TargetPath:    "/downloads/tv/S01E02.ass",
						MediaFileName: "S01E02.mkv",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "同名前缀匹配-匹配成功",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-prefix").Return(Task{
					TaskID:       "task-prefix",
					Status:       TaskStatusInitSuccess,
					DownloadType: downloader.DownloadTypeTV,
					Torrent: Torrent{
						Hash: "prefix-hash",
						Files: []TorrentFile{
							{FileName: "a.mkv", Episode: 1, Media: true},
						},
					},
				}, nil)
				top := downloader.NewMockTorrentOperator(ctrl)
				top.EXPECT().Get(gomock.Any(), "prefix-hash").Return(downloader.Torrent{
					Path: "/downloads",
				}, nil)

				return Dependency{
					Repository: repo,
					TorrentOp:  top,
				}
			},
			req: PreviewAddSubtitlesReq{
				TaskID:       "task-prefix",
				SubtitlePath: "testdata/[VCB-Studio] test/a.ass",
				DstPath:      "",
			},
			want: PreviewAddSubtitlesResp{
				SubtileFiles: map[string]AddSubtitlesResult{
					"testdata/[VCB-Studio] test/a.ass": {
						SubtitleFile:  "testdata/[VCB-Studio] test/a.ass",
						NewFileName:   "a.ass",
						TargetPath:    "/downloads/a.ass",
						MediaFileName: "a.mkv",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "任务未就绪-返回错误",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-not-ready").Return(Task{
					TaskID: "task-not-ready",
					Status: TaskStatusWaitingForConfirmation,
				}, nil)

				return Dependency{
					Repository: repo,
				}
			},
			req: PreviewAddSubtitlesReq{
				TaskID: "task-not-ready",
			},
			want:    PreviewAddSubtitlesResp{},
			wantErr: assert.Error,
		},
		{
			name: "剧场版-多个媒体文件且未匹配-返回错误",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-movie-multi").Return(Task{
					TaskID:       "task-movie-multi",
					Status:       TaskStatusInitSuccess,
					DownloadType: downloader.DownloadTypeMovie,
					Torrent: Torrent{
						Hash: "multi-movie-hash",
						Files: []TorrentFile{
							{FileName: "movie1.mkv", Media: true},
							{FileName: "movie2.mkv", Media: true},
						},
					},
				}, nil)
				top := downloader.NewMockTorrentOperator(ctrl)
				top.EXPECT().Get(gomock.Any(), "multi-movie-hash").Return(downloader.Torrent{
					Path: "/downloads",
				}, nil)

				return Dependency{
					Repository: repo,
					TorrentOp:  top,
				}
			},
			req: PreviewAddSubtitlesReq{
				TaskID:       "task-movie-multi",
				SubtitlePath: "testdata/[VCB-Studio] test/a.ass",
				DstPath:      ".",
			},
			want: PreviewAddSubtitlesResp{
				SubtileFiles: map[string]AddSubtitlesResult{
					"testdata/[VCB-Studio] test/a.ass": {
						SubtitleFile: "testdata/[VCB-Studio] test/a.ass",
						Error:        "movie 模式目标目录包含 2 个媒体文件，且字幕文件名无法唯一对应媒体文件",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "TV-未找到匹配集数-返回错误",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-tv-no-match").Return(Task{
					TaskID:       "task-tv-no-match",
					Status:       TaskStatusInitSuccess,
					DownloadType: downloader.DownloadTypeTV,
					Torrent: Torrent{
						Hash: "tv-hash",
						Files: []TorrentFile{
							{FileName: "S01E01.mkv", Episode: 1, Media: true},
						},
					},
				}, nil)
				top := downloader.NewMockTorrentOperator(ctrl)
				top.EXPECT().Get(gomock.Any(), "tv-hash").Return(downloader.Torrent{
					Path: "/downloads",
				}, nil)
				parser := bangumifile.NewMockParser(ctrl)
				parser.EXPECT().Parse(gomock.Any(), "testdata/[VCB-Studio] test/a.ass", gomock.Any(), gomock.Any()).
					Return(bangumifile.BangumiFile{Episode: 10}, nil)

				return Dependency{
					Repository:        repo,
					TorrentOp:         top,
					BangumiFileParser: parser,
				}
			},
			req: PreviewAddSubtitlesReq{
				TaskID:       "task-tv-no-match",
				SubtitlePath: "testdata/[VCB-Studio] test/a.ass",
				DstPath:      "",
			},
			want: PreviewAddSubtitlesResp{
				SubtileFiles: map[string]AddSubtitlesResult{
					"testdata/[VCB-Studio] test/a.ass": {
						SubtitleFile: "testdata/[VCB-Studio] test/a.ass",
						Error:        "未找到第 10 集媒体文件",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "多级目录-匹配成功",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-nested").Return(Task{
					TaskID:       "task-nested",
					Status:       TaskStatusInitSuccess,
					DownloadType: downloader.DownloadTypeMovie,
					Torrent: Torrent{
						Hash: "nested-hash",
						Files: []TorrentFile{
							{FileName: "Movies/Action/hero.mkv", Media: true},
						},
					},
				}, nil)
				top := downloader.NewMockTorrentOperator(ctrl)
				top.EXPECT().Get(gomock.Any(), "nested-hash").Return(downloader.Torrent{
					Path: "/data",
				}, nil)

				return Dependency{
					Repository: repo,
					TorrentOp:  top,
				}
			},
			req: PreviewAddSubtitlesReq{
				TaskID:       "task-nested",
				SubtitlePath: "testdata/[VCB-Studio] test/a.ass",
				DstPath:      "Movies/Action",
			},
			want: PreviewAddSubtitlesResp{
				SubtileFiles: map[string]AddSubtitlesResult{
					"testdata/[VCB-Studio] test/a.ass": {
						SubtitleFile:  "testdata/[VCB-Studio] test/a.ass",
						NewFileName:   "hero.ass",
						TargetPath:    "/data/Movies/Action/hero.ass",
						MediaFileName: "hero.mkv",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "剧场版-指定具体文件一对一匹配",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-movie-file").Return(Task{
					TaskID:       "task-movie-file",
					Status:       TaskStatusInitSuccess,
					DownloadType: downloader.DownloadTypeMovie,
					Torrent: Torrent{
						Hash: "movie-file-hash",
						Files: []TorrentFile{
							{FileName: "movie1.mkv", Media: true},
							{FileName: "movie2.mkv", Media: true},
						},
					},
				}, nil)
				top := downloader.NewMockTorrentOperator(ctrl)
				top.EXPECT().Get(gomock.Any(), "movie-file-hash").Return(downloader.Torrent{
					Path: "/downloads",
				}, nil)

				return Dependency{
					Repository: repo,
					TorrentOp:  top,
				}
			},
			req: PreviewAddSubtitlesReq{
				TaskID:       "task-movie-file",
				SubtitlePath: "testdata/[VCB-Studio] test/a.ass",
				DstPath:      "movie2.mkv", // 直接指定文件，而非目录
			},
			want: PreviewAddSubtitlesResp{
				SubtileFiles: map[string]AddSubtitlesResult{
					"testdata/[VCB-Studio] test/a.ass": {
						SubtitleFile:  "testdata/[VCB-Studio] test/a.ass",
						NewFileName:   "movie2.ass",
						TargetPath:    "/downloads/movie2.ass",
						MediaFileName: "movie2.mkv",
					},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "目标目录无媒体文件-返回错误",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-no-media").Return(Task{
					TaskID:       "task-no-media",
					Status:       TaskStatusInitSuccess,
					DownloadType: downloader.DownloadTypeMovie,
					Torrent: Torrent{
						Hash: "no-media-hash",
						Files: []TorrentFile{
							{FileName: "readme.txt", Media: false},
						},
					},
				}, nil)
				top := downloader.NewMockTorrentOperator(ctrl)
				top.EXPECT().Get(gomock.Any(), "no-media-hash").Return(downloader.Torrent{
					Path: "/downloads",
				}, nil)

				return Dependency{
					Repository: repo,
					TorrentOp:  top,
				}
			},
			req: PreviewAddSubtitlesReq{
				TaskID:       "task-no-media",
				SubtitlePath: "testdata/[VCB-Studio] test/a.ass",
				DstPath:      "",
			},
			want:    PreviewAddSubtitlesResp{},
			wantErr: assert.Error,
		},
		{
			name: "TV-目标目录下没有季度匹配的媒体文件-返回错误",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-tv-season-mismatch").Return(Task{
					TaskID:       "task-tv-season-mismatch",
					Status:       TaskStatusInitSuccess,
					DownloadType: downloader.DownloadTypeTV,
					Torrent: Torrent{
						Hash: "tv-hash",
						Files: []TorrentFile{
							{FileName: "S01E01.mkv", Season: 1, Episode: 1, Media: true},
						},
					},
				}, nil)
				top := downloader.NewMockTorrentOperator(ctrl)
				top.EXPECT().Get(gomock.Any(), "tv-hash").Return(downloader.Torrent{
					Path: "/downloads",
				}, nil)

				return Dependency{
					Repository: repo,
					TorrentOp:  top,
				}
			},
			req: PreviewAddSubtitlesReq{
				TaskID:       "task-tv-season-mismatch",
				SubtitlePath: "testdata/[VCB-Studio] test/a.ass",
				DstPath:      "",
				Season:       lo.ToPtr(2),
			},
			want:    PreviewAddSubtitlesResp{},
			wantErr: assert.Error,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := New(tc.fake(t))

			resp, err := m.PreviewAddSubtitles(context.Background(), tc.req)

			tc.wantErr(t, err)
			assert.Equal(t, tc.want, resp)
		})
	}
}

func TestManager_getAllExtensions(t *testing.T) {
	testCases := []struct {
		name           string
		filePath       string
		extensionLevel *int
		want           string
	}{
		{
			name:           "普通一层后缀",
			filePath:       "test1.ass",
			extensionLevel: nil,
			want:           ".ass",
		},
		{
			name:           "普通两层后缀",
			filePath:       "test1.default.chs.ass",
			extensionLevel: nil,
			want:           ".default.chs.ass",
		},
		{
			name:           "普通双语后缀",
			filePath:       "test1.chs&jpn.ass",
			extensionLevel: nil,
			want:           ".chs&jpn.ass",
		},
		{
			name:           "文件名本身包含.不加层级时获取错误",
			filePath:       "a.b.ass",
			extensionLevel: nil,
			want:           ".b.ass",
		},
		{
			name:           "文件名本身包含.加层级时获取正确",
			filePath:       "a.b.ass",
			extensionLevel: lo.ToPtr(1),
			want:           ".ass",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &Manager{}

			got := m.getAllExtensions(tc.filePath, tc.extensionLevel)

			assert.Equal(t, tc.want, got)
		})
	}
}

func TestManager_FindTaskSimilarFiles(t *testing.T) {
	testCases := []struct {
		name    string
		fake    func(t *testing.T) Dependency
		taskID  string
		file    string
		want    []string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "同层级同标题匹配",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-similar").Return(Task{
					TaskID: "task-similar",
					Torrent: Torrent{
						Files: []TorrentFile{
							{FileName: "A/Show S01E01.mkv", Media: true},
							{FileName: "A/Show S01E02.mkv", Media: true},
							{FileName: "A/Show S01E03.mkv", Media: false},
							{FileName: "A/Other S01E01.mkv", Media: true},
							{FileName: "A/Show/Show S01E03.mkv", Media: true},
							{FileName: "B/Show S01E01.mkv", Media: true},
						},
					},
				}, nil)
				parser := anito.NewParser()
				return Dependency{
					Repository:        repo,
					BangumiFileParser: parser,
				}
			},
			taskID: "task-similar",
			file:   "A/Show S01E01.mkv",
			want: []string{
				"A/Show S01E01.mkv",
				"A/Show S01E02.mkv",
			},
			wantErr: assert.NoError,
		},
		{
			name: "只有自身文件匹配",
			fake: func(t *testing.T) Dependency {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)
				repo := NewMockRepository(ctrl)
				repo.EXPECT().GetTask(gomock.Any(), "task-similar").Return(Task{
					TaskID: "task-similar",
					Torrent: Torrent{
						Files: []TorrentFile{
							{FileName: "A/Show S01E01.mkv", Media: true},
						},
					},
				}, nil)
				parser := anito.NewParser()
				return Dependency{
					Repository:        repo,
					BangumiFileParser: parser,
				}
			},
			taskID: "task-similar",
			file:   "A/Show S01E01.mkv",
			want: []string{
				"A/Show S01E01.mkv",
			},
			wantErr: assert.NoError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := New(tc.fake(t))

			got, err := m.FindTaskSimilarFiles(context.Background(), tc.taskID, tc.file)

			tc.wantErr(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
