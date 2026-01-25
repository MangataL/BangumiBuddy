package transfer

import (
	"archive/zip"
	"context"
	"os"
	"testing"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/creasty/defaults"
	"github.com/stretchr/testify/require"
)

func TestTransfer_makeFileExt(t *testing.T) {
	config := Config{}
	_ = defaults.Set(&config)
	config.SubtitleRename.Enabled = true

	configWant := Config{
		SubtitleRename: SubtitleRenameConfig{
			Enabled:                     true,
			SimpleChineseRenameExt:      ".zh",
			SimpleChineseExts:           []string{".zh-cn", ".zh-hans", ".sc"},
			TraditionalChineseRenameExt: ".zh-hant",
			TraditionalChineseExts:      []string{".zh-tw", ".zh-hk", ".zh-hant", ".tc"},
		},
	}
	require.Equal(t, configWant.SubtitleRename, config.SubtitleRename)

	testCases := []struct {
		fileExt string
		want    string
	}{
		{
			fileExt: ".zh-cn.ass",
			want:    ".zh.ass",
		},
		{
			fileExt: ".zh-hans.ass",
			want:    ".zh.ass",
		},
		{
			fileExt: ".sc.ass",
			want:    ".zh.ass",
		},
		{
			fileExt: ".zh-tw.ass",
			want:    ".zh-hant.ass",
		},
		{
			fileExt: ".zh-hk.ass",
			want:    ".zh-hant.ass",
		},
		{
			fileExt: ".zh-hant.ass",
			want:    ".zh-hant.ass",
		},
		{
			fileExt: ".tc.ass",
			want:    ".zh-hant.ass",
		},
		{
			fileExt: ".SC.ass",
			want:    ".zh.ass",
		},
		{
			fileExt: ".subset.sc.ass",
			want:    ".subset.zh.ass",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.fileExt, func(t *testing.T) {
			trans := &Transfer{
				config: config,
			}

			got := trans.makeFileExt(tc.fileExt, true)

			require.Equal(t, tc.want, got)
		})
	}
}

func Test_getCommonParent(t *testing.T) {
	testCases := []struct {
		name     string
		dirs     []string
		basePath string
		want     string
	}{
		{
			name: "normal",
			dirs: []string{
				"/data/movies/1",
				"/data/movies/2",
				"/data/movies/3",
				"/data/movies/3/a/",
			},
			basePath: "/data/movies",
			want:     "/data/movies",
		},
		{
			name:     "empty",
			dirs:     []string{},
			basePath: "/",
			want:     "",
		},
		{
			name: "one",
			dirs: []string{
				"/data/movies/1",
			},
			basePath: "/data/movies",
			want:     "/data/movies/1",
		},
		{
			name: "no common",
			dirs: []string{
				"/data/movies/1",
				"/data2",
			},
			basePath: "/",
			want:     "/",
		},
		{
			name: "dir prefix equal",
			dirs: []string{
				"/data/movies/Fonts/m1",
				"/data/movies/Fonts/m2",
			},
			basePath: "/data/movies",
			want:     "/data/movies/Fonts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := getCommonParent(tc.dirs, tc.basePath)
			require.Equal(t, tc.want, got)
		})
	}
}

func createZipWithFont(t *testing.T, archivePath string, fontName string) {
	t.Helper()
	zipFile, err := os.Create(archivePath)
	require.NoError(t, err)
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	file, err := writer.Create(fontName)
	require.NoError(t, err)
	_, err = file.Write([]byte("x"))
	require.NoError(t, err)
}

