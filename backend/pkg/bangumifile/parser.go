package bangumifile

import "context"

//go:generate mockgen -destination parser_mock.go -source $GOFILE -package $GOPACKAGE

type Parser interface {
	Parse(ctx context.Context, fileName string, opts ...ParserOption) (BangumiFile, error)
}

type ParserOption func(options *ParseOptions)

type ParseOptions struct {
	IgnoreValidateEpisode bool
	EpisodeLocation       string
	EpisodeOffset         int
	PreserveOriginName    bool
}

func IgnoreValidateEpisode() ParserOption {
	return func(options *ParseOptions) {
		options.IgnoreValidateEpisode = true
	}
}

func WithEpisodeLocation(location string) ParserOption {
	return func(options *ParseOptions) {
		options.EpisodeLocation = location
	}
}

func WithEpisodeOffset(offset int) ParserOption {
	return func(options *ParseOptions) {
		options.EpisodeOffset = offset
	}
}

func PreserveOriginName() ParserOption {
	return func(options *ParseOptions) {
		options.PreserveOriginName = true
	}
}