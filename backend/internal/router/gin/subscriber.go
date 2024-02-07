package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

const (
	boolTrue = "true"
)

// ParseRSS parses RSS link and returns bangumi info
// GET /apis/v1/bangumis/rss?link=xxx
func (r *Router) ParseRSS(c *gin.Context) {
	link := c.Query("link")

	rsp, err := r.subscriber.ParseRSS(c.Request.Context(), link)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, rsp)
}

// Subscribe 处理订阅番剧请求
// POST /apis/v1/bangumis
func (r *Router) Subscribe(c *gin.Context) {
	var req subscriber.SubscribeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}

	bangumi, err := r.subscriber.Subscribe(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusCreated, bangumi)
}

// GetBangumi 获取番剧信息
// GET /apis/v1/bangumis/:id
func (r *Router) GetBangumi(c *gin.Context) {
	id := c.Param("id")
	bangumi, err := r.subscriber.Get(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, bangumi)
}

// ListBangumisBase 列出所有订阅
// GET /apis/v1/bangumis/base?fuzz_name=xxx&active=true
func (r *Router) ListBangumisBase(c *gin.Context) {
	var req subscriber.ListBangumiReq
	if err := c.ShouldBindQuery(&req); err != nil {
		writeError(c, err)
		return
	}
	bangumiList, err := r.web.ListBangumis(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, bangumiList)
}

// ListBangumis 列出所有订阅
// GET /apis/v1/bangumis?fuzz_name=xxx&active=true
func (r *Router) ListBangumis(c *gin.Context) {
	var req subscriber.ListBangumiReq
	if err := c.ShouldBindQuery(&req); err != nil {
		writeError(c, err)
		return
	}

	bangumiList, err := r.subscriber.List(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, bangumiList)
}

// UpdateSubscription 更新订阅
// PUT /apis/v1/bangumis/:id
func (r *Router) UpdateSubscription(c *gin.Context) {
	id := c.Param("id")

	var req subscriber.UpdateSubscribeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	req.SubscriptionID = id

	err := r.subscriber.UpdateSubscription(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

// DeleteSubscription 删除订阅
// DELETE /apis/v1/bangumis/:id?delete_files=true
func (r *Router) DeleteSubscription(c *gin.Context) {
	id := c.Param("id")
	deleteFiles := c.Query("delete_files") == boolTrue
	err := r.web.DeleteSubscription(c.Request.Context(), id, deleteFiles)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// GetRSSMatch 获取RSS匹配
// GET /apis/v1/bangumis/:id/rss_match
func (r *Router) GetRSSMatch(c *gin.Context) {
	id := c.Param("id")
	match, err := r.subscriber.GetRSSMatch(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, match)
}

type markRSSRecordReq struct {
	GUIDs     []string `json:"guids"`
	Processed bool     `json:"processed"`
}

// MarkRSSRecord 标记RSS记录
// POST /apis/v1/bangumis/:id/rss_match
func (r *Router) MarkRSSRecord(c *gin.Context) {
	id := c.Param("id")
	var req markRSSRecordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	err := r.subscriber.MarkRSSRecord(c.Request.Context(), subscriber.MarkRSSRecordReq{
		SubscriptionID: id,
		GUIDs:          req.GUIDs,
		Processed:      req.Processed,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// HandleBangumiSubscription 处理番剧订阅
// POST /apis/v1/bangumis/:id/download
func (r *Router) HandleBangumiSubscription(c *gin.Context) {
	id := c.Param("id")
	err := r.subscriber.HandleBangumiSubscription(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// GetBangumiTorrents 获取番剧的torrents
// GET /apis/v1/bangumis/:id/torrents
func (r *Router) GetBangumiTorrents(c *gin.Context) {
	id := c.Param("id")
	torrents, err := r.web.GetBangumiTorrents(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, torrents)
}

// GetSubscriptionCalendar 获取订阅日历
// GET /apis/v1/bangumis/calendar
func (r *Router) GetSubscriptionCalendar(c *gin.Context) {
	calendar, err := r.subscriber.GetSubscriptionCalendar(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, calendar)
}
