package mikan

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MangataL/BangumiBuddy/internal/discovery"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

type captureHTTPClientProvider struct {
	request *http.Request
}

func (c *captureHTTPClientProvider) HTTPClient(time.Duration) *http.Client {
	return &http.Client{Transport: fakeRoundTripper(func(req *http.Request) (*http.Response, error) {
		c.request = req
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`<html><body></body></html>`)),
			Header:     make(http.Header),
		}, nil
	})}
}

func TestListBangumisUsesMikanSeasonToken(t *testing.T) {
	provider := &captureHTTPClientProvider{}
	service := New(discovery.Config{MikanHost: discovery.DefaultMikanHost}, &staticSubscriptionLister{}, provider)

	_, err := service.ListBangumis(context.Background(), discovery.ListBangumiReq{
		Year:   2026,
		Season: discovery.SeasonSummer,
	})

	require.NoError(t, err)
	require.NotNil(t, provider.request)
	assert.Equal(t, "/Home/BangumiCoverFlowByDayOfWeek", provider.request.URL.Path)
	assert.Equal(t, "2026", provider.request.URL.Query().Get("year"))
	assert.Equal(t, "夏", provider.request.URL.Query().Get("seasonStr"))
}

type routeHTTPClientProvider struct {
	responses map[string]string
	requests  map[string]int
	statuses  map[string]int
}

