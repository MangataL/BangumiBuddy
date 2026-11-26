package subscriber

import (
	"context"
	"errors"
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

func rssItem(title, date string) RSSItem {
	published, _ := time.Parse(time.DateOnly, date)
	return RSSItem{Title: title, PublishedAt: published}
}

func weeklyAirDates(first string, count int) []string {
	date, _ := time.Parse(time.DateOnly, first)
	dates := make([]string, count)
	for i := range dates {
		dates[i] = date.AddDate(0, 0, 7*i).Format(time.DateOnly)
	}
	return dates
}

func TestParseRSSCallOfTheNight(t *testing.T) {
	const batch = "[星空字幕组&LoliHouse] 彻夜之歌 / Yofukashi no Uta [01-13] [WebRip 1080p]"
	const single = "[LoliHouse] 彻夜之歌 第二季 / Yofukashi no Uta S2 - 01 [WebRip 1080p]"
	cases := []struct {
		name       string
		tmdbID     int
		items      []RSSItem
		wantSeason int
		wantDate   string
	}{
		{"第一季仅有完结后发布的合集", 0, []RSSItem{rssItem(batch, "2022-12-01")}, 1, "2022-12-01"},
		{"第二季第一条是合集，中间缺日期，最早的有效单集排在最后", callOfTheNight.TMDBID, []RSSItem{
			rssItem(batch, "2025-09-20"), rssItem(single, ""),
			rssItem(single, "2025-07-19"), rssItem(single, "2025-07-05"),
		}, 2, "2025-07-05"},
		{"两季皆已播出的合集", 0, []RSSItem{rssItem(batch, "2025-09-20")}, 0, ""},
		{"单集补档明显晚于所有播出日期", 0, []RSSItem{rssItem(single, "2025-12-01")}, 0, ""},
		{"单集缺少日期时不得改用合集日期", 0, []RSSItem{rssItem(batch, "2022-12-01"), rssItem(single, "")}, 0, ""},
		{"单集标题写第二季但日期属于第一季", 0, []RSSItem{rssItem(single, "2022-07-09")}, 1, "2022-07-09"},
		{"晚发单集不能回退到可推断的合集", 0, []RSSItem{rssItem(batch, "2022-12-01"), rssItem(single, "2022-12-02")}, 0, ""},
		{"无法识别条目类型", 0, []RSSItem{rssItem("[Group] Yofukashi no Uta [1080p]", "2022-07-09")}, 0, ""},
		{"尚未开播的合集", 0, []RSSItem{rssItem(batch, "2022-07-07")}, 0, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			rss := NewMockRSSParser(ctrl)
			parser := meta.NewMockParser(ctrl)
			ctx := context.Background()
			rss.EXPECT().Parse(ctx, "rss").Return(RSS{BangumiName: "彻夜之歌", Items: tc.items}, nil)
			if tc.tmdbID == 0 {
				parser.EXPECT().SearchTV(ctx, "彻夜之歌").Return(callOfTheNight, nil)
			} else {
				parser.EXPECT().ParseTV(ctx, tc.tmdbID).Return(callOfTheNight, nil)
			}
			for _, season := range callOfTheNight.Seasons {
				parser.EXPECT().GetSeasonAirDates(ctx, callOfTheNight.TMDBID, season.Number).
					Return(weeklyAirDates(season.AirDate, season.EpisodeTotalNum), nil).MaxTimes(1)
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

func TestParseRSSEpisodeSchedules(t *testing.T) {
	seasons := []meta.Season{
		{Number: 1, AirDate: "2023-01-05", EpisodeTotalNum: 2, PosterURL: "s1"},
		{Number: 2, AirDate: "2023-03-03", EpisodeTotalNum: 2, PosterURL: "s2"},
	}
	cases := []struct {
		name       string
		firstDates []string
		nextDates  []string
		nextErr    error
		want       int
	}{
		{"第一季分割放送，下半篇晚于第二季首播", []string{"2023-01-05", "2023-10-05"}, []string{"2023-03-03", "2023-03-10"}, nil, 1},
		{"两季都有同时期单集", []string{"2023-01-05", "2023-10-05"}, []string{"2023-03-03", "2023-10-02"}, nil, 0},
		{"另一季播出日期不全", []string{"2023-01-05", "2023-10-05"}, []string{"2023-03-03", ""}, nil, 0},
		{"另一季详情请求失败，仍然可以手选", []string{"2023-01-05", "2023-10-05"}, nil, errors.New("unavailable"), 0},
		{"落在两段放送之间", []string{"2023-01-05", "2023-12-15"}, []string{"2023-03-03", "2023-03-10"}, nil, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			rss := NewMockRSSParser(ctrl)
			parser := meta.NewMockParser(ctrl)
			ctx := context.Background()
			rss.EXPECT().Parse(ctx, "rss").Return(RSS{Items: []RSSItem{rssItem("[Group] Some Anime S2 - 01", "2023-10-06")}}, nil)
			parser.EXPECT().ParseTV(ctx, 1).Return(meta.Meta{TMDBID: 1, Seasons: seasons}, nil)
			parser.EXPECT().GetSeasonAirDates(ctx, 1, 1).Return(tc.firstDates, nil)
			parser.EXPECT().GetSeasonAirDates(ctx, 1, 2).Return(tc.nextDates, tc.nextErr)
			s := Subscriber{rssParser: rss, metaParser: parser}
			got, err := s.ParseRSS(ctx, ParserRSSReq{RSSLink: "rss", TMDBID: 1})
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.Season)
			assert.Equal(t, seasons, got.Seasons)
			if tc.want == 0 {
				assert.Empty(t, got.RecommendationDate)
			}
		})
	}
}

func TestParseRSSWithoutSeasonRequests(t *testing.T) {
	var manySeasons []meta.Season
	for i := 1; i <= 30; i++ {
		manySeasons = append(manySeasons, meta.Season{Number: i, AirDate: "2020-01-01"})
	}
	cases := []struct {
		name    string
		seasons []meta.Season
		items   []RSSItem
		want    int
	}{
		{"无正季", nil, nil, 0},
		{"仅一个正季可以默认", callOfTheNight.Seasons[:1], nil, 1},
		{"候选过多时跳过查询", manySeasons, []RSSItem{rssItem("[Group] Anime - 01", "2022-01-01")}, 0},
		{"某季首播日未知的合集", []meta.Season{{Number: 1, AirDate: "2022-01-01"}, {Number: 2}}, []RSSItem{rssItem("Anime [01-13]", "2022-06-01")}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			rss := NewMockRSSParser(ctrl)
			parser := meta.NewMockParser(ctrl)
			ctx := context.Background()
			rss.EXPECT().Parse(ctx, "rss").Return(RSS{Items: tc.items}, nil)
			parser.EXPECT().ParseTV(ctx, 1).Return(meta.Meta{TMDBID: 1, Seasons: tc.seasons}, nil)
			s := Subscriber{rssParser: rss, metaParser: parser}
			got, err := s.ParseRSS(ctx, ParserRSSReq{RSSLink: "rss", TMDBID: 1})
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.Season)
			assert.Empty(t, got.RecommendationDate)
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
