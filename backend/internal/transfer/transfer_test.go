package transfer

import (
	"testing"

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
