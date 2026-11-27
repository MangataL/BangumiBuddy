package subtitle

import "context"

// PreparedSubset writes the subtitle after its fonts have been resolved.
type PreparedSubset func() (string, error)

type Subsetter interface {
	PrepareSubset(ctx context.Context, filePath string) (PreparedSubset, error)
	SubsetFont(ctx context.Context, filePath string) (string, error)
	InitFontMetaSet(ctx context.Context) error
	GetFontMetaSetStats(ctx context.Context) (FontMetaSetStats, error)
	UsingTempFontDir(ctx context.Context, fontDir string) (Subsetter, error)
	ListFonts(ctx context.Context, req ListFontsReq) ([]Font, error)
}
