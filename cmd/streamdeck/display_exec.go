package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"maps"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/v2/env"
	log "github.com/sirupsen/logrus"
)

const minLoopInterval = 100 * time.Millisecond

type displayElementExec struct {
	running bool
}

func init() {
	registerDisplayElement("exec", &displayElementExec{})
}

func (d displayElementExec) Display(ctx context.Context, idx int, attributes attributeCollection) (err error) {
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

	attributes.Text = strings.TrimSpace(buf.String())

	tmpAttrs := attributes.Clone()
	if err = json.Unmarshal(buf.Bytes(), &tmpAttrs); err == nil {
		// Reset text to empty as it was parsable json
		attributes = tmpAttrs
	}

	if !d.running && d.NeedsLoop(attributes) {
		return nil
	}

	return displayElementText{}.Display(ctx, idx, attributes)
}

func (displayElementExec) NeedsLoop(attributes attributeCollection) bool {
	if attributes.Interval > 0 && attributes.Interval < minLoopInterval {
		log.WithFields(log.Fields{
			"type":     "exec",
			"interval": attributes.Interval,
		}).Warn("Ignoring interval below 100ms")
	}

	return attributes.Interval > minLoopInterval
}

func (d *displayElementExec) StartLoopDisplay(ctx context.Context, idx int, attributes attributeCollection) error {
	d.running = true

	go func() {
		for tick := time.NewTicker(attributes.Interval); ; <-tick.C {
			if !d.running {
				return
			}

			if err := d.Display(ctx, idx, attributes); err != nil {
				log.WithError(err).Error("Unable to refresh element")
			}
		}
	}()

	return nil
}

func (d *displayElementExec) StopLoopDisplay() error {
	d.running = false
	return nil
}
