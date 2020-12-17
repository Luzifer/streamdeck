package main

import (
	"math"

	"github.com/pkg/errors"
)

func init() {
	registerAction("page", actionPage{})
}

type actionPage struct{}

func (actionPage) Execute(attributes attributeCollection) error {
	if attributes.Name != "" {
		return errors.Wrap(togglePage(attributes.Name), "switch page")
	}

	if absRel := int(math.Abs(float64(attributes.Relative))); absRel != 0 && absRel < len(pageStack) {
		nextPage := pageStack[absRel]
		pageStack = pageStack[absRel+1:]
		return errors.Wrap(togglePage(nextPage), "switch relative page")
	}

	return errors.New("no page name or relative move supplied")
}
