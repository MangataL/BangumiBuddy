package anito

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/nssteinbrenner/anitogo"

	"github.com/MangataL/BangumiBuddy/pkg/bangumifile"
	"github.com/MangataL/BangumiBuddy/pkg/utils"
)

type parser struct{}

func NewParser() bangumifile.Parser {
	return &parser{}
}


func (p *parser) Parse(ctx context.Context, fileName string, opts ...bangumifile.ParserOption) (bangumifile.BangumiFile, error) {
	options := &bangumifile.ParseOptions{}
	for _, opt := range opts {
		opt(options)
	}
	if !options.PreserveOriginName {
		fileName = filepath.Base(fileName)
	}

	anitogoMeta := anitogo.Parse(normalizeFilenameForAnitogo(fileName), anitogo.DefaultOptions)

	// 解析季数
	season := 1 // 默认第一季
	if len(anitogoMeta.AnimeSeason) > 0 {
		if s, err := strconv.Atoi(anitogoMeta.AnimeSeason[0]); err == nil {
			season = s
		}
	}

	var (
		episode int
		err     error
	)
	if options.EpisodeLocation != "" {
		episode, err = p.parseEpisodeWithLocation(fileName, options.EpisodeLocation)
	} else {
		episode, err = p.parseEpisodeFromAnitogoMeta(anitogoMeta)
	}

	if !options.IgnoreValidateEpisode && err != nil {
		return bangumifile.BangumiFile{}, err
	}

	episode += options.EpisodeOffset

	return bangumifile.BangumiFile{
		Season:       season,
		Episode:      episode,
		AnimeTitle:   anitogoMeta.AnimeTitle,
		ReleaseGroup: anitogoMeta.ReleaseGroup,
	}, nil
}

var (
	reEpisodeWithPrefixHua = regexp.MustCompile(`第(\d{1,4})话`)
	reEpisodeNoPrefixHua   = regexp.MustCompile(`(\d{1,4})话`)
	reEpisodeWithPrefixJi  = regexp.MustCompile(`第(\d{1,4})集`)
	reEpisodeNoPrefixJi    = regexp.MustCompile(`(\d{1,4})集`)
)

// normalizeFilenameForAnitogo makes simplified Chinese episode counters parsable by anitogo.
// It rewrites "话/集" (simplified) into the traditional "話" that the library recognizes,
// and strips the optional "第" prefix so the upstream regex can match.
func normalizeFilenameForAnitogo(name string) string {
	n := reEpisodeWithPrefixHua.ReplaceAllString(name, "${1}話")
	n = reEpisodeNoPrefixHua.ReplaceAllString(n, "${1}話")
	n = reEpisodeWithPrefixJi.ReplaceAllString(n, "${1}話")
	n = reEpisodeNoPrefixJi.ReplaceAllString(n, "${1}話")
	return n
}

func (p *parser) parseEpisodeFromAnitogoMeta(anitogoMeta *anitogo.Elements) (int, error) {
	if len(anitogoMeta.EpisodeNumber) == 0 {
		return 0, errors.New("无法识别集数信息")
	}
	episode, err := strconv.Atoi(anitogoMeta.EpisodeNumber[0])
	if err != nil {
		return 0, errors.New("不是有效的集数信息")
	}
	return episode, nil
}

func (p *parser) parseEpisodeWithLocation(name string, location string) (int, error) {
	pattern := regexp.QuoteMeta(location)
	pattern = strings.ReplaceAll(pattern, `\{ep\}`, `(\d+|[一二三四五六七八九十百千]+)`)

	reg := regexp.MustCompile(pattern)
	matches := reg.FindStringSubmatch(name)
	if len(matches) < 2 {
		return 0, errors.New("无法从文件名中解析出集数信息")
	}

	epStr := matches[1]

	// 如果是数字直接转换
	if num, err := strconv.Atoi(epStr); err == nil {
		return num, nil
	}

	ep, err := utils.ChineseNumberToInt(epStr)
	if err != nil {
		return 0, fmt.Errorf("无效的集数信息(%s): %w", epStr, err)
	}

	return ep, nil
}
