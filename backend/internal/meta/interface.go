package meta

import "context"

//go:generate mockgen -destination interface_mock.go -source $GOFILE -package $GOPACKAGE

// Parser 番剧元数据解析器
type Parser interface {
	Search(ctx context.Context, name string) (Meta, error)
	Parse(ctx context.Context, id int) (Meta, error)
}
