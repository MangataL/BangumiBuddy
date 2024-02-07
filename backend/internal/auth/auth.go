package auth

import (
	"context"
	"time"

	"github.com/MangataL/BangumiBuddy/pkg/errs"
	"github.com/MangataL/BangumiBuddy/pkg/log"
)

const (
	defaultUsername  = "admin"
	defaultPassword  = "admin123"
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

//go:generate mockgen -destination auth_mock.go -source $GOFILE -package $GOPACKAGE

// New 生成鉴权器
func New(dep Dependencies) Authenticator {
	return &authenticator{
		config:        dep.Config,
		cipher:        dep.Cipher,
		tokenOperator: dep.TokenOperator,
	}
}

// Dependencies 依赖
type Dependencies struct {
	Config
	Cipher
	TokenOperator
}

// Cipher is the interface for token operation
type Cipher interface {
	GenerateKey(ctx context.Context) (string, error)
	Encrypt(ctx context.Context, key, text string) (string, error)
	Check(ctx context.Context, key, text, cipher string) error
}

// TokenOperator token操作器
type TokenOperator interface {
	Generate(ctx context.Context, tokenType, key string, expireAt time.Time) (string, error)
	Check(ctx context.Context, tokenType, key, token string) error
}

// Config 配置项
type Config interface {
	GetUsername() (string, error)
	SetUsername(username string) error
	GetPassword() (string, error)
	SetPassword(password string) error
	GetUserToken() (string, error)
	SetUserToken(token string) error
}

// authenticator
type authenticator struct {
	config        Config
	cipher        Cipher
	tokenOperator TokenOperator
}

const (
	defaultAccessTokenExpire  = time.Hour * 24
	defaultRefreshTokenExpire = time.Hour * 24 * 7
)

func (a *authenticator) Authorize(ctx context.Context, username, password string) (Credentials, error) {
	token := a.getKey()
	if err := a.validateLogin(ctx, username, password, token); err != nil {
		return Credentials{}, err
	}
	credentials, err := a.generateCredentials(ctx, token)
	if err != nil {
		return Credentials{}, err
	}
	return credentials, nil
}

func (a *authenticator) validateLogin(ctx context.Context, username, password, token string) error {
	if username == "" || password == "" {
		return errs.NewBadRequest("用户名或密码不能为空")
	}
	if username != a.getUsername() {
		return ErrUsernameOrPasswordError
	}
	pw := a.getPassword()
	if err := a.cipher.Check(ctx, token, password, pw); err != nil {
		log.Error(ctx, "检查密码错误: ", err.Error())
		return ErrUsernameOrPasswordError
	}
	return nil
}

func (a *authenticator) getUsername() string {
	username, _ := a.config.GetUsername()
	if username == "" {
		username = defaultUsername
		_ = a.config.SetUsername(username)
	}
	return username
}

func (a *authenticator) getPassword() string {
	password, _ := a.config.GetPassword()
	if password == "" {
		_ = a.setPassword(context.Background(), defaultPassword)
		password, _ = a.config.GetPassword()
	}
	return password
}

func (a *authenticator) getKey() string {
	token, _ := a.config.GetUserToken()
	if token == "" {
		token, _ = a.cipher.GenerateKey(context.Background())
		_ = a.config.SetUserToken(token)
	}
	return token
}

func (a *authenticator) generateCredentials(ctx context.Context, token string) (Credentials, error) {
	accessToken, err := a.tokenOperator.Generate(ctx, accessTokenType, token, time.Now().Add(defaultAccessTokenExpire))
	if err != nil {
		return Credentials{}, err
	}
	refreshToken, err := a.tokenOperator.Generate(ctx, refreshTokenType, token, time.Now().Add(defaultRefreshTokenExpire))
	if err != nil {
		return Credentials{}, err
	}
	return Credentials{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *authenticator) UpdateUser(ctx context.Context, username, password string) error {
	if err := a.setPassword(ctx, password); err != nil {
		return err
	}
	_ = a.config.SetUsername(username)
	return nil
}

func (a *authenticator) setPassword(ctx context.Context, password string) error {
	cipherText, err := a.cipher.Encrypt(ctx, a.getKey(), password)
	if err != nil {
		return err
	}
	_ = a.config.SetPassword(cipherText)
	return nil
}

func (a *authenticator) checkToken(ctx context.Context, tokenType, token string) error {
	if token == "" {
		return errs.NewUnauthorized("token不能为空")
	}
	if err := a.tokenOperator.Check(ctx, tokenType, a.getKey(), token); err != nil {
		return errs.NewUnauthorizedf("token无效: %v", err)
	}
	return nil
}

func (a *authenticator) CheckAccessToken(ctx context.Context, accessToken string) error {
	return a.checkToken(ctx, accessTokenType, accessToken)
}

func (a *authenticator) RefreshCredentials(ctx context.Context, refreshToken string) (Credentials, error) {
	if err := a.checkToken(ctx, refreshTokenType, refreshToken); err != nil {
		return Credentials{}, err
	}
	return a.generateCredentials(ctx, a.getKey())
}