package meta

import (
	"context"
	"time"
)

//go:generate mockgen -destination interface_mock.go -source $GOFILE -package $GOPACKAGE

// Parser 番剧元数据解析器
type Parser interface {
	SearchTV(ctx context.Context, name string) (Meta, error)
	SearchTVs(ctx context.Context, name string) ([]Meta, error)
	ParseTV(ctx context.Context, id int) (Meta, error)
	SearchMovie(ctx context.Context, name string) (Meta, error)
	SearchMovies(ctx context.Context, name string) ([]Meta, error)
	ParseMovie(ctx context.Context, id int) (Meta, error)
	GetSeasonEpisodeTotalNum(ctx context.Context, tmdbID, season int, opts ...MetaOption) (int, error)
	GetEpisodeDetails(ctx context.Context, tmdbID, season, episode int) (EpisodeDetails, error)
}

type Options struct {
	CacheTTL time.Duration
}

type MetaOption func(*Options)

func WithCacheTTL(ttl time.Duration) MetaOption {
	return func(o *Options) {
		o.CacheTTL = ttl
	}
}

func NewOptions(opts ...MetaOption) Options {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}
	return options
}
