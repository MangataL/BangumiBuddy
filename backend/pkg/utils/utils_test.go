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

func TestParseEpisodeWithLocation(t *testing.T) {
	testCases := []struct {
		name string
		location string
		expected int
	}{
		{
			name: "Ranma.1-2.2024.S02E05.Its.Fast.or.Its.Free.1080p.NF.WEB-DL.DUAL.AAC2.0.H.264-VARYG.mkv",
			location: "S02E{ep}",
			expected: 5,
		},
		{
			name: "Fujimoto Tatsuki 17-26 - 08 [WebRip 1080p HEVC-10bit AAC EAC3 ASSx2].mkv",
			location: "Fujimoto Tatsuki 17-26 - {ep}",
			expected: 8,
		},
	}

	for _, testCase := range testCases {
		episode, err := ParseEpisodeWithLocation(testCase.name, testCase.location)
		if err != nil {
			t.Fatalf("ParseEpisodeWithLocation failed: %v", err)
		}
		assert.Equal(t, testCase.expected, episode)
	}
}