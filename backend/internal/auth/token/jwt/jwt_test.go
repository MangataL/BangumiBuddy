package jwt

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTokenOperator(t *testing.T) {
	to := NewTokenOperator()
	ctx := context.Background()
	key := "key"
	tokenType1 := "type1"
	tokenType2 := "type2"

	token, err := to.Generate(ctx, tokenType1, key, time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	require.NoError(t, to.Check(ctx, tokenType1, key, token))

	token2, err := to.Generate(ctx, tokenType2, key, time.Now().Add(time.Hour))
	require.NoError(t, err)

	require.Error(t, to.Check(ctx, tokenType1, key, token2))

	tokenExpired, err := to.Generate(ctx, tokenType1, key, time.Now().Add(-time.Hour))
	require.Nil(t, err)

	require.Error(t, to.Check(ctx, tokenType1, key, tokenExpired))
}
