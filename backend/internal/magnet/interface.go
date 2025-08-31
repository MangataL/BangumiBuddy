package magnet

import "context"

// Interface 种子下载功能接口
type Interface interface {
	AddTask(ctx context.Context, req AddTaskReq) (Task, error)
	InitTask(ctx context.Context, taskID string, tmdbID int) (Task, error)
	UpdateTask(ctx context.Context, req UpdateTaskReq) error
	AddSubtitles(ctx context.Context, req AddSubtitlesReq) error
	ListTasks(ctx context.Context, req ListTasksReq) ([]Task, int, error)
	GetTask(ctx context.Context, taskID string) (Task, error)
	DeleteTask(ctx context.Context, taskID string) error
}
