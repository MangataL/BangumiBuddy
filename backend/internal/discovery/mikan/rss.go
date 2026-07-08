package mikan

import (
	"fmt"
	"net/url"
	"strings"
)

type rssKey struct {
	BangumiID  string
	SubgroupID string
}

func buildRSSLink(baseURL, bangumiID, subgroupID string) string {
	return fmt.Sprintf("%s/RSS/Bangumi?bangumiId=%s&subgroupid=%s", strings.TrimRight(baseURL, "/"), bangumiID, subgroupID)
}

func normalizeRSSKey(link string) (rssKey, bool) {
	u, err := url.Parse(link)
	if err != nil {
		return rssKey{}, false
	}

	query := normalizeQuery(u.Query())
	key := rssKey{
		BangumiID:  query.Get("bangumiid"),
		SubgroupID: query.Get("subgroupid"),
	}
	if key.BangumiID == "" || key.SubgroupID == "" {
		return rssKey{}, false
	}
	return key, true
}

func normalizeQuery(query url.Values) url.Values {
	normalized := make(url.Values, len(query))
	for key, values := range query {
		normalized[strings.ToLower(key)] = values
	}
	return normalized
}
