package viper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserRepo(t *testing.T) {
	f, clo := newFile(t)
	defer clo()

	r, err := NewRepo(f.Name())
	require.Nil(t, err)

	username, err := r.GetUsername()
	require.Nil(t, err)
	require.Empty(t, username)

	wantUsername := "user"
	err = r.SetUsername(wantUsername)
	require.Nil(t, err)

	username, err = r.GetUsername()
	require.Nil(t, err)
	require.Equal(t, wantUsername, username)

	password, err := r.GetPassword()
	require.Nil(t, err)
	require.Empty(t, password)

	wantPassword := "password"
	err = r.SetPassword(wantPassword)
	require.Nil(t, err)

	password, err = r.GetPassword()
	require.Nil(t, err)
	require.Equal(t, wantPassword, password)

	token, err := r.GetUserToken()
	require.Nil(t, err)
	require.Empty(t, token)

	wantToken := "token"
	err = r.SetUserToken(wantToken)
	require.Nil(t, err)

	token, err = r.GetUserToken()
	require.Nil(t, err)
	require.Equal(t, wantToken, token)
}

func newFile(t *testing.T) (*os.File, func()) {
	dir := "./"
	f, err := os.CreateTemp(dir, "config-*.yaml")
	require.Nil(t, err)
	return f, func() {
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		if err := os.Remove(f.Name()); err != nil {
			t.Fatal(err)
		}
	}
}
