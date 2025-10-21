package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseReleaseGroup(t *testing.T) {
	parser := NewParser(nil)
	releaseGroup, err := parser.ParseReleaseGroup(context.Background(), "[新Sub&萌樱字幕组]青春猪头少年不会梦到娇伶妹 Seishun Buta Yarou wa Odekake Sister no Yume o Minai[BDRip][1080p][HEVC+FLAC][简繁内封].mkv.mkv")
	if err != nil {
		t.Fatalf("ParseReleaseGroup failed: %v", err)
	}
	assert.Equal(t, "新Sub&萌樱字幕组", releaseGroup)
}

func TestParseFile(t *testing.T) {
	parser := NewParser(nil)
	season, episode, err := parser.ParseFile(context.Background(), "[LoliHouse] Food Court de, Mata Ashita. [01-06][WebRip 1080p HEVC-10bit AAC]/[LoliHouse] Food Court de, Mata Ashita. - 02 [WebRip 1080p HEVC-10bit AAC SRTx2].mkv")
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}
	assert.Equal(t, 1, season)
	assert.Equal(t, 2, episode)
}
