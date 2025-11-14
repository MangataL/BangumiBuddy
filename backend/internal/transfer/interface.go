package transfer

import (
	"context"
)

type Interface interface {
	Transfer(ctx context.Context, hash string) error
	// DeleteTransferFile 删除转移文件
	DeleteTransferFile(ctx context.Context, filePath string) error
	// GetTransferFile 获取转移文件，返回完整文件路径，如果文件不存在，返回 ErrTransferFileNotFound
	GetTransferFile(ctx context.Context, filePath string) (string, error)
	// DeleteTransferCache 删除转移缓存
	DeleteTransferCache(ctx context.Context, req DeleteFileTransferredReq) error
	// ParseEpisode 解析集数
	ParseEpisode(ctx context.Context, fileName string, epLocation string) (int, error)
}
