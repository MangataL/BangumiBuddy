package scrape

import (
	"context"
)

type Interface interface {
	// AddMetadataFillTask 添加元数据填充任务
	AddMetadataFillTask(ctx context.Context, req AddMetadataFillTaskReq) error
	// Enable 是否启用元数据填充
	Enable() bool
}
