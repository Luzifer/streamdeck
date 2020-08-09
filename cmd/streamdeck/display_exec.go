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

func (d displayElementExec) Display(ctx context.Context, idx int, attributes map[string]interface{}) error {
	var (
		err         error
		imgRenderer = newTextOnImageRenderer()
	)

	// Initialize command
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

	// Execute command and parse it
	var buf = new(bytes.Buffer)

	processEnv := env.ListToMap(os.Environ())

	if e, ok := attributes["env"].(map[interface{}]interface{}); ok {
		for k, v := range e {
			key, ok := k.(string)
			if !ok {
				continue
			}
			value, ok := v.(string)
			if !ok {
				continue
			}

			processEnv[key] = value
		}
	}

	command := exec.Command(args[0], args[1:]...)
	command.Env = env.MapToList(processEnv)
	command.Stdout = buf

	if err := command.Run(); err != nil {
		return errors.Wrap(err, "Command has exit != 0")
	}

	attributes["text"] = strings.TrimSpace(buf.String())

	tmpAttrs := map[string]interface{}{}
	if err = json.Unmarshal(buf.Bytes(), &tmpAttrs); err == nil {
		// Reset text to empty as it was parsable json
		attributes["text"] = ""

		for k, v := range tmpAttrs {
			attributes[k] = v
		}
	}

	// Initialize background
	if filename, ok := attributes["image"].(string); ok {
		if err = imgRenderer.DrawBackgroundFromFile(filename); err != nil {
			return errors.Wrap(err, "Unable to draw background from disk")
		}
	}

	// Initialize color
	var textColor color.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	if rgba, ok := attributes["color"].([]interface{}); ok {
		if len(rgba) != 4 {
			return errors.New("RGBA color definition needs 4 hex values")
		}

		tmpCol := color.RGBA{}

		for idx, vp := range []*uint8{&tmpCol.R, &tmpCol.G, &tmpCol.B, &tmpCol.A} {
			switch rgba[idx].(type) {
			case int:
				*vp = uint8(rgba[idx].(int))
			case float64:
				*vp = uint8(rgba[idx].(float64))
			}
		}

		textColor = tmpCol
	}

	// Initialize fontsize
	var fontsize float64 = 120
	if v, ok := attributes["font_size"].(float64); ok {
		fontsize = v
	}

	var border = 10
	if v, ok := attributes["border"].(int); ok {
		border = v
	}

	if strings.TrimSpace(attributes["text"].(string)) != "" {
		if err = imgRenderer.DrawBigText(strings.TrimSpace(attributes["text"].(string)), fontsize, border, textColor); err != nil {
			return errors.Wrap(err, "Unable to render text")
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

func (d displayElementExec) NeedsLoop(attributes map[string]interface{}) bool {
	if v, ok := attributes["interval"].(int); ok {
		return v > 0
	}

	return false
}

func (d *displayElementExec) StartLoopDisplay(ctx context.Context, idx int, attributes map[string]interface{}) error {
	d.running = true

	var interval = 5 * time.Second
	if v, ok := attributes["interval"].(int); ok {
		interval = time.Duration(v) * time.Second
	}

	go func() {
		for tick := time.NewTicker(interval); ; <-tick.C {
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
