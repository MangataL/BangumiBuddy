package viper

import "github.com/MangataL/BangumiBuddy/internal/network"

const (
	ComponentNameNetwork = ComponentName("network")
)

func (r *Repo) GetNetworkConfig() (network.Config, error) {
	var config network.Config
	if err := r.GetComponentConfig(ComponentNameNetwork, &config); err != nil {
		return network.Config{}, err
	}
	return config, nil
}

func (r *Repo) SetNetworkConfig(config *network.Config) error {
	if err := network.Validate(*config); err != nil {
		return err
	}
	return r.SetComponentConfig(ComponentNameNetwork, config)
}
