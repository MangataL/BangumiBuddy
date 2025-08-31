package qbittorrent

import (
	"context"
	"testing"
)

func TestQBittorrent_GetTorrentName(t *testing.T) {
	q := NewQBittorrent(Config{
		Host:     "http://localhost:8080",
		Username: "admin",
		Password: "admin123",
	})
	name, err := q.GetTorrentName(context.Background(), "fde26cbf73ddd256f196f8dcb9483102217ebbd6")
	if err != nil {
		t.Fatalf("获取种子名称失败: %s", err)
	}
	t.Logf("种子名称: %s", name)
}