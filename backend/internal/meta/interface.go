package meta

import "context"

//go:generate mockgen -destination interface_mock.go -source $GOFILE -package $GOPACKAGE

// Parser 番剧元数据解析器
type Parser interface {
	SearchTV(ctx context.Context, name string) (Meta, error)
	SearchTVs(ctx context.Context, name string) ([]Meta, error)
	ParseTV(ctx context.Context, id int) (Meta, error)
	SearchMovie(ctx context.Context, name string) (Meta, error)
	SearchMovies(ctx context.Context, name string) ([]Meta, error)
	ParseMovie(ctx context.Context, id int) (Meta, error)
}
