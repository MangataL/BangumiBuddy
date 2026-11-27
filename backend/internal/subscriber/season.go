package subscriber

import (
	"time"

	"github.com/MangataL/BangumiBuddy/internal/meta"
)

func recommendSeason(seasons []meta.Season, items []RSSItem) (meta.Season, string) {
	now := time.Now()
	var earliest time.Time
	for _, item := range items {
		published := item.PublishedAt
		if published.IsZero() || published.After(now) {
			continue
		}
		if earliest.IsZero() || published.Before(earliest) {
			earliest = published
		}
	}
	if earliest.IsZero() {
		return meta.Season{}, ""
	}

	date := earliest.Format(time.DateOnly)
	var match meta.Season
	tied := false
	for _, season := range seasons {
		if season.Number <= 0 {
			continue
		}
		if _, err := time.Parse(time.DateOnly, season.AirDate); err != nil {
			return meta.Season{}, ""
		}
		if season.AirDate > date {
			continue
		}
		switch {
		case season.AirDate > match.AirDate:
			match = season
			tied = false
		case season.AirDate == match.AirDate:
			tied = true
		}
	}
	if match.Number == 0 || tied {
		return meta.Season{}, ""
	}
	return match, date
}
