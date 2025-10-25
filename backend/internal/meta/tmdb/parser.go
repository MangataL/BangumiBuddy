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

// SearchMovie implements meta.Parser.
func (t *Client) SearchMovie(ctx context.Context, name string) (meta.Meta, error) {
	movies, err := t.SearchMovies(ctx, name)
	if err != nil {
		return meta.Meta{}, err
	}
	if len(movies) == 0 {
		return meta.Meta{}, errs.NewNotFoundf("未搜索到剧场版，剧场版名称: %s", name)
	}
	return movies[0], nil
}

func (t *Client) SearchMovies(ctx context.Context, name string) ([]meta.Meta, error) {
	if t.client == nil {
		return nil, ErrTMDBTokenNotSet
	}
	movies, err := t.client.GetSearchMovies(name, map[string]string{
		"language": "zh",
		"page":     "1",
	})
	if err != nil {
		return nil, err
	}
	log.Debugf(ctx, "search %s got movies: %+v", name, movies)
	if len(movies.Results) == 0 {
		return nil, errs.NewNotFoundf("未搜索到剧场版，剧场版名称: %s", name)
	}
	metas := make([]meta.Meta, 0, len(movies.Results))
	for _, movie := range movies.Results {
		m, err := t.ParseMovie(ctx, int(movie.ID))
		if err != nil {
			return nil, err
		}
		metas = append(metas, m)
	}
	return metas, nil
}

func (t *Client) ParseMovie(ctx context.Context, id int) (meta.Meta, error) {
	if t.client == nil {
		return meta.Meta{}, ErrTMDBTokenNotSet
	}
	movie, err := t.client.GetMovieDetails(id, map[string]string{
		"language": "zh",
	})
	if err != nil {
		return meta.Meta{}, err
	}
	return meta.Meta{
		ChineseName: movie.Title,
		Year:        getYear(ctx, movie.ReleaseDate),
		TMDBID:      int(movie.ID),
		PosterURL:   getImageURL(movie.PosterPath),
		BackdropURL: getImageURL(movie.BackdropPath),
		Overview:    movie.Overview,
		Genres:      getGeneres(movie.Genres),
	}, nil
}

func (t *Client) ParseTV(ctx context.Context, id int) (meta.Meta, error) {
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
		PosterURL:       getImageURL(tv.PosterPath),
		BackdropURL:     getImageURL(tv.BackdropPath),
		Overview:        tv.Overview,
		Genres:          getGeneres(tv.Genres),
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

func getImageURL(url string) string {
	if url == "" {
		return ""
	}
	return imageBaseURL + url
}

const (
	animeGenreID = 16
)

func getGeneres(generes []struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}) string {
	genres := make([]string, 0, len(generes))
	for _, genre := range generes {
		if genre.ID == animeGenreID {
			continue
		}
		genres = append(genres, genre.Name)
	}
	return strings.Join(genres, ", ")
}

func (t *Client) SearchTVs(ctx context.Context, name string) ([]meta.Meta, error) {
	if t.client == nil {
		return nil, ErrTMDBTokenNotSet
	}
	tvs, err := t.client.GetSearchTVShow(name, map[string]string{
		"language": "zh",
		"page":     "1",
	})
	if err != nil {
		return nil, err
	}
	log.Debugf(ctx, "search %s got tvs: %+v", name, tvs)
	if len(tvs.Results) == 0 {
		return nil, errs.NewNotFoundf("未搜索到番剧，番剧名称: %s", name)
	}
	metas := make([]meta.Meta, 0, len(tvs.Results))
	for _, tv := range tvs.Results {
		for _, genre := range tv.GenreIDs {
			if genre == animeGenreID {
				m, err := t.ParseTV(ctx, int(tv.ID))
				if err != nil {
					return nil, err
				}
				metas = append(metas, m)
			}
		}
	}
	return metas, nil
}

// TODO: 后期有必要可以改为多次分页去查找满足条件的，当前从简只查一页
func (t *Client) SearchTV(ctx context.Context, name string) (meta.Meta, error) {
	tvs, err := t.SearchTVs(ctx, name)
	if err != nil {
		return meta.Meta{}, err
	}
	if len(tvs) == 0 {
		return meta.Meta{}, errs.NewNotFoundf("未搜索到番剧，番剧名称: %s", name)
	}
	return tvs[0], nil
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

func (t *Client) GetEpisodeDetails(ctx context.Context, tmdbID, season, episode int) (meta.EpisodeDetails, error) {
	if t.client == nil {
		return meta.EpisodeDetails{}, ErrTMDBTokenNotSet
	}
	episodeDetails, err := t.client.GetTVEpisodeDetails(tmdbID, season, episode, map[string]string{
		"language": "zh",
	})
	if err != nil {
		return meta.EpisodeDetails{}, err
	}
	log.Debugf(ctx, "获取单集元数据: %+v", episodeDetails)
	return meta.EpisodeDetails{
		Name:      episodeDetails.Name,
		Overview:  episodeDetails.Overview,
		StillPath: getImageURL(episodeDetails.StillPath),
		AirDate:   episodeDetails.AirDate,
	}, nil
}
