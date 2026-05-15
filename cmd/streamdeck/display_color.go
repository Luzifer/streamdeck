package main

import (
	"context"
	"fmt"
	"image/color"
)

type displayElementColor struct{}

func init() {
	registerDisplayElement("color", displayElementColor{})
}

func (d displayElementColor) Display(ctx context.Context, idx int, attributes attributeCollection) (err error) {
	if attributes.Color != "" {
		return d.displayPredefinedColor(idx, attributes.Color)
	}

	if attributes.RGBA != nil {
		if err := ctx.Err(); err != nil {
			// Page context was cancelled, do not draw
			return fmt.Errorf("page context cancelled: %w", err)
		}

		fillColor, err := int4ToRGBA(attributes.RGBA)
		if err != nil {
			return fmt.Errorf("invalid 'rgba' color definition: %w", err)
		}

		if err = sd.FillColor(idx, fillColor); err != nil {
			return fmt.Errorf("filling with color: %w", err)
		}

		return nil
	}

	return fmt.Errorf("no valid attributes specified for type color")
}

func (displayElementColor) displayPredefinedColor(idx int, name string) (err error) {
	c, ok := map[string]color.RGBA{
		"blue": {0x0, 0x0, 0xff, 0xff}, //revive:disable-line:add-constant // color definition
	}[name]

	if !ok {
		return fmt.Errorf("unknown color %q", name)
	}

	if err = sd.FillColor(idx, c); err != nil {
		return fmt.Errorf("filling with color: %w", err)
	}

	return nil
}
