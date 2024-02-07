package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/MangataL/BangumiBuddy/internal/web"
)

// DeleteTorrent 删除torrent
// DELETE /api/v1/torrents/:hash?delete_origin_files=true&delete_transfer_files=true
func (r *Router) DeleteTorrent(c *gin.Context) {
	hash := c.Param("hash")
	deleteOriginFiles := c.Query("delete_origin_files") == boolTrue
	err := r.web.DeleteTorrent(c.Request.Context(), web.DeleteTorrentReq{
		Hash:               hash,
		DeleteOriginFiles:  deleteOriginFiles,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// Transfer 转移文件
// POST /api/v1/torrents/:hash/transfer
func (r *Router) Transfer(c *gin.Context) {
	hash := c.Param("hash")
	err := r.transfer.Transfer(c.Request.Context(), hash)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// GetTorrentFiles 获取torrent文件
// GET /api/v1/torrents/:hash/files
func (r *Router) GetTorrentFiles(c *gin.Context) {
	hash := c.Param("hash")
	files, err := r.web.GetTorrentFiles(c.Request.Context(), hash)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, files)
}

// ListRecentUpdatedTorrents 获取最近更新的torrent
// GET /api/v1/torrents/recent?start_time=2025-04-12T00:00:00Z&end_time=2025-04-12T23:59:59Z&page=1&page_size=10
func (r *Router) ListRecentUpdatedTorrents(c *gin.Context) {
	var req web.ListRecentUpdatedTorrentsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		writeError(c, err)
		return
	}
	resp, err := r.web.ListRecentUpdatedTorrents(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
