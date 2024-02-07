package viper

import (
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
)

const (
	ComponentNameSubscriber = ComponentName("subscriber")
)

func (r *Repo) GetSubscriberConfig() (subscriber.Config, error) {
	var config subscriber.Config
	if err := r.GetComponentConfig(ComponentNameSubscriber, &config); err != nil {
		return subscriber.Config{}, err
	}
	return config, nil
}

func (r *Repo) SetSubscriberConfig(config *subscriber.Config) error {
	return r.SetComponentConfig(ComponentNameSubscriber, config)
}
