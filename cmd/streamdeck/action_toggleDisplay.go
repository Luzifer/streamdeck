package main

import "github.com/pkg/errors"

func init() {
	registerAction("toggle_display", actionToggleDisplay{})
}

var actionToggleDisplayPreviousBrightness int

type actionToggleDisplay struct{}

func (actionToggleDisplay) Execute(attributes attributeCollection) error {
	var newB int
	if currentBrightness > 0 {
		actionToggleDisplayPreviousBrightness = currentBrightness
	} else {
		newB = actionToggleDisplayPreviousBrightness
	}

	if err := sd.SetBrightness(newB); err != nil {
		return errors.Wrap(err, "Unable to set brightness")
	}
	currentBrightness = newB

	return nil
}
