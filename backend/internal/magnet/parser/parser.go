package parser

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nssteinbrenner/anitogo"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/magnet"
	"github.com/MangataL/BangumiBuddy/internal/meta"
)

func NewParser(metaParser meta.Parser) magnet.MetaParser {
	return &Parser{
		metaParser: metaParser,
	}
}

// Parser MetaParser接口实现
type Parser struct {
	metaParser meta.Parser
}

// ParseMetaByID implements magnet.MetaParser.
func (p *Parser) ParseMetaByID(ctx context.Context, tmdbID int, downloadType downloader.DownloadType) (meta.Meta, error) {
	var (
		tmdbMeta meta.Meta
		err      error
	)
	if downloadType == downloader.DownloadTypeTV {
		tmdbMeta, err = p.metaParser.ParseTV(ctx, tmdbID)
	} else {
		tmdbMeta, err = p.metaParser.ParseMovie(ctx, tmdbID)
	}
	if err != nil {
		return meta.Meta{}, fmt.Errorf("解析TMDB元数据失败: %w", err)
	}
	return tmdbMeta, nil
}

// ParseReleaseGroup implements magnet.MetaParser.
func (p *Parser) ParseReleaseGroup(ctx context.Context, torrentName string) (string, error) {
	anitogoMeta := anitogo.Parse(torrentName, anitogo.DefaultOptions)
	return anitogoMeta.ReleaseGroup, nil
}

func (p *Parser) ParseMetaByTorrent(ctx context.Context, torrentName string, downloadType downloader.DownloadType) (meta.Meta, error) {
	// 解析种子名称，提取番剧名称
	anitogoMeta := anitogo.Parse(torrentName, anitogo.DefaultOptions)
	name := anitogoMeta.AnimeTitle
	var (
		tmdbMeta meta.Meta
		err      error
	)
	if name != "" {
		if downloadType == downloader.DownloadTypeTV {
			tmdbMeta, err = p.metaParser.SearchTV(ctx, name)
		} else {
			tmdbMeta, err = p.metaParser.SearchMovie(ctx, name)
		}
	} else {
		return meta.Meta{}, fmt.Errorf("无法从种子名称中解析出番剧名称")
	}

	if err != nil {
		return meta.Meta{}, fmt.Errorf("搜索TMDB失败: %w", err)
	}
	return tmdbMeta, nil
}

func (p *Parser) ParseFile(ctx context.Context, fileName string) (season, episode int, err error) {
	anitogoMeta := anitogo.Parse(fileName, anitogo.DefaultOptions)

	// 解析季数
	season = 1 // 默认第一季
	if len(anitogoMeta.AnimeSeasonPrefix) > 0 {
		if s, err := strconv.Atoi(anitogoMeta.AnimeSeasonPrefix[0]); err == nil {
			season = s
		}
	}

	// 解析集数
	if len(anitogoMeta.EpisodeNumber) == 0 {
		return 0, 0, fmt.Errorf("无法识别集数信息")
	}
	episode, err = strconv.Atoi(anitogoMeta.EpisodeNumber[0])
	if err != nil {
		return 0, 0, fmt.Errorf("不是有效的集数信息")
	}

	return season, episode, nil
}
