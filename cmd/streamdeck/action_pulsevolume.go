// +build linux

package main

import (
	"strconv"

	"github.com/pkg/errors"
)

func init() {
	registerAction("pulsevolume", actionPulseVolume{})
}

type actionPulseVolume struct{}

func (actionPulseVolume) Execute(attributes map[string]interface{}) error {
	if pulseClient == nil {
		return errors.New("PulseAudio client not initialized")
	}

	devType, ok := attributes["device"].(string)
	if !ok {
		return errors.New("Missing 'device' attribute")
	}

	match, ok := attributes["match"].(string)
	if !ok {
		return errors.New("Missing 'match' attribute")
	}

	// Read mute value
	var (
		mute  string
		mutev = attributes["mute"]
	)
	switch mutev.(type) {
	case string:
		mute = mutev.(string)

	case bool:
		mute = strconv.FormatBool(mutev.(bool))
	}

	// Read volume
	var (
		volAbs bool
		volVal float64
	)
	for attr, abs := range map[string]bool{
		"set_volume":    true,
		"change_volume": false,
	} {
		val, ok := attributes[attr]
		if !ok {
			continue
		}

		switch val.(type) {
		case float64:
			volVal = val.(float64) / 100

		case int:
			volVal = float64(val.(int)) / 100

		case int64:
			volVal = float64(val.(int64)) / 100
		}

		volAbs = abs
		break
	}

	// Execute change
	switch devType {

	case "input":
		return errors.Wrap(
			pulseClient.SetSinkInputVolume(match, mute, volVal, volAbs),
			"Unable to set sink input volume",
		)

	case "sink":
		return errors.Wrap(
			pulseClient.SetSinkVolume(match, mute, volVal, volAbs),
			"Unable to set sink volume",
		)

	case "source":
		return errors.Wrap(
			pulseClient.SetSourceVolume(match, mute, volVal, volAbs),
			"Unable to set source volume",
		)

	default:
		return errors.Errorf("Unsupported device type: %q", devType)

	}
}
