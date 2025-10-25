package meta

import "time"

type Meta struct {
	ChineseName     string        `json:"chineseName"`
	Year            string        `json:"year"`
	TMDBID          int           `json:"tmdbID"`
	Season          int           `json:"season"`
	EpisodeTotalNum int           `json:"episodeTotalNum"`
	AirWeekday      *time.Weekday `json:"airWeekday"`
	PosterURL       string        `json:"posterURL"`
	BackdropURL     string        `json:"backdropURL"`
	Overview        string        `json:"overview"`
	Genres          string        `json:"genres"`
}

type EpisodeDetails struct {
	Name      string `json:"name"`
	Overview  string `json:"overview"`
	StillPath string `json:"stillPath"` // 单集图片路径
	AirDate   string `json:"airDate"`
}
