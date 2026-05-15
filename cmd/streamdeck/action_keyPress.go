package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/sashko/go-uinput"
)

type actionKeyPress struct{}

func init() {
	registerAction("key_press", actionKeyPress{})
}

//nolint:gocyclo // only pressing a few keys
func (a actionKeyPress) Execute(attributes attributeCollection) (err error) {
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
			return fmt.Errorf("no key_codes or keys array present")
		}
		useKeyNames = true
	}

	if useKeyNames {
		for _, k := range keyNames {
			kc, ok := uinputKeyMapping[k]
			if !ok {
				return fmt.Errorf("unknown key %q", k)
			}
			execCodes = append(execCodes, kc)
		}
	} else {
		for _, k := range keyCodes {
			if k < 0 || k > 65535 { //revive:disable-line:add-constant // single-use boundary
				return fmt.Errorf("key-code out of bounds 0..65535: %d", k)
			}

			execCodes = append(execCodes, uint16(k))
		}
	}

	if attributes.ModShift {
		if err := kbd.KeyDown(uinput.KeyLeftShift); err != nil {
			return fmt.Errorf("setting Shift key: %w", err)
		}
		defer a.releaseKey(uinput.KeyLeftShift, "shift", &err)
	}

	if attributes.ModAlt {
		if err := kbd.KeyDown(uinput.KeyLeftAlt); err != nil {
			return fmt.Errorf("setting Alt key: %w", err)
		}
		defer a.releaseKey(uinput.KeyLeftAlt, "alt", &err)
	}

	if attributes.ModCtrl {
		if err := kbd.KeyDown(uinput.KeyLeftCtrl); err != nil {
			return fmt.Errorf("setting Ctrl key: %w", err)
		}
		defer a.releaseKey(uinput.KeyLeftCtrl, "ctrl", &err)
	}

	if attributes.ModMeta {
		if err := kbd.KeyDown(uinput.KeyLeftMeta); err != nil {
			return fmt.Errorf("setting Meta key: %w", err)
		}
		defer a.releaseKey(uinput.KeyLeftMeta, "meta", &err)
	}

	for _, kc := range execCodes {
		if err := kbd.KeyPress(kc); err != nil {
			return fmt.Errorf("pressing key: %w", err)
		}
		time.Sleep(attributes.Delay)
	}

	return nil
}

func (actionKeyPress) releaseKey(key uint16, name string, err *error) {
	if releaseErr := kbd.KeyUp(key); releaseErr != nil {
		*err = errors.Join(*err, fmt.Errorf("releasing %s key: %w", name, releaseErr))
	}
}
