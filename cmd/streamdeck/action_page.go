package main

import (
	"fmt"
	"math"
)

type actionPage struct{}

func init() {
	registerAction("page", actionPage{})
}

func (actionPage) Execute(attributes attributeCollection) (err error) {
	if attributes.Name != "" {
		if err = togglePage(attributes.Name); err != nil {
			return fmt.Errorf("switching page: %w", err)
		}

		return nil
	}

	if absRel := int(math.Abs(float64(attributes.Relative))); absRel != 0 && absRel < len(pageStack) {
		nextPage := pageStack[absRel]
		pageStack = pageStack[absRel+1:]

		if err = togglePage(nextPage); err != nil {
			return fmt.Errorf("switching relative page: %w", err)
		}

		return nil
	}

	return fmt.Errorf("no page name or relative move supplied")
}
