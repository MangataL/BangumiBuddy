package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChineseNumberToInt(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "零",
			input:   "零",
			want:    0,
			wantErr: assert.NoError,
		},
		{
			name:    "两位数-十",
			input:   "十",
			want:    10,
			wantErr: assert.NoError,
		},
		{
			name:    "两位数-十二",
			input:   "十二",
			want:    12,
			wantErr: assert.NoError,
		},
				{
			name:    "两位数-一十二",
			input:   "一十二",
			want:    12,
			wantErr: assert.NoError,
		},
		{
			name:    "两位数-二十",
			input:   "二十",
			want:    20,
			wantErr: assert.NoError,
		},
		{
			name:    "两位数-二十五",
			input:   "二十五",
			want:    25,
			wantErr: assert.NoError,
		},
		{
			name:    "百位-一百零二",
			input:   "一百零二",
			want:    102,
			wantErr: assert.NoError,
		},
		{
			name:    "千位-一千零一",
			input:   "一千零一",
			want:    1001,
			wantErr: assert.NoError,
		},
		{
			name:    "万位-一万零三",
			input:   "一万零三",
			want:    10003,
			wantErr: assert.NoError,
		},
		{
			name:    "万位-十二万三千四百五十六",
			input:   "十二万三千四百五十六",
			want:    123456,
			wantErr: assert.NoError,
		},
		{
			name:    "亿位-一亿零三万零五",
			input:   "一亿零三万零五",
			want:    100030005,
			wantErr: assert.NoError,
		},
		{
			name:    "两百三",
			input:   "两百三",
			want:    203,
			wantErr: assert.NoError,
		},
		{
			name:    "空字符串",
			input:   "",
			want:    0,
			wantErr: assert.Error,
		},
		{
			name:    "包含非中文数字",
			input:   "第十二话",
			want:    0,
			wantErr: assert.Error,
		},
		{
			name:    "包含英文",
			input:   "一百A",
			want:    0,
			wantErr: assert.Error,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ChineseNumberToInt(tc.input)

			tc.wantErr(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
