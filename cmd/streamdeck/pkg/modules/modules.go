// Package modules registers and invokes configured actions and displays.
package modules

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/displays/color"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/modules/opts"
	log "github.com/sirupsen/logrus"
)

const errorDisplayElementType = "color"

type (
	// Action is implemented by executable StreamDeck actions.
	Action interface {
		// Execute runs the action with the provided runtime and attributes.
		Execute(dev opts.Runtime, attributes config.DynamicAttributes) error
	}

	// DisplayElement is implemented by key display renderers.
	DisplayElement interface {
		// Display renders the element on the selected key.
		Display(ctx context.Context, idx int, dev opts.Runtime, attributes config.DynamicAttributes) error
	}

	// RefreshingDisplayElement is implemented by displays needing periodic refresh.
	RefreshingDisplayElement interface {
		// NeedsLoop reports whether the display should be refreshed periodically.
		NeedsLoop(attributes config.DynamicAttributes) bool
		// StartLoopDisplay starts periodic rendering until the context is cancelled.
		StartLoopDisplay(ctx context.Context, idx int, dev opts.Runtime, attributes config.DynamicAttributes) error
	}
)

var (
	registeredActions             = make(map[string]reflect.Type)
	registeredActionsLock         sync.Mutex
	registeredDisplayElements     = make(map[string]reflect.Type)
	registeredDisplayElementsLock sync.Mutex
)

func registerAction(name string, handler Action) {
	registeredActionsLock.Lock()
	defer registeredActionsLock.Unlock()

	registeredActions[name] = reflect.TypeOf(handler)
}

func registerDisplayElement(name string, handler DisplayElement) {
	registeredDisplayElementsLock.Lock()
	defer registeredDisplayElementsLock.Unlock()

	registeredDisplayElements[name] = reflect.TypeOf(handler)
}

// CallAction instantiates and executes a registered action.
func CallAction(dev opts.Runtime, a config.DynamicElement) (err error) {
	t, ok := registeredActions[a.Type]
	if !ok {
		return fmt.Errorf("unknown action type %q", a.Type)
	}

	inst := reflect.New(t).Interface().(Action)
	if err = inst.Execute(dev, a.Attributes); err != nil {
		return fmt.Errorf("calling action: %w", err)
	}

	return nil
}

// CallDisplayElement instantiates and renders a registered display element.
func CallDisplayElement(ctx context.Context, idx int, dev opts.Runtime, kd config.KeyDefinition) (err error) {
	t, ok := registeredDisplayElements[kd.Display.Type]
	if !ok {
		return fmt.Errorf("unknown display type %q", kd.Display.Type)
	}

	var inst any
	if t.Kind() == reflect.Pointer {
		inst = reflect.New(t.Elem()).Interface()
	} else {
		inst = reflect.New(t).Interface()
	}

	if loop, ok := inst.(RefreshingDisplayElement); ok && loop.NeedsLoop(kd.Display.Attributes) {
		log.WithFields(log.Fields{
			"key":          idx,
			"display_type": kd.Display.Type,
		}).Debug("starting loop")

		if err = loop.StartLoopDisplay(ctx, idx, dev, kd.Display.Attributes); err != nil {
			return fmt.Errorf("starting display-loop: %w", err)
		}

		return nil
	}

	if err = inst.(DisplayElement).
		Display(ctx, idx, dev, kd.Display.Attributes); err != nil {
		return fmt.Errorf("displaying element: %w", err)
	}

	return nil
}

// CallErrorDisplayElement renders the fallback error display on a key.
func CallErrorDisplayElement(ctx context.Context, idx int, dev opts.Runtime) (err error) {
	t, ok := registeredDisplayElements[errorDisplayElementType]
	if !ok {
		return fmt.Errorf("unknown display type %q", errorDisplayElementType)
	}

	var inst any
	if t.Kind() == reflect.Pointer {
		inst = reflect.New(t.Elem()).Interface()
	} else {
		inst = reflect.New(t).Interface()
	}

	attrs, err := config.EncodeAttributes(color.Attrs{
		RGBA: []int{0xff, 0x0, 0x0, 0xff},
	})
	if err != nil {
		return fmt.Errorf("encoding attributes: %w", err)
	}

	if err = inst.(DisplayElement).Display(ctx, idx, dev, attrs); err != nil {
		return fmt.Errorf("displaying error element: %w", err)
	}

	return nil
}
