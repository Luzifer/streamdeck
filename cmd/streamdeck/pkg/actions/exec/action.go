// Package exec provides the command execution action.
package exec

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/exec"

	"github.com/Luzifer/go_helpers/env"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/modules/opts"
)

type (
	// Action executes a configured command.
	Action struct{}

	// Attrs contains configuration for the exec action.
	Attrs struct {
		AttachStderr bool              `json:"attach_stderr,omitempty" yaml:"attach_stderr,omitempty"`
		AttachStdout bool              `json:"attach_stdout,omitempty" yaml:"attach_stdout,omitempty"`
		Command      []string          `json:"command,omitempty" yaml:"command,omitempty"`
		Env          map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
		Wait         bool              `json:"wait,omitempty" yaml:"wait,omitempty"`
	}
)

// Execute runs the configured command.
func (Action) Execute(_ opts.Runtime, atts config.DynamicAttributes) (err error) {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return fmt.Errorf("decoding attributes: %w", err)
	}

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
		command.Stderr = os.Stderr
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
