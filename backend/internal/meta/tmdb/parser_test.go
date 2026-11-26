package tmdb

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MangataL/BangumiBuddy/internal/meta"
)

func testTMDBClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	c, err := tmdb.Init("test")
	require.NoError(t, err)
	c.SetClientConfig(http.Client{Transport: &CustomRoundTripper{
		Transport: server.Client().Transport,
		NewURL:    strings.TrimPrefix(server.URL, "https://"),
	}})
	return &Client{client: c}
}

func TestClientTVSeasons(t *testing.T) {
	const tv = `{"id":1,"name":"彻夜之歌","first_air_date":"2022-07-08","poster_path":"/show.jpg",
		"seasons":[{"season_number":0,"episode_count":1},{"season_number":1,"air_date":"2022-07-08","episode_count":13},
		{"season_number":2,"air_date":"2025-07-04","episode_count":12,"poster_path":"/s2.jpg"}],
		"last_episode_to_air":{"season_number":2}}`
	var paths []string
	client := testTMDBClient(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		switch r.URL.Path {
		case "/3/search/tv":
			fmt.Fprint(w, `{"results":[{"id":2,"genre_ids":[18]},{"id":1,"genre_ids":[16]}]}`)
		case "/3/tv/1":
			fmt.Fprint(w, tv)
		case "/3/tv/2":
			fmt.Fprint(w, `{"id":2,"name":"Another show","first_air_date":"2021-01-01","seasons":[]}`)
		case "/3/tv/1/season/2":
			fmt.Fprint(w, `{"episodes":[{"air_date":"2025-07-04"},{"air_date":"2025-07-11"},{"air_date":null}]}`)
		default:
			http.NotFound(w, r)
		}
	})

	expected := meta.Meta{
		ChineseName: "彻夜之歌", Year: "2022", TMDBID: 1, Season: 2, EpisodeTotalNum: 12, PosterURL: imageBaseURL + "/show.jpg",
		Seasons: []meta.Season{
			{Number: 1, AirDate: "2022-07-08", EpisodeTotalNum: 13, PosterURL: imageBaseURL + "/show.jpg"},
			{Number: 2, AirDate: "2025-07-04", EpisodeTotalNum: 12, PosterURL: imageBaseURL + "/s2.jpg"},
		},
	}
	ctx := context.Background()
	got, err := client.SearchTV(ctx, "彻夜之歌")
	require.NoError(t, err)
	assert.Equal(t, expected, got)
	assert.Equal(t, []string{"/3/search/tv", "/3/tv/1"}, paths)

	got, err = client.ParseTV(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
	gotAll, err := client.SearchTVs(ctx, "彻夜之歌")
	require.NoError(t, err)
	require.Len(t, gotAll, 2)
	assert.Equal(t, expected, gotAll[0])
	assert.Equal(t, []meta.Season{}, gotAll[1].Seasons)

	paths = nil
	dates, err := client.GetSeasonAirDates(ctx, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, []string{"2025-07-04", "2025-07-11", ""}, dates)
	assert.Equal(t, []string{"/3/tv/1/season/2"}, paths)

	total, err := client.GetSeasonEpisodeTotalNum(ctx, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 12, total)
}

func Test_cacheAdapter_GetSeasonEpisodeTotalNum(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name      string
		opts      []meta.MetaOption
		fake      func(t *testing.T) *meta.MockParser
		wantFirst int
		wantNext  int
	}{
		{
			name: "when call option ttl is positive then caches season total",
			opts: []meta.MetaOption{
				meta.WithCacheTTL(time.Hour),
			},
			fake: func(t *testing.T) *meta.MockParser {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				parser := meta.NewMockParser(ctrl)
				parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2, gomock.Any()).Return(13, nil).Times(1)
				return parser
			},
			wantFirst: 13,
			wantNext:  13,
		},
		{
			name: "when call option is absent then does not cache",
			fake: func(t *testing.T) *meta.MockParser {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				parser := meta.NewMockParser(ctrl)
				gomock.InOrder(
					parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2).Return(13, nil),
					parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2).Return(14, nil),
				)
				return parser
			},
			wantFirst: 13,
			wantNext:  14,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := newCacheAdapter(reloadableTestParser{Parser: tc.fake(t)})

			gotFirst, err := adapter.GetSeasonEpisodeTotalNum(ctx, 95231, 2, tc.opts...)
			assert.NoError(t, err)
			gotNext, err := adapter.GetSeasonEpisodeTotalNum(ctx, 95231, 2, tc.opts...)
			assert.NoError(t, err)

			assert.Equal(t, tc.wantFirst, gotFirst)
			assert.Equal(t, tc.wantNext, gotNext)
		})
	}
}

type reloadableTestParser struct {
	meta.Parser
}

func (reloadableTestParser) Reload(config interface{}) error {
	return nil
}

// CustomRoundTripper 是一个自定义的 RoundTripper，它将所有请求转发到另一个地址
type CustomRoundTripper struct {
	Transport http.RoundTripper
	NewURL    string
}

func (c *CustomRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Host = c.NewURL
	return c.Transport.RoundTrip(req)
}
