// Package opts contains runtime dependencies passed to modules.
package opts

import (
	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/config"
	"github.com/sashko/go-uinput"

	"github.com/Luzifer/streamdeck/v2"
)

type (
	// Runtime contains device handles and callbacks available to modules.
	Runtime struct {
		Conf     config.File
		Deck     *streamdeck.Client
		Keyboard uinput.Keyboard

		ReloadConfig       func() error
		TogglePage         func(string) error
		ToggleRelativePage func(int) error
	}
)