func (r routeHTTPClientProvider) HTTPClient(time.Duration) *http.Client {
	return &http.Client{Transport: fakeRoundTripper(func(req *http.Request) (*http.Response, error) {
		r.requests[req.URL.Path]++
		body := r.responses[req.URL.Path]
		status := r.statuses[req.URL.Path]
		if status == 0 {
			status = http.StatusOK
		}
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}
}

func TestListBangumisDoesNotFetchDetails(t *testing.T) {
	requests := map[string]int{}
	service := New(
		discovery.Config{MikanHost: discovery.DefaultMikanHost},
		&staticSubscriptionLister{},
		routeHTTPClientProvider{responses: map[string]string{
			"/Home/BangumiCoverFlowByDayOfWeek": `
				<html><body>
					<div class="sk-bangumi" data-dayofweek="1">
						<ul>
							<li>
								<div class="num-node text-center">2</div>
								<div class="an-info">
									<div class="date-text">2026/07/07 更新</div>
									<a href="/Home/Bangumi/999" class="an-text">欺诈游戏</a>
								</div>
							</li>
						</ul>
					</div>
				</body></html>`,
		}, requests: requests},
	)

	got, err := service.ListBangumis(context.Background(), discovery.ListBangumiReq{
		Year:   2026,
		Season: discovery.SeasonSummer,
	})

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 1, requests["/Home/BangumiCoverFlowByDayOfWeek"])
	assert.Zero(t, requests["/Home/Bangumi/999"])
}

func TestSearchPaginatesBangumisAndResources(t *testing.T) {
	service := New(
		discovery.Config{MikanHost: discovery.DefaultMikanHost},
		&staticSubscriptionLister{},
		routeHTTPClientProvider{
			responses: map[string]string{
				"/Home/Search": `
					<html><body>
						<a href="/Home/Bangumi/1">番剧一</a>
						<a href="/Home/Bangumi/2">番剧二</a>
						<a href="/Home/Bangumi/3">番剧三</a>
						<table>
							<tr><td><a href="/Home/Episode/1">[组一] 资源一</a></td><td><button data-clipboard-text="magnet:?xt=urn:btih:1">复制</button></td></tr>
							<tr><td><a href="/Home/Episode/2">[组二] 资源二</a></td><td><button data-clipboard-text="magnet:?xt=urn:btih:2">复制</button></td></tr>
							<tr><td><a href="/Home/Episode/3">[组三] 资源三</a></td><td><button data-clipboard-text="magnet:?xt=urn:btih:3">复制</button></td></tr>
						</table>
					</body></html>`,
			},
			requests: map[string]int{},
		},
	)
	service.batchConcurrency = 1

	got, err := service.Search(context.Background(), discovery.SearchReq{
		Query:    "番剧",
		Page:     2,
		PageSize: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, 3, got.BangumiTotal)
	assert.Equal(t, 3, got.ResourceTotal)
	assert.Equal(t, 2, got.ResourcePage)
	assert.Equal(t, 1, got.ResourcePageSize)
	require.Len(t, got.Bangumis, 3)
	assert.Equal(t, []string{"1", "2", "3"}, []string{
		got.Bangumis[0].MikanBangumiID,
		got.Bangumis[1].MikanBangumiID,
		got.Bangumis[2].MikanBangumiID,
	})
	require.Len(t, got.Resources, 1)
	assert.Equal(t, "[组二] 资源二", got.Resources[0].Title)
}

func TestSearchEnrichesBangumiAirStartDateAndMissingPoster(t *testing.T) {
	requests := map[string]int{}
	service := New(
		discovery.Config{MikanHost: discovery.DefaultMikanHost},
		&staticSubscriptionLister{},
		routeHTTPClientProvider{
			responses: map[string]string{
				"/Home/Search": `<a href="/Home/Bangumi/681">哆啦A梦</a>`,
				"/Home/Bangumi/681": `
					<html><body>
						<section class="bangumi-info">
							<div class="bangumi-poster" style="background-image: url('/images/doraemon.jpg');"></div>
							<h1>哆啦A梦</h1>
							<p>放送开始：4/15/2005</p>
						</section>
					</body></html>`,
			},
			requests: requests,
		},
	)

	got, err := service.Search(context.Background(), discovery.SearchReq{
		Query: "哆啦A梦",
	})

	require.NoError(t, err)
	require.Len(t, got.Bangumis, 1)
	assert.Equal(t, "4/15/2005", got.Bangumis[0].AirStartDate)
	assert.Equal(t, "https://mikanani.me/images/doraemon.jpg", got.Bangumis[0].PosterURL)
	assert.Equal(t, 1, requests["/Home/Bangumi/681"])
}

func TestPaginateHandlesPagesBeyondIntegerMultiplicationRange(t *testing.T) {
	assert.Empty(t, paginate([]int{1, 2, 3}, int(^uint(0)>>1), 100))
}

type staticSubscriptionLister struct {
	items []subscriber.Bangumi
	calls int
	mu    sync.Mutex
}

func (s *staticSubscriptionLister) List(
	context.Context,
	subscriber.ListBangumiReq,
) ([]subscriber.Bangumi, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return s.items, nil
}

func TestBatchReleaseGroupsDeduplicatesCachesAndAppliesSubscriptions(t *testing.T) {
	requests := map[string]int{}
	subscriptions := &staticSubscriptionLister{items: []subscriber.Bangumi{
		{
			SubscriptionID: "subscription-1",
			RSSLink:        "https://mikanani.me/RSS/Bangumi?bangumiId=1&subgroupid=10",
		},
	}}
	service := New(
		discovery.Config{MikanHost: discovery.DefaultMikanHost},
		subscriptions,
		routeHTTPClientProvider{
			responses: map[string]string{
				"/Home/Bangumi/1": detailDocument("番剧一", "10", "字幕组一"),
				"/Home/Bangumi/2": detailDocument("番剧二", "20", "字幕组二"),
			},
			requests: requests,
		},
	)
	service.batchConcurrency = 1

	got, err := service.BatchReleaseGroups(context.Background(), discovery.BatchReleaseGroupsReq{
		BangumiIDs: []string{"1", "1", "2"},
	})
	require.NoError(t, err)
	require.Len(t, got.Items, 2)
	require.Empty(t, got.Failures)
	assert.Equal(t, "1", got.Items[0].MikanBangumiID)
	require.Len(t, got.Items[0].ReleaseGroups, 1)
	assert.True(t, got.Items[0].ReleaseGroups[0].Subscribed)
	assert.Equal(t, "subscription-1", got.Items[0].ReleaseGroups[0].SubscriptionID)

	_, err = service.BatchReleaseGroups(context.Background(), discovery.BatchReleaseGroupsReq{
		BangumiIDs: []string{"1", "2"},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, requests["/Home/Bangumi/1"])
	assert.Equal(t, 1, requests["/Home/Bangumi/2"])
	assert.Equal(t, 2, subscriptions.calls)
}

func TestBatchReleaseGroupsReturnsPartialFailures(t *testing.T) {
	service := New(
		discovery.Config{MikanHost: discovery.DefaultMikanHost},
		&staticSubscriptionLister{},
		routeHTTPClientProvider{
			responses: map[string]string{
				"/Home/Bangumi/1": detailDocument("番剧一", "10", "字幕组一"),
			},
			requests: map[string]int{},
			statuses: map[string]int{"/Home/Bangumi/2": http.StatusBadGateway},
		},
	)
	service.batchConcurrency = 1

	got, err := service.BatchReleaseGroups(context.Background(), discovery.BatchReleaseGroupsReq{
		BangumiIDs: []string{"1", "2"},
	})

	require.NoError(t, err)
	require.Len(t, got.Items, 1)
	assert.Equal(t, "1", got.Items[0].MikanBangumiID)
	require.Len(t, got.Failures, 1)
	assert.Equal(t, "2", got.Failures[0].MikanBangumiID)
	assert.NotEmpty(t, got.Failures[0].Message)
}

type controlledConcurrencyProvider struct {
	mu       sync.Mutex
	active   int
	max      int
	requests int
	started  chan struct{}
	gate     <-chan struct{}
}

func (p *controlledConcurrencyProvider) HTTPClient(time.Duration) *http.Client {
	return &http.Client{Transport: fakeRoundTripper(func(req *http.Request) (*http.Response, error) {
		p.mu.Lock()
		p.active++
		p.requests++
		if p.active > p.max {
			p.max = p.active
		}
		p.mu.Unlock()
		p.started <- struct{}{}
		<-p.gate
		p.mu.Lock()
		p.active--
		p.mu.Unlock()
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(detailDocument(req.URL.Path, "10", "字幕组"))),
			Header:     make(http.Header),
		}, nil
	})}
}

func TestBatchReleaseGroupsLimitsConcurrency(t *testing.T) {
	gate := make(chan struct{})
	provider := &controlledConcurrencyProvider{
		started: make(chan struct{}, 4),
		gate:    gate,
	}
	service := New(discovery.Config{MikanHost: discovery.DefaultMikanHost}, &staticSubscriptionLister{}, provider)
	service.batchConcurrency = 2

	done := make(chan error, 1)
	go func() {
		_, err := service.BatchReleaseGroups(context.Background(), discovery.BatchReleaseGroupsReq{
			BangumiIDs: []string{"1", "2", "3", "4"},
		})
		done <- err
	}()

	for range 2 {
		select {
		case <-provider.started:
		case <-time.After(time.Second):
			t.Fatal("批量请求未按预期并发启动")
		}
	}
	close(gate)
	require.NoError(t, <-done)
	provider.mu.Lock()
	defer provider.mu.Unlock()
	assert.Equal(t, 2, provider.max)
}

func TestGetBangumiDetailDeduplicatesConcurrentRequests(t *testing.T) {
	gate := make(chan struct{})
	provider := &controlledConcurrencyProvider{
		started: make(chan struct{}, 2),
		gate:    gate,
	}
	service := New(discovery.Config{MikanHost: discovery.DefaultMikanHost}, &staticSubscriptionLister{}, provider)

	done := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := service.GetBangumiDetail(context.Background(), "1")
			done <- err
		}()
	}
	select {
	case <-provider.started:
	case <-time.After(time.Second):
		t.Fatal("详情请求未启动")
	}
	time.Sleep(20 * time.Millisecond)
	provider.mu.Lock()
	assert.Equal(t, 1, provider.requests)
	provider.mu.Unlock()
	close(gate)
	require.NoError(t, <-done)
	require.NoError(t, <-done)
}

func detailDocument(name, subgroupID, subgroupName string) string {
	return `<html><body>
		<h1>` + name + `</h1>
		<section class="subgroup" data-subgroupid="` + subgroupID + `">
			<a href="/Home/PublishGroup/` + subgroupID + `">` + subgroupName + `</a>
		</section>
	</body></html>`
}
