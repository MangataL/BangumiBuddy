package viper

import "github.com/MangataL/BangumiBuddy/internal/discovery"

const (
	ComponentNameDiscovery = ComponentName("discovery")
)

func (r *Repo) GetDiscoveryConfig() (discovery.Config, error) {
	var config discovery.Config
	if err := r.GetComponentConfig(ComponentNameDiscovery, &config); err != nil {
		return discovery.Config{}, err
	}
	return config, nil
}

func (r *Repo) SetDiscoveryConfig(config *discovery.Config) error {
	return r.SetComponentConfig(ComponentNameDiscovery, config)
}
