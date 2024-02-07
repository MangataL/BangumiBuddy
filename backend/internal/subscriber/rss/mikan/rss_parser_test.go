package mikan

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

func TestParser_Parse(t *testing.T) {
	testCases := []struct {
		name    string
		stub    func(t *testing.T) (url string, clo func())
		wantErr bool
		want    subscriber.RSS
	}{
		{
			name: "standard season 1",
			stub: stubParse("gbc"),
			want: subscriber.RSS{
				BangumiName:  "GIRLS BAND CRY",
				ReleaseGroup: "喵萌Production&LoliHouse",
				Items: []subscriber.RSSItem{
					{
						GUID:        "[喵萌Production&LoliHouse] GIRLS BAND CRY - 02 [WebRip 1080p HEVC-10bit AAC][简繁日内封字幕]",
						TorrentLink: "https://mikanime.tv/Download/20240414/3a2e456a689ead23ca8f49fdc74ba1872c6f0c12.torrent",
					},
					{
						GUID:        "[喵萌Production&LoliHouse] GIRLS BAND CRY - 01 [WebRip 1080p HEVC-10bit AAC][简繁日内封字幕]",
						TorrentLink: "https://mikanime.tv/Download/20240407/b13d145d95d9acdd5fc50784a6906007b540b468.torrent",
					},
				},
			},
		},
		{
			name: "not season 1, and season not standard",
			stub: stubParse("euphonium3"),
			want: subscriber.RSS{
				BangumiName:  "吹响！悠风号",
				ReleaseGroup: "千夏字幕组&LoliHouse",
				Items: []subscriber.RSSItem{
					{
						GUID:        "[千夏字幕组&LoliHouse] 吹响吧！上低音号 3 / Hibike! Euphonium 3 - 09 [WebRip 1080p HEVC-10bit AAC][简繁内封字幕]",
						TorrentLink: "https://mikanani.me/Download/20240614/0486ae4aafa5fe9406e61e9289b7d81f874fc7fa.torrent",
					},
				},
			},
		},
		{
			name:    "standard season 2",
			stub:    stubParse("kono3"),
			wantErr: false,
			want: subscriber.RSS{
				BangumiName:  "为美好的世界献上祝福！",
				ReleaseGroup: "LoliHouse",
				Items: []subscriber.RSSItem{
					{
						GUID:        "[LoliHouse] 为美好的世界献上祝福！3 / Kono Subarashii Sekai ni Shukufuku wo! S3 - 10 [WebRip 1080p HEVC-10bit AAC][简繁内封字幕]",
						TorrentLink: "https://mikanani.me/Download/20240613/305bcdb1dc367d1684d8350188beab851c8d75e1.torrent",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			url, clo := tc.stub(t)
			defer clo()
			p := NewParser()

			rss, err := p.Parse(context.Background(), url)
			t.Log(err)

			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, tc.want, rss)
		})
	}
}

func stubParse(name string) func(t *testing.T) (url string, clo func()) {
	return func(t *testing.T) (url string, clo func()) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write(getRSSContent(t, name))
		}))
		return ts.URL, ts.Close
	}
}

func getRSSContent(t *testing.T, name string) []byte {
	data, err := os.ReadFile(fmt.Sprintf("./testdata/%s.xml", name))
	require.Nil(t, err)
	return data
}

func TestParseReleaseGroup(t *testing.T) {
	items := []subscriber.RSSItem{
		{
			GUID:        "[ANi] 小市民系列 第二季 - 11 [1080P][Baha][WEB-DL][AAC AVC][CHT][MP4]",
			TorrentLink: "https://mikanani.me/Download/20240414/3a2e456a689ead23ca8f49fdc74ba1872c6f0c12.torrent",
		},
	}

	rg := parseReleaseGroup(context.Background(), items)
	assert.Equal(t, "ANi", rg)
}