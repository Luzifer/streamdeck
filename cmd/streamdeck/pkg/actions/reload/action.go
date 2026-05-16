// Package reload provides configuration reload actions.
package reload

import (
	"fmt"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/modules/opts"
)

// Action reloads the StreamDeck configuration.
type Action struct{}

func init() {
}

// Execute reloads the configuration through the runtime.
func (Action) Execute(devs opts.Runtime, _ config.DynamicAttributes) (err error) {
	if err = devs.ReloadConfig(); err != nil {
		return fmt.Errorf("reloading config: %w", err)
	}

	return nil
}
