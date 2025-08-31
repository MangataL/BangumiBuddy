package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListDirs 列出目录
// GET /apis/v1/utils/dirs?path=xxx
func (r *Router) ListDirs(c *gin.Context) {
	path := c.Query("path")
	dirs, err := r.web.ListDirs(c.Request.Context(), path)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dirs)
}
