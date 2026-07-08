package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MangataL/BangumiBuddy/internal/discovery"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

type fakeDiscovery struct {
	rssLink  string
	rssReq   discovery.CandidateRSSReq
	batchReq discovery.BatchReleaseGroupsReq
}

func (f *fakeDiscovery) ListBangumis(
	context.Context,
	discovery.ListBangumiReq,
) ([]discovery.BangumiCandidate, error) {
	return nil, nil
}

func (f *fakeDiscovery) Search(context.Context, discovery.SearchReq) (discovery.SearchResp, error) {
	return discovery.SearchResp{}, nil
}

func (f *fakeDiscovery) BatchReleaseGroups(
	_ context.Context,
	req discovery.BatchReleaseGroupsReq,
) (discovery.BatchReleaseGroupsResp, error) {
	f.batchReq = req
	return discovery.BatchReleaseGroupsResp{
		Items: []discovery.ReleaseGroupSummary{
			{
				MikanBangumiID: "681",
				ReleaseGroups: []discovery.ReleaseGroupCandidate{
					{MikanSubgroupID: "370", Name: "夜莺工作室"},
				},
			},
		},
		Failures: []discovery.ReleaseGroupSummaryFailure{},
	}, nil
}

func (f *fakeDiscovery) GetBangumiDetail(context.Context, string) (discovery.BangumiDetail, error) {
	return discovery.BangumiDetail{}, nil
}

func (f *fakeDiscovery) BuildCandidateRSS(req discovery.CandidateRSSReq) (string, error) {
	f.rssReq = req
	return f.rssLink, nil
}

type fakeSubscriber struct {
	req subscriber.ParserRSSReq
}

func (f *fakeSubscriber) ParseRSS(ctx context.Context, req subscriber.ParserRSSReq) (subscriber.ParseRSSRsp, error) {
	f.req = req
	return subscriber.ParseRSSRsp{
		Name:         "哆啦A梦",
		TMDBID:       req.TMDBID,
		RSSLink:      req.RSSLink,
		ReleaseGroup: "夜莺工作室",
	}, nil
}

func (f *fakeSubscriber) Subscribe(context.Context, subscriber.SubscribeReq) (subscriber.Bangumi, error) {
	return subscriber.Bangumi{}, nil
}

func (f *fakeSubscriber) Get(context.Context, string) (subscriber.Bangumi, error) {
	return subscriber.Bangumi{}, nil
}

func (f *fakeSubscriber) List(context.Context, subscriber.ListBangumiReq) ([]subscriber.Bangumi, error) {
	return nil, nil
}

func (f *fakeSubscriber) UpdateSubscription(context.Context, subscriber.UpdateSubscribeReq) error {
	return nil
}

func (f *fakeSubscriber) DeleteSubscription(context.Context, string) error {
	return nil
}

func (f *fakeSubscriber) UpdateLastAirEpisode(context.Context, string, int) error {
	return nil
}

func (f *fakeSubscriber) HandleEpisodeTransferred(context.Context, string, int) error {
	return nil
}

func (f *fakeSubscriber) GetRSSMatch(context.Context, string) ([]subscriber.RSSMatch, error) {
	return nil, nil
}

func (f *fakeSubscriber) PreviewRSSMatch(
	context.Context,
	subscriber.PreviewRSSMatchReq,
) ([]subscriber.RSSMatch, error) {
	return nil, nil
}

func (f *fakeSubscriber) HandleBangumiSubscription(context.Context, string) error {
	return nil
}

func (f *fakeSubscriber) MarkRSSRecord(context.Context, subscriber.MarkRSSRecordReq) error {
	return nil
}

func (f *fakeSubscriber) GetSubscriptionCalendar(context.Context) (map[time.Weekday][]subscriber.CalendarItem, error) {
	return nil, nil
}

func (f *fakeSubscriber) StopSubscription(context.Context, string) error {
	return nil
}

func (f *fakeSubscriber) AutoStopSubscription(context.Context, string) bool {
	return false
}

func TestRouter_ParseDiscoveryCandidateRSS(t *testing.T) {
	d := &fakeDiscovery{rssLink: "https://mikanani.me/RSS/Bangumi?bangumiId=681&subgroupid=370"}
	s := &fakeSubscriber{}
	router := New(Dependency{
		Discovery:  d,
		Subscriber: s,
	})
	body, err := json.Marshal(discovery.CandidateRSSReq{
		MikanBangumiID:  "681",
		MikanSubgroupID: "370",
		TMDBID:          18692,
	})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/discovery/mikan/rss/parse", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	router.ParseDiscoveryCandidateRSS(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, discovery.CandidateRSSReq{
		MikanBangumiID:  "681",
		MikanSubgroupID: "370",
		TMDBID:          18692,
	}, d.rssReq)
	assert.Equal(t, subscriber.ParserRSSReq{
		RSSLink: "https://mikanani.me/RSS/Bangumi?bangumiId=681&subgroupid=370",
		TMDBID:  18692,
	}, s.req)
	assert.Contains(t, w.Body.String(), `"releaseGroup":"夜莺工作室"`)
}

func TestRouter_BatchDiscoveryReleaseGroups(t *testing.T) {
	d := &fakeDiscovery{}
	router := New(Dependency{Discovery: d})
	body, err := json.Marshal(discovery.BatchReleaseGroupsReq{
		BangumiIDs: []string{"681", "999"},
	})
	require.NoError(t, err)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(
		http.MethodPost,
		"/discovery/mikan/bangumis/release-groups/batch",
		bytes.NewReader(body),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	router.BatchDiscoveryReleaseGroups(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []string{"681", "999"}, d.batchReq.BangumiIDs)
	assert.Contains(t, w.Body.String(), `"mikanBangumiID":"681"`)
	assert.Contains(t, w.Body.String(), `"mikanSubgroupID":"370"`)
}
