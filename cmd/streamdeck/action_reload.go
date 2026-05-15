package main

import "fmt"

type actionReloadConfig struct{}

func init() {
	registerAction("reload_config", actionReloadConfig{})
}

func (actionReloadConfig) Execute(_ attributeCollection) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("reloading config: %w", err)
	}

	return togglePage(userConfig.DefaultPage)
}
