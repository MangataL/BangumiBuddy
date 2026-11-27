package mikan

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/errors"

	"github.com/MangataL/BangumiBuddy/internal/network"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
	"github.com/MangataL/BangumiBuddy/pkg/bangumifile"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

func NewParser(bfParser bangumifile.Parser, provider network.HTTPClientProvider) subscriber.RSSParser {
	return &parser{
		bfParser: bfParser,
		client:   newHTTPClient(provider),
	}
}

type parser struct {
	bfParser bangumifile.Parser
	client   *http.Client
}

func (p *parser) Parse(ctx context.Context, link string) (subscriber.RSS, error) {
	feed, err := p.feedParser().ParseURLWithContext(link, ctx)
	if err != nil {
		return subscriber.RSS{}, errors.WithMessage(err, "解析RSS失败")
	}
	items := getItems(ctx, feed.Items)
	rg := p.parseReleaseGroup(ctx, items)
	return subscriber.RSS{
		BangumiName:  getBangumiName(ctx, feed.Title),
		ReleaseGroup: rg,
		Items:        items,
	}, nil
}

func (p *parser) feedParser() *gofeed.Parser {
	fp := gofeed.NewParser()
	fp.Client = p.client
	return fp
}

func newHTTPClient(provider network.HTTPClientProvider) *http.Client {
	return provider.HTTPClient(30 * time.Second)
}

func (p *parser) parseReleaseGroup(ctx context.Context, items []subscriber.RSSItem) string {
	if len(items) == 0 {
		return ""
	}
	bf, err := p.bfParser.Parse(ctx, items[0].GUID,
		bangumifile.IgnoreValidateEpisode(),
		bangumifile.PreserveOriginName(),
	)
	if err != nil {
		return ""
	}
	return bf.ReleaseGroup
}

var (
	mikanTitleRegex = regexp.MustCompile(`Mikan Project - (.+?)(?:\s*第|$)`)
	shanghai        = func() *time.Location {
		loc, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			return time.FixedZone("Asia/Shanghai", 8*60*60)
		}
		return loc
	}()
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
		if err := xml.Unmarshal([]byte(wrappedData), &pubDate); err == nil {
			if t, err := parseMikanTimestamp(pubDate.PubDate); err == nil {
				return t
			}
		}
		log.Warnf(ctx, "无法解析 torrent 发布时间: %s", pubDate.PubDate)
	}
	if item.PublishedParsed != nil {
		// gofeed 将标准日期转换为 UTC，推荐使用发布者原时区的日历日。
		if t, err := mail.ParseDate(item.Published); err == nil {
			return t
		}
		if t, err := parseMikanTimestamp(item.Published); err == nil {
			return t
		}
		return *item.PublishedParsed
	}
	return time.Time{}
}

func parseMikanTimestamp(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	return time.ParseInLocation("2006-01-02T15:04:05", value, shanghai)
}
