package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
)

// InitSubtitleMetaSet 初始化字幕元数据集
// POST /apis/v1/subtitle/meta-sets
func (r *Router) InitSubtitleMetaSet(c *gin.Context) {
	err := r.subtitleSubsetter.InitFontMetaSet(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// CountSubtitleMetaSet 统计字幕元数据集数量
// GET /apis/v1/subtitle/meta-sets/stats
func (r *Router) GetSubtitleMetaSetStats(c *gin.Context) {
	stats, err := r.subtitleSubsetter.GetFontMetaSetStats(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

// ListFonts 列出所有字体
// GET /apis/v1/subtitle/meta-sets/fonts?page=1&page_size=10
func (r *Router) ListFonts(c *gin.Context) {
	var req subtitle.ListFontsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		writeError(c, err)
		return
	}
	fonts, err := r.subtitleSubsetter.ListFonts(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, fonts)
}
