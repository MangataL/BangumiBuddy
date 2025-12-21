package web

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestWeb_ListDirs(t *testing.T) {
	tempDir := t.TempDir()

	// 创建测试目录结构
	// tempDir/
	//   empty_dir/
	//   dir_with_sub/
	//     sub_dir/
	//   dir_with_subs/
	//     test.ass
	//     test.srt
	//   dir_with_both/
	//     sub_dir2/
	//     test2.ass

	// 1. 空目录
	os.Mkdir(filepath.Join(tempDir, "empty_dir"), 0755)

	// 2. 只有子目录
	dirWithSub := filepath.Join(tempDir, "dir_with_sub")
	os.Mkdir(dirWithSub, 0755)
	os.Mkdir(filepath.Join(dirWithSub, "sub_dir"), 0755)

	// 3. 只有字幕文件
	dirWithSubs := filepath.Join(tempDir, "dir_with_subs")
	os.Mkdir(dirWithSubs, 0755)
	os.WriteFile(filepath.Join(dirWithSubs, "test.ass"), []byte(""), 0644)
	os.WriteFile(filepath.Join(dirWithSubs, "test.srt"), []byte(""), 0644)

	// 4. 既有子目录又有字幕
	dirWithBoth := filepath.Join(tempDir, "dir_with_both")
	os.Mkdir(dirWithBoth, 0755)
	os.Mkdir(filepath.Join(dirWithBoth, "sub_dir2"), 0755)
	os.WriteFile(filepath.Join(dirWithBoth, "test2.ass"), []byte(""), 0644)

	testCases := []struct {
		name    string
		fake    func(t *testing.T) Dependency
		path    string
		want    func(w *Web, base string) ListDirsResp
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "空路径-报错",
			fake: func(t *testing.T) Dependency {
				return Dependency{}
			},
			path: "",
			want: func(w *Web, base string) ListDirsResp {
				return ListDirsResp{}
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err) && assert.Contains(t, err.Error(), "有效")
			},
		},
		{
			name: "路径不存在-报错",
			fake: func(t *testing.T) Dependency {
				return Dependency{}
			},
			path: filepath.Join(tempDir, "non_existent"),
			want: func(w *Web, base string) ListDirsResp {
				return ListDirsResp{}
			},
			wantErr: assert.Error,
		},
		{
			name: "正常列出目录",
			fake: func(t *testing.T) Dependency {
				return Dependency{}
			},
			path: tempDir,
			want: func(w *Web, base string) ListDirsResp {
				return ListDirsResp{
					Dirs: []DirInfo{
						{
							Path:          filepath.Join(base, "dir_with_both"),
							Name:          "dir_with_both",
							HasDir:        true,
							SubtitleCount: 1,
						},
						{
							Path:          filepath.Join(base, "dir_with_sub"),
							Name:          "dir_with_sub",
							HasDir:        true,
							SubtitleCount: 0,
						},
						{
							Path:          filepath.Join(base, "dir_with_subs"),
							Name:          "dir_with_subs",
							HasDir:        false,
							SubtitleCount: 2,
						},
						{
							Path:          filepath.Join(base, "empty_dir"),
							Name:          "empty_dir",
							HasDir:        false,
							SubtitleCount: 0,
						},
					},
					FilePathSplit: string(filepath.Separator),
					FileRoots:     w.getFileRoots(),
					Files:         []FileInfo{},
				}
			},
			wantErr: assert.NoError,
		},
		{
			name: "列出包含字幕的目录",
			fake: func(t *testing.T) Dependency {
				return Dependency{}
			},
			path: dirWithSubs,
			want: func(w *Web, base string) ListDirsResp {
				return ListDirsResp{
					Dirs:          []DirInfo{},
					FilePathSplit: string(filepath.Separator),
					FileRoots:     w.getFileRoots(),
					Files: []FileInfo{
						{
							Path: filepath.Join(dirWithSubs, "test.ass"),
							Name: "test.ass",
							Size: 0,
						},
						{
							Path: filepath.Join(dirWithSubs, "test.srt"),
							Name: "test.srt",
							Size: 0,
						},
					},
				}
			},
			wantErr: assert.NoError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			w := New(tc.fake(t)).(*Web)

			resp, err := w.ListDirs(context.Background(), tc.path)

			tc.wantErr(t, err)
			if err == nil {
				expected := tc.want(w, tempDir)
				assert.Equal(t, expected.FilePathSplit, resp.FilePathSplit)
				assert.Equal(t, expected.FileRoots, resp.FileRoots)
				// 目录顺序可能受系统影响，使用 ElementsMatch 比较
				assert.ElementsMatch(t, expected.Dirs, resp.Dirs)
				assert.ElementsMatch(t, expected.Files, resp.Files)
			}
		})
	}
}
