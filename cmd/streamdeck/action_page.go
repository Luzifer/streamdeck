package main

import "github.com/pkg/errors"

func init() {
	registerAction("page", actionPage{})
}

type actionPage struct{}

func (actionPage) Execute(attributes map[string]interface{}) error {
	name, ok := attributes["name"].(string)
	if !ok {
		return errors.New("No page name supplied")
	}

	return errors.Wrap(togglePage(name), "Unable to switch page")
}
