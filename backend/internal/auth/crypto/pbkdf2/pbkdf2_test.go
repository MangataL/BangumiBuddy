package pbkdf2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCipher(t *testing.T) {
	c := NewCipher()
	ctx := context.Background()

	key, err := c.GenerateKey(ctx)
	require.Nil(t, err)
	t.Log(key)

	text := "password"
	cipherText, err := c.Encrypt(ctx, key, text)
	require.Nil(t, err)
	t.Log(cipherText)

	err = c.Check(ctx, key, text, cipherText)
	require.Nil(t, err)

	err = c.Check(ctx, key, "wrong", cipherText)
	require.NotNil(t, err)
}
