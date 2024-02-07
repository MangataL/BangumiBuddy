package viper

import (
	"github.com/MangataL/BangumiBuddy/internal/auth"
)

var _ auth.Config = (*Repo)(nil)

func (r *Repo) GetUsername() (string, error) {
	return r.file.GetString("user.username"), nil
}

func (r *Repo) SetUsername(username string) error {
	r.file.Set("user.username", username)
	return r.file.WriteConfig()
}

func (r *Repo) GetPassword() (string, error) {
	return r.file.GetString("user.password"), nil
}

func (r *Repo) SetPassword(password string) error {
	r.file.Set("user.password", password)
	return r.file.WriteConfig()
}

func (r *Repo) GetUserToken() (string, error) {
	return r.file.GetString("user.token"), nil
}

func (r *Repo) SetUserToken(token string) error {
	r.file.Set("user.token", token)
	return r.file.WriteConfig()
}
