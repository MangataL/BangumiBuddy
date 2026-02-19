package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/MangataL/BangumiBuddy/internal/scrape"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	return db
}

func TestRepository_AddGetListAndUpsert(t *testing.T) {
	ctx := context.Background()
	repo := New(setupTestDB(t))

	err := repo.Add(ctx, scrape.MetadataCheckTask{
		TMDBID:      1,
		FilePath:    "/tmp/ep01.mkv",
		BangumiName: "测试番剧",
		PosterURL:   "https://example.com/poster.jpg",
		Season:      1,
		Episode:     1,
		Statuses:    nil,
	})
	require.NoError(t, err)

	tasks, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, []scrape.ScrapeStatus{scrape.ScrapeStatusPending}, tasks[0].Statuses)
	assert.Equal(t, 1, tasks[0].TMDBID)

	err = repo.Add(ctx, scrape.MetadataCheckTask{
		TMDBID:      2,
		FilePath:    "/tmp/ep01.mkv",
		BangumiName: "更新番剧",
		PosterURL:   "https://example.com/new-poster.jpg",
		Season:      2,
		Episode:     3,
		Statuses:    []scrape.ScrapeStatus{scrape.ScrapeStatusMissingPlot},
	})
	require.NoError(t, err)

	tasks, err = repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, 2, tasks[0].TMDBID)
	assert.Equal(t, "更新番剧", tasks[0].BangumiName)
	assert.Equal(t, 2, tasks[0].Season)
	assert.Equal(t, 3, tasks[0].Episode)
	assert.Equal(t, []scrape.ScrapeStatus{scrape.ScrapeStatusMissingPlot}, tasks[0].Statuses)

	got, err := repo.Get(ctx, tasks[0].ID)
	require.NoError(t, err)
	assert.Equal(t, tasks[0], got)
}

func TestRepository_UpdateStatusesAndDelete(t *testing.T) {
	ctx := context.Background()
	repo := New(setupTestDB(t))

	filePath := "/tmp/ep02.mkv"
	require.NoError(t, repo.Add(ctx, scrape.MetadataCheckTask{
		TMDBID:      3,
		FilePath:    filePath,
		BangumiName: "测试番剧",
		Statuses:    []scrape.ScrapeStatus{scrape.ScrapeStatusPending},
	}))

	newStatuses := []scrape.ScrapeStatus{
		scrape.ScrapeStatusMissingTitle,
		scrape.ScrapeStatusMissingImage,
	}
	require.NoError(t, repo.UpdateStatuses(ctx, filePath, newStatuses))

	tasks, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, newStatuses, tasks[0].Statuses)

	require.NoError(t, repo.Delete(ctx, filePath))

	tasks, err = repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 0)
}

func TestRepository_GetNotFound(t *testing.T) {
	ctx := context.Background()
	repo := New(setupTestDB(t))

	_, err := repo.Get(ctx, 999)
	require.Error(t, err)
	assert.ErrorIs(t, err, scrape.ErrTaskNotFound)
}

func TestRepository_Clean(t *testing.T) {
	ctx := context.Background()
	repo := New(setupTestDB(t))

	require.NoError(t, repo.Add(ctx, scrape.MetadataCheckTask{
		TMDBID:      1,
		FilePath:    "/tmp/a.mkv",
		BangumiName: "A",
		Statuses:    []scrape.ScrapeStatus{scrape.ScrapeStatusPending},
	}))
	require.NoError(t, repo.Add(ctx, scrape.MetadataCheckTask{
		TMDBID:      2,
		FilePath:    "/tmp/b.mkv",
		BangumiName: "B",
		Statuses:    []scrape.ScrapeStatus{scrape.ScrapeStatusMissingPlot},
	}))

	require.NoError(t, repo.Clean(ctx))

	tasks, err := repo.List(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, 0)
}
