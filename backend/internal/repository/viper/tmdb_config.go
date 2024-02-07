package viper

import (
	"github.com/MangataL/BangumiBuddy/internal/meta/tmdb"
)

const (
	ComponentNameTMDB = ComponentName("tmdb")
)

func (r *Repo) GetTMDBConfig() (tmdb.Config, error) {
	var config tmdb.Config
	if err := r.GetComponentConfig(ComponentNameTMDB, &config); err != nil {
		return tmdb.Config{}, err
	}
	return config, nil
}

func (r *Repo) SetTMDBConfig(config *tmdb.Config) error {
	return r.SetComponentConfig(ComponentNameTMDB, config)
}
