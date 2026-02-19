package gin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListScraperTasks 获取刮削任务列表
// GET /apis/v1/scraper/tasks
func (r *Router) ListScraperTasks(ctx *gin.Context) {
	tasks, err := r.scraper.ListTasks(ctx.Request.Context())
	if err != nil {
		writeError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, tasks)
}

// TriggerScrapeTask 触发单个任务刮削
// POST /apis/v1/scraper/tasks/:id/scrape
func (r *Router) TriggerScrapeTask(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}

	if err := r.scraper.TriggerScrape(ctx.Request.Context(), uint(id)); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}

// TriggerScrapeAll 触发全部任务刮削
// POST /apis/v1/scraper/tasks/scrape
func (r *Router) TriggerScrapeAll(ctx *gin.Context) {
	if err := r.scraper.TriggerScrapeAll(ctx.Request.Context()); err != nil {
		writeError(ctx, err)
		return
	}
	ctx.Status(http.StatusOK)
}
