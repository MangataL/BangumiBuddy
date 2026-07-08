package mikan

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/sync/singleflight"

	"github.com/MangataL/BangumiBuddy/internal/discovery"
	"github.com/MangataL/BangumiBuddy/internal/network"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
	"github.com/MangataL/BangumiBuddy/pkg/errs"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

type subscriptionLister interface {
	List(ctx context.Context, req subscriber.ListBangumiReq) ([]subscriber.Bangumi, error)
}

const (
	defaultSearchPageSize       = 20
	maxSearchPageSize           = 100
	defaultBatchConcurrency     = 6
	defaultDetailCacheTTL       = 5 * time.Minute
	defaultDetailCacheSize      = 512
	maxBatchReleaseGroupEntries = 30
)

type Service struct {
	mu               sync.RWMutex
	baseURL          string
	network          network.HTTPClientProvider
	subscriptions    subscriptionLister
	detailCache      *expirable.LRU[string, discovery.BangumiDetail]
	detailRequests   singleflight.Group
	detailSemaphore  chan struct{}
	batchConcurrency int
}

func New(
	config discovery.Config,
	subscriptions subscriptionLister,
	provider network.HTTPClientProvider,
) *Service {
	return &Service{
		baseURL:          normalizeBaseURL(config.MikanHost),
		network:          provider,
		subscriptions:    subscriptions,
		detailCache:      expirable.NewLRU[string, discovery.BangumiDetail](defaultDetailCacheSize, nil, defaultDetailCacheTTL),
		detailSemaphore:  make(chan struct{}, defaultBatchConcurrency),
		batchConcurrency: defaultBatchConcurrency,
	}
}

func (s *Service) Reload(config interface{}) error {
	discoveryConfig, ok := config.(*discovery.Config)
	if !ok {
		return errs.NewBadRequest("invalid discovery config")
	}
	s.mu.Lock()
	s.baseURL = normalizeBaseURL(discoveryConfig.MikanHost)
	s.mu.Unlock()
	s.detailCache.Purge()
	return nil
}

func (s *Service) ListBangumis(
	ctx context.Context,
	req discovery.ListBangumiReq,
) ([]discovery.BangumiCandidate, error) {
	if req.Year == 0 || req.Season == "" {
		return nil, errs.NewBadRequest("年份和季度不能为空")
	}
	baseURL := s.currentBaseURL()
	body, err := s.get(ctx, baseURL, "/Home/BangumiCoverFlowByDayOfWeek", url.Values{
		"year":      {fmt.Sprintf("%d", req.Year)},
		"seasonStr": {seasonLabel(req.Season)},
	})
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return parseSeasonDocument(body, baseURL)
}

func (s *Service) Search(ctx context.Context, req discovery.SearchReq) (discovery.SearchResp, error) {
	if strings.TrimSpace(req.Query) == "" {
		return discovery.SearchResp{}, errs.NewBadRequest("搜索关键词不能为空")
	}
	baseURL := s.currentBaseURL()
	body, err := s.get(ctx, baseURL, "/Home/Search", url.Values{"searchstr": {req.Query}})
	if err != nil {
		return discovery.SearchResp{}, err
	}
	defer body.Close()
	resp, err := parseSearchDocument(body, baseURL)
	if err != nil {
		return discovery.SearchResp{}, err
	}
	resp.Bangumis = s.enrichSearchBangumiMetadata(ctx, baseURL, resp.Bangumis)
	page, pageSize := normalizeSearchPage(req.Page, req.PageSize)
	resp.BangumiTotal = len(resp.Bangumis)
	resp.ResourceTotal = len(resp.Resources)
	resp.ResourcePage = page
	resp.ResourcePageSize = pageSize
	resp.Resources = paginate(resp.Resources, page, pageSize)
	return resp, nil
}

func (s *Service) enrichSearchBangumiMetadata(
	ctx context.Context,
	baseURL string,
	bangumis []discovery.BangumiCandidate,
) []discovery.BangumiCandidate {
	if len(bangumis) == 0 {
		return bangumis
	}

	enriched := append([]discovery.BangumiCandidate(nil), bangumis...)
	jobs := make(chan int)
	workerCount := s.batchConcurrency
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(enriched) {
		workerCount = len(enriched)
	}

	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range jobs {
				detail, loadErr := s.getCachedBangumiDetail(
					ctx,
					baseURL,
					enriched[index].MikanBangumiID,
				)
				if loadErr != nil {
					continue
				}
				enriched[index].AirStartDate = detail.AirStartDate
				if enriched[index].PosterURL == "" {
					enriched[index].PosterURL = detail.PosterURL
				}
			}
		}()
	}
	for index := range enriched {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
	return enriched
}

