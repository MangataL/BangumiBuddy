package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// 自动迁移表结构
	err = db.AutoMigrate(&bangumiSchema{})
	require.NoError(t, err)

	return db
}

func TestRepository_SaveAndGet(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	// 测试数据
	bangumi := subscriber.Bangumi{
		Name:          "测试番剧",
		RSSLink:       "https://example.com/rss",
		Active:        true,
		IncludeRegs:   []string{".*1080p.*", ".*HEVC.*"},
		ExcludeRegs:   []string{".*720p.*", ".*预告.*"},
		Priority:      10,
		EpisodeOffset: 0,
		Season:        1,
		Year:          "2023",
		TMDBID:        12345,
		ReleaseGroup:  "测试字幕组",
	}

	// 保存
	err := repo.Save(ctx, bangumi)
	assert.NoError(t, err)

	// 获取
	got, err := repo.Get(ctx, bangumi.RSSLink)
	assert.NoError(t, err)
	assert.Equal(t, bangumi, got)

	// 更新
	bangumi.Name = "更新后的番剧名"
	err = repo.Save(ctx, bangumi)
	assert.NoError(t, err)

	// 再次获取
	got, err = repo.Get(ctx, bangumi.RSSLink)
	assert.NoError(t, err)
	assert.Equal(t, bangumi, got)
}

func TestRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	// 插入测试数据
	bangumis := []subscriber.Bangumi{
		{
			Name:        "番剧1",
			RSSLink:     "https://example.com/rss1",
			Active:      true,
			IncludeRegs: []string{".*1080p.*"},
			Season:      1,
			Year:        "2021",
		},
		{
			Name:        "番剧2",
			RSSLink:     "https://example.com/rss2",
			Active:      true,
			IncludeRegs: []string{".*1080p.*", ".*HEVC.*"},
			Season:      2,
			Year:        "2022",
		},
		{
			Name:        "番剧3",
			RSSLink:     "https://example.com/rss3",
			Active:      false,
			ExcludeRegs: []string{".*720p.*"},
			Season:      1,
			Year:        "2023",
		},
	}

	for _, b := range bangumis {
		err := repo.Save(ctx, b)
		require.NoError(t, err)
	}

	// 测试列表查询
	got, err := repo.List(ctx, subscriber.ListBangumiReq{})
	assert.NoError(t, err)
	assert.Len(t, got, 3)
}

func TestRepository_RSSRecord(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	subscriptionID := "test-subscription-id"
	guid1 := "test-guid-1"
	guid2 := "test-guid-2"

	// 测试标记处理
	err := repo.MarkProcessed(ctx, subscriptionID, guid1, guid2)
	assert.NoError(t, err)

	// 测试检查处理状态
	processed, err := repo.IsProcessed(ctx, subscriptionID, guid1)
	assert.NoError(t, err)
	assert.True(t, processed)

	// 测试获取所有已处理GUID
	guids, err := repo.ListProcessedGUIDs(ctx, subscriptionID)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{guid1, guid2}, guids)

	// 测试删除处理记录
	err = repo.DeleteProcessed(ctx, subscriptionID, guid1)
	assert.NoError(t, err)

	// 验证删除结果
	processed, err = repo.IsProcessed(ctx, subscriptionID, guid1)
	assert.NoError(t, err)
	assert.False(t, processed)

	processed, err = repo.IsProcessed(ctx, subscriptionID, guid2)
	assert.NoError(t, err)
	assert.True(t, processed)
}
