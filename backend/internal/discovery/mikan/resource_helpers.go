package mikan

import (
	"regexp"
	"strings"
	"time"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/PuerkitoBio/goquery"
)

func pickMagnetLink(sel *goquery.Selection) string {
	if href, ok := sel.Attr("href"); ok && strings.HasPrefix(href, "magnet:") {
		return href
	}
	if value, ok := sel.Attr("data-clipboard-text"); ok && strings.HasPrefix(value, "magnet:") {
		return value
	}
	return ""
}

func closestResourceContainer(sel *goquery.Selection) *goquery.Selection {
	if row := sel.Closest("tr"); row.Length() > 0 {
		return row
	}
	if item := sel.Closest("li,.episode,.resource,.torrent-item"); item.Length() > 0 {
		return item
	}
	return nil
}

func pickResourceTitle(container *goquery.Selection) string {
	var title string
	container.Find("a").EachWithBreak(func(_ int, link *goquery.Selection) bool {
		if magnet := pickMagnetLink(link); magnet != "" {
			return true
		}
		title = cleanText(link.Text())
		return title == ""
	})
	if title != "" {
		return title
	}
	return cleanText(container.Text())
}

func pickResourceSubgroupID(container *goquery.Selection) string {
	if group := container.Closest("[data-subgroupid]"); group.Length() > 0 {
		return extractSubgroupID(group)
	}
	if group := container.Closest(".subgroup-text"); group.Length() > 0 {
		return pickSubgroupTextID(group)
	}
	if table := container.Closest(".episode-table"); table.Length() > 0 {
		if group := table.PrevAll().Filter(".subgroup-text").First(); group.Length() > 0 {
			return pickSubgroupTextID(group)
		}
	}
	return ""
}

func firstMatch(pattern *regexp.Regexp, text string) string {
	return pattern.FindString(text)
}

func parseTimePtr(raw string) *time.Time {
	for _, layout := range []string{
		"2006/1/2 15:04",
		"2006/01/02 15:04",
		"2006/1/2",
		"2006/01/02",
		"1/2/2006",
		"01/02/2006",
	} {
		parsed, err := time.ParseInLocation(layout, raw, time.Local)
		if err == nil {
			return &parsed
		}
	}
	return nil
}

func parseReleaseGroup(title string) string {
	if matches := releaseGroupPat.FindStringSubmatch(title); len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func suggestDownloadType(title string) downloader.DownloadType {
	switch {
	case strings.Contains(title, "剧场版"), strings.Contains(title, "电影"), strings.Contains(strings.ToLower(title), "movie"):
		return downloader.DownloadTypeMovie
	default:
		return downloader.DownloadTypeTV
	}
}

func closestReleaseGroupContainer(sel *goquery.Selection) *goquery.Selection {
	if section := sel.Closest("section,.subgroup,.subgroup-item,.sk-bangumi"); section.Length() > 0 {
		return section
	}
	return sel.Parent()
}

func pickReleaseGroupName(sel, container *goquery.Selection) string {
	if text := cleanText(container.Find("a[href*='PublishGroup']").First().Text()); text != "" {
		return text
	}
	if text := cleanText(sel.Text()); text != "" && !strings.Contains(text, "订阅") {
		return text
	}
	if text := cleanText(container.Find("a").First().Text()); text != "" {
		return text
	}
	return cleanText(container.Text())
}

func normalizeReleaseGroupName(name string) string {
	return strings.ToLower(cleanText(name))
}
