package tmdb

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/liuzl/gocc"

	"github.com/MangataL/BangumiBuddy/internal/meta"
	"github.com/MangataL/BangumiBuddy/internal/network"
	"github.com/MangataL/BangumiBuddy/pkg/errs"
	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/MangataL/BangumiBuddy/pkg/utils"
)

var _ meta.Parser = (*Client)(nil)

type Config struct {
	TMDBToken    string `mapstructure:"tmdb_token" json:"tmdbToken"`
	AlternateURL bool   `mapstructure:"alternate_url" json:"alternateURL"`
}

var ErrTMDBTokenNotSet = errors.New("请先设置TMDB Token")

func NewParser(config Config, provider network.HTTPClientProvider) *cacheAdapter {
	client := &Client{
		client:  newTMDBClient(config, provider),
		network: provider,
	}
	return newCacheAdapter(client)
}

func newTMDBClient(config Config, provider network.HTTPClientProvider) *tmdb.Client {
	c, err := tmdb.InitV4(config.TMDBToken)
	if err != nil {
		return nil
	}
	if provider != nil {
		// XXX: tmdb客户端只提供了set方法，这里不想每次获取client时都走一次set方法，就只在初始化时set一次
		// 		实际上，持有的是HTTP Transport是指针，可以动态加载
		c.SetClientConfig(*provider.HTTPClient(30 * time.Second))
	}
	c.SetClientAutoRetry()
	if config.AlternateURL {
		c.SetAlternateBaseURL()
	}
	return c
}

type Client struct {
	mu      sync.RWMutex
	client  *tmdb.Client
	network network.HTTPClientProvider
}


