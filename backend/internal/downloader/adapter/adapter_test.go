package adapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/MangataL/BangumiBuddy/internal/downloader/qbittorrent"
)

func TestReloadReplacesInitializedQBittorrentClient(t *testing.T) {
	firstServer, firstLoginCount := newQBittorrentServer(t)
	secondServer, secondLoginCount := newQBittorrentServer(t)

	adapter := NewAdapter(Config{
		DownloadType: "qbittorrent",
		QBitTorrent: qbittorrent.Config{
			Host:     firstServer.URL,
			Username: "first-user",
			Password: "first-password",
		},
	})

	_, err := adapter.GetTorrentFileNames(context.Background(), "first-hash")
	require.NoError(t, err)
	_, err = adapter.GetTorrentFileNames(context.Background(), "first-hash")
	require.NoError(t, err)
	require.Equal(t, int32(1), firstLoginCount.Load())

	err = adapter.Reload(&Config{
		DownloadType: "qbittorrent",
		QBitTorrent: qbittorrent.Config{
			Host:     secondServer.URL,
			Username: "second-user",
			Password: "second-password",
		},
	})
	require.NoError(t, err)

	_, err = adapter.GetTorrentFileNames(context.Background(), "second-hash")
	require.NoError(t, err)
	_, err = adapter.GetTorrentFileNames(context.Background(), "second-hash")
	require.NoError(t, err)
	require.Equal(t, int32(1), firstLoginCount.Load())
	require.Equal(t, int32(1), secondLoginCount.Load())
}

func newQBittorrentServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	var loginCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			if loginCount.Add(1) == 1 {
				http.SetCookie(w, &http.Cookie{Name: "SID", Value: "test-session"})
			}
			w.WriteHeader(http.StatusNoContent)
		case "/api/v2/torrents/tags":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`["BangumiBuddy"]`))
		case "/api/v2/torrents/files":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[]`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	return server, &loginCount
}
