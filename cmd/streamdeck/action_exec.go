package main

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"

	"github.com/Luzifer/go_helpers/env"
)

type actionExec struct{}

func init() {
	registerAction("exec", actionExec{})
}

func (actionExec) Execute(attributes attributeCollection) (err error) {
	if attributes.Command == nil {
		return fmt.Errorf("no command supplied")
	}

	processEnv := env.ListToMap(os.Environ())

	maps.Copy(processEnv, attributes.Env)

	//#nosec:G204 // intended to run user-provided command
	command := exec.CommandContext(context.Background(), attributes.Command[0], attributes.Command[1:]...)
	command.Env = env.MapToList(processEnv)

	if attributes.AttachStdout {
		command.Stdout = os.Stdout
	}

	if attributes.AttachStderr {
		command.Stdout = os.Stderr
	}

	if err = command.Start(); err != nil {
		return fmt.Errorf("starting command: %w", err)
	}

	// If "wait" is set to true start command and wait for execution
	if attributes.Wait {
		if err = command.Wait(); err != nil {
			return fmt.Errorf("waiting for command: %w", err)
		}

		return nil
	}

	// We don't wait so we release the process and don't care anymore
	if err = command.Process.Release(); err != nil {
		return fmt.Errorf("releasing process: %w", err)
	}

	return nil
}
