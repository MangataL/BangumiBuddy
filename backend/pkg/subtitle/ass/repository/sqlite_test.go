package repository

import (
	"path/filepath"
	"testing"

	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass/freetype"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpgradeAndRebuildFontNames(t *testing.T) {
	ctx := t.Context()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "fonts.db")))
	require.NoError(t, err)
	// 创建旧版本的表，验证升级、重复启动和重建前的名称查询。
	require.NoError(t, db.AutoMigrate(&fontMetaSchema{}))
	require.NoError(t, db.Create(&fontMetaSchema{FullName: "Old", ChineseFullName: "旧字体", Type: "ttf"}).Error)
	for range 2 {
		repo, err := New(db)
		require.NoError(t, err)
		stats, err := repo.Stats(ctx)
		require.NoError(t, err)
		require.True(t, stats.NeedsRebuild)
		found, err := repo.Find(ctx, ass.FindFontMetaReq{FullName: "旧字体"})
		require.NoError(t, err)
		require.Len(t, found, 1)
	}
	repo, err := New(db)
	require.NoError(t, err)
	first := ass.FontMeta{FullName: "Test", Type: ass.FontTypeTTF, Location: ass.FontLocation{Path: "first.ttf"},
		Aliases: []freetype.FontName{{Kind: freetype.NameIDFullName, Value: "日本語"}, {Kind: freetype.NameIDFamily, Value: "Семья"}, {Kind: freetype.NameIDFullName, Value: "ΛΟΓΟΣ"}}}
	second := ass.FontMeta{FullName: "Test", Type: ass.FontTypeTTF, Location: ass.FontLocation{Path: "second.ttf"},
		Aliases: []freetype.FontName{{Kind: freetype.NameIDFullName, Value: "別の名前"}}}
	require.NoError(t, repo.Replace(ctx, []ass.FontMeta{first, second}))
	stats, err := repo.Stats(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 1, stats.Total)
	require.False(t, stats.NeedsRebuild)
	for _, req := range []ass.FindFontMetaReq{{FullName: "日本語"}, {FamilyName: "семья"}, {FullName: "λογος"}} {
		found, err := repo.Find(ctx, req)
		require.NoError(t, err)
		require.Len(t, found, 1)
		require.Equal(t, "first.ttf", found[0].Location.Path)
	}
	for _, req := range []ass.FindFontMetaReq{{FullName: "別の名前"}, {FullName: "旧字体"}, {FullName: "Семья"}} {
		found, err := repo.Find(ctx, req)
		require.NoError(t, err)
		require.Empty(t, found)
	}
	// 在别名落库时模拟失败，旧索引仍完整可查。
	require.NoError(t, db.Exec("CREATE TRIGGER fail_alias BEFORE INSERT ON font_names BEGIN SELECT RAISE(ABORT,'test'); END").Error)
	require.Error(t, repo.Replace(ctx, []ass.FontMeta{second}))
	found, err := repo.Find(ctx, ass.FindFontMetaReq{FullName: "日本語"})
	require.NoError(t, err)
	require.Len(t, found, 1)
	require.NoError(t, repo.Replace(ctx, nil))
	stats, err = repo.Stats(ctx)
	require.NoError(t, err)
	require.Zero(t, stats.Total)
	require.False(t, stats.NeedsRebuild)
}

func TestTempFontNameOnlyResolvesToKeptFile(t *testing.T) {
	repo := ass.NewMemoryRepository()
	first := ass.FontMeta{FullName: "A", Aliases: []freetype.FontName{{Kind: freetype.NameIDFullName, Value: "旧名称"}}}
	last := ass.FontMeta{FullName: "A", Location: ass.FontLocation{Path: "last.ttf"}, Aliases: []freetype.FontName{{Kind: freetype.NameIDFullName, Value: "別名"}}}
	require.NoError(t, repo.Replace(t.Context(), []ass.FontMeta{first, last}))
	found, err := repo.Find(t.Context(), ass.FindFontMetaReq{FullName: "旧名称"})
	require.NoError(t, err)
	require.Empty(t, found)
	found, err = repo.Find(t.Context(), ass.FindFontMetaReq{FullName: "別名"})
	require.NoError(t, err)
	require.Equal(t, []ass.FontMeta{last}, found)
}

