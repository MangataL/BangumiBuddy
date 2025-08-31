package viper

import "github.com/MangataL/BangumiBuddy/pkg/subtitle/ass"

const (
	ComponentNameSubtitleOperator = ComponentName("subtitle")
)

func (r *Repo) GetSubtitleOperatorConfig() (ass.FontSubsetterConfig, error) {
	var config ass.FontSubsetterConfig
	if err := r.GetComponentConfig(ComponentNameSubtitleOperator, &config); err != nil {
		return ass.FontSubsetterConfig{}, err
	}
	return config, nil
}

func (r *Repo) SetSubtitleOperatorConfig(config *ass.FontSubsetterConfig) error {
	return r.SetComponentConfig(ComponentNameSubtitleOperator, config)
}