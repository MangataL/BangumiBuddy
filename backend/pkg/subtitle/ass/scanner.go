package ass

import (
	"bufio"
	"io"
)

// newBufferScanner 创建一个支持大 token 的 Scanner
// 设置最大 token 大小为 10MB，足以处理嵌入字体的长行
func newBufferScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	// 设置初始缓冲区大小为 64KB（默认值）
	// 设置最大 token 大小为 10MB，足以处理嵌入字体的长行
	const maxTokenSize = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxTokenSize)
	return scanner
}
