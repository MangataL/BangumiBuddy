package subtitle

import "context"

// PreparedSubset uses subsetPath when configured to generate a separate file.
type PreparedSubset func(subsetPath string) (string, error)

type Subsetter interface {
	PrepareSubset(ctx context.Context, filePath string) (PreparedSubset, error)
	SubsetFont(ctx context.Context, filePath, subsetPath string) (string, error)
	InitFontMetaSet(ctx context.Context) error
	GetFontMetaSetStats(ctx context.Context) (FontMetaSetStats, error)
	UsingTempFontDir(ctx context.Context, fontDir string) (Subsetter, error)
	ListFonts(ctx context.Context, req ListFontsReq) ([]Font, error)
}
