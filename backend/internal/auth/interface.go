package auth

import (
	"context"

	"github.com/MangataL/BangumiBuddy/pkg/errs"
)

//go:generate mockgen -destination interface_mock.go -source $GOFILE -package $GOPACKAGE

var (
	// ErrUsernameOrPasswordError 用户名或密码错误
	ErrUsernameOrPasswordError = errs.NewBadRequest("用户名或密码错误")
)

// Authenticator 鉴权接口
type Authenticator interface {
	// Authorize 用户密码方式进行授权
	Authorize(ctx context.Context, username, password string) (Credentials, error)
	// UpdateUser 更新用户名和密码
	UpdateUser(ctx context.Context, username, password string) error
	// CheckAccessToken 检查token是否有效
	CheckAccessToken(ctx context.Context, token string) error
	// RefreshCredentials 刷新凭证
	RefreshCredentials(ctx context.Context, refreshToken string) (Credentials, error)
}
