package pbkdf2

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/pbkdf2"

	"github.com/MangataL/BangumiBuddy/internal/auth"
)

var _ auth.Cipher = (*Cipher)(nil)

// Cipher 使用 pbkdf2 实现的加密解密器
type Cipher struct{}

// NewCipher 生成 pbkdf2 实现的加密解密工具
func NewCipher() *Cipher {
	return &Cipher{}
}

const (
	keyLen    = 32
	iterTimes = 100000
)

func (o *Cipher) GenerateKey(_ context.Context) (string, error) {
	salt := make([]byte, keyLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return encode(salt), nil
}

func encode(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func (o *Cipher) Encrypt(_ context.Context, key, text string) (string, error) {
	salt, err := decode(key)
	if err != nil {
		return "", err
	}
	return encode(pbkdf2.Key([]byte(text), salt, iterTimes, keyLen, sha512.New)), nil
}

func decode(str string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(str)
}

func (o *Cipher) Check(ctx context.Context, key, text, cipher string) error {
	got, _ := o.Encrypt(ctx, key, text)
	if got != cipher {
		return errors.New("cipher not match")
	}
	return nil
}
