package tmdb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tmdb "github.com/cyruzin/golang-tmdb"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/MangataL/BangumiBuddy/internal/meta"
)

func TestClient_Search(t *testing.T) {
	testCases := []struct {
		name        string
		fake        func() (*tmdb.Client, func())
		bangumiName string
		wantErr     bool
		want        meta.Meta
	}{
		{
			name: "success",
			fake: func() (*tmdb.Client, func()) {
				ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if strings.Contains(r.URL.Path, "search/tv") {
						rsp := `{"results":[{"id":1,"name":"test","first_air_date":"2021-01-01"}]}`
						_, _ = w.Write([]byte(rsp))
					} else {
						rsp := `{"id":1,"name":"test","first_air_date":"2021-01-01"}`
						_, _ = w.Write([]byte(rsp))
					}
				}))
				certPool := x509.NewCertPool()
				certPool.AddCert(ts.Certificate())
				customTransport := &CustomRoundTripper{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							RootCAs: certPool,
						},
					},
					NewURL: ts.URL[len("https://"):], // 提取主机和端口部分
				}

				httpClient := http.Client{
					Transport: customTransport,
				}
				c, _ := tmdb.Init("test")
				c.SetClientConfig(httpClient)
				return c, ts.Close
			},
			bangumiName: "test1",
			wantErr:     false,
			want: meta.Meta{
				ChineseName: "test",
				Year:        "2021",
				TMDBID:      1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, clo := tc.fake()
			defer clo()
			p := &Client{
				client: c,
			}

			got, err := p.SearchTV(context.Background(), tc.bangumiName)
			t.Log(err)

			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestClient_Parse(t *testing.T) {
	testCases := []struct {
		name    string
		fake    func() (*tmdb.Client, func())
		id      int
		wantErr bool
		want    meta.Meta
	}{
		{
			name: "success",
			fake: func() (*tmdb.Client, func()) {
				ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					rsp := `{"id":1,"name":"test","first_air_date":"2021-01-01"}`
					_, _ = w.Write([]byte(rsp))
				}))
				certPool := x509.NewCertPool()
				certPool.AddCert(ts.Certificate())
				customTransport := &CustomRoundTripper{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							RootCAs: certPool,
						},
					},
					NewURL: ts.URL[len("https://"):], // 提取主机和端口部分
				}

				httpClient := http.Client{
					Transport: customTransport,
				}
				c, _ := tmdb.Init("test")
				c.SetClientConfig(httpClient)
				return c, ts.Close
			},
			id:      1,
			wantErr: false,
			want: meta.Meta{
				ChineseName: "test",
				Year:        "2021",
				TMDBID:      1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, clo := tc.fake()
			defer clo()
			p := &Client{
				client: c,
			}

			got, err := p.ParseTV(context.Background(), tc.id)
			t.Log(err)

			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestClient_GetSeasonEpisodeTotalNum(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rsp := `{"id":1,"name":"test","seasons":[{"season_number":1,"episode_count":12},{"season_number":2,"episode_count":13}]}`
		_, _ = w.Write([]byte(rsp))
	}))
	defer ts.Close()
	certPool := x509.NewCertPool()
	certPool.AddCert(ts.Certificate())
	customTransport := &CustomRoundTripper{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		},
		NewURL: ts.URL[len("https://"):],
	}
	httpClient := http.Client{
		Transport: customTransport,
	}
	c, _ := tmdb.Init("test")
	c.SetClientConfig(httpClient)
	p := &Client{client: c}

	got, err := p.GetSeasonEpisodeTotalNum(context.Background(), 1, 2)

	assert.NoError(t, err)
	assert.Equal(t, 13, got)
}

func Test_cacheAdapter_GetSeasonEpisodeTotalNum(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name      string
		opts      []meta.MetaOption
		fake      func(t *testing.T) *meta.MockParser
		wantFirst int
		wantNext  int
	}{
		{
			name: "when call option ttl is positive then caches season total",
			opts: []meta.MetaOption{
				meta.WithCacheTTL(time.Hour),
			},
			fake: func(t *testing.T) *meta.MockParser {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				parser := meta.NewMockParser(ctrl)
				parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2, gomock.Any()).Return(13, nil).Times(1)
				return parser
			},
			wantFirst: 13,
			wantNext:  13,
		},
		{
			name: "when call option is absent then does not cache",
			fake: func(t *testing.T) *meta.MockParser {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				parser := meta.NewMockParser(ctrl)
				gomock.InOrder(
					parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2).Return(13, nil),
					parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2).Return(14, nil),
				)
				return parser
			},
			wantFirst: 13,
			wantNext:  14,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			adapter := newCacheAdapter(reloadableTestParser{Parser: tc.fake(t)})

			gotFirst, err := adapter.GetSeasonEpisodeTotalNum(ctx, 95231, 2, tc.opts...)
			assert.NoError(t, err)
			gotNext, err := adapter.GetSeasonEpisodeTotalNum(ctx, 95231, 2, tc.opts...)
			assert.NoError(t, err)

			assert.Equal(t, tc.wantFirst, gotFirst)
			assert.Equal(t, tc.wantNext, gotNext)
		})
	}
}

type reloadableTestParser struct {
	meta.Parser
}

func (reloadableTestParser) Reload(config interface{}) error {
	return nil
}

// CustomRoundTripper 是一个自定义的 RoundTripper，它将所有请求转发到另一个地址
type CustomRoundTripper struct {
	Transport http.RoundTripper
	NewURL    string
}

func (c *CustomRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Host = c.NewURL
	return c.Transport.RoundTrip(req)
}
