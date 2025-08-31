package subtitle

type FontMetaSetStats struct {
	Total    int64 `json:"total"`
	InitDone bool  `json:"initDone"`
}

const (
	SubtitleSubsetExt = ".subset"
)