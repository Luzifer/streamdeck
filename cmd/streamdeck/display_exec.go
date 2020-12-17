package main

import (
	"bytes"
	"context"
	"encoding/json"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Luzifer/go_helpers/v2/env"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func init() {
	registerDisplayElement("exec", &displayElementExec{})
}

type displayElementExec struct {
	running bool
}

func (d displayElementExec) Display(ctx context.Context, idx int, attributes attributeCollection) error {
	var (
		err         error
		imgRenderer = newTextOnImageRenderer()
	)

	// Initialize command
	if attributes.Command == nil {
		return errors.New("No command supplied")
	}

	// Execute command and parse it
	buf := new(bytes.Buffer)

	processEnv := env.ListToMap(os.Environ())

	for k, v := range attributes.Env {
		processEnv[k] = v
	}

	command := exec.Command(attributes.Command[0], attributes.Command[1:]...)
	command.Env = env.MapToList(processEnv)
	command.Stdout = buf

	if err := command.Run(); err != nil {
		return errors.Wrap(err, "Command has exit != 0")
	}

	attributes.Text = strings.TrimSpace(buf.String())

	tmpAttrs := attributes.Clone()
	if err = json.Unmarshal(buf.Bytes(), &tmpAttrs); err == nil {
		// Reset text to empty as it was parsable json
		attributes = tmpAttrs
	}

	// Initialize background
	if attributes.Image != "" {
		if err = imgRenderer.DrawBackgroundFromFile(attributes.Image); err != nil {
			return errors.Wrap(err, "Unable to draw background from disk")
		}
	}

	// Initialize color
	var textColor color.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	if attributes.RGBA != nil {
		if len(attributes.RGBA) != 4 {
			return errors.New("RGBA color definition needs 4 hex values")
		}

		textColor = color.RGBA{attributes.RGBA[0], attributes.RGBA[1], attributes.RGBA[2], attributes.RGBA[3]}
	}

	// Initialize fontsize
	var fontsize float64 = 120
	if attributes.FontSize != nil {
		fontsize = *attributes.FontSize
	}

	border := 10
	if attributes.Border != nil {
		border = *attributes.Border
	}

	if strings.TrimSpace(attributes.Text) != "" {
		if err = imgRenderer.DrawBigText(strings.TrimSpace(attributes.Text), fontsize, border, textColor); err != nil {
			return errors.Wrap(err, "Unable to render text")
		}
	}

	if strings.TrimSpace(attributes.Caption) != "" {
		if err = imgRenderer.DrawCaptionText(strings.TrimSpace(attributes.Caption)); err != nil {
			return errors.Wrap(err, "Unable to render caption")
		}
	}

	if !d.running && d.NeedsLoop(attributes) {
		return nil
	}

	if err := ctx.Err(); err != nil {
		// Page context was cancelled, do not draw
		return err
	}

	return errors.Wrap(sd.FillImage(idx, imgRenderer.GetImage()), "Unable to set image")
}

func (d displayElementExec) NeedsLoop(attributes attributeCollection) bool {
	return attributes.Interval > 0
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
