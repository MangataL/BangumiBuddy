package viper

import "github.com/MangataL/BangumiBuddy/internal/notice/adapter"

const (
	ComponentNameNotice = ComponentName("notice")
)

func (r *Repo) GetNoticeConfig() (adapter.Config, error) {
	var config adapter.Config
	if err := r.GetComponentConfig(ComponentNameNotice, &config); err != nil {
		return adapter.Config{}, err
	}
	return config, nil
}

func (r *Repo) SetNoticeConfig(config *adapter.Config) error {
	return r.SetComponentConfig(ComponentNameNotice, config)
}
