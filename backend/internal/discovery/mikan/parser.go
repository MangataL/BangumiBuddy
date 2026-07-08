package mikan

import (
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/MangataL/BangumiBuddy/internal/discovery"
)

var (
	bangumiIDPattern  = regexp.MustCompile(`/Home/Bangumi/(\d+)`)
	subgroupIDPattern = regexp.MustCompile(`(?i)(?:subgroupid|subgroup_id|subgroup)=(\d+)`)
	sizePattern       = regexp.MustCompile(`(?i)\d+(?:\.\d+)?\s*(?:GB|MB|GiB|MiB)`)
	dateTimePattern   = regexp.MustCompile(`\d{4}/\d{1,2}/\d{1,2}(?:\s+\d{1,2}:\d{2})?`)
	releaseGroupPat   = regexp.MustCompile(`^\[([^\]]+)\]`)
	cssURLPattern     = regexp.MustCompile(`url\(['"]?([^'")]+)['"]?\)`)
)

func parseSearchDocument(reader io.Reader, baseURL string) (discovery.SearchResp, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return discovery.SearchResp{}, err
	}
	return discovery.SearchResp{
		Bangumis:  parseBangumiCandidates(doc, baseURL),
		Resources: parseResourceCandidates(doc),
	}, nil
}

func parseSeasonDocument(reader io.Reader, baseURL string) ([]discovery.BangumiCandidate, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}
	candidates := parseSeasonBangumiCandidates(doc, baseURL)
	if len(candidates) > 0 {
		return candidates, nil
	}
	return parseBangumiCandidates(doc, baseURL), nil
}

func parseDetailDocument(reader io.Reader, bangumiID, baseURL string) (discovery.BangumiDetail, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return discovery.BangumiDetail{}, err
	}
	detail := discovery.BangumiDetail{
		MikanBangumiID: bangumiID,
		Name:           firstText(doc, "h1", ".bangumi-title", "title"),
		PosterURL:      parsePosterURL(doc, baseURL),
		Overview:       parseOverview(doc),
		ReleaseGroups:  parseReleaseGroups(doc),
		Resources:      parseResourceCandidates(doc),
	}
	fillBangumiDetailMeta(&detail, doc.Text())
	return detail, nil
}

func parseBangumiCandidates(doc *goquery.Document, baseURL string) []discovery.BangumiCandidate {
	seen := map[string]bool{}
	candidates := make([]discovery.BangumiCandidate, 0)
	doc.Find("a[href*='/Home/Bangumi/']").Each(func(_ int, sel *goquery.Selection) {
		href, _ := sel.Attr("href")
		id := extractBangumiID(href)
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		candidate := discovery.BangumiCandidate{
			MikanBangumiID: id,
			Name:           cleanText(pickBangumiName(sel)),
			PosterURL:      absoluteMikanURL(pickImageURL(sel), baseURL),
		}
		if weekday, ok := detectWeekday(sel); ok {
			candidate.Weekday = weekday
		}
		candidates = append(candidates, candidate)
	})
	return candidates
}

func parseSeasonBangumiCandidates(doc *goquery.Document, baseURL string) []discovery.BangumiCandidate {
	seen := map[string]bool{}
	candidates := make([]discovery.BangumiCandidate, 0)
	doc.Find(".sk-bangumi").Each(func(_ int, day *goquery.Selection) {
		weekday := parseMikanDayOfWeek(day)
		day.Find("li").Each(func(_ int, item *goquery.Selection) {
			link := item.Find("a[href*='/Home/Bangumi/']").First()
			id := pickSeasonBangumiID(item, link)
			if id == "" || seen[id] {
				return
			}
			name := pickSeasonBangumiName(item, link)
			if name == "" {
				return
			}
			seen[id] = true
			candidates = append(candidates, discovery.BangumiCandidate{
				MikanBangumiID:          id,
				Name:                    name,
				PosterURL:               absoluteMikanURL(pickImageURL(item), baseURL),
				Weekday:                 weekday,
				ReleaseGroupsKnownEmpty: isSeasonBangumiWithoutWorks(item),
			})
		})
	})
	return candidates
}

func pickSeasonBangumiID(item, link *goquery.Selection) string {
	if link != nil && link.Length() > 0 {
		href, _ := link.Attr("href")
		return extractBangumiID(href)
	}
	if id, ok := item.Find("[data-bangumiid]").First().Attr("data-bangumiid"); ok {
		return strings.TrimSpace(id)
	}
	return ""
}

func pickSeasonBangumiName(item, link *goquery.Selection) string {
	if link != nil && link.Length() > 0 {
		return cleanText(pickBangumiName(link))
	}
	if title, ok := item.Find("[title]").Last().Attr("title"); ok {
		return cleanText(title)
	}
	return cleanText(item.Find(".date-text").Last().Text())
}

func isSeasonBangumiWithoutWorks(item *goquery.Selection) bool {
	if item.Find(".greyout").Length() > 0 {
		return true
	}
	return strings.Contains(cleanText(item.Text()), "此番组下暂无作品")
}

