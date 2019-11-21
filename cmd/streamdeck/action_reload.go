package main

import "github.com/pkg/errors"

func init() {
	registerAction("reload_config", actionReloadConfig{})
}

type actionReloadConfig struct{}

func (actionReloadConfig) Execute(attributes map[string]interface{}) error {
	if err := loadConfig(); err != nil {
		return errors.Wrap(err, "Unable to reload config")
	}

	return togglePage(userConfig.DefaultPage)
}
