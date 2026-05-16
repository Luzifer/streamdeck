// Package exec provides command-backed display elements.
package exec

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/env"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/displays/text"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/modules/opts"
)

const minLoopInterval = 100 * time.Millisecond

type (
	// Display renders text attributes produced by a command.
	Display struct{}

	// Attrs contains configuration for the exec display.
	Attrs struct {
		Command  []string          `json:"command,omitempty" yaml:"command,omitempty"`
		Env      map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
		Interval time.Duration     `json:"interval,omitempty" yaml:"interval,omitempty"`

		text.Attrs `yaml:",inline"`
	}
)

// Display executes the command and renders its output as a text display.
func (Display) Display(ctx context.Context, idx int, devs opts.Runtime, atts config.DynamicAttributes) (err error) {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return fmt.Errorf("decoding attributes: %w", err)
	}

	// Initialize command
	if attributes.Command == nil {
		return fmt.Errorf("no command supplied")
	}

	// Execute command and parse it
	buf := new(bytes.Buffer)

	processEnv := env.ListToMap(os.Environ())

	maps.Copy(processEnv, attributes.Env)

	command := exec.CommandContext(ctx, attributes.Command[0], attributes.Command[1:]...) //#nosec:G204 // intended to run user-defined command
	command.Env = env.MapToList(processEnv)
	command.Stdout = buf

	if err := command.Run(); err != nil {
		return fmt.Errorf("running command: %w", err)
	}

	forwardAtts := attributes.Attrs
	forwardAtts.Text = strings.TrimSpace(buf.String())

	tmpAttrs := attributes.Attrs
	if err = json.Unmarshal(buf.Bytes(), &tmpAttrs); err == nil {
		// Reset text to empty as it was parsable json
		forwardAtts = tmpAttrs
	}

	return new(text.Display).Render(ctx, idx, devs, forwardAtts) //nolint:wrapcheck // fine for this as that's a normal render module itself
}

// NeedsLoop reports whether the display should refresh periodically.
func (Display) NeedsLoop(atts config.DynamicAttributes) bool {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return false
	}

	if attributes.Interval > 0 && attributes.Interval < minLoopInterval {
		log.WithFields(log.Fields{
			"type":     "exec",
			"interval": attributes.Interval,
		}).Warn("Ignoring interval below 100ms")
	}

	return attributes.Interval > minLoopInterval
}

// StartLoopDisplay starts periodic display refresh until the context is cancelled.
func (d *Display) StartLoopDisplay(ctx context.Context, idx int, devs opts.Runtime, atts config.DynamicAttributes) error {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return fmt.Errorf("decoding attributes: %w", err)
	}

	go func() {
		tick := time.NewTicker(attributes.Interval)
		defer tick.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-tick.C:
				if err := d.Display(ctx, idx, devs, atts); err != nil {
					if errors.Is(ctx.Err(), context.Canceled) {
						return
					}

					log.WithError(err).Error("refreshing element")
				}
			}
		}
	}()

	return nil
}
