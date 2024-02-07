package gin

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/MangataL/BangumiBuddy/internal/auth"
	"github.com/MangataL/BangumiBuddy/pkg/errs"
)

type args struct {
	ctx    *gin.Context
	writer *httptest.ResponseRecorder
}

type authTestCases struct {
	name       string
	args       args
	fake       func(t *testing.T) (auth.Authenticator, func())
	wantStatus int
	wantBody   string
}

func TestAuth_CheckToken(t *testing.T) {
	initArgs := func() args {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "", nil)
		c.Request.Header.Set("Authorization", "Bearer token")
		return args{
			ctx:    c,
			writer: w,
		}
	}
	testCases := []authTestCases{
		{
			name: "success",
			args: initArgs(),
			fake: func(t *testing.T) (auth.Authenticator, func()) {
				ctrl := gomock.NewController(t)
				mockAuth := auth.NewMockAuthenticator(ctrl)
				mockAuth.EXPECT().CheckAccessToken(gomock.Any(), gomock.Any()).Return(nil)
				return mockAuth, ctrl.Finish
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "auth error",
			args: initArgs(),
			fake: func(t *testing.T) (auth.Authenticator, func()) {
				ctrl := gomock.NewController(t)
				mockAuth := auth.NewMockAuthenticator(ctrl)
				mockAuth.EXPECT().CheckAccessToken(gomock.Any(), gomock.Any()).Return(errs.NewUnauthorized("错误"))
				return mockAuth, ctrl.Finish
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"错误"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dep, clo := tc.fake(t)
			defer clo()
			a := New(Dependency{
				Authenticator: dep,
			})

			a.CheckToken(tc.args.ctx)

			assert.Equal(t, tc.wantStatus, tc.args.writer.Code)
			assert.Equal(t, tc.wantBody, tc.args.writer.Body.String())
		})
	}
}

func TestAuth_Token(t *testing.T) {
	initArgs := func(body string) args {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodPost, "/token", bytes.NewBufferString(body))
		c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return args{
			ctx:    c,
			writer: w,
		}
	}
	testCases := []authTestCases{
		{
			name: "login-success",
			args: initArgs("grant_type=password&username=user&password=pass"),
			fake: func(t *testing.T) (auth.Authenticator, func()) {
				ctrl := gomock.NewController(t)
				mockAuth := auth.NewMockAuthenticator(ctrl)
				mockAuth.EXPECT().Authorize(gomock.Any(), "user", "pass").Return(auth.Credentials{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
				}, nil)
				return mockAuth, ctrl.Finish
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"access_token":"access_token","token_type":"Bearer","refresh_token":"refresh_token"}`,
		},
		{
			name: "login-error",
			args: initArgs("grant_type=password&username=user&password=pass"),
			fake: func(t *testing.T) (auth.Authenticator, func()) {
				ctrl := gomock.NewController(t)
				mockAuth := auth.NewMockAuthenticator(ctrl)
				mockAuth.EXPECT().Authorize(gomock.Any(), gomock.Any(), gomock.Any()).Return(auth.Credentials{}, auth.ErrUsernameOrPasswordError)
				return mockAuth, ctrl.Finish
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"invalid_grant","error_description":"用户名或密码错误"}`,
		},
		{
			name: "login-error-server",
			args: initArgs("grant_type=password&username=user&password=pass"),
			fake: func(t *testing.T) (auth.Authenticator, func()) {
				ctrl := gomock.NewController(t)
				mockAuth := auth.NewMockAuthenticator(ctrl)
				mockAuth.EXPECT().Authorize(gomock.Any(), gomock.Any(), gomock.Any()).Return(auth.Credentials{}, errors.New("err"))
				return mockAuth, ctrl.Finish
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"server_error","error_description":"err"}`,
		},
		{
			name: "refresh-success",
			args: initArgs("grant_type=refresh_token&refresh_token=token"),
			fake: func(t *testing.T) (auth.Authenticator, func()) {
				ctrl := gomock.NewController(t)
				mockAuth := auth.NewMockAuthenticator(ctrl)
				mockAuth.EXPECT().RefreshCredentials(gomock.Any(), "token").Return(auth.Credentials{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
				}, nil)
				return mockAuth, ctrl.Finish
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"access_token":"access_token","token_type":"Bearer","refresh_token":"refresh_token"}`,
		},
		{
			name: "refresh-error",
			args: initArgs("grant_type=refresh_token&refresh_token=token"),
			fake: func(t *testing.T) (auth.Authenticator, func()) {
				ctrl := gomock.NewController(t)
				mockAuth := auth.NewMockAuthenticator(ctrl)
				mockAuth.EXPECT().RefreshCredentials(gomock.Any(), gomock.Any()).Return(auth.Credentials{}, errs.NewBadRequest("错误"))
				return mockAuth, ctrl.Finish
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"invalid_request","error_description":"错误"}`,
		},
		{
			name: "unsupported-grant-type",
			args: initArgs("grant_type=unsupported"),
			fake: func(t *testing.T) (auth.Authenticator, func()) {
				return nil, func() {}
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"unsupported_response_type","error_description":"不支持的授权类型"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dep, clo := tc.fake(t)
			defer clo()
			a := New(Dependency{
				Authenticator: dep,
			})

			a.Token(tc.args.ctx)

			assert.Equal(t, tc.wantStatus, tc.args.writer.Code)
			assert.Equal(t, tc.wantBody, tc.args.writer.Body.String())
		})
	}
}

func TestAuth_UpdateUser(t *testing.T) {
	initArgs := func() args {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := `{"username": "user", "password": "pass"}`
		c.Request, _ = http.NewRequest(http.MethodPut, "/user", bytes.NewBufferString(body))
		return args{
			ctx:    c,
			writer: w,
		}
	}
	testCases := []authTestCases{
		{
			name: "success",
			args: initArgs(),
			fake: func(t *testing.T) (auth.Authenticator, func()) {
				ctrl := gomock.NewController(t)
				mockAuth := auth.NewMockAuthenticator(ctrl)
				mockAuth.EXPECT().UpdateUser(gomock.Any(), "user", "pass").Return(nil)
				return mockAuth, ctrl.Finish
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "error",
			args: initArgs(),
			fake: func(t *testing.T) (auth.Authenticator, func()) {
				ctrl := gomock.NewController(t)
				mockAuth := auth.NewMockAuthenticator(ctrl)
				mockAuth.EXPECT().UpdateUser(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("err"))
				return mockAuth, ctrl.Finish
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"err"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dep, clo := tc.fake(t)
			defer clo()
			a := New(Dependency{
				Authenticator: dep,
			})

			a.UpdateUser(tc.args.ctx)

			assert.Equal(t, tc.wantStatus, tc.args.writer.Code)
			assert.Equal(t, tc.wantBody, tc.args.writer.Body.String())
		})
	}
}
