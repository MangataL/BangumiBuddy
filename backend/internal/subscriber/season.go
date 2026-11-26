package subscriber

import (
	"context"
	"time"

	"github.com/MangataL/BangumiBuddy/internal/meta"
	"github.com/MangataL/BangumiBuddy/pkg/bangumifile/anito"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

const (
	// 字幕组在播出后一个周播周期内发布，可作为该季的日期依据。
	maxReleaseDelayDays = 7
	maxSeasonLookups    = 5
)

func (s *Subscriber) recommendSeason(ctx context.Context, tv meta.Meta, items []RSSItem) (meta.Season, string) {
	if len(tv.Seasons) == 1 {
		return tv.Seasons[0], ""
	}
	published, kind := earliestRelease(items)
	if published.IsZero() {
		return meta.Season{}, ""
	}
	date := published.Format(time.DateOnly)
	var eligible []meta.Season
	for _, season := range tv.Seasons {
		_, err := time.Parse(time.DateOnly, season.AirDate)
		if err != nil && kind == anito.EpisodeCollection {
			return meta.Season{}, ""
		}
		if err != nil || season.AirDate <= date {
			eligible = append(eligible, season)
		}
	}
	if kind == anito.EpisodeCollection {
		if len(eligible) == 1 {
			return eligible[0], date
		}
		return meta.Season{}, ""
	}

	// 季可能分割放送，不能仅按首播日排除较早的季。
	if len(eligible) > maxSeasonLookups {
		return meta.Season{}, ""
	}
	earliestAirDate := published.AddDate(0, 0, -maxReleaseDelayDays).Format(time.DateOnly)
	var match meta.Season
	for _, season := range eligible {
		dates, err := s.metaParser.GetSeasonAirDates(ctx, tv.TMDBID, season.Number)
		if err != nil {
			log.Warnf(ctx, "获取 TMDB %d 第 %d 季播出日期失败: %v", tv.TMDBID, season.Number, err)
			return meta.Season{}, ""
		}
		complete := len(dates) > 0 && len(dates) >= season.EpisodeTotalNum
		matched := false
		for _, airDate := range dates {
			if _, err := time.Parse(time.DateOnly, airDate); err != nil {
				complete = false
			} else if earliestAirDate <= airDate && airDate <= date {
				matched = true
				break
			}
		}
		if matched {
			if match.Number != 0 {
				return meta.Season{}, ""
			}
			match = season
		} else if !complete {
			return meta.Season{}, ""
		}
	}
	if match.Number == 0 {
		return meta.Season{}, ""
	}
	return match, date
}

func earliestRelease(items []RSSItem) (time.Time, anito.ReleaseKind) {
	var single, collection time.Time
	onlyCollections := true
	now := time.Now()
	for _, item := range items {
		kind := anito.ClassifyRelease(item.Title)
		if kind != anito.EpisodeCollection {
			onlyCollections = false
		}
		if item.PublishedAt.IsZero() || item.PublishedAt.After(now) {
			continue
		}
		switch kind {
		case anito.SingleEpisode:
			if single.IsZero() || item.PublishedAt.Before(single) {
				single = item.PublishedAt
			}
		case anito.EpisodeCollection:
			if collection.IsZero() || item.PublishedAt.Before(collection) {
				collection = item.PublishedAt
			}
		}
	}
	if onlyCollections {
		return collection, anito.EpisodeCollection
	}
	return single, anito.SingleEpisode
}
