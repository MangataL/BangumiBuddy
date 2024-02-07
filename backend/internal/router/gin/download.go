package gin

import (
	"fmt"
	"net/http"

	"github.com/autobrr/go-qbittorrent"
	"github.com/gin-gonic/gin"
)

type checkQBittorrentConnectionReq struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// CheckQBittorrentConnection 检查qBittorrent连通性
// POST /apis/v1/downloader/qbittorrent/check
func (r *Router) CheckQBittorrentConnection(c *gin.Context) {
	var req checkQBittorrentConnectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}

	if err := r.checkQBittorrentConnection(qbittorrent.Config{
		Host:     req.Host,
		Username: req.Username,
		Password: req.Password,
	}); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (r *Router) checkQBittorrentConnection(config qbittorrent.Config) error {
	client := qbittorrent.NewClient(config)
	if err := client.Login(); err != nil {
		return fmt.Errorf("qbittorrent 登录失败，请检查配置的账号密码是否正确: %w", err)
	}
	return nil
}	