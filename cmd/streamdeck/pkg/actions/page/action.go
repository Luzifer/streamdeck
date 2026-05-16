// Package page provides page navigation actions.
package page

import (
	"fmt"
	"math"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/modules/opts"
)

type (
	// Action switches to another page.
	Action struct{}

	// Attrs contains configuration for the page action.
	Attrs struct {
		Name     string `json:"name,omitempty" yaml:"name,omitempty"`
		Relative int    `json:"relative,omitempty" yaml:"relative,omitempty"`
	}
)

// Execute switches to the configured page or relative page.
func (Action) Execute(dev opts.Runtime, atts config.DynamicAttributes) (err error) {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return fmt.Errorf("decoding attributes: %w", err)
	}

	if attributes.Name != "" {
		if err = dev.TogglePage(attributes.Name); err != nil {
			return fmt.Errorf("switching page: %w", err)
		}

		return nil
	}

	if absRel := int(math.Abs(float64(attributes.Relative))); absRel != 0 {
		if err = dev.ToggleRelativePage(absRel); err != nil {
			return fmt.Errorf("switching relative page: %w", err)
		}

		return nil
	}

	return fmt.Errorf("no page name or relative move supplied")
}
