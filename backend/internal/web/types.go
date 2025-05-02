package web

import (
	"time"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
)

type BangumiBase struct {
	BangumiName   string                    `json:"bangumiName"`
	Season        int                       `json:"season"`
	PosterURL     string                    `json:"posterURL"`
	BackdropURL   string                    `json:"backdropURL"`
	Overview      string                    `json:"overview"`
	Genres        string                    `json:"genres"`
	AirWeekday    time.Weekday              `json:"airWeekday"`
	ReleaseGroups []ReleaseGroupSubsription `json:"releaseGroups"`
	CreatedAt     time.Time                 `json:"-"`
}

type ReleaseGroupSubsription struct {
	SubscriptionID  string `json:"subscriptionID"`
	ReleaseGroup    string `json:"releaseGroup"`
	EpisodeTotalNum int    `json:"episodeTotalNum"`
	LastAirEpisode  int    `json:"lastAirEpisode"`
	Priority        int    `json:"priority"`
	Active          bool   `json:"active"`
}

type Torrent struct {
	Name          string                   `json:"name"`
	Hash          string                   `json:"hash"`
	Status        downloader.TorrentStatus `json:"status"`
	StatusDetail  string                   `json:"statusDetail"`
	DownloadSpeed int64                    `json:"downloadSpeed"`
	Progress      float64                  `json:"progress"`
	RSSGUID       string                   `json:"rssGUID"`
	CreatedAt     time.Time                `json:"createdAt"`
}

type File struct {
	FileName string `json:"fileName"`
	LinkName string `json:"linkName"`
}

type DeleteTorrentReq struct {
	Hash              string
	DeleteOriginFiles bool
}

type ListRecentUpdatedTorrentsReq struct {
	StartTime time.Time `form:"start_time"`
	EndTime   time.Time `form:"end_time"`
	Page      int       `form:"page"`
	PageSize  int       `form:"page_size"`
}

type ListRecentUpdatedTorrentsResp struct {
	Total    int                    `json:"total"`
	Torrents []RecentUpdatedTorrent `json:"torrents"`
}

type RecentUpdatedTorrent struct {
	PosterURL    string                   `json:"posterURL"`
	BangumiName  string                   `json:"bangumiName"`
	Season       int                      `json:"season"`
	CreatedAt    time.Time                `json:"createdAt"`
	RSSItem      string                   `json:"rssItem"`
	Status       downloader.TorrentStatus `json:"status"`
	StatusDetail string                   `json:"statusDetail"`
}
