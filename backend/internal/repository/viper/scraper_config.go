package viper

import (
	"github.com/MangataL/BangumiBuddy/internal/scrape"
)

const (
	ComponentNameScraper = ComponentName("scraper")
)

func (r *Repo) GetScraperConfig() (scrape.Config, error) {
	var config scrape.Config
	if err := r.GetComponentConfig(ComponentNameScraper, &config); err != nil {
		return scrape.Config{}, err
	}
	return config, nil
}

func (r *Repo) SetScraperConfig(config *scrape.Config) error {
	return r.SetComponentConfig(ComponentNameScraper, config)
}