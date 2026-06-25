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
		SubscriptionID: "test-subscription-id",
		Name:           "测试番剧",
		RSSLink:        "https://example.com/rss",
		Active:         true,
		IncludeRegs:    []string{".*1080p.*", ".*HEVC.*"},
		ExcludeRegs:    []string{".*720p.*", ".*预告.*"},
		Priority:       10,
		EpisodeOffset:  0,
		Season:         1,
		Year:           "2023",
		TMDBID:         12345,
		ReleaseGroup:   "测试字幕组",
	}

	// 保存
	err := repo.Save(ctx, bangumi)
	assert.NoError(t, err)

	// 获取
	got, err := repo.Get(ctx, bangumi.SubscriptionID)
	assert.NoError(t, err)
	// 忽略 CreatedAt 进行比较，因为数据库存储可能存在精度差异
	bangumi.CreatedAt = got.CreatedAt
	assert.Equal(t, bangumi, got)

	// 更新
	bangumi.Name = "更新后的番剧名"
	err = repo.Save(ctx, bangumi)
	assert.NoError(t, err)

	// 再次获取
	got, err = repo.Get(ctx, bangumi.SubscriptionID)
	assert.NoError(t, err)
	bangumi.CreatedAt = got.CreatedAt
	assert.Equal(t, bangumi, got)
}

func TestRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)
	ctx := context.Background()

	// 插入测试数据
	bangumis := []subscriber.Bangumi{
		{
			SubscriptionID: "sub1",
			Name:           "番剧1",
			RSSLink:        "https://example.com/rss1",
			Active:         true,
			IncludeRegs:    []string{".*1080p.*"},
			Season:         1,
			Year:           "2021",
		},
		{
			SubscriptionID: "sub2",
			Name:           "番剧2",
			RSSLink:        "https://example.com/rss2",
			Active:         true,
			IncludeRegs:    []string{".*1080p.*", ".*HEVC.*"},
			Season:         2,
			Year:           "2022",
		},
		{
			SubscriptionID: "sub3",
			Name:           "番剧3",
			RSSLink:        "https://example.com/rss3",
			Active:         false,
			ExcludeRegs:    []string{".*720p.*"},
			Season:         1,
			Year:           "2023",
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

func TestRepository_UpdateEpisodeTotalNum(t *testing.T) {
	testCases := []struct {
		name      string
		oldTotal  int
		newTotal  int
		wantTotal int
	}{
		{
			name:      "when new total is greater then overwrites",
			oldTotal:  12,
			newTotal:  13,
			wantTotal: 13,
		},
		{
			name:      "when new total is less then overwrites",
			oldTotal:  13,
			newTotal:  12,
			wantTotal: 12,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := New(db)
			ctx := context.Background()
			bangumi := subscriber.Bangumi{
				SubscriptionID:  "sub-1",
				Name:            "测试番剧",
				RSSLink:         "https://example.com/rss",
				Active:          true,
				EpisodeTotalNum: tc.oldTotal,
			}
			require.NoError(t, repo.Save(ctx, bangumi))

			err := repo.UpdateEpisodeTotalNum(ctx, bangumi.SubscriptionID, tc.newTotal)

			require.NoError(t, err)
			got, err := repo.Get(ctx, bangumi.SubscriptionID)
			require.NoError(t, err)
			assert.Equal(t, tc.wantTotal, got.EpisodeTotalNum)
		})
	}
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