func parseResourceCandidates(doc *goquery.Document) []discovery.ResourceCandidate {
	resources := make([]discovery.ResourceCandidate, 0)
	doc.Find("a[href^='magnet:'], [data-clipboard-text^='magnet:']").Each(func(_ int, sel *goquery.Selection) {
		magnetLink := pickMagnetLink(sel)
		if magnetLink == "" {
			return
		}
		container := closestResourceContainer(sel)
		if container == nil || container.Length() == 0 {
			return
		}
		title := pickResourceTitle(container)
		if title == "" {
			return
		}
		resources = append(resources, discovery.ResourceCandidate{
			Title:                 title,
			MikanSubgroupID:       pickResourceSubgroupID(container),
			MagnetLink:            magnetLink,
			Size:                  firstMatch(sizePattern, container.Text()),
			PublishedAt:           parseTimePtr(firstMatch(dateTimePattern, container.Text())),
			ReleaseGroup:          parseReleaseGroup(title),
			SuggestedDownloadType: suggestDownloadType(title),
		})
	})
	return resources
}

func parseMikanDayOfWeek(sel *goquery.Selection) time.Weekday {
	if raw, ok := sel.Attr("data-dayofweek"); ok {
		parsed, err := strconv.Atoi(strings.TrimSpace(raw))
		if err == nil && parsed >= 0 {
			return time.Weekday(parsed)
		}
	}
	if weekday, ok := detectWeekday(sel); ok {
		return weekday
	}
	return time.Sunday
}

func parseReleaseGroups(doc *goquery.Document) []discovery.ReleaseGroupCandidate {
	if groups := parseSubgroupTextReleaseGroups(doc); len(groups) > 0 {
		return groups
	}

	seenIDs := map[string]bool{}
	seenNames := map[string]bool{}
	groups := make([]discovery.ReleaseGroupCandidate, 0)
	doc.Find("[data-subgroupid], a[href*='subgroup'], a[href*='PublishGroup']").Each(func(_ int, sel *goquery.Selection) {
		id := extractSubgroupID(sel)
		if id == "" || seenIDs[id] {
			return
		}
		container := closestReleaseGroupContainer(sel)
		name := pickReleaseGroupName(sel, container)
		nameKey := normalizeReleaseGroupName(name)
		if nameKey != "" && seenNames[nameKey] {
			seenIDs[id] = true
			return
		}
		seenIDs[id] = true
		if nameKey != "" {
			seenNames[nameKey] = true
		}
		groups = append(groups, discovery.ReleaseGroupCandidate{
			MikanSubgroupID: id,
			Name:            name,
		})
	})
	return groups
}

func parseSubgroupTextReleaseGroups(doc *goquery.Document) []discovery.ReleaseGroupCandidate {
	seenIDs := map[string]bool{}
	seenNames := map[string]bool{}
	groups := make([]discovery.ReleaseGroupCandidate, 0)
	doc.Find(".subgroup-text").Each(func(_ int, sel *goquery.Selection) {
		id := pickSubgroupTextID(sel)
		if id == "" || seenIDs[id] {
			return
		}
		name := pickSubgroupTextName(sel)
		nameKey := normalizeReleaseGroupName(name)
		if nameKey != "" && seenNames[nameKey] {
			seenIDs[id] = true
			return
		}
		seenIDs[id] = true
		if nameKey != "" {
			seenNames[nameKey] = true
		}
		groups = append(groups, discovery.ReleaseGroupCandidate{
			MikanSubgroupID: id,
			Name:            name,
		})
	})
	return groups
}

func pickSubgroupTextID(sel *goquery.Selection) string {
	if href, ok := sel.Find(".mikan-rss").First().Attr("href"); ok {
		if matches := subgroupIDPattern.FindStringSubmatch(href); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	if id, ok := sel.Attr("id"); ok {
		return strings.TrimSpace(id)
	}
	return ""
}

func pickSubgroupTextName(sel *goquery.Selection) string {
	header := sel.Children().First()
	if header.Length() > 0 {
		clone := header.Clone()
		clone.Find(".mikan-rss,.subscribed,.dropdown,a[href*='PublishGroup']").Remove()
		if text := cleanText(clone.Text()); text != "" {
			return text
		}
	}
	if text := cleanText(sel.Find(".dropdown span").First().Text()); text != "" {
		return text
	}
	if text := cleanText(sel.Find("a[href*='PublishGroup']").First().Text()); text != "" {
		return text
	}
	return "生肉/不明字幕"
}

func fillBangumiDetailMeta(detail *discovery.BangumiDetail, text string) {
	detail.AirStartDate = extractLabelValue(text, "放送开始")
	detail.EpisodeTotalText = extractLabelValue(text, "总集数")
}

func parseOverview(doc *goquery.Document) string {
	if overview := cleanOverviewText(doc.Find(".header2-desc").First()); overview != "" {
		return overview
	}
	if overview := cleanOverviewText(doc.Find(".overview").First()); overview != "" {
		return overview
	}
	var got string
	doc.Find("h1,h2,h3,section-title,div").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		if !strings.Contains(cleanText(sel.Text()), "概况") {
			return true
		}
		got = cleanOverviewText(sel.Next())
		return got == ""
	})
	return got
}

func cleanOverviewText(sel *goquery.Selection) string {
	if sel == nil || sel.Length() == 0 {
		return ""
	}
	clone := sel.Clone()
	clone.Find("section,.subgroup,.subgroup-item,.resource,.torrent-item,table,ul,ol").Remove()
	clone.Find("a[href*='PublishGroup'],a[href*='subgroup'],a[href^='magnet:'],[data-clipboard-text^='magnet:']").Remove()
	if text := cleanText(clone.Find("p").First().Text()); text != "" {
		return text
	}
	return cleanText(clone.Text())
}

func pickBangumiName(sel *goquery.Selection) string {
	if text := cleanText(sel.Text()); text != "" {
		return text
	}
	if alt, ok := sel.Find("img").First().Attr("alt"); ok {
		return alt
	}
	return cleanText(sel.Parent().Text())
}
