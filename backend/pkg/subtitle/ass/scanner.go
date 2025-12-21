package ass

import (
	"bufio"
	"io"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// newBufferScanner 创建一个支持大 token 的 Scanner
// 设置最大 token 大小为 10MB，足以处理嵌入字体的长行
func newBufferScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(normalizeTextReader(r))
	// 设置初始缓冲区大小为 64KB（默认值）
	// 设置最大 token 大小为 10MB，足以处理嵌入字体的长行
	const maxTokenSize = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxTokenSize)
	return scanner
}

// normalizeTextReader 处理带 BOM 的 UTF-8/UTF-16 字幕文件，统一输出 UTF-8。
func normalizeTextReader(r io.Reader) io.Reader {
	br := bufio.NewReader(r)

	bom2, _ := br.Peek(2)
	if len(bom2) >= 2 {
		if bom2[0] == 0xFF && bom2[1] == 0xFE {
			return transform.NewReader(br, unicode.UTF16(unicode.LittleEndian, unicode.ExpectBOM).NewDecoder())
		}
		if bom2[0] == 0xFE && bom2[1] == 0xFF {
			return transform.NewReader(br, unicode.UTF16(unicode.BigEndian, unicode.ExpectBOM).NewDecoder())
		}
	}

	bom3, _ := br.Peek(3)
	if len(bom3) >= 3 && bom3[0] == 0xEF && bom3[1] == 0xBB && bom3[2] == 0xBF {
		_, _ = br.Discard(3)
	}

	return br
}
