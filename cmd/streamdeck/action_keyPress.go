package main

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sashko/go-uinput"
)

func init() {
	registerAction("key_press", actionKeyPress{})
}

type actionKeyPress struct{}

func (actionKeyPress) Execute(attributes map[string]interface{}) error {
	var (
		delay     time.Duration
		execCodes []uint16
		ok        bool

		keyNames    []interface{}
		keyCodes    []interface{}
		useKeyNames bool
	)

	keyCodes, ok = attributes["key_codes"].([]interface{})
	if !ok {
		keyNames, ok = attributes["keys"].([]interface{})
		if !ok {
			return errors.New("No key_codes or keys array present")
		}
		useKeyNames = true
	}

	if v, ok := attributes["delay"].(string); ok {
		if d, err := time.ParseDuration(v); err == nil {
			delay = d
		}
	}

	if useKeyNames {
		for _, k := range keyNames {
			// Convert misdetections into strings
			switch k.(type) {
			case int:
				k = strconv.Itoa(k.(int))
			}

			if kv, ok := k.(string); ok {
				if kc, ok := uinputKeyMapping[kv]; ok {
					execCodes = append(execCodes, kc)
				} else {
					return errors.Errorf("Unknown key %q", kv)
				}
			} else {
				return errors.New("Unknown key type detected")
			}
		}
	} else {
		for _, k := range keyCodes {
			execCodes = append(execCodes, uint16(k.(int)))
		}
	}

	if v, ok := attributes["mod_shift"].(bool); ok && v {
		if err := kbd.KeyDown(uinput.KeyLeftShift); err != nil {
			return errors.Wrap(err, "Unable to set shift")
		}
		defer kbd.KeyUp(uinput.KeyLeftShift)
	}

	if v, ok := attributes["mod_alt"].(bool); ok && v {
		if err := kbd.KeyDown(uinput.KeyLeftAlt); err != nil {
			return errors.Wrap(err, "Unable to set shift")
		}
		defer kbd.KeyUp(uinput.KeyLeftAlt)
	}

	if v, ok := attributes["mod_ctrl"].(bool); ok && v {
		if err := kbd.KeyDown(uinput.KeyLeftCtrl); err != nil {
			return errors.Wrap(err, "Unable to set shift")
		}
		defer kbd.KeyUp(uinput.KeyLeftCtrl)
	}

	for _, kc := range execCodes {
		if err := kbd.KeyPress(kc); err != nil {
			return errors.Wrap(err, "Unable to press key")
		}
		time.Sleep(delay)
	}

	return nil
}
