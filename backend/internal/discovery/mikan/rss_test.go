package mikan

import (
	"testing"

	"github.com/MangataL/BangumiBuddy/internal/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeRSSKey(t *testing.T) {
	testCases := []struct {
		name string
		link string
		want rssKey
	}{
		{
			name: "mikanani host",
			link: "https://mikanani.me/RSS/Bangumi?bangumiId=681&subgroupid=370",
			want: rssKey{BangumiID: "681", SubgroupID: "370"},
		},
		{
			name: "mikanime host with different query order",
			link: "https://mikanime.tv/RSS/Bangumi?subgroupid=370&bangumiId=681",
			want: rssKey{BangumiID: "681", SubgroupID: "370"},
		},
		{
			name: "case-insensitive query keys",
			link: "https://mikanani.me/RSS/Bangumi?BangumiId=681&SubGroupID=370",
			want: rssKey{BangumiID: "681", SubgroupID: "370"},
		},
		{
			name: "different host",
			link: "https://example.com/RSS/Bangumi?bangumiId=681&subgroupid=370",
			want: rssKey{BangumiID: "681", SubgroupID: "370"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := normalizeRSSKey(tc.link)

			require.True(t, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNormalizeRSSKeyRejectsInvalidLinks(t *testing.T) {
	testCases := []string{
		"",
		"https://mikanani.me/RSS/Bangumi?bangumiId=681",
	}

	for _, link := range testCases {
		t.Run(link, func(t *testing.T) {
			_, ok := normalizeRSSKey(link)

			assert.False(t, ok)
		})
	}
}

func TestBuildRSSLink(t *testing.T) {
	got := buildRSSLink(normalizeBaseURL(discovery.DefaultMikanHost), "681", "370")

	assert.Equal(t, "https://mikanani.me/RSS/Bangumi?bangumiId=681&subgroupid=370", got)
}

func TestBuildCandidateRSSUsesConfiguredHost(t *testing.T) {
	service := New(discovery.Config{MikanHost: "mikanime.tv"}, &staticSubscriptionLister{}, nil)

	got, err := service.BuildCandidateRSS(discovery.CandidateRSSReq{
		MikanBangumiID:  "681",
		MikanSubgroupID: "370",
	})

	require.NoError(t, err)
	assert.Equal(t, "https://mikanime.tv/RSS/Bangumi?bangumiId=681&subgroupid=370", got)
}

func TestBuildCandidateRSSUsesDefaultHost(t *testing.T) {
	service := New(discovery.Config{}, &staticSubscriptionLister{}, nil)

	got, err := service.BuildCandidateRSS(discovery.CandidateRSSReq{
		MikanBangumiID:  "681",
		MikanSubgroupID: "370",
	})

	require.NoError(t, err)
	assert.Equal(t, "https://mikanani.me/RSS/Bangumi?bangumiId=681&subgroupid=370", got)
}
