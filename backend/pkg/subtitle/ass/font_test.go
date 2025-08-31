package ass_test

import (
	"context"
	"testing"

	sqlitedriver "github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass/repository"
)

func TestFontSubset(t *testing.T) {
	db, err := gorm.Open(sqlitedriver.Open("/Users/ljc/mygo/github.com/BangumiBuddy/test/data/data.db"), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)
	fontMetaSet := ass.NewFontMetaSet("/Users/ljc/mygo/github.com/BangumiBuddy/test/data/fonts", repository.New(db))
	subsetter := ass.NewSubsetter(fontMetaSet, ass.FontSubsetterConfig{
		UseSystemFontsDir: false,
		GenerateNewFile:   true,
		CheckGlyphs:       true,
		UseOTF:            true,
		UseSimilarFont:    true,
		CoverExistSubFont: true,
	})

	ctx := context.Background()
	// require.NoError(t, subsetter.InitFontMetaSet(ctx))
	outputPath, err := subsetter.SubsetFont(ctx, "/Users/ljc/bangumi/download/tv/[Nekomoe kissaten&VCB-Studio] Sono Bisque Doll wa Koi wo Suru [Ma10p_1080p]/[Nekomoe kissaten&VCB-Studio] Sono Bisque Doll wa Koi wo Suru [01][Ma10p_1080p][x265_flac].SC.ass")
	require.NoError(t, err)
	require.NotEmpty(t, outputPath)
}
