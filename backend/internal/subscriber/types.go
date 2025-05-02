package subscriber

import "time"

// Bangumi 番剧信息
type Bangumi struct {
	SubscriptionID  string       `json:"subscriptionID"`  // 订阅ID
	Name            string       `json:"name"`            // 番剧名称
	RSSLink         string       `json:"rssLink"`         // 番剧RSS链接
	Active          bool         `json:"active"`          // 订阅状态，true表示订阅中，false表示已停止
	IncludeRegs     []string     `json:"includeRegs"`     // 包含匹配，多个正则表达式，作用于RSS标题
	ExcludeRegs     []string     `json:"excludeRegs"`     // 排除匹配，多个正则表达式，作用于RSS标题
	Priority        int          `json:"priority"`        // 优先级，同一个番剧，优先级高的会覆盖优先级低的
	EpisodeOffset   int          `json:"episodeOffset"`   // 集数偏移
	Season          int          `json:"season"`          // 季数
	Year            string       `json:"year"`            // 年份
	TMDBID          int          `json:"tmdbID"`          // TMDB ID
	ReleaseGroup    string       `json:"releaseGroup"`    // 发布组
	EpisodeLocation string       `json:"episodeLocation"` // 集数位置
	PosterURL       string       `json:"posterURL"`       // 海报URL
	BackdropURL     string       `json:"backdropURL"`     // 背景图URL
	Overview        string       `json:"overview"`        // 简介
	Genres          string       `json:"genres"`          // 类型
	AirWeekday      time.Weekday `json:"airWeekday"`      // 播出时间
	EpisodeTotalNum int          `json:"episodeTotalNum"` // 集数总数
	LastAirEpisode  int          `json:"lastAirEpisode"`  // 最新下载集数
	CreatedAt       time.Time    `json:"-"`               // 创建时间
}

// ParseRSSRsp 解析RSS返回的番剧信息
type ParseRSSRsp struct {
	Name            string        `json:"name"`
	Season          int           `json:"season"`
	Year            string        `json:"year"`
	TMDBID          int           `json:"tmdbID"`
	RSSLink         string        `json:"rssLink"`
	ReleaseGroup    string        `json:"releaseGroup"`
	EpisodeTotalNum int           `json:"episodeTotalNum"`
	AirWeekday      *time.Weekday `json:"airWeekday"`
}

// RSS RSS信息
type RSS struct {
	BangumiName  string
	ReleaseGroup string
	Season       int
	Items        []RSSItem
}

// RSSItem RSS节点信息
type RSSItem struct {
	GUID        string
	TorrentLink string
	PublishedAt time.Time
}

// SubscribeReq 订阅请求
type SubscribeReq struct {
	RSSLink         string       `json:"rssLink" binding:"required"`      // RSS链接
	Season          int          `json:"season" binding:"gt=0"`           // 季数
	IncludeRegs     []string     `json:"includeRegs"`                     // 包含匹配，多个正则表达式，作用于RSS标题
	ExcludeRegs     []string     `json:"excludeRegs"`                     // 排除匹配，多个正则表达式，作用于RSS标题
	EpisodeOffset   int          `json:"episodeOffset"`                   // 集数偏移
	Priority        int          `json:"priority"`                        // 优先级，同一个番剧，优先级高的会覆盖优先级低的
	TMDBID          int          `json:"tmdbID" binding:"required"`       // TMDB ID
	ReleaseGroup    string       `json:"releaseGroup" binding:"required"` // 发布组
	EpisodeLocation string       `json:"episodeLocation"`                 // 集数位置
	EpisodeTotalNum int          `json:"episodeTotalNum" binding:"gt=0"`  // 集数总数
	AirWeekday      time.Weekday `json:"airWeekday"`                      // 播出时间
}

// ListBangumiReq 查询番剧请求
type ListBangumiReq struct {
	Active          *bool    `json:"active" form:"active"`                              // 是否只查询活跃的订阅，nil表示不过滤
	FuzzName        string   `json:"fuzzName" form:"fuzz_name,omitempty"`               // 番剧名称，模糊搜索
	Name            string   `json:"name" form:"name,omitempty"`                        // 番剧名称，精确搜索
	Season          int      `json:"season" form:"season,omitempty"`                    // 季数
	SubscriptionIDs []string `json:"subscriptionIDs" form:"subscription_ids,omitempty"` // 订阅ID
}

// UpdateSubscribeReq 更新订阅请求
type UpdateSubscribeReq struct {
	SubscriptionID  string       `json:"-"`               // 订阅ID
	Active          bool         `json:"active"`          // 订阅状态
	IncludeRegs     []string     `json:"includeRegs"`     // 包含匹配，多个正则表达式，作用于RSS标题
	ExcludeRegs     []string     `json:"excludeRegs"`     // 排除匹配，多个正则表达式，作用于RSS标题
	EpisodeOffset   int          `json:"episodeOffset"`   // 集数偏移
	Priority        int          `json:"priority"`        // 优先级，同一个番剧，优先级高的会覆盖优先级低的
	EpisodeLocation string       `json:"episodeLocation"` // 集数位置
	EpisodeTotalNum int          `json:"episodeTotalNum"` // 集数总数
	AirWeekday      time.Weekday `json:"airWeekday"`      // 播出时间
}

// RSSMatch RSS匹配
type RSSMatch struct {
	GUID        string    `json:"guid"`
	Match       bool      `json:"match"`
	Processed   bool      `json:"processed"`
	PublishedAt time.Time `json:"publishedAt"`
}

// MarkRSSRecordReq 标记RSS记录请求
type MarkRSSRecordReq struct {
	SubscriptionID string   // 订阅ID
	GUIDs          []string // RSS条目GUID
	Processed      bool     // 是否已处理
}

// CalendarItem 订阅日历条目
type CalendarItem struct {
	BangumiName string `json:"bangumiName"`
	PosterURL   string `json:"posterURL"`
	Season      int    `json:"season"`
}