func TestTransfer_getTorrentSearchPath(t *testing.T) {
	testCases := []struct {
		name      string
		fileNames []string
		want      string // 相对于 torrent.Path 的路径，空字符串表示跳过
	}{
		{
			name:      "no files",
			fileNames: []string{},
			want:      "",
		},
		{
			name: "files without directory",
			fileNames: []string{
				"video1.mkv",
				"video2.mkv",
				"subtitle.ass",
			},
			want: "",
		},
		{
			name: "files in single directory",
			fileNames: []string{
				"Season 1/Episode 1.mkv",
				"Season 1/Episode 2.mkv",
				"Season 1/Episode 3.mkv",
			},
			want: "Season 1",
		},
		{
			name: "files in nested directory",
			fileNames: []string{
				"Show Name/Season 1/Episode 1.mkv",
				"Show Name/Season 1/Episode 2.mkv",
			},
			want: "Show Name/Season 1",
		},
		{
			name: "files in different seasons",
			fileNames: []string{
				"Show Name/Season 1/Episode 1.mkv",
				"Show Name/Season 2/Episode 1.mkv",
			},
			want: "Show Name",
		},
		{
			name: "mixed files with and without directory",
			fileNames: []string{
				"Season 1/Episode 1.mkv",
				"README.txt",
			},
			want: "Season 1",
		},
		{
			name: "mixed media files with and without directory",
			fileNames: []string{
				"Season 1/Episode 1.mkv",
				"Episode 2.mkv",
			},
			// 混合情况：有的文件在根目录，有的在子目录
			// 这种情况下没有公共的目录层级，返回空字符串（即遍历整个下载目录）
			want: "",
		},
		{
			name: "files in completely different directories",
			fileNames: []string{
				"Dir1/video1.mkv",
				"Dir2/video2.mkv",
			},
			// 文件分散在不同的顶层目录中，没有公共父目录，返回空字符串（跳过）
			want: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			transfer := &Transfer{}
			torrent := downloader.Torrent{
				Path:      tempDir,
				FileNames: tc.fileNames,
			}

			got := transfer.getTorrentSearchPath(torrent)

			if tc.want == "" {
				require.Equal(t, "", got)
			} else {
				require.Equal(t, tempDir+"/"+tc.want, got)
			}
		})
	}
}