func (t *Client) currentClient() *tmdb.Client {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.client
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
	client := t.currentClient()
	if client == nil {
		return nil, ErrTMDBTokenNotSet
	}
	movies, err := client.GetSearchMovies(name, map[string]string{
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
	animeMetas := make([]meta.Meta, 0)
	otherMetas := make([]meta.Meta, 0)
	for _, movie := range movies.Results {
		hasAnimeGenre := false
		for _, genreID := range movie.GenreIDs {
			if genreID == animeGenreID {
				hasAnimeGenre = true
				break
			}
		}
		m, err := t.ParseMovie(ctx, int(movie.ID))
		if err != nil {
			return nil, err
		}
		if hasAnimeGenre {
			animeMetas = append(animeMetas, m)
		} else {
			otherMetas = append(otherMetas, m)
		}
	}
	return append(animeMetas, otherMetas...), nil
}

func (t *Client) ParseMovie(ctx context.Context, id int) (meta.Meta, error) {
	client := t.currentClient()
	if client == nil {
		return meta.Meta{}, ErrTMDBTokenNotSet
	}
	movie, err := client.GetMovieDetails(id, map[string]string{
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
	client := t.currentClient()
	if client == nil {
		return meta.Meta{}, ErrTMDBTokenNotSet
	}
	tv, err := client.GetTVDetails(id, map[string]string{
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

func (t *Client) GetSeasonEpisodeTotalNum(
	ctx context.Context,
	tmdbID, season int,
	opts ...meta.MetaOption,
) (int, error) {
	client := t.currentClient()
	if client == nil {
		return 0, ErrTMDBTokenNotSet
	}
	tv, err := client.GetTVDetails(tmdbID, map[string]string{
		"language": "zh",
	})
	if err != nil {
		return 0, err
	}
	log.Debugf(ctx, "获取tmdb季总集数: tmdbID=%d season=%d tv=%+v", tmdbID, season, tv)
	return getSeasonEpisodeTotalNum(tv, season), nil
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
	client := t.currentClient()
	if client == nil {
		return nil, ErrTMDBTokenNotSet
	}
	tvs, err := client.GetSearchTVShow(name, map[string]string{
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
	animeMetas := make([]meta.Meta, 0)
	otherMetas := make([]meta.Meta, 0)
	for _, tv := range tvs.Results {
		hasAnimeGenre := false
		for _, genreID := range tv.GenreIDs {
			if genreID == animeGenreID {
				hasAnimeGenre = true
				break
			}
		}
		m, err := t.ParseTV(ctx, int(tv.ID))
		if err != nil {
			return nil, err
		}
		if hasAnimeGenre {
			animeMetas = append(animeMetas, m)
		} else {
			otherMetas = append(otherMetas, m)
		}
	}
	return append(animeMetas, otherMetas...), nil
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
	c.mu.Lock()
	c.client = newTMDBClient(*cfg, c.network)
	c.mu.Unlock()
	return nil
}

func (t *Client) GetEpisodeDetails(ctx context.Context, tmdbID, season, episode int) (meta.EpisodeDetails, error) {
	var result meta.EpisodeDetails
	var err error

	// 尝试按语言优先级获取数据
	for _, lang := range languagePriorities {
		log.Debugf(ctx, "尝试获取单集元数据 (language=%s)", lang.Language)
		details, err := t.getEpisodeDetailsWithLanguage(ctx, tmdbID, season, episode, lang)
		if err != nil {
			continue
		}

		// 如果是第一个语言，直接赋值
		if result.Empty() {
			result = details
		} else {
			// 合并无效字段
			result.Merge(details)
		}

		// 如果所有字段都有效，提前返回
		if result.AllValid() {
			break
		}
	}

	// 如果没有获取到任何数据，返回最后一个错误
	if result.Empty() {
		return meta.EpisodeDetails{}, err
	}

	return result, nil
}

// languagePriority 语言优先级结构
type languagePriority struct {
	Language      string
	IsTraditional bool
}

// languagePriorities 语言优先级列表
var languagePriorities = []languagePriority{
	{Language: "zh", IsTraditional: false},
	{Language: "zh-SG", IsTraditional: false},
	{Language: "zh-HK", IsTraditional: true},
	{Language: "zh-TW", IsTraditional: true},
}

// getEpisodeDetailsWithLanguage 根据指定语言获取单集元数据
func (t *Client) getEpisodeDetailsWithLanguage(ctx context.Context, tmdbID, season, episode int, language languagePriority) (meta.EpisodeDetails, error) {
	client := t.currentClient()
	if client == nil {
		return meta.EpisodeDetails{}, ErrTMDBTokenNotSet
	}
	episodeDetails, err := client.GetTVEpisodeDetails(tmdbID, season, episode, map[string]string{
		"language": language.Language,
	})
	if err != nil {
		return meta.EpisodeDetails{}, err
	}
	log.Debugf(ctx, "获取单集元数据 (language=%s): %+v", language.Language, episodeDetails)

	// 获取原始数据
	name := episodeDetails.Name
	overview := episodeDetails.Overview

	// 如果是繁体语言，先进行繁转简
	if language.IsTraditional {
		simplifiedName, err := traditionalToSimplified(name)
		if err != nil {
			return meta.EpisodeDetails{}, fmt.Errorf("繁体转简体失败 (language=%s): %w", language.Language, err)
		}
		name = simplifiedName
		simplifiedOverview, err := traditionalToSimplified(overview)
		if err != nil {
			return meta.EpisodeDetails{}, fmt.Errorf("繁体转简体失败 (language=%s): %w", language.Language, err)
		}
		overview = simplifiedOverview
	}

	// 如果名称无效（"第 x 集"格式），返回空字符串
	if utils.EpisodeNameInvalid(name) {
		name = ""
	}

	return meta.EpisodeDetails{
		Name:      name,
		Overview:  overview,
		StillPath: getImageURL(episodeDetails.StillPath),
		AirDate:   episodeDetails.AirDate,
	}, nil
}

// traditionalToSimplified 繁体转简体
var t2sConverter, _ = gocc.New("t2s")

func traditionalToSimplified(text string) (string, error) {
	if text == "" {
		return "", nil
	}
	if t2sConverter == nil {
		return text, errors.New("繁体转简体转换器初始化失败")
	}
	return t2sConverter.Convert(text)
}
