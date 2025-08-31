package gin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SearchTV 搜索电视剧
// GET /apis/v1/meta/tvs?name=xxx
func (r *Router) SearchTVs(c *gin.Context) {
	name := c.Query("name")
	meta, err := r.metaParser.SearchTVs(c.Request.Context(), name)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, meta)
}

// SearchMovie 搜索电影
// GET /apis/v1/meta/movies?name=xxx
func (r *Router) SearchMovies(c *gin.Context) {
	name := c.Query("name")
	meta, err := r.metaParser.SearchMovies(c.Request.Context(), name)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, meta)
}

// GetTVMeta 根据TMDB ID获取电视剧元数据
// GET /apis/v1/meta/tv/:id
func (r *Router) GetTVMeta(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(c, err)
		return
	}

	meta, err := r.metaParser.ParseTV(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, meta)
}

// GetMovieMeta 根据TMDB ID获取电影元数据
// GET /apis/v1/meta/movie/:id
func (r *Router) GetMovieMeta(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(c, err)
		return
	}

	meta, err := r.metaParser.ParseMovie(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, meta)
}
