package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/MangataL/BangumiBuddy/internal/discovery"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

func (r *Router) ListDiscoveryBangumis(c *gin.Context) {
	var req discovery.ListBangumiReq
	if err := c.ShouldBindQuery(&req); err != nil {
		writeError(c, err)
		return
	}
	bangumis, err := r.discovery.ListBangumis(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, bangumis)
}

func (r *Router) SearchDiscovery(c *gin.Context) {
	var req discovery.SearchReq
	if err := c.ShouldBindQuery(&req); err != nil {
		writeError(c, err)
		return
	}
	resp, err := r.discovery.Search(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (r *Router) BatchDiscoveryReleaseGroups(c *gin.Context) {
	var req discovery.BatchReleaseGroupsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	resp, err := r.discovery.BatchReleaseGroups(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (r *Router) GetDiscoveryBangumi(c *gin.Context) {
	detail, err := r.discovery.GetBangumiDetail(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (r *Router) ParseDiscoveryCandidateRSS(c *gin.Context) {
	var req discovery.CandidateRSSReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	rssLink, err := r.discovery.BuildCandidateRSS(req)
	if err != nil {
		writeError(c, err)
		return
	}
	resp, err := r.subscriber.ParseRSS(c.Request.Context(), subscriber.ParserRSSReq{
		RSSLink: rssLink,
		TMDBID:  req.TMDBID,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
