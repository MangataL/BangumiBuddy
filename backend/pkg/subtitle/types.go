package subtitle

type FontMetaSetStats struct {
	Total    int64 `json:"total"`
	InitDone bool  `json:"initDone"`
}

const (
	SubtitleSubsetExt = ".subset"
)

// ListFontsReq 查询字体请求
type ListFontsReq struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"pageSize"`
}

// Font 字体信息
type Font struct {
	FamilyName            string `json:"familyName"`
	FullName              string `json:"fullName"`
	PostScriptName        string `json:"postScriptName"`
	ChineseFamilyName     string `json:"chineseFamilyName"`
	ChineseFullName       string `json:"chineseFullName"`
	ChinesePostScriptName string `json:"chinesePostScriptName"`
	BoldWeight            int    `json:"boldWeight"` // 字重
	Italic                bool   `json:"italic"`     // 是否斜体
	FontFileName          string `json:"fontFileName"`
}
