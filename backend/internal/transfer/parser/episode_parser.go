package parser

import (
	"context"
	"path/filepath"
	"strconv"

	"github.com/MangataL/BangumiBuddy/internal/transfer"
	"github.com/MangataL/BangumiBuddy/pkg/errs"
	"github.com/nssteinbrenner/anitogo"
)

type episode struct {}

func NewEpisodeParser() transfer.EpisodeParser {
	return &episode{}
}

func (e *episode) Parse(ctx context.Context, fileName string) (int, error) {
	meta := anitogo.Parse(filepath.Base(fileName), anitogo.DefaultOptions)
	if len(meta.EpisodeNumber) == 0 {
		return 0, errs.NewBadRequest("无法识别集数信息")
	}
	episode, err := strconv.Atoi(meta.EpisodeNumber[0])
	if err != nil {
		return 0, errs.NewBadRequest("无法识别集数信息")
	}
	return episode, nil
}