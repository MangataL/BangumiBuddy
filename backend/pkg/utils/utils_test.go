package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEpisodeNameInvalid(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected bool
	}{
		{
			name:     "有效标题",
			title:    "东京BLADE",
			expected: false,
		},
		{
			name:     "无效标题 - 第x集单空格",
			title:    "第 12 集",
			expected: true,
		},
		{
			name:     "无效标题 - 第x集无空格",
			title:    "第12集",
			expected: true,
		},
		{
			name:     "无效标题 - 第x集多空格",
			title:    "第  12  集",
			expected: true,
		},
		{
			name:     "空标题",
			title:    "",
			expected: true,
		},
		{
			name:     "无效标题 - SxEx格式",
			title:    "ABC S01E01",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EpisodeNameInvalid(tt.title)
			assert.Equal(t, tt.expected, result)
		})
	}
}