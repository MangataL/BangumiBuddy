package tmdb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"testing"

	tmdb "github.com/cyruzin/golang-tmdb"
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
					rsp := `{"results":[{"id":1,"name":"test","first_air_date":"2021-01-01"}]}`
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

			got, err := p.Search(context.Background(), tc.bangumiName)
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

			got, err := p.Parse(context.Background(), tc.id)
			t.Log(err)

			assert.Equal(t, tc.wantErr, err != nil)
			assert.Equal(t, tc.want, got)
		})
	}
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