package notice

import "time"

type NoticeReq struct {
	Title   string
	Content string
}

type NoticeSubscriptionUpdatedReq struct {
	BangumiName  string
	Poster       string
	Season       int
	ReleaseGroup string
	RSSGUID      string
	Error        error
}

type DownloadStatus string

type NoticeDownloadedReq struct {
	RSSGUID     string
	TorrentName string
	Failed      bool
	FailDetail  string
	Cost        time.Duration
	Size        int64
}

type NoticeSubscriptionTransferredReq struct {
	RSSGUID       string
	FileName      string
	BangumiName   string
	Season        int
	ReleaseGroup  string
	Poster        string
	MediaFilePath string
	Error         error
}

type NoticeTaskTransferredReq struct {
	BangumiName    string
	TorrentName    string
	Error          error
	MediaFilePaths map[string]string
}
