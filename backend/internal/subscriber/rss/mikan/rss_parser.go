package mikan

import (
	"context"
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
	"github.com/nssteinbrenner/anitogo"
	"github.com/pkg/errors"

	"github.com/MangataL/BangumiBuddy/internal/subscriber"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

func NewParser() subscriber.RSSParser {
	return &parser{
		fp: gofeed.NewParser(),
	}
}

type parser struct {
	fp *gofeed.Parser
}

func (p *parser) Parse(ctx context.Context, link string) (subscriber.RSS, error) {
	feed, err := p.fp.ParseURLWithContext(link, ctx)
	if err != nil {
		return subscriber.RSS{}, errors.WithMessage(err, "解析RSS失败")
	}
	items := getItems(ctx, feed.Items)
	rg := parseReleaseGroup(ctx, items)
	return subscriber.RSS{
		BangumiName:  getBangumiName(ctx, feed.Title),
		ReleaseGroup: rg,
		Items:        items,
	}, nil
}

func parseReleaseGroup(ctx context.Context, items []subscriber.RSSItem) string {
	if len(items) == 0 {
		return ""
	}
	meta := anitogo.Parse(items[0].GUID, anitogo.DefaultOptions)
	return meta.ReleaseGroup
}

var (
	mikanTitleRegex = regexp.MustCompile(`Mikan Project - (.+?)(?:\s*第|$)`)
)

func getBangumiName(ctx context.Context, title string) string {
	if matches := mikanTitleRegex.FindStringSubmatch(title); len(matches) > 1 {
		return matches[1]
	}

	log.Warnf(ctx, "无法解析番剧名称: %s", title)
	// 兜底措施
	const mikanPrefix = "Mikan Project - "
	if strings.HasPrefix(title, mikanPrefix) {
		return title[len(mikanPrefix):]
	}
	return title
}

func getItems(ctx context.Context, items []*gofeed.Item) []subscriber.RSSItem {
	rssItems := make([]subscriber.RSSItem, 0, len(items))
	for _, item := range items {
		if len(item.Enclosures) == 0 {
			log.Warnf(ctx, "item %s has no enclosures", item.GUID)
			continue
		}
		rssItems = append(rssItems, subscriber.RSSItem{
			GUID:        item.GUID,
			TorrentLink: item.Enclosures[0].URL,
		})
	}
	return rssItems
}
