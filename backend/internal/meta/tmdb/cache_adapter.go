package tmdb

import (
	"context"
	"sync"
	"time"

	"github.com/MangataL/BangumiBuddy/internal/meta"
)

var _ reloaderParser = (*cacheAdapter)(nil)

type reloaderParser interface {
	meta.Parser
	Reload(config interface{}) error
}

type seasonEpisodeTotalCacheKey struct {
	tmdbID int
	season int
}

type seasonEpisodeTotalCacheEntry struct {
	total     int
	expiresAt time.Time
}

type cacheAdapter struct {
	reloaderParser
	now func() time.Time

	mu               sync.RWMutex
	seasonTotalCache map[seasonEpisodeTotalCacheKey]seasonEpisodeTotalCacheEntry
}

func newCacheAdapter(reloaderParser reloaderParser) *cacheAdapter {
	return &cacheAdapter{
		reloaderParser:   reloaderParser,
		now:              time.Now,
		seasonTotalCache: make(map[seasonEpisodeTotalCacheKey]seasonEpisodeTotalCacheEntry),
	}
}

func (c *cacheAdapter) GetSeasonEpisodeTotalNum(
	ctx context.Context,
	tmdbID, season int,
	opts ...meta.MetaOption,
) (int, error) {
	key := seasonEpisodeTotalCacheKey{tmdbID: tmdbID, season: season}
	options := meta.NewOptions(opts...)
	if total, ok := c.getCachedSeasonTotal(key, options.CacheTTL); ok {
		return total, nil
	}

	total, err := c.reloaderParser.GetSeasonEpisodeTotalNum(ctx, tmdbID, season, opts...)
	if err != nil {
		return 0, err
	}
	if total > 0 {
		c.setCachedSeasonTotal(key, total, options.CacheTTL)
	}
	return total, nil
}

func (c *cacheAdapter) getCachedSeasonTotal(key seasonEpisodeTotalCacheKey, ttl time.Duration) (int, bool) {
	if ttl <= 0 {
		return 0, false
	}

	c.mu.RLock()
	entry, ok := c.seasonTotalCache[key]
	c.mu.RUnlock()
	if !ok || !c.now().Before(entry.expiresAt) {
		return 0, false
	}
	return entry.total, true
}

func (c *cacheAdapter) setCachedSeasonTotal(key seasonEpisodeTotalCacheKey, total int, ttl time.Duration) {
	if ttl <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.seasonTotalCache[key] = seasonEpisodeTotalCacheEntry{
		total:     total,
		expiresAt: c.now().Add(ttl),
	}
}
