package tmdb

import (
	"context"
	"errors"
	"strings"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"

	"github.com/MangataL/BangumiBuddy/internal/meta"
	"github.com/MangataL/BangumiBuddy/pkg/errs"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

var _ meta.Parser = (*Client)(nil)

type Config struct {
	TMDBToken    string `mapstructure:"tmdb_token" json:"tmdbToken"`
	AlternateURL bool   `mapstructure:"alternate_url" json:"alternateURL"`
}

var ErrTMDBTokenNotSet = errors.New("请先设置TMDB Token")

func NewParser(config Config) *Client {
	return &Client{
		client: newTMDBClient(config),
	}
}

func newTMDBClient(config Config) *tmdb.Client {
	c, err := tmdb.InitV4(config.TMDBToken)
	if err != nil {
		return nil
	}
	c.SetClientAutoRetry()
	if config.AlternateURL {
		c.SetAlternateBaseURL()
	}
	return c
}

type Client struct {
	client *tmdb.Client
}

func (t *Client) Parse(ctx context.Context, id int) (meta.Meta, error) {
	if t.client == nil {
		return meta.Meta{}, ErrTMDBTokenNotSet
	}
	tv, err := t.client.GetTVDetails(id, map[string]string{
		"language": "zh",
	})
	if err != nil {
		return meta.Meta{}, err
	}
	log.Debugf(ctx, "获取tmdb元数据: %+v", tv)
	season := getSeason(tv)
	return meta.Meta{
		ChineseName:     tv.Name,
		Year:            getYear(ctx, tv.FirstAirDate),
		TMDBID:          int(tv.ID),
		Season:          season,
		EpisodeTotalNum: getSeasonEpisodeTotalNum(tv, season),
		AirWeekday:      getAirWeekday(tv),
		PosterURL:       getImageURL(tv),
		BackdropURL:     getBackdropURL(tv),
		Overview:        tv.Overview,
		Genres:          getGenres(tv),
	}, nil
}

func getSeason(tv *tmdb.TVDetails) int {
	if tv.NextEpisodeToAir.ID == 0 && tv.NextEpisodeToAir.AirDate == "" {
		return tv.LastEpisodeToAir.SeasonNumber
	}
	return tv.NextEpisodeToAir.SeasonNumber
}

func getSeasonEpisodeTotalNum(tv *tmdb.TVDetails, season int) int {
	for _, episode := range tv.Seasons {
		if episode.SeasonNumber == season {
			return episode.EpisodeCount
		}
	}
	return 0
}

func getAirWeekday(tv *tmdb.TVDetails) *time.Weekday {
	airDay := tv.NextEpisodeToAir.AirDate
	if tv.NextEpisodeToAir.ID == 0 && tv.NextEpisodeToAir.AirDate == "" {
		airDay = tv.LastEpisodeToAir.AirDate
	}
	airDate, err := time.Parse(time.DateOnly, airDay)
	if err != nil {
		return nil
	}
	weekday := airDate.Weekday()
	return &weekday
}

const (
	imageBaseURL = "https://image.tmdb.org/t/p/w500"
)

func getImageURL(tv *tmdb.TVDetails) string {
	if tv.PosterPath == "" {
		return ""
	}
	return imageBaseURL + tv.PosterPath
}

func getBackdropURL(tv *tmdb.TVDetails) string {
	if tv.BackdropPath == "" {
		return ""
	}
	return imageBaseURL + tv.BackdropPath
}

const (
	animeGenreID = 16
)

func getGenres(tv *tmdb.TVDetails) string {
	genres := make([]string, 0, len(tv.Genres))
	for _, genre := range tv.Genres {
		if genre.ID == animeGenreID {
			continue
		}
		genres = append(genres, genre.Name)
	}
	return strings.Join(genres, ", ")
}

// TODO: 后期有必要可以改为多次分页去查找满足条件的，当前从简只查一页
func (t *Client) Search(ctx context.Context, name string) (meta.Meta, error) {
	if t.client == nil {
		return meta.Meta{}, ErrTMDBTokenNotSet
	}
	tvs, err := t.client.GetSearchTVShow(name, map[string]string{
		"language": "zh",
		"page":     "1",
	})
	if err != nil {
		return meta.Meta{}, err
	}
	log.Debugf(ctx, "search %s got tvs: %+v", name, tvs)
	if len(tvs.Results) == 0 {
		return meta.Meta{}, errs.NewNotFoundf("未搜索到番剧，解析出的番剧名称: %s", name)
	}
	id := tvs.Results[0].ID
	for _, tv := range tvs.Results {
		for _, genre := range tv.GenreIDs {
			if genre == animeGenreID {
				id = tv.ID
				break
			}
		}
	}
	return t.Parse(ctx, int(id))
}

func getYear(ctx context.Context, date string) string {
	if len(date) < 4 {
		log.Warnf(ctx, "invalid date: %s", date)
		return ""
	}
	return date[:4]
}

func (c *Client) Reload(config interface{}) error {
	cfg, ok := config.(*Config)
	if !ok {
		return errors.New("配置类型错误")
	}
	c.client = newTMDBClient(*cfg)
	return nil
}
