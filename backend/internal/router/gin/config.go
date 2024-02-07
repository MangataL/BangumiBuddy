package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/MangataL/BangumiBuddy/internal/downloader"
	downloadadapter "github.com/MangataL/BangumiBuddy/internal/downloader/adapter"
	"github.com/MangataL/BangumiBuddy/internal/meta/tmdb"
	noticeadapter "github.com/MangataL/BangumiBuddy/internal/notice/adapter"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
	"github.com/MangataL/BangumiBuddy/internal/transfer"
)

// GetDownloaderConfig 获取下载器配置
// GET /apis/v1/config/download/downloader
func (r *Router) GetDownloaderConfig(ctx *gin.Context) {
	config, err := r.repo.GetDownloaderConfig()
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, config)
}

// SetDownloaderConfig 设置下载器配置
// PUT /apis/v1/config/download/downloader
func (r *Router) SetDownloaderConfig(ctx *gin.Context) {
	var config downloadadapter.Config
	if err := ctx.ShouldBindJSON(&config); err != nil {
		writeError(ctx, err)
		return
	}
	if err := r.repo.SetDownloaderConfig(&config); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}

// GetDownloadManagerConfig 获取下载管理器配置
// GET /apis/v1/config/download/manager
func (r *Router) GetDownloadManagerConfig(ctx *gin.Context) {
	config, err := r.repo.GetDownloadManagerConfig()
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, config)
}

// SetDownloadManagerConfig 设置下载管理器配置
// PUT /apis/v1/config/download/manager
func (r *Router) SetDownloadManagerConfig(ctx *gin.Context) {
	var config downloader.Config
	if err := ctx.ShouldBindJSON(&config); err != nil {
		writeError(ctx, err)
		return
	}
	if err := r.repo.SetDownloadManagerConfig(&config); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}

// GetTMDBConfig 获取TMDB配置
// GET /apis/v1/config/tmdb
func (r *Router) GetTMDBConfig(ctx *gin.Context) {
	config, err := r.repo.GetTMDBConfig()
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, config)
}

// SetTMDBConfig 设置TMDB配置
// PUT /apis/v1/config/tmdb
func (r *Router) SetTMDBConfig(ctx *gin.Context) {
	var config tmdb.Config
	if err := ctx.ShouldBindJSON(&config); err != nil {
		writeError(ctx, err)
		return
	}
	if err := r.repo.SetTMDBConfig(&config); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}

// GetSubscriberConfig 获取订阅器配置
// GET /apis/v1/config/subscriber
func (r *Router) GetSubscriberConfig(ctx *gin.Context) {
	config, err := r.repo.GetSubscriberConfig()
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, config)
}

// SetSubscriberConfig 设置订阅器配置
// PUT /apis/v1/config/subscriber
func (r *Router) SetSubscriberConfig(ctx *gin.Context) {
	var config subscriber.Config
	if err := ctx.ShouldBindJSON(&config); err != nil {
		writeError(ctx, err)
		return
	}
	if err := r.repo.SetSubscriberConfig(&config); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}

// GetTransferConfig 获取文件转移配置
// GET /apis/v1/config/transfer
func (r *Router) GetTransferConfig(ctx *gin.Context) {
	config, err := r.repo.GetTransferConfig()
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, config)
}

// SetTransferConfig 设置文件转移配置
// PUT /apis/v1/config/transfer
func (r *Router) SetTransferConfig(ctx *gin.Context) {
	var config transfer.Config
	if err := ctx.ShouldBindJSON(&config); err != nil {
		writeError(ctx, err)
		return
	}
	if err := r.repo.SetTransferConfig(&config); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}

// GetNoticeConfig 获取通知配置
// GET /apis/v1/config/notice
func (r *Router) GetNoticeConfig(ctx *gin.Context) {
	config, err := r.repo.GetNoticeConfig()
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, config)
}

// SetNoticeConfig 设置通知配置
// PUT /apis/v1/config/notice
func (r *Router) SetNoticeConfig(ctx *gin.Context) {
	var config noticeadapter.Config
	if err := ctx.ShouldBindJSON(&config); err != nil {
		writeError(ctx, err)
		return
	}
	if err := r.repo.SetNoticeConfig(&config); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}


