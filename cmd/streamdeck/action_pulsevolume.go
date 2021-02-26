// +build linux

package main

import (
	"github.com/pkg/errors"
)

func init() {
	registerAction("pulsevolume", actionPulseVolume{})
}

type actionPulseVolume struct{}

func (actionPulseVolume) Execute(attributes attributeCollection) error {
	if pulseClient == nil {
		return errors.New("PulseAudio client not initialized")
	}

	if attributes.Device == "" {
		return errors.New("Missing 'device' attribute")
	}

	if attributes.Match == "" {
		return errors.New("Missing 'match' attribute")
	}

	// Read volume
	var (
		volAbs bool
		volVal float64
	)

	if attributes.SetVolume != nil {
		volVal = *attributes.SetVolume / 100
		volAbs = true
	} else if attributes.ChangeVolume != nil {
		volVal = *attributes.ChangeVolume / 100
		volAbs = false
	}

	// Execute change
	switch attributes.Device {

	case "input":
		return errors.Wrap(
			pulseClient.SetSinkInputVolume(attributes.Match, attributes.Mute, volVal, volAbs),
			"Unable to set sink input volume",
		)

	case "sink":
		return errors.Wrap(
			pulseClient.SetSinkVolume(attributes.Match, attributes.Mute, volVal, volAbs),
			"Unable to set sink volume",
		)

	case "source":
		return errors.Wrap(
			pulseClient.SetSourceVolume(attributes.Match, attributes.Mute, volVal, volAbs),
			"Unable to set source volume",
		)

	default:
		return errors.Errorf("Unsupported device type: %q", attributes.Device)

	}
}
