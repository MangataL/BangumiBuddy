package discovery

import (
	"time"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
)

type Season string

const (
	SeasonWinter Season = "winter"
	SeasonSpring Season = "spring"
	SeasonSummer Season = "summer"
	SeasonFall   Season = "fall"
)

type BangumiCandidate struct {
	MikanBangumiID          string       `json:"mikanBangumiID"`
	Name                    string       `json:"name"`
	PosterURL               string       `json:"posterURL"`
	Weekday                 time.Weekday `json:"weekday"`
	AirStartDate            string       `json:"airStartDate"`
	ReleaseGroupsKnownEmpty bool         `json:"releaseGroupsKnownEmpty,omitempty"`
}

type ListBangumiReq struct {
	Year   int    `form:"year" binding:"required"`
	Season Season `form:"season" binding:"required"`
}

type SearchReq struct {
	Query    string `form:"q" binding:"required"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

type ReleaseGroupCandidate struct {
	MikanSubgroupID     string `json:"mikanSubgroupID"`
	Name                string `json:"name"`
	Subscribed          bool   `json:"subscribed"`
	SubscriptionID      string `json:"subscriptionID,omitempty"`
	PreviousMaxPriority int    `json:"previousMaxPriority"`
}

type ResourceCandidate struct {
	Title                 string                  `json:"title"`
	MikanSubgroupID       string                  `json:"mikanSubgroupID,omitempty"`
	MagnetLink            string                  `json:"magnetLink"`
	Size                  string                  `json:"size"`
	PublishedAt           *time.Time              `json:"publishedAt"`
	ReleaseGroup          string                  `json:"releaseGroup"`
	SuggestedDownloadType downloader.DownloadType `json:"suggestedDownloadType"`
}

type SearchResp struct {
	Bangumis         []BangumiCandidate  `json:"bangumis"`
	Resources        []ResourceCandidate `json:"resources"`
	BangumiTotal     int                 `json:"bangumiTotal"`
	ResourceTotal    int                 `json:"resourceTotal"`
	ResourcePage     int                 `json:"resourcePage"`
	ResourcePageSize int                 `json:"resourcePageSize"`
}

type BatchReleaseGroupsReq struct {
	BangumiIDs []string `json:"bangumiIDs" binding:"required,min=1,max=30,dive,required"`
}

type ReleaseGroupSummary struct {
	MikanBangumiID string                  `json:"mikanBangumiID"`
	ReleaseGroups  []ReleaseGroupCandidate `json:"releaseGroups"`
}

type ReleaseGroupSummaryFailure struct {
	MikanBangumiID string `json:"mikanBangumiID"`
	Message        string `json:"message"`
}

type BatchReleaseGroupsResp struct {
	Items    []ReleaseGroupSummary        `json:"items"`
	Failures []ReleaseGroupSummaryFailure `json:"failures"`
}

type BangumiDetail struct {
	MikanBangumiID   string                  `json:"mikanBangumiID"`
	Name             string                  `json:"name"`
	PosterURL        string                  `json:"posterURL"`
	AirStartDate     string                  `json:"airStartDate"`
	EpisodeTotalText string                  `json:"episodeTotalText"`
	Overview         string                  `json:"overview"`
	ReleaseGroups    []ReleaseGroupCandidate `json:"releaseGroups"`
	Resources        []ResourceCandidate     `json:"resources"`
}

type CandidateRSSReq struct {
	MikanBangumiID  string `json:"mikanBangumiID" binding:"required"`
	MikanSubgroupID string `json:"mikanSubgroupID" binding:"required"`
	TMDBID          int    `json:"tmdbID"`
}