func TestFullNameBeatsLocalizedAlias(t *testing.T) {
	ctx := t.Context()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "fonts.db")))
	require.NoError(t, err)
	global, err := New(db)
	require.NoError(t, err)
	temp := ass.NewMemoryRepository()
	alias := ass.FontMeta{FullName: "Alias Face", Type: ass.FontTypeTTF, Aliases: []freetype.FontName{{Kind: freetype.NameIDFullName, Value: "Requested"}}}
	exact := ass.FontMeta{FullName: "Requested", Type: ass.FontTypeTTF}
	req := ass.FindFontReq{Font: ass.Font{FontName: "requested"}}
	for _, repo := range []ass.FontMetaRepository{temp, global} {
		require.NoError(t, repo.Replace(ctx, []ass.FontMeta{alias, exact}))
		set := ass.NewFontMetaSet("", repo)
		for range 32 {
			got, err := set.FindFontMeta(ctx, req)
			require.NoError(t, err)
			require.Equal(t, exact.FullName, got.FullName)
		}
	}
	require.NoError(t, temp.Replace(ctx, []ass.FontMeta{alias}))
	set := ass.NewFontMetaSet("", ass.NewTempFontMetaRepository(temp, global))
	got, err := set.FindFontMeta(ctx, req)
	require.NoError(t, err)
	require.Equal(t, exact.FullName, got.FullName)

	exact.Type = ass.FontTypeOTF
	require.NoError(t, global.Replace(ctx, []ass.FontMeta{exact}))
	got, err = set.FindFontMeta(ctx, req)
	require.NoError(t, err)
	require.Equal(t, alias.FullName, got.FullName)
	req.UseOTF = true
	got, err = set.FindFontMeta(ctx, req)
	require.NoError(t, err)
	require.Equal(t, exact.FullName, got.FullName)
}

func TestPostScriptNameBeatsLocalizedAlias(t *testing.T) {
	ctx := t.Context()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "fonts.db")))
	require.NoError(t, err)
	global, err := New(db)
	require.NoError(t, err)
	temp := ass.NewMemoryRepository()
	const name = "StudioSans-Regular"
	req := ass.FindFontReq{Font: ass.Font{FontName: name}, UseOTF: true}
	exact := ass.FontMeta{FullName: "Studio Sans Regular", PostScriptName: name, Type: ass.FontTypeTTF}
	alias := ass.FontMeta{FullName: "Other Face", Type: ass.FontTypeTTF}
	for _, kind := range []uint16{freetype.NameIDFullName, freetype.NameIDPostScript} {
		alias.Aliases = []freetype.FontName{{Kind: kind, Value: name}}
		for _, repo := range []ass.FontMetaRepository{temp, global} {
			require.NoError(t, repo.Replace(ctx, []ass.FontMeta{alias, exact}))
			got, err := ass.NewFontMetaSet("", repo).FindFontMeta(ctx, req)
			require.NoError(t, err)
			require.Equal(t, exact.FullName, got.FullName)
		}
	}

	// 种子内的 TTF 别名和全局库中 OTF 的 PostScript 名重合。
	exact.Type = ass.FontTypeOTF
	require.NoError(t, global.Replace(ctx, []ass.FontMeta{exact}))
	require.NoError(t, temp.Replace(ctx, []ass.FontMeta{alias}))
	set := ass.NewFontMetaSet("", ass.NewTempFontMetaRepository(temp, global))
	got, err := set.FindFontMeta(ctx, req)
	require.NoError(t, err)
	require.Equal(t, exact.FullName, got.FullName)
}
