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

func (e EpisodeDetails) Empty() bool {
	return e.Name == "" && e.Overview == "" && e.StillPath == "" && e.AirDate == ""
}

func (e EpisodeDetails) AllValid() bool {
	return e.Name != "" && e.Overview != "" && e.StillPath != "" && e.AirDate != ""
}

// Merge 合并另一个 EpisodeDetails 的非空字段
func (e *EpisodeDetails) Merge(next EpisodeDetails) {
	if e.Name == "" && next.Name != "" {
		e.Name = next.Name
	}
	if e.Overview == "" && next.Overview != "" {
		e.Overview = next.Overview
	}
	if e.StillPath == "" && next.StillPath != "" {
		e.StillPath = next.StillPath
	}
	if e.AirDate == "" && next.AirDate != "" {
		e.AirDate = next.AirDate
	}
}
