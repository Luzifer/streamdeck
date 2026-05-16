package main

import "github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/modules/opts"

func moduleRuntime() opts.Runtime {
	return opts.Runtime{
		Conf:               userConfig,
		Deck:               sd,
		Keyboard:           kbd,
		ReloadConfig:       reloadConfig,
		TogglePage:         togglePage,
		ToggleRelativePage: toggleRelativePage,
	}
}
