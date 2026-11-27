package notice

import (
	"strings"
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

func TestSplitTaskMessage(t *testing.T) {
	line := "[字幕组]_🥔_日本語.ass\n"
	text := strings.Repeat("字", 4095) + "🥔" + strings.Repeat(line, 500)
	parts := SplitMessage(text, 4096, utf16.RuneLen)
	require.Equal(t, text, strings.Join(parts, ""))
	require.Greater(t, len(parts), 1)
	for _, part := range parts {
		require.LessOrEqual(t, len(utf16.Encode([]rune(part))), 4096)
	}
	parts = SplitMessage(text, 2400, utf8.RuneLen)
	require.Equal(t, text, strings.Join(parts, ""))
	for _, part := range parts {
		require.LessOrEqual(t, len(part), 2400)
		require.True(t, utf8.ValidString(part))
	}
}
