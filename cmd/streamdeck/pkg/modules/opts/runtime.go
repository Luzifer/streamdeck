// Package opts contains runtime dependencies passed to modules.
package opts

import (
	"github.com/sashko/go-uinput"

	"github.com/Luzifer/streamdeck"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/config"
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
