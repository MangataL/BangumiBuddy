package gin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

func TestSubscriber_ParseRSS(t *testing.T) {
	initArgs := func(req subscriber.ParserRSSReq) args {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		u, _ := url.Parse("/bangumi/rss")
		q := u.Query()
		if req.RSSLink != "" {
			q.Set("rss_link", req.RSSLink)
		}
		if req.TMDBID != 0 {
			q.Set("tmdb_id", strconv.Itoa(req.TMDBID))
		}
		u.RawQuery = q.Encode()
		c.Request, _ = http.NewRequest("GET", u.String(), nil)
		return args{
			ctx:    c,
			writer: w,
		}
	}

	testCases := []struct {
		name       string
		args       args
		fake       func(t *testing.T) (subscriber.Interface, func())
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			args: initArgs(subscriber.ParserRSSReq{
				RSSLink: "link",
				TMDBID:  1,
			}),
			fake: func(t *testing.T) (subscriber.Interface, func()) {
				ctrl := gomock.NewController(t)
				sm := subscriber.NewMockInterface(ctrl)
				sm.EXPECT().ParseRSS(gomock.Any(), subscriber.ParserRSSReq{
					RSSLink: "link",
					TMDBID:  1,
				}).Return(subscriber.ParseRSSRsp{
					Name:    "name",
					Season:  2,
					Year:    "2024",
					TMDBID:  1,
					RSSLink: "link",
				}, nil).AnyTimes()
				return sm, ctrl.Finish
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"name":"name","season":2,"year":"2024","tmdbID":1,"rssLink":"link","releaseGroup":"","episodeTotalNum":0,"airWeekday":null}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dep, clo := tc.fake(t)
			defer clo()
			a := New(Dependency{
				Subscriber: dep,
			})

			a.ParseRSS(tc.args.ctx)

			assert.Equal(t, tc.wantStatus, tc.args.writer.Code)
			assert.Equal(t, tc.wantBody, tc.args.writer.Body.String())
		})
	}
}
