package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
