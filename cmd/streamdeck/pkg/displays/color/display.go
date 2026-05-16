// Package color provides solid color display elements.
package color

import (
	"context"
	"fmt"
	"image/color"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/helpers"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/modules/opts"
)

type (
	// Display renders a solid color on a key.
	Display struct{}

	// Attrs contains configuration for the color display.
	Attrs struct {
		Color string `json:"color,omitempty" yaml:"color,omitempty"`
		RGBA  []int  `json:"rgba,omitempty" yaml:"rgba,omitempty"`
	}
)

// Display renders the configured color on the selected key.
func (d Display) Display(ctx context.Context, idx int, devs opts.Runtime, atts config.DynamicAttributes) (err error) {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return fmt.Errorf("decoding attributes: %w", err)
	}

	if attributes.Color != "" {
		return d.displayPredefinedColor(devs, idx, attributes.Color)
	}

	if attributes.RGBA != nil {
		if err := ctx.Err(); err != nil {
			// Page context was cancelled, do not draw
			return fmt.Errorf("page context cancelled: %w", err)
		}

		fillColor, err := helpers.Int4ToRGBA(attributes.RGBA)
		if err != nil {
			return fmt.Errorf("invalid 'rgba' color definition: %w", err)
		}

		if err = devs.Deck.FillColor(idx, fillColor); err != nil {
			return fmt.Errorf("filling with color: %w", err)
		}

		return nil
	}

	return fmt.Errorf("no valid attributes specified for type color")
}

func (Display) displayPredefinedColor(devs opts.Runtime, idx int, name string) (err error) {
	c, ok := map[string]color.RGBA{
		"blue": {0x0, 0x0, 0xff, 0xff}, //revive:disable-line:add-constant // color definition
	}[name]

	if !ok {
		return fmt.Errorf("unknown color %q", name)
	}

	if err = devs.Deck.FillColor(idx, c); err != nil {
		return fmt.Errorf("filling with color: %w", err)
	}

	return nil
}
