package main

import "fmt"

type actionToggleDisplay struct{}

var actionToggleDisplayPreviousBrightness int

func init() {
	registerAction("toggle_display", actionToggleDisplay{})
}

func (actionToggleDisplay) Execute(_ attributeCollection) error {
	var newB int
	if currentBrightness > 0 {
		actionToggleDisplayPreviousBrightness = currentBrightness
	} else {
		newB = actionToggleDisplayPreviousBrightness
	}

	if err := sd.SetBrightness(newB); err != nil {
		return fmt.Errorf("setting brightness: %w", err)
	}
	currentBrightness = newB

	return nil
}
