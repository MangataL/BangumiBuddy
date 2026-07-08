package mikan

import (
	"strings"

	"github.com/MangataL/BangumiBuddy/internal/discovery"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

type subscriptionIndex struct {
	subscriptionIDsByRSSKey map[rssKey]string
	maxPriorityBySubgroupID map[string]int
	maxPriorityByGroupName  map[string]int
}

func newSubscriptionIndex(subscriptions []subscriber.Bangumi) subscriptionIndex {
	index := subscriptionIndex{
		subscriptionIDsByRSSKey: make(map[rssKey]string, len(subscriptions)),
		maxPriorityBySubgroupID: make(map[string]int),
		maxPriorityByGroupName:  make(map[string]int),
	}
	for _, subscription := range subscriptions {
		if key, ok := normalizeRSSKey(subscription.RSSLink); ok {
			index.subscriptionIDsByRSSKey[key] = subscription.SubscriptionID
			recordMaxPriority(index.maxPriorityBySubgroupID, key.SubgroupID, subscription.Priority)
		}
		if name := normalizePriorityGroupName(subscription.ReleaseGroup); name != "" {
			recordMaxPriority(index.maxPriorityByGroupName, name, subscription.Priority)
		}
	}
	return index
}

func (index subscriptionIndex) enrichReleaseGroups(
	bangumiID string,
	groups []discovery.ReleaseGroupCandidate,
) []discovery.ReleaseGroupCandidate {
	result := make([]discovery.ReleaseGroupCandidate, 0, len(groups))
	for _, group := range groups {
		if id, ok := index.subscriptionIDsByRSSKey[rssKey{
			BangumiID:  bangumiID,
			SubgroupID: group.MikanSubgroupID,
		}]; ok {
			group.Subscribed = true
			group.SubscriptionID = id
		}
		group.PreviousMaxPriority = index.previousMaxPriorityFor(group)
		result = append(result, group)
	}
	return result
}

func (index subscriptionIndex) previousMaxPriorityFor(
	group discovery.ReleaseGroupCandidate,
) int {
	if priority, ok := index.maxPriorityBySubgroupID[group.MikanSubgroupID]; ok {
		return priority
	}
	return index.maxPriorityByGroupName[normalizePriorityGroupName(group.Name)]
}

func recordMaxPriority(priorities map[string]int, key string, priority int) {
	current, exists := priorities[key]
	if !exists || priority > current {
		priorities[key] = priority
	}
}

func normalizePriorityGroupName(name string) string {
	replacer := strings.NewReplacer(
		" ", "",
		"\t", "",
		"\n", "",
		"【", "",
		"】", "",
		"[", "",
		"]", "",
		"(", "",
		")", "",
		"（", "",
		"）", "",
	)
	return strings.ToLower(replacer.Replace(strings.TrimSpace(name)))
}
