package main

import (
	"math"

	"github.com/pkg/errors"
)

func init() {
	registerAction("page", actionPage{})
}

type actionPage struct{}

func (actionPage) Execute(attributes map[string]interface{}) error {
	name, nameOk := attributes["name"].(string)
	relative, relativeOk := attributes["relative"].(int)

	if nameOk && name != "" {
		return errors.Wrap(togglePage(name), "switch page")
	}

	if absRel := int(math.Abs(float64(relative))); relativeOk && absRel != 0 && absRel < len(pageStack) {
		nextPage := pageStack[absRel]
		pageStack = pageStack[absRel+1:]
		return errors.Wrap(togglePage(nextPage), "switch relative page")
	}

	return errors.New("no page name or relative move supplied")
}
