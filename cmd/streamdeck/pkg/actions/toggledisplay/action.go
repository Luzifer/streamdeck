// Package toggledisplay provides display brightness toggle actions.
package toggledisplay

import (
	"fmt"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/modules/opts"
)

// Action toggles the StreamDeck display brightness.
type Action struct{}

var (
	actionToggleDisplayPreviousBrightness int
	currentBrightness                     = -1
)

// Execute toggles between the previous brightness and display-off.
func (Action) Execute(devs opts.Runtime, _ config.DynamicAttributes) error {
	if currentBrightness == -1 {
		// Initialize on first call
		currentBrightness = devs.Conf.DefaultBrightness
	}

	var newB int
	if currentBrightness > 0 {
		actionToggleDisplayPreviousBrightness = currentBrightness
	} else {
		newB = actionToggleDisplayPreviousBrightness
	}

	if err := devs.Deck.SetBrightness(newB); err != nil {
		return fmt.Errorf("setting brightness: %w", err)
	}
	currentBrightness = newB

	return nil
}
