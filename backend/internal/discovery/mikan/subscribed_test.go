package mikan

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MangataL/BangumiBuddy/internal/discovery"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

type fakeSubscriptionLister struct {
	subscriptions []subscriber.Bangumi
}

func (f fakeSubscriptionLister) List(
	context.Context,
	subscriber.ListBangumiReq,
) ([]subscriber.Bangumi, error) {
	return f.subscriptions, nil
}

type fakeHTTPClientProvider struct {
	response string
}

func (f fakeHTTPClientProvider) HTTPClient(time.Duration) *http.Client {
	return &http.Client{Transport: fakeRoundTripper(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(f.response)),
			Header:     make(http.Header),
		}, nil
	})}
}

type fakeRoundTripper func(*http.Request) (*http.Response, error)

func (f fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestApplySubscribedStateMatchesByNormalizedRSSKey(t *testing.T) {
	groups := []discovery.ReleaseGroupCandidate{
		{MikanSubgroupID: "370", Name: "LoliHouse"},
		{MikanSubgroupID: "371", Name: "Other"},
	}
	subscriptions := []subscriber.Bangumi{
		{
			SubscriptionID: "sub-1",
			RSSLink:        "https://mikanime.tv/RSS/Bangumi?subgroupid=370&bangumiId=681",
		},
	}

	got := newSubscriptionIndex(subscriptions).enrichReleaseGroups("681", groups)

	assert.True(t, got[0].Subscribed)
	assert.Equal(t, "sub-1", got[0].SubscriptionID)
	assert.False(t, got[1].Subscribed)
	assert.Empty(t, got[1].SubscriptionID)
}

func TestApplySubscribedStateIgnoresRSSHost(t *testing.T) {
	groups := []discovery.ReleaseGroupCandidate{
		{MikanSubgroupID: "370", Name: "LoliHouse"},
	}
	subscriptions := []subscriber.Bangumi{
		{
			SubscriptionID: "sub-1",
			RSSLink:        "https://custom.example/RSS/Bangumi?subgroupid=370&bangumiId=681",
		},
	}

	got := newSubscriptionIndex(subscriptions).enrichReleaseGroups("681", groups)

	require.Len(t, got, 1)
	assert.True(t, got[0].Subscribed)
	assert.Equal(t, "sub-1", got[0].SubscriptionID)
}

func TestSubscriptionIndexEnrichesMultipleBangumis(t *testing.T) {
	index := newSubscriptionIndex([]subscriber.Bangumi{
		{
			SubscriptionID: "sub-681",
			RSSLink:        "https://mikanani.me/RSS/Bangumi?bangumiId=681&subgroupid=370",
			Priority:       3,
		},
		{
			SubscriptionID: "sub-682",
			RSSLink:        "https://mikanani.me/RSS/Bangumi?bangumiId=682&subgroupid=370",
			Priority:       8,
		},
	})
	groups := []discovery.ReleaseGroupCandidate{
		{MikanSubgroupID: "370", Name: "LoliHouse"},
	}

	first := index.enrichReleaseGroups("681", groups)
	second := index.enrichReleaseGroups("682", groups)

	require.Len(t, first, 1)
	require.Len(t, second, 1)
	assert.Equal(t, "sub-681", first[0].SubscriptionID)
	assert.Equal(t, "sub-682", second[0].SubscriptionID)
	assert.Equal(t, 8, first[0].PreviousMaxPriority)
	assert.Equal(t, 8, second[0].PreviousMaxPriority)
}

func TestSubscriptionIndexFallsBackToNormalizedGroupNameForPriority(t *testing.T) {
	index := newSubscriptionIndex([]subscriber.Bangumi{{
		RSSLink:      "https://mikanani.me/RSS/Bangumi?bangumiId=681&subgroupid=370",
		ReleaseGroup: "【Loli House】",
		Priority:     9,
	}})

	got := index.enrichReleaseGroups("682", []discovery.ReleaseGroupCandidate{{
		MikanSubgroupID: "999",
		Name:            "LoliHouse",
	}})

	require.Len(t, got, 1)
	assert.Equal(t, 9, got[0].PreviousMaxPriority)
}

func TestGetBangumiDetailAppliesSubscribedState(t *testing.T) {
	service := New(
		discovery.Config{MikanHost: discovery.DefaultMikanHost},
		fakeSubscriptionLister{
			subscriptions: []subscriber.Bangumi{{
				SubscriptionID: "sub-1",
				RSSLink:        "https://mikanime.tv/RSS/Bangumi?subgroupid=370&bangumiId=681",
			}},
		},
		fakeHTTPClientProvider{response: `
			<html><body>
				<h1>哆啦A梦</h1>
				<section class="subgroup" data-subgroupid="370">
					<a href="/Home/PublishGroup/370">夜莺工作室</a>
				</section>
				<section class="subgroup" data-subgroupid="371">
					<a href="/Home/PublishGroup/371">其他发布组</a>
				</section>
			</body></html>`},
	)

	got, err := service.GetBangumiDetail(context.Background(), "681")

	require.NoError(t, err)
	require.Len(t, got.ReleaseGroups, 2)
	assert.True(t, got.ReleaseGroups[0].Subscribed)
	assert.Equal(t, "sub-1", got.ReleaseGroups[0].SubscriptionID)
	assert.False(t, got.ReleaseGroups[1].Subscribed)
}
