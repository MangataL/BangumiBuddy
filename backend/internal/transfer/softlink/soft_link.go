package softlink

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/MangataL/BangumiBuddy/internal/transfer"
)

func init() {
	transfer.RegisterFileTransfer("softlink", &softLink{})
}

func NewSoftLink() transfer.FileTransfer {
	return &softLink{}
}

type softLink struct{}

func (h *softLink) Transfer(ctx context.Context, src string, dst string) (originFile string, err error) {
	// 确保目标目录存在
	targetDir := filepath.Dir(dst)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0777); err != nil {
			return "", fmt.Errorf("创建目标目录失败: %w", err)
		}
	}
	if _, err := os.Stat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			return "", fmt.Errorf("删除目标文件失败: %w", err)
		}
	}
	if err := os.Symlink(src, dst); err != nil {
		return "", fmt.Errorf("创建软链接失败: %w", err)
	}
	return src, nil
}
