// Package keypress provides keyboard input actions.
package keypress

import (
	"errors"
	"fmt"
	"time"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/modules/opts"
	"github.com/sashko/go-uinput"
)

type (
	// Action sends configured key presses through the runtime keyboard.
	Action struct{}

	// Attrs contains configuration for the key press action.
	Attrs struct {
		Delay    time.Duration `json:"delay,omitempty" yaml:"delay,omitempty"`
		KeyCodes []int         `json:"key_codes,omitempty" yaml:"key_codes,omitempty"`
		ModAlt   bool          `json:"mod_alt,omitempty" yaml:"mod_alt,omitempty"`
		ModCtrl  bool          `json:"mod_ctrl,omitempty" yaml:"mod_ctrl,omitempty"`
		ModShift bool          `json:"mod_shift,omitempty" yaml:"mod_shift,omitempty"`
		ModMeta  bool          `json:"mod_meta,omitempty" yaml:"mod_meta,omitempty"`
	}
)

// Execute sends the configured key sequence.
//
//nolint:gocyclo // only pressing a few keys
func (a Action) Execute(devs opts.Runtime, atts config.DynamicAttributes) (err error) {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return fmt.Errorf("decoding attributes: %w", err)
	}

	if attributes.KeyCodes == nil {
		return fmt.Errorf("no key_codes array present")
	}

	var execCodes []uint16
	for _, k := range attributes.KeyCodes {
		if k < 0 || k > 65535 { //revive:disable-line:add-constant // single-use boundary
			return fmt.Errorf("key-code out of bounds 0..65535: %d", k)
		}

		execCodes = append(execCodes, uint16(k))
	}

	if attributes.ModShift {
		if err := devs.Keyboard.KeyDown(uinput.KeyLeftShift); err != nil {
			return fmt.Errorf("setting Shift key: %w", err)
		}
		defer a.releaseKey(devs.Keyboard, uinput.KeyLeftShift, "shift", &err)
	}

	if attributes.ModAlt {
		if err := devs.Keyboard.KeyDown(uinput.KeyLeftAlt); err != nil {
			return fmt.Errorf("setting Alt key: %w", err)
		}
		defer a.releaseKey(devs.Keyboard, uinput.KeyLeftAlt, "alt", &err)
	}

	if attributes.ModCtrl {
		if err := devs.Keyboard.KeyDown(uinput.KeyLeftCtrl); err != nil {
			return fmt.Errorf("setting Ctrl key: %w", err)
		}
		defer a.releaseKey(devs.Keyboard, uinput.KeyLeftCtrl, "ctrl", &err)
	}

	if attributes.ModMeta {
		if err := devs.Keyboard.KeyDown(uinput.KeyLeftMeta); err != nil {
			return fmt.Errorf("setting Meta key: %w", err)
		}
		defer a.releaseKey(devs.Keyboard, uinput.KeyLeftMeta, "meta", &err)
	}

	for _, kc := range execCodes {
		if err := devs.Keyboard.KeyPress(kc); err != nil {
			return fmt.Errorf("pressing key: %w", err)
		}
		time.Sleep(attributes.Delay)
	}

	return nil
}

func (Action) releaseKey(kbd uinput.Keyboard, key uint16, name string, err *error) {
	if releaseErr := kbd.KeyUp(key); releaseErr != nil {
		*err = errors.Join(*err, fmt.Errorf("releasing %s key: %w", name, releaseErr))
	}
}
