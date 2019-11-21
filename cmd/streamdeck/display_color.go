package main

import (
	"image/color"

	"github.com/pkg/errors"
)

func init() {
	registerDisplayElement("color", displayElementColor{})
}

type displayElementColor struct{}

func (d displayElementColor) Display(idx int, attributes map[string]interface{}) error {
	if name, ok := attributes["color"].(string); ok {
		return d.displayPredefinedColor(idx, name)
	}

	if rgba, ok := attributes["rgba"].([]interface{}); ok {
		if len(rgba) != 4 {
			return errors.New("RGBA color definition needs 4 hex values")
		}

		return sd.FillColor(idx, color.RGBA{uint8(rgba[0].(int)), uint8(rgba[1].(int)), uint8(rgba[2].(int)), uint8(rgba[3].(int))})
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