func (s *Service) BatchReleaseGroups(
	ctx context.Context,
	req discovery.BatchReleaseGroupsReq,
) (discovery.BatchReleaseGroupsResp, error) {
	ids, err := normalizeBatchBangumiIDs(req.BangumiIDs)
	if err != nil {
		return discovery.BatchReleaseGroupsResp{}, err
	}

	subscriptions, err := s.subscriptions.List(ctx, subscriber.ListBangumiReq{})
	if err != nil {
		return discovery.BatchReleaseGroupsResp{}, err
	}
	subscriptionIndex := newSubscriptionIndex(subscriptions)

	type batchResult struct {
		summary *discovery.ReleaseGroupSummary
		failure *discovery.ReleaseGroupSummaryFailure
	}
	results := make([]batchResult, len(ids))
	jobs := make(chan int)
	workerCount := s.batchConcurrency
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(ids) {
		workerCount = len(ids)
	}

	baseURL := s.currentBaseURL()
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for range workerCount {
		go func() {
			defer workers.Done()
			for index := range jobs {
				id := ids[index]
				detail, loadErr := s.getCachedBangumiDetail(ctx, baseURL, id)
				if loadErr != nil {
					results[index].failure = &discovery.ReleaseGroupSummaryFailure{
						MikanBangumiID: id,
						Message:        loadErr.Error(),
					}
					continue
				}
				groups := subscriptionIndex.enrichReleaseGroups(id, detail.ReleaseGroups)
				results[index].summary = &discovery.ReleaseGroupSummary{
					MikanBangumiID: id,
					ReleaseGroups:  groups,
				}
			}
		}()
	}
	for index := range ids {
		jobs <- index
	}
	close(jobs)
	workers.Wait()

	resp := discovery.BatchReleaseGroupsResp{
		Items:    make([]discovery.ReleaseGroupSummary, 0, len(ids)),
		Failures: make([]discovery.ReleaseGroupSummaryFailure, 0),
	}
	for _, result := range results {
		if result.summary != nil {
			resp.Items = append(resp.Items, *result.summary)
		}
		if result.failure != nil {
			resp.Failures = append(resp.Failures, *result.failure)
		}
	}
	return resp, nil
}

func (s *Service) GetBangumiDetail(ctx context.Context, id string) (discovery.BangumiDetail, error) {
	if strings.TrimSpace(id) == "" {
		return discovery.BangumiDetail{}, errs.NewBadRequest("蜜柑番剧ID不能为空")
	}
	baseURL := s.currentBaseURL()
	detail, err := s.getCachedBangumiDetail(ctx, baseURL, id)
	if err != nil {
		return discovery.BangumiDetail{}, err
	}
	return s.enrichWithSubscriptions(ctx, detail)
}

func (s *Service) fetchBangumiDetail(
	ctx context.Context,
	baseURL string,
	id string,
) (discovery.BangumiDetail, error) {
	body, err := s.get(ctx, baseURL, "/Home/Bangumi/"+id, nil)
	if err != nil {
		return discovery.BangumiDetail{}, err
	}
	defer body.Close()
	detail, err := parseDetailDocument(body, id, baseURL)
	if err != nil {
		return discovery.BangumiDetail{}, err
	}
	return detail, nil
}

func (s *Service) getCachedBangumiDetail(
	ctx context.Context,
	baseURL string,
	id string,
) (discovery.BangumiDetail, error) {
	cacheKey := baseURL + "\x00" + id
	if detail, ok := s.detailCache.Get(cacheKey); ok {
		return detail, nil
	}

	value, err, _ := s.detailRequests.Do(cacheKey, func() (any, error) {
		if detail, ok := s.detailCache.Get(cacheKey); ok {
			return detail, nil
		}
		select {
		case s.detailSemaphore <- struct{}{}:
			defer func() { <-s.detailSemaphore }()
		case <-ctx.Done():
			return discovery.BangumiDetail{}, ctx.Err()
		}
		detail, fetchErr := s.fetchBangumiDetail(ctx, baseURL, id)
		if fetchErr != nil {
			return discovery.BangumiDetail{}, fetchErr
		}
		s.detailCache.Add(cacheKey, detail)
		return detail, nil
	})
	if err != nil {
		return discovery.BangumiDetail{}, err
	}
	detail, ok := value.(discovery.BangumiDetail)
	if !ok {
		return discovery.BangumiDetail{}, fmt.Errorf("invalid cached bangumi detail")
	}
	return detail, nil
}

