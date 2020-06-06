package main

import (
	"context"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type action interface {
	Execute(attributes map[string]interface{}) error
}

type displayElement interface {
	Display(ctx context.Context, idx int, attributes map[string]interface{}) error
}

type refreshingDisplayElement interface {
	NeedsLoop(attributes map[string]interface{}) bool
	StartLoopDisplay(ctx context.Context, idx int, attributes map[string]interface{}) error
	StopLoopDisplay() error
}

var (
	registeredActions             = map[string]reflect.Type{}
	registeredActionsLock         sync.Mutex
	registeredDisplayElements     = map[string]reflect.Type{}
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

func callAction(a dynamicElement) error {
	t, ok := registeredActions[a.Type]
	if !ok {
		return errors.Errorf("Unknown action type %q", a.Type)
	}

	inst := reflect.New(t).Interface().(action)

	return inst.Execute(a.Attributes)
}

func callDisplayElement(ctx context.Context, idx int, kd keyDefinition) error {
	t, ok := registeredDisplayElements[kd.Display.Type]
	if !ok {
		return errors.Errorf("Unknown display type %q", kd.Display.Type)
	}

	var inst interface{}
	if t.Kind() == reflect.Ptr {
		inst = reflect.New(t.Elem()).Interface()
	} else {
		inst = reflect.New(t).Interface()
	}

	if t.Implements(reflect.TypeOf((*refreshingDisplayElement)(nil)).Elem()) &&
		inst.(refreshingDisplayElement).NeedsLoop(kd.Display.Attributes) {
		log.WithFields(log.Fields{
			"key":          idx,
			"display_type": kd.Display.Type,
		}).Debug("Starting loop")
		activeLoops = append(activeLoops, inst.(refreshingDisplayElement))
		return inst.(refreshingDisplayElement).StartLoopDisplay(ctx, idx, kd.Display.Attributes)
	}

	return inst.(displayElement).Display(ctx, idx, kd.Display.Attributes)
}
