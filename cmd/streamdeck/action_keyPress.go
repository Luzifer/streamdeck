package main

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sashko/go-uinput"
)

func init() {
	registerAction("key_press", actionKeyPress{})
}

type actionKeyPress struct{}

func (actionKeyPress) Execute(attributes attributeCollection) error {
	var (
		execCodes []uint16

		keyNames    []string
		keyCodes    []int
		useKeyNames bool
	)

	keyCodes = attributes.KeyCodes
	if keyCodes == nil {
		keyNames = attributes.Keys
		if keyNames == nil {
			return errors.New("No key_codes or keys array present")
		}
		useKeyNames = true
	}

	if useKeyNames {
		for _, k := range keyNames {
			if kc, ok := uinputKeyMapping[k]; ok {
				execCodes = append(execCodes, kc)
			} else {
				return errors.Errorf("Unknown key %q", k)
			}
		}
	} else {
		for _, k := range keyCodes {
			execCodes = append(execCodes, uint16(k))
		}
	}

	if attributes.ModShift {
		if err := kbd.KeyDown(uinput.KeyLeftShift); err != nil {
			return errors.Wrap(err, "Unable to set Shift key")
		}
		defer kbd.KeyUp(uinput.KeyLeftShift)
	}

	if attributes.ModAlt {
		if err := kbd.KeyDown(uinput.KeyLeftAlt); err != nil {
			return errors.Wrap(err, "Unable to set Alt key")
		}
		defer kbd.KeyUp(uinput.KeyLeftAlt)
	}

	if attributes.ModCtrl {
		if err := kbd.KeyDown(uinput.KeyLeftCtrl); err != nil {
			return errors.Wrap(err, "Unable to set Ctrl key")
		}
		defer kbd.KeyUp(uinput.KeyLeftCtrl)
	}

	if attributes.ModMeta {
		if err := kbd.KeyDown(uinput.KeyLeftMeta); err != nil {
			return errors.Wrap(err, "Unable to set Meta key")
		}
		defer kbd.KeyUp(uinput.KeyLeftMeta)
	}

	for _, kc := range execCodes {
		if err := kbd.KeyPress(kc); err != nil {
			return errors.Wrap(err, "Unable to press key")
		}
		time.Sleep(attributes.Delay)
	}

	return nil
}
