package parser

import (
	"context"
	"testing"
)

func TestParseReleaseGroup(t *testing.T) {
	parser := NewParser(nil)
	releaseGroup, err := parser.ParseReleaseGroup(context.Background(), "[新Sub&萌樱字幕组]青春猪头少年不会梦到娇伶妹 Seishun Buta Yarou wa Odekake Sister no Yume o Minai[BDRip][1080p][HEVC+FLAC][简繁内封].mkv.mkv")
	if err != nil {
		t.Fatalf("ParseReleaseGroup failed: %v", err)
	}
	t.Logf("ParseReleaseGroup success: %s", releaseGroup)
}