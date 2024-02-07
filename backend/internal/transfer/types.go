package transfer

type TransferReq struct {
}

type Meta struct {
	ChineseName     string
	Year            string
	Season          int
	EpisodeLocation string
	EpisodeOffset   int
	FileName        string
	FilePath        string
	SubscriptionID  string
	ReleaseGroup    string
}

type DeletePriorityReq struct {
	NewFile        string
	SubscriptionID string
}

type Priority struct {
	NewFile        string
	SubscriptionID string
}

type FileTransferred struct {
	OriginFile     string
	BangumiName    string
	Season         int
	SubscriptionID string
	NewFile        string
	NewFileID      string
}

type GetFileTransferredReq struct {
	OriginFile string
	NewFileID  string
}

type ListFileTransferredReq struct {
	BangumiName string
	Season      int
}

type DeleteFileTransferredReq struct {
	OriginFile     string
	NewFile        string
	NewFileID      string
	SubscriptionID string
}