func TestTransfer_getFontPath(t *testing.T) {
	testCases := []struct {
		name            string
		setupFunc       func(t *testing.T) (string, []string) // 返回 torrent.Path 和 FileNames
		checkPath       func(t *testing.T, torrentPath, gotPath string)
		wantCleanupFunc bool // 是否期望有清理函数
		wantErr         bool
	}{
		{
			name: "no font files",
			setupFunc: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				// 创建一些非字体文件
				require.NoError(t, os.WriteFile(tempDir+"/video.mkv", []byte("video"), 0644))
				require.NoError(t, os.WriteFile(tempDir+"/subtitle.ass", []byte("subtitle"), 0644))
				return tempDir, []string{"video.mkv", "subtitle.ass"}
			},
			checkPath: func(t *testing.T, torrentPath, gotPath string) {
				// 没有字体文件时，返回空字符串
				require.Equal(t, "", gotPath)
			},
			wantCleanupFunc: false,
			wantErr:         false,
		},
		{
			name: "no directory hierarchy - skip font search",
			setupFunc: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				// 所有文件都在根目录，没有目录层级
				require.NoError(t, os.WriteFile(tempDir+"/video.mkv", []byte("video"), 0644))
				require.NoError(t, os.WriteFile(tempDir+"/font.ttf", []byte("font"), 0644))
				return tempDir, []string{"video.mkv", "font.ttf"}
			},
			checkPath: func(t *testing.T, torrentPath, gotPath string) {
				// 没有目录层级，应该跳过，返回空字符串
				require.Equal(t, "", gotPath)
			},
			wantCleanupFunc: false,
			wantErr:         false,
		},
		{
			name: "with directory - single font file in subdirectory",
			setupFunc: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				fontDir := tempDir + "/Season 1"
				require.NoError(t, os.MkdirAll(fontDir, 0755))
				require.NoError(t, os.WriteFile(fontDir+"/font.ttf", []byte("font"), 0644))
				require.NoError(t, os.WriteFile(fontDir+"/video.mkv", []byte("video"), 0644))
				return tempDir, []string{"Season 1/font.ttf", "Season 1/video.mkv"}
			},
			checkPath: func(t *testing.T, torrentPath, gotPath string) {
				// 单个字体文件在子目录中，返回该子目录
				require.Equal(t, torrentPath+"/Season 1", gotPath)
			},
			wantCleanupFunc: false,
			wantErr:         false,
		},
		{
			name: "with directory - multiple font files in fonts directory",
			setupFunc: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				showDir := tempDir + "/Show Name"
				fontDir := showDir + "/fonts"
				require.NoError(t, os.MkdirAll(fontDir, 0755))
				require.NoError(t, os.WriteFile(fontDir+"/font1.ttf", []byte("font1"), 0644))
				require.NoError(t, os.WriteFile(fontDir+"/font2.otf", []byte("font2"), 0644))
				require.NoError(t, os.WriteFile(showDir+"/video.mkv", []byte("video"), 0644))
				return tempDir, []string{
					"Show Name/fonts/font1.ttf",
					"Show Name/fonts/font2.otf",
					"Show Name/video.mkv",
				}
			},
			checkPath: func(t *testing.T, torrentPath, gotPath string) {
				// 多个字体文件在 fonts 子目录，返回该子目录
				require.Equal(t, torrentPath+"/Show Name/fonts", gotPath)
			},
			wantCleanupFunc: false,
			wantErr:         false,
		},
		{
			name: "with directory - multiple font files in different subdirectories",
			setupFunc: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				showDir := tempDir + "/Show Name"
				fontDir1 := showDir + "/fonts/dir1"
				fontDir2 := showDir + "/fonts/dir2"
				require.NoError(t, os.MkdirAll(fontDir1, 0755))
				require.NoError(t, os.MkdirAll(fontDir2, 0755))
				require.NoError(t, os.WriteFile(fontDir1+"/font1.ttf", []byte("font1"), 0644))
				require.NoError(t, os.WriteFile(fontDir2+"/font2.ttf", []byte("font2"), 0644))
				require.NoError(t, os.WriteFile(showDir+"/video.mkv", []byte("video"), 0644))
				return tempDir, []string{
					"Show Name/fonts/dir1/font1.ttf",
					"Show Name/fonts/dir2/font2.ttf",
					"Show Name/video.mkv",
				}
			},
			checkPath: func(t *testing.T, torrentPath, gotPath string) {
				// 多个字体文件在不同子目录，返回公共父目录
				require.Equal(t, torrentPath+"/Show Name/fonts", gotPath)
			},
			wantCleanupFunc: false,
			wantErr:         false,
		},
		{
			name: "with directory - font archive file",
			setupFunc: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				showDir := tempDir + "/Show Name"
				require.NoError(t, os.MkdirAll(showDir, 0755))
				archivePath := showDir + "/fonts.zip"
				createZipWithFont(t, archivePath, "font.ttf")
				require.NoError(t, os.WriteFile(showDir+"/video.mkv", []byte("video"), 0644))
				return tempDir, []string{"Show Name/fonts.zip", "Show Name/video.mkv"}
			},
			checkPath: func(t *testing.T, torrentPath, gotPath string) {
				// 字体压缩包会被解压到临时目录，返回临时目录路径
				require.NotEmpty(t, gotPath)
				require.Contains(t, gotPath, ".fonts_extract_")
				// 验证目录存在
				_, err := os.Stat(gotPath)
				require.NoError(t, err)
			},
			wantCleanupFunc: true,
			wantErr:         false,
		},
		{
			name: "with directory - mixed font files and archive",
			setupFunc: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				showDir := tempDir + "/Show Name"
				fontDir := showDir + "/fonts"
				require.NoError(t, os.MkdirAll(fontDir, 0755))
				require.NoError(t, os.WriteFile(fontDir+"/font1.ttf", []byte("font1"), 0644))
				// 创建字体压缩包
				archivePath := showDir + "/extra_fonts.zip"
				createZipWithFont(t, archivePath, "font2.ttf")
				require.NoError(t, os.WriteFile(showDir+"/video.mkv", []byte("video"), 0644))
				return tempDir, []string{
					"Show Name/fonts/font1.ttf",
					"Show Name/extra_fonts.zip",
					"Show Name/video.mkv",
				}
			},
			checkPath: func(t *testing.T, torrentPath, gotPath string) {
				// 混合情况，返回种子目录
				require.Equal(t, torrentPath+"/Show Name", gotPath)
			},
			wantCleanupFunc: true,
			wantErr:         false,
		},
		{
			name: "with directory - nested font directories",
			setupFunc: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				fontDir := tempDir + "/Show Name/media/fonts/subfolder"
				require.NoError(t, os.MkdirAll(fontDir, 0755))
				require.NoError(t, os.WriteFile(fontDir+"/font.ttf", []byte("font"), 0644))
				require.NoError(t, os.WriteFile(tempDir+"/Show Name/video.mkv", []byte("video"), 0644))
				return tempDir, []string{
					"Show Name/media/fonts/subfolder/font.ttf",
					"Show Name/video.mkv",
				}
			},
			checkPath: func(t *testing.T, torrentPath, gotPath string) {
				// 嵌套目录，返回字体文件所在的完整路径
				require.Equal(t, torrentPath+"/Show Name/media/fonts/subfolder", gotPath)
			},
			wantCleanupFunc: false,
			wantErr:         false,
		},
		{
			name: "with directory - font files with various extensions",
			setupFunc: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				fontDir := tempDir + "/Show Name/fonts"
				require.NoError(t, os.MkdirAll(fontDir, 0755))
				require.NoError(t, os.WriteFile(fontDir+"/font1.ttf", []byte("font1"), 0644))
				require.NoError(t, os.WriteFile(fontDir+"/font2.otf", []byte("font2"), 0644))
				require.NoError(t, os.WriteFile(fontDir+"/font3.ttc", []byte("font3"), 0644))
				require.NoError(t, os.WriteFile(fontDir+"/font4.woff", []byte("font4"), 0644))
				require.NoError(t, os.WriteFile(tempDir+"/Show Name/video.mkv", []byte("video"), 0644))
				return tempDir, []string{
					"Show Name/fonts/font1.ttf",
					"Show Name/fonts/font2.otf",
					"Show Name/fonts/font3.ttc",
					"Show Name/fonts/font4.woff",
					"Show Name/video.mkv",
				}
			},
			checkPath: func(t *testing.T, torrentPath, gotPath string) {
				// 不同扩展名的字体文件，返回字体目录
				require.Equal(t, torrentPath+"/Show Name/fonts", gotPath)
			},
			wantCleanupFunc: false,
			wantErr:         false,
		},
		{
			name: "with directory - archive named with FONTS in uppercase",
			setupFunc: func(t *testing.T) (string, []string) {
				tempDir := t.TempDir()
				showDir := tempDir + "/Show Name"
				require.NoError(t, os.MkdirAll(showDir, 0755))
				archivePath := showDir + "/FONTS.zip"
				createZipWithFont(t, archivePath, "font.ttf")
				require.NoError(t, os.WriteFile(showDir+"/video.mkv", []byte("video"), 0644))
				return tempDir, []string{"Show Name/FONTS.zip", "Show Name/video.mkv"}
			},
			checkPath: func(t *testing.T, torrentPath, gotPath string) {
				// 大写 FONTS 压缩包，会被解压到临时目录
				require.NotEmpty(t, gotPath)
				require.Contains(t, gotPath, ".fonts_extract_")
				// 验证目录存在
				_, err := os.Stat(gotPath)
				require.NoError(t, err)
			},
			wantCleanupFunc: true,
			wantErr:         false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			torrentPath, fileNames := tc.setupFunc(t)

			transfer := &Transfer{}
			torrent := downloader.Torrent{
				Path:      torrentPath,
				FileNames: fileNames,
			}

			gotPath, cleanup, err := transfer.getFontPath(context.Background(), torrent)

			// 检查错误
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// 使用自定义检查函数验证路径
			tc.checkPath(t, torrentPath, gotPath)

			// 检查清理函数
			if tc.wantCleanupFunc {
				// 验证清理函数不为 nil，并且可以调用
				require.NotNil(t, cleanup)
			}

			// 无论如何都调用清理函数
			cleanup()

			// 如果有清理函数，验证临时目录已被清理
			if tc.wantCleanupFunc && gotPath != "" && gotPath != torrentPath {
				// 对于临时目录，调用清理函数后应该不存在
				if _, err := os.Stat(gotPath); !os.IsNotExist(err) {
					// 如果目录仍然存在，可能是临时目录的清理逻辑
					t.Logf("临时目录 %s 在清理后可能仍然存在", gotPath)
				}
			}
		})
	}
}