func (s *Service) BuildCandidateRSS(req discovery.CandidateRSSReq) (string, error) {
	bangumiID := strings.TrimSpace(req.MikanBangumiID)
	subgroupID := strings.TrimSpace(req.MikanSubgroupID)
	if bangumiID == "" || subgroupID == "" {
		return "", errs.NewBadRequest("蜜柑番剧ID和发布组ID不能为空")
	}
	return buildRSSLink(s.currentBaseURL(), bangumiID, subgroupID), nil
}

func (s *Service) get(ctx context.Context, baseURL, path string, query url.Values) (io.ReadCloser, error) {
	reqURL, err := s.buildURL(baseURL, path, query)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.httpClient().Do(req)
	if err != nil {
		log.Warnf(ctx, "fetch mikan page failed: %v", err)
		return nil, errs.NewBadRequest("蜜柑计划暂时不可用，请稍后重试")
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		log.Warnf(ctx, "fetch mikan page returned status %d, location: %s", resp.StatusCode, resp.Header.Get("Location"))
		_ = resp.Body.Close()
		return nil, errs.NewBadRequest("蜜柑计划暂时不可用，请稍后重试")
	}
	return resp.Body, nil
}

var defaultHTTPClient = &http.Client{Timeout: 10 * time.Second}

func (s *Service) httpClient() *http.Client {
	if s.network == nil {
		return defaultHTTPClient
	}
	return s.network.HTTPClient(10 * time.Second)
}

func (s *Service) currentBaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.baseURL
}

func (s *Service) buildURL(baseURL, path string, query url.Values) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	ref := &url.URL{Path: path, RawQuery: query.Encode()}
	return base.ResolveReference(ref).String(), nil
}

func (s *Service) enrichWithSubscriptions(
	ctx context.Context,
	detail discovery.BangumiDetail,
) (discovery.BangumiDetail, error) {
	if len(detail.ReleaseGroups) == 0 {
		return detail, nil
	}
	subscriptions, err := s.subscriptions.List(ctx, subscriber.ListBangumiReq{})
	if err != nil {
		return discovery.BangumiDetail{}, err
	}
	subscriptionIndex := newSubscriptionIndex(subscriptions)
	detail.ReleaseGroups = subscriptionIndex.enrichReleaseGroups(
		detail.MikanBangumiID,
		detail.ReleaseGroups,
	)
	return detail, nil
}

func normalizeSearchPage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultSearchPageSize
	}
	if pageSize > maxSearchPageSize {
		pageSize = maxSearchPageSize
	}
	return page, pageSize
}

func paginate[T any](items []T, page, pageSize int) []T {
	if len(items) == 0 {
		return []T{}
	}
	maxPage := (len(items)-1)/pageSize + 1
	if page > maxPage {
		return []T{}
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func normalizeBatchBangumiIDs(ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, errs.NewBadRequest("至少需要一个蜜柑番剧ID")
	}
	if len(ids) > maxBatchReleaseGroupEntries {
		return nil, errs.NewBadRequest("单次最多加载30部番剧的字幕组")
	}
	seen := make(map[string]bool, len(ids))
	result := make([]string, 0, len(ids))
	for _, value := range ids {
		id := strings.TrimSpace(value)
		if id == "" {
			return nil, errs.NewBadRequest("蜜柑番剧ID不能为空")
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, id)
	}
	return result, nil
}

func seasonLabel(season discovery.Season) string {
	switch season {
	case discovery.SeasonWinter:
		return "冬"
	case discovery.SeasonSpring:
		return "春"
	case discovery.SeasonSummer:
		return "夏"
	case discovery.SeasonFall:
		return "秋"
	default:
		return string(season)
	}
}

func normalizeBaseURL(host string) string {
	mikanHost := strings.Trim(strings.TrimSpace(host), "/")
	if mikanHost == "" {
		mikanHost = discovery.DefaultMikanHost
	}
	return "https://" + mikanHost
}
