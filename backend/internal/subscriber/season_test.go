package subscriber

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MangataL/BangumiBuddy/internal/meta"
)

var callOfTheNight = meta.Meta{
	ChineseName: "彻夜之歌", Year: "2022", TMDBID: 1,
	Season: 2, EpisodeTotalNum: 12, PosterURL: "show.jpg",
	Seasons: []meta.Season{
		{Number: 1, AirDate: "2022-07-08", EpisodeTotalNum: 13, PosterURL: "s1.jpg"},
		{Number: 2, AirDate: "2025-07-04", EpisodeTotalNum: 12, PosterURL: "s2.jpg"},
	},
}

func TestParseRSSCallOfTheNight(t *testing.T) {
	cases := []struct {
		name       string
		tmdbID     int
		dates      []string
		wantSeason int
		wantDate   string
	}{
		{"第一季完结后的合集", 0, []string{"2022-12-01"}, 1, "2022-12-01"},
		{"第二季首播当天", 1, []string{"2025-07-04"}, 2, "2025-07-04"},
		{"第二季开播三周后才更新", 1, []string{"2025-07-25"}, 2, "2025-07-25"},
		{"两季均已完结的补档默认最近开播的季", 0, []string{"2025-12-01"}, 2, "2025-12-01"},
		{"按最早的有效日期选季", 1, []string{"2025-07-25", "", "2022-07-09"}, 1, "2022-07-09"},
		{"所有正季都尚未开播", 0, []string{"2022-07-07"}, 0, ""},
		{"日期缺失或在未来", 0, []string{"", time.Now().AddDate(0, 0, 2).Format(time.DateOnly)}, 0, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var items []RSSItem
			for _, date := range tc.dates {
				var item RSSItem
				if date != "" {
					var err error
					item.PublishedAt, err = time.Parse(time.DateOnly, date)
					require.NoError(t, err)
				}
				items = append(items, item)
			}
			ctx := t.Context()
			ctrl := gomock.NewController(t)
			rss := NewMockRSSParser(ctrl)
			parser := meta.NewMockParser(ctrl)
			rss.EXPECT().Parse(ctx, "rss").Return(RSS{BangumiName: "彻夜之歌", Items: items}, nil)
			if tc.tmdbID == 0 {
				parser.EXPECT().SearchTV(ctx, "彻夜之歌").Return(callOfTheNight, nil)
			} else {
				parser.EXPECT().ParseTV(ctx, tc.tmdbID).Return(callOfTheNight, nil)
			}
			s := Subscriber{rssParser: rss, metaParser: parser}
			got, err := s.ParseRSS(ctx, ParserRSSReq{RSSLink: "rss", TMDBID: tc.tmdbID})
			require.NoError(t, err)
			assert.Equal(t, tc.wantSeason, got.Season)
			assert.Equal(t, tc.wantDate, got.RecommendationDate)
			assert.Equal(t, "2022", got.Year)
			assert.Equal(t, callOfTheNight.Seasons, got.Seasons)
			assert.Equal(t, []int{0, 13, 12}[tc.wantSeason], got.EpisodeTotalNum)
			assert.Equal(t, []string{"show.jpg", "s1.jpg", "s2.jpg"}[tc.wantSeason], got.PosterURL)
		})
	}
}

func TestRecommendSeasonByPremiereDate(t *testing.T) {
	published, err := time.Parse(time.RFC3339, "2025-07-04T00:12:00+08:00")
	require.NoError(t, err)
	items := []RSSItem{{PublishedAt: published}}
	older := meta.Season{Number: 1, AirDate: "2022-07-08"}
	latest := meta.Season{Number: 2, AirDate: "2025-07-04"}
	sameDay := meta.Season{Number: 3, AirDate: latest.AirDate}
	for _, tc := range []struct {
		name    string
		seasons []meta.Season
		items   []RSSItem
		want    int
	}{
		{"忽略特别篇并按日期比较乱序的季度", []meta.Season{latest, {Number: 0}, older}, items, 2},
		{"最新首播日并列", []meta.Season{latest, older, sameDay}, items, 0},
		{"较早的两季同日开播不影响最近的季", []meta.Season{older, {Number: 3, AirDate: older.AirDate}, latest}, items, 2},
		{"正季首播日未知", []meta.Season{older, {Number: 2}}, items, 0},
		{"没有正季", nil, items, 0},
		{"只有一季但没有有效发布日期", []meta.Season{older}, nil, 0},
		{"只有一季但还没开播", []meta.Season{{Number: 1, AirDate: "2025-07-05"}}, items, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, date := recommendSeason(tc.seasons, tc.items)
			require.Equal(t, tc.want, got.Number)
			if tc.want == 0 {
				require.Empty(t, date)
			} else {
				require.Equal(t, "2025-07-04", date)
			}
		})
	}
}

func TestSubscribeSavesSelectedSeason(t *testing.T) {
	for _, tc := range []struct {
		season int
		poster string
	}{{1, "s1.jpg"}, {2, "s2.jpg"}} {
		ctrl := gomock.NewController(t)
		ctx := context.Background()
		parser := meta.NewMockParser(ctrl)
		repo := NewMockRepository(ctrl)
		s := Subscriber{metaParser: parser, repo: repo}
		req := SubscribeReq{RSSLink: "rss", TMDBID: callOfTheNight.TMDBID, Season: tc.season, EpisodeTotalNum: 10, ReleaseGroup: "Group"}
		parser.EXPECT().ParseTV(ctx, req.TMDBID).Return(callOfTheNight, nil)
		repo.EXPECT().Get(ctx, req.RSSLink).Return(Bangumi{}, ErrSubscriberNotFound)
		var saved Bangumi
		repo.EXPECT().Save(ctx, gomock.Any()).Do(func(_ context.Context, b Bangumi) { saved = b })
		got, err := s.Subscribe(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, saved, got)
		assert.Equal(t, tc.season, saved.Season)
		assert.Equal(t, tc.poster, saved.PosterURL)
		assert.Equal(t, 10, saved.EpisodeTotalNum)
		assert.Equal(t, "2022", saved.Year)
	}
}
