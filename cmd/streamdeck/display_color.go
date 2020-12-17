package main

import (
	"context"
	"image/color"

	"github.com/pkg/errors"
)

func init() {
	registerDisplayElement("color", displayElementColor{})
}

type displayElementColor struct{}

func (d displayElementColor) Display(ctx context.Context, idx int, attributes attributeCollection) error {
	if attributes.Color != "" {
		return d.displayPredefinedColor(idx, attributes.Color)
	}

	if attributes.RGBA != nil {
		if len(attributes.RGBA) != 4 {
			return errors.New("RGBA color definition needs 4 hex values")
		}

		if err := ctx.Err(); err != nil {
			// Page context was cancelled, do not draw
			return err
		}

		return sd.FillColor(idx, color.RGBA{attributes.RGBA[0], attributes.RGBA[1], attributes.RGBA[2], attributes.RGBA[3]})
	}

	return errors.New("No valid attributes specified for type color")
}

func (displayElementColor) displayPredefinedColor(idx int, name string) error {
	c, ok := map[string]color.RGBA{
		"blue": {0x0, 0x0, 0xff, 0xff},
	}[name]

	if !ok {
		return errors.Errorf("Unknown color %q", name)
	}

	return sd.FillColor(idx, c)
}
