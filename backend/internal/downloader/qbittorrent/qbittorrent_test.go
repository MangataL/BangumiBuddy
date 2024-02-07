package qbittorrent

import (
	"context"
	"testing"
)

func TestQBittorrent_GetTorrentStatus(t *testing.T) {
	q := NewQBittorrent(Config{
		Host:     "http://localhost:8080",
		Username: "admin",
		Password: "admin123",
	})
	status, err := q.GetDownloadStatuses(context.Background(), []string{"6c44213510b1756c55ec48c1670701a571f84e3d"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(status)
}
