package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/MangataL/BangumiBuddy/internal/auth"
	"github.com/MangataL/BangumiBuddy/pkg/errs"
)

func NewTokenOperator() auth.TokenOperator {
	return &tokenOperator{}
}

type tokenOperator struct {
}

type jwtClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (t *tokenOperator) Generate(_ context.Context, tokenType, key string, expireAt time.Time) (string, error) {
	claim := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "BangumiBuddy",
			Audience:  []string{"BangumiBuddy-Web"},
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	return claim.SignedString([]byte(key))
}

func (t *tokenOperator) Check(ctx context.Context, tokenType, key, token string) error {
	var claims jwtClaims
	if _, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	}); err != nil {
		return err
	}
	if claims.TokenType != tokenType {
		return errs.NewBadRequest("token类型不匹配")
	}
	return nil
}
