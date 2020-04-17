package main

import (
	"os"
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
	command.Env = os.Environ()

	if err := command.Start(); err != nil {
		return errors.Wrap(err, "Unable to start command")
	}

	// If "wait" is set and set to true start command and wait for execution
	if v, ok := attributes["wait"].(bool); ok && v {
		return errors.Wrap(command.Wait(), "Unable to execute command")
	}

	// We don't wait so we release the process and don't care anymore
	return errors.Wrap(command.Process.Release(), "Unable to release process")
}
