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

type taskSidecars struct {
	files    []string
	err      error
	prepared map[string]preparedSubtitle
}

func needsFontSubset(file string) bool {
	ext := strings.ToLower(filepath.Ext(file))
	return (ext == ".ass" || ext == ".ssa") && notSubsetFile(filepath.Base(file))
}

// Inspect exactly the files the transfer will process, before replacing anything in the library.
func prepareTaskSubtitles(ctx context.Context, root string, media []taskMedia, subsetter subtitle.Subsetter) subtitle.MissingFonts {
	missing := make(subtitle.MissingFonts)
	prepared := make(map[string]preparedSubtitle)
	for i := range media {
		files, readErr := utils.FindSameBaseFiles(media[i].meta.FilePath)
		media[i].meta.sidecars = &taskSidecars{files: files, err: readErr, prepared: prepared}
		for _, file := range files {
			if !needsFontSubset(file) {
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
