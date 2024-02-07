package meta

import "time"

type Meta struct {
	ChineseName     string
	Year            string
	TMDBID          int
	Season          int
	EpisodeTotalNum int
	AirWeekday      *time.Weekday
	PosterURL       string
	BackdropURL     string
	Overview        string
	Genres          string
}
