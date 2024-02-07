package viper

import (
	"github.com/MangataL/BangumiBuddy/internal/transfer"
)

const (
	ComponentNameTransfer = ComponentName("transfer")
)

func (r *Repo) GetTransferConfig() (transfer.Config, error) {
	var config transfer.Config
	if err := r.GetComponentConfig(ComponentNameTransfer, &config); err != nil {
		return transfer.Config{}, err
	}
	return config, nil
}

func (r *Repo) SetTransferConfig(config *transfer.Config) error {
	return r.SetComponentConfig(ComponentNameTransfer, config)
}
