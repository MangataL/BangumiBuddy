package gin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/MangataL/BangumiBuddy/internal/magnet"
	"github.com/MangataL/BangumiBuddy/internal/web"
)

// AddTorrentTask 添加磁力任务
// POST /apis/v1/magnets
func (r *Router) AddMagnetTask(c *gin.Context) {
	var req magnet.AddTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}

	task, err := r.magnet.AddTask(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// InitMagnetTask 初始化磁力任务
// PUT /apis/v1/magnet/init/:id?tmdb_id=123
func (r *Router) InitMagnetTask(c *gin.Context) {
	taskID := c.Param("id")
	var tmdbID int
	tmdbIDParam := c.Query("tmdb_id")
	if tmdbIDParam != "" {
		var err error
		tmdbID, err = strconv.Atoi(tmdbIDParam)
		if err != nil {
			writeError(c, err)
			return
		}
	}
	task, err := r.magnet.InitTask(c.Request.Context(), taskID, tmdbID)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateMagnetTask 更新磁力任务
// PUT /apis/v1/magnets/:id
func (r *Router) UpdateMagnetTask(c *gin.Context) {
	taskID := c.Param("id")
	var req magnet.UpdateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	req.TaskID = taskID

	err := r.magnet.UpdateTask(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusOK)
}

// ListMagnetTasks 列出磁力任务
// GET /apis/v1/magnets?page=1&page_size=10
func (r *Router) ListMagnetTasks(c *gin.Context) {
	var req web.ListMagnetTasksReq
	if err := c.ShouldBindQuery(&req); err != nil {
		writeError(c, err)
		return
	}

	resp, err := r.web.ListMagnetTasks(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMagnetTask 获取磁力任务
// GET /apis/v1/magnet/:id
func (r *Router) GetMagnetTask(c *gin.Context) {
	taskID := c.Param("id")
	task, err := r.web.GetMagnetTask(c.Request.Context(), taskID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

// DeleteMagnetTask 删除磁力任务
// DELETE /apis/v1/magnets/:id?delete_files=true
func (r *Router) DeleteMagnetTask(c *gin.Context) {
	taskID := c.Param("id")
	deleteFiles := c.Query("delete_files") == boolTrue
	err := r.web.DeleteMagnetTask(c.Request.Context(), taskID, deleteFiles)
	if err != nil {
		writeError(c, err)
		return
	}
}

// AddSubtitles 添加字幕
// POST /apis/v1/magnets/subtitles
func (r *Router) AddSubtitles(c *gin.Context) {
	var req magnet.AddSubtitlesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, r.magnet.AddSubtitles(c.Request.Context(), req))
}

// PreviewAddSubtitles 预览添加字幕
// POST /apis/v1/magnets/:id/subtitles/preview
func (r *Router) PreviewAddSubtitles(c *gin.Context) {
	taskID := c.Param("id")
	var req magnet.PreviewAddSubtitlesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, err)
		return
	}
	req.TaskID = taskID
	resp, err := r.magnet.PreviewAddSubtitles(c.Request.Context(), req)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// FindTaskSimilarFiles 查找任务中相似的文件
// GET /apis/v1/magnets/:id/files?similar_file_path=xxx
func (r *Router) FindTaskSimilarFiles(c *gin.Context) {
	taskID := c.Param("id")
	filePath := c.Query("similar_file_path")
	files, err := r.magnet.FindTaskSimilarFiles(c.Request.Context(), taskID, filePath)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, files)
}