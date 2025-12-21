package anito

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MangataL/BangumiBuddy/pkg/bangumifile"
)

func TestAntiParser_Parse(t *testing.T) {
	testCases := []struct {
		name     string
		fileName string
		opts     []bangumifile.ParserOption
		want     bangumifile.BangumiFile
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:     "目录本身包含集数信息",
			fileName: "[LoliHouse] Food Court de, Mata Ashita. [01-06][WebRip 1080p HEVC-10bit AAC]/[LoliHouse] Food Court de, Mata Ashita. - 02 [WebRip 1080p HEVC-10bit AAC SRTx2].mkv",
			want: bangumifile.BangumiFile{
				Season:       1,
				Episode:      2,
				AnimeTitle:   "Food Court de, Mata Ashita.",
				ReleaseGroup: "LoliHouse",
			},
			wantErr: assert.NoError,
		},
		{
			name:     "基础解析 - 带有季数信息",
			fileName: "[Group] My Anime S02E05 [1080p].mkv",
			want: bangumifile.BangumiFile{
				Season:       2,
				Episode:      5,
				AnimeTitle:   "My Anime",
				ReleaseGroup: "Group",
			},
			wantErr: assert.NoError,
		},
		{
			name:     "自定义集数位置 - 中文数字",
			fileName: "[Group] My Anime.01 第十二话 [1080p].mkv",
			opts:     []bangumifile.ParserOption{bangumifile.WithEpisodeLocation("第{ep}话")},
			want: bangumifile.BangumiFile{
				Season:       1,
				Episode:      12,
				AnimeTitle:   "My Anime",
				ReleaseGroup: "Group",
			},
			wantErr: assert.NoError,
		},
		{
			name:     "简体中文集数默认解析",
			fileName: "[Group] My Anime - 第08话 [1080p].mkv",
			want: bangumifile.BangumiFile{
				Season:       1,
				Episode:      8,
				AnimeTitle:   "My Anime",
				ReleaseGroup: "Group",
			},
			wantErr: assert.NoError,
		},
		{
			name:     "带有集数偏移 (EpisodeOffset)",
			fileName: "[Group] My Anime - 01 [1080p].mkv",
			opts:     []bangumifile.ParserOption{bangumifile.WithEpisodeOffset(12)},
			want: bangumifile.BangumiFile{
				Season:       1,
				Episode:      13,
				AnimeTitle:   "My Anime",
				ReleaseGroup: "Group",
			},
			wantErr: assert.NoError,
		},
		{
			name:     "忽略无法识别的集数 (IgnoreValidateEpisode)",
			fileName: "[Group] My Anime - 特别篇 [1080p].mkv",
			opts:     []bangumifile.ParserOption{bangumifile.IgnoreValidateEpisode()},
			want: bangumifile.BangumiFile{
				Season:       1,
				Episode:      0,
				AnimeTitle:   "My Anime - 特别篇",
				ReleaseGroup: "Group",
			},
			wantErr: assert.NoError,
		},
		{
			name:     "无法识别集数且未忽略 (默认行为)",
			fileName: "[Group] My Anime - 特别篇 [1080p].mkv",
			want:     bangumifile.BangumiFile{},
			wantErr:  assert.Error,
		},
		{
			name:     "自定义位置匹配失败",
			fileName: "[Group] My Anime - 01 [1080p].mkv",
			opts:     []bangumifile.ParserOption{bangumifile.WithEpisodeLocation("第{ep}话")},
			want:     bangumifile.BangumiFile{},
			wantErr:  assert.Error,
		},
		{
			name:     "保留原始文件名",
			fileName: "[LoliHouse] 为美好的世界献上祝福！3 / Kono Subarashii Sekai ni Shukufuku wo! S3 - 10 [WebRip 1080p HEVC-10bit AAC][简繁内封字幕]",
			opts:     []bangumifile.ParserOption{bangumifile.PreserveOriginName()},
			want: bangumifile.BangumiFile{
				Season:       3,
				Episode:      10,
				AnimeTitle:   "为美好的世界献上祝福！3 / Kono Subarashii Sekai ni Shukufuku wo!",
				ReleaseGroup: "LoliHouse",
			},
			wantErr: assert.NoError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewParser()

			got, err := parser.Parse(context.Background(), tc.fileName, tc.opts...)

			tc.wantErr(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestAntiParser_parseEpisodeWithLocation(t *testing.T) {
	testCases := []struct {
		name     string
		fileName string
		location string
		want     int
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name:     "标准数字",
			fileName: "My Anime - 05.mkv",
			location: " - {ep}.",
			want:     5,
			wantErr:  assert.NoError,
		},
		{
			name:     "中文数字 - 单个",
			fileName: "My Anime 第五话.mkv",
			location: "第{ep}话",
			want:     5,
			wantErr:  assert.NoError,
		},
		{
			name:     "中文数字 - 十二",
			fileName: "My Anime 第十二话.mkv",
			location: "第{ep}话",
			want:     12,
			wantErr:  assert.NoError,
		},
		{
			name:     "带括号的定位符",
			fileName: "[Group] My Anime [08].mkv",
			location: "[{ep}]",
			want:     8,
			wantErr:  assert.NoError,
		},
		{
			name:     "多处数字匹配 - 取定位符处",
			fileName: "[2023] My Anime - 12 [1080p].mkv",
			location: " - {ep} [",
			want:     12,
			wantErr:  assert.NoError,
		},
		{
			name:     "匹配失败 - 格式不符",
			fileName: "My Anime - 01.mkv",
			location: "第{ep}话",
			want:     0,
			wantErr:  assert.Error,
		},
		{
			name:     "匹配失败 - 无数字",
			fileName: "My Anime 第话.mkv",
			location: "第{ep}话",
			want:     0,
			wantErr:  assert.Error,
		},
	}

	p := &parser{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := p.parseEpisodeWithLocation(tc.fileName, tc.location)

			tc.wantErr(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
