package scrape

import (
	"context"
)

type Interface interface {
	// AddMetadataFillTask 添加元数据填充任务
	AddMetadataFillTask(ctx context.Context, req AddMetadataFillTaskReq) error
	// Enable 是否启用元数据填充
	Enable() bool
	// ListTasks 获取刮削任务列表
	ListTasks(ctx context.Context) ([]MetadataCheckTask, error)
	// TriggerScrape 触发单个任务刮削
	TriggerScrape(ctx context.Context, id uint) error
	// TriggerScrapeAll 触发全部任务刮削
	TriggerScrapeAll(ctx context.Context) error
}
