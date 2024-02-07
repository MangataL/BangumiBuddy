package mikan

import (
	"context"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
	"time"

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
	rg := parseReleaseGroup(items)
	return subscriber.RSS{
		BangumiName:  getBangumiName(ctx, feed.Title),
		ReleaseGroup: rg,
		Items:        items,
	}, nil
}

func parseReleaseGroup(items []subscriber.RSSItem) string {
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
			PublishedAt: parsePublishedAt(ctx, item),
		})
	}
	return rssItems
}

func parsePublishedAt(ctx context.Context, item *gofeed.Item) time.Time {
	if torrent, ok := item.Custom["torrent"]; ok {
		wrappedData := fmt.Sprintf("<torrent>%s</torrent>", strings.TrimSpace(torrent))
		var pubDate struct {
			PubDate string `xml:"pubDate"`
		}
		if err := xml.Unmarshal([]byte(wrappedData), &pubDate); err != nil {
			log.Warnf(ctx, "解析发布时间失败: %v", err)
			return time.Time{}
		}
		t, err := time.ParseInLocation("2006-01-02T15:04:05.999", pubDate.PubDate, time.Local)
		if err != nil {
			log.Warnf(ctx, "解析发布时间失败: %v", err)
			return time.Time{}
		}
		return t
	}
	return time.Time{}
}
