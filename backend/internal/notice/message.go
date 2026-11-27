package notice

// SplitMessage 按通知渠道的长度单位分段，避免截断多字节字符。
func SplitMessage(text string, limit int, runeSize func(rune) int) []string {
	var parts []string
	start, length := 0, 0
	for i, r := range text {
		size := runeSize(r)
		if length+size > limit {
			parts = append(parts, text[start:i])
			start, length = i, 0
		}
		length += size
	}
	if start < len(text) {
		parts = append(parts, text[start:])
	}
	return parts
}
