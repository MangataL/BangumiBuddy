package mikan

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func extractBangumiID(href string) string {
	if matches := bangumiIDPattern.FindStringSubmatch(href); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func extractSubgroupID(sel *goquery.Selection) string {
	for node := sel; node.Length() > 0; node = node.Parent() {
		if value, ok := node.Attr("data-subgroupid"); ok && value != "" {
			return value
		}
	}
	href, _ := sel.Attr("href")
	if matches := subgroupIDPattern.FindStringSubmatch(href); len(matches) > 1 {
		return matches[1]
	}
	if matches := regexp.MustCompile(`/Home/PublishGroup/(\d+)`).FindStringSubmatch(href); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func absoluteMikanURL(raw, baseURL string) string {
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "//") {
		return "https:" + raw
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	base, _ := url.Parse(baseURL)
	ref, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	return base.ResolveReference(ref).String()
}

func cleanText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func firstText(doc *goquery.Document, selectors ...string) string {
	for _, selector := range selectors {
		if text := cleanText(doc.Find(selector).First().Text()); text != "" {
			return strings.TrimPrefix(text, "Mikan Project - ")
		}
	}
	return ""
}

func parsePosterURL(doc *goquery.Document, baseURL string) string {
	if style, ok := doc.Find(".bangumi-poster").First().Attr("style"); ok {
		if matches := cssURLPattern.FindStringSubmatch(style); len(matches) > 1 {
			return absoluteMikanURL(matches[1], baseURL)
		}
	}
	if src := pickImageURL(doc.Selection); src != "" {
		return absoluteMikanURL(src, baseURL)
	}
	return ""
}

func pickImageURL(sel *goquery.Selection) string {
	if src, ok := sel.Attr("data-src"); ok {
		return src
	}
	img := sel.Find("img").First()
	if img.Length() == 0 {
		img = sel.Find("[data-src]").First()
	}
	if img.Length() == 0 && goquery.NodeName(sel) == "img" {
		img = sel
	}
	if src, ok := img.Attr("src"); ok {
		return src
	}
	if src, ok := img.Attr("data-src"); ok {
		return src
	}
	return ""
}

func detectWeekday(sel *goquery.Selection) (time.Weekday, bool) {
	for node := sel; node.Length() > 0; node = node.Parent() {
		if weekday, ok := parseWeekday(node.Text()); ok {
			return weekday, true
		}
	}
	return time.Sunday, false
}

func parseWeekday(text string) (time.Weekday, bool) {
	switch {
	case strings.Contains(text, "星期日"):
		return time.Sunday, true
	case strings.Contains(text, "星期一"):
		return time.Monday, true
	case strings.Contains(text, "星期二"):
		return time.Tuesday, true
	case strings.Contains(text, "星期三"):
		return time.Wednesday, true
	case strings.Contains(text, "星期四"):
		return time.Thursday, true
	case strings.Contains(text, "星期五"):
		return time.Friday, true
	case strings.Contains(text, "星期六"):
		return time.Saturday, true
	default:
		return time.Sunday, false
	}
}

func extractLabelValue(text, label string) string {
	pattern := regexp.MustCompile(label + `[:：]\s*([^\n\r]+)`)
	if matches := pattern.FindStringSubmatch(text); len(matches) > 1 {
		return cleanText(matches[1])
	}
	return ""
}
