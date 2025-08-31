package subtitle

import "context"

type Subsetter interface {
	SubsetFont(ctx context.Context, filePath string) (string, error)
	InitFontMetaSet(ctx context.Context) error
	GetFontMetaSetStats(ctx context.Context) (FontMetaSetStats, error)
	UsingTempFontDir(ctx context.Context, fontDir string) (Subsetter, error)
}
