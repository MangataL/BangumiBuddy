package transfer

import (
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
	"github.com/MangataL/BangumiBuddy/pkg/utils"
)

type preparedSubtitle struct {
	write subtitle.PreparedSubset
	err   error
}

func (t *Transfer) subsetSubtitle(ctx context.Context, meta Meta, file, subsetPath string) (string, error) {
	if prepared, ok := meta.preparedSubsets[file]; ok {
		if prepared.err != nil {
			return "", prepared.err
		}
		return prepared.write(subsetPath)
	}
	subsetter := meta.FontSubsetter
	if subsetter == nil {
		subsetter = t.fontSubsetter
	}
	return subsetter.SubsetFont(ctx, file, subsetPath)
}

func prepareTaskSubtitles(ctx context.Context, root string, media []taskMedia, subsetter subtitle.Subsetter) subtitle.MissingFonts {
	missing := make(subtitle.MissingFonts)
	prepared := make(map[string]preparedSubtitle)
	for i := range media {
		media[i].meta.preparedSubsets = prepared
		files, err := utils.FindSameBaseFiles(media[i].meta.FilePath)
		if err != nil {
			continue // 转移该视频时会返回目录读取错误。
		}
		mediaBase := utils.GetFileBaseName(media[i].meta.FilePath)
		for _, file := range files {
			if !needsFontSubset(strings.TrimPrefix(file, mediaBase)) {
				continue
			}
			if _, seen := prepared[file]; seen {
				continue
			}
			write, err := subsetter.PrepareSubset(ctx, file)
			prepared[file] = preparedSubtitle{write, err}
			var absent subtitle.MissingFonts
			if errors.As(err, &absent) {
				source, relErr := filepath.Rel(root, file)
				if relErr != nil {
					source = file
				}
				for font := range absent {
					missing[font] = append(missing[font], source)
				}
			}
		}
	}
	return missing
}
