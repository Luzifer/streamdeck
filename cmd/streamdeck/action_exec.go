package main

import (
	"os/exec"

	"github.com/pkg/errors"
)

func init() {
	registerAction("exec", actionExec{})
}

type actionExec struct{}

func (actionExec) Execute(attributes map[string]interface{}) error {
	cmd, ok := attributes["command"].([]interface{})
	if !ok {
		return errors.New("No command supplied")
	}

	var args []string
	for _, c := range cmd {
		if v, ok := c.(string); ok {
			args = append(args, v)
			continue
		}
		return errors.New("Command conatins non-string argument")
	}

	command := exec.Command(args[0], args[1:]...)
	if err := command.Start(); err != nil {
		return errors.Wrap(err, "Unable to start command")
	}

	return errors.Wrap(command.Process.Release(), "Unable to release process")
}
