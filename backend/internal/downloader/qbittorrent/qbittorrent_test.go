package qbittorrent

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitLogsInOnlyOnce(t *testing.T) {
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
			_, err := w.Write([]byte(`["BangumiBuddy"]`))
			require.NoError(t, err)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client := NewQBittorrent(Config{
		Host:     server.URL,
		Username: "user",
		Password: "password",
	})

	require.NoError(t, client.init())
	require.NoError(t, client.init())
	require.Equal(t, int32(1), loginCount.Load())
}
