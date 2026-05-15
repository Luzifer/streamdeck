package main

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	log "github.com/sirupsen/logrus"
)

const errorDisplayElementType = "color"

type (
	action interface {
		Execute(attributes attributeCollection) error
	}

	displayElement interface {
		Display(ctx context.Context, idx int, attributes attributeCollection) error
	}

	refreshingDisplayElement interface {
		NeedsLoop(attributes attributeCollection) bool
		StartLoopDisplay(ctx context.Context, idx int, attributes attributeCollection) error
		StopLoopDisplay() error
	}
)

var errorDisplayElementAttributes = attributeCollection{
	RGBA: []int{0xff, 0x0, 0x0, 0xff},
}

var (
	registeredActions             = make(map[string]reflect.Type)
	registeredActionsLock         sync.Mutex
	registeredDisplayElements     = make(map[string]reflect.Type)
	registeredDisplayElementsLock sync.Mutex
)

func registerAction(name string, handler action) {
	registeredActionsLock.Lock()
	defer registeredActionsLock.Unlock()

	registeredActions[name] = reflect.TypeOf(handler)
}

func registerDisplayElement(name string, handler displayElement) {
	registeredDisplayElementsLock.Lock()
	defer registeredDisplayElementsLock.Unlock()

	registeredDisplayElements[name] = reflect.TypeOf(handler)
}

func callAction(a dynamicElement) (err error) {
	t, ok := registeredActions[a.Type]
	if !ok {
		return fmt.Errorf("unknown action type %q", a.Type)
	}

	inst := reflect.New(t).Interface().(action)
	if err = inst.Execute(a.Attributes); err != nil {
		return fmt.Errorf("calling action: %w", err)
	}

	return nil
}

func callDisplayElement(ctx context.Context, idx int, kd keyDefinition) (err error) {
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

	if t.Implements(reflect.TypeFor[refreshingDisplayElement]()) &&
		inst.(refreshingDisplayElement).NeedsLoop(kd.Display.Attributes) {
		log.WithFields(log.Fields{
			"key":          idx,
			"display_type": kd.Display.Type,
		}).Debug("Starting loop")
		activeLoops = append(activeLoops, inst.(refreshingDisplayElement))

		if err = inst.(refreshingDisplayElement).StartLoopDisplay(ctx, idx, kd.Display.Attributes); err != nil {
			return fmt.Errorf("starting display-loop: %w", err)
		}

		return nil
	}

	if err = inst.(displayElement).Display(ctx, idx, kd.Display.Attributes); err != nil {
		return fmt.Errorf("displaying element: %w", err)
	}

	return nil
}

func callErrorDisplayElement(ctx context.Context, idx int) (err error) {
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

	if err = inst.(displayElement).Display(ctx, idx, errorDisplayElementAttributes); err != nil {
		return fmt.Errorf("displaying error element: %w", err)
	}

	return nil
}
