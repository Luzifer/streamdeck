// +build linux

package main

import (
	"context"
	"fmt"
	"image/color"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func init() {
	registerDisplayElement("pulsevolume", &displayElementPulseVolume{})
}

type displayElementPulseVolume struct{}

func (d displayElementPulseVolume) Display(ctx context.Context, idx int, attributes map[string]interface{}) error {
	if pulseClient == nil {
		return errors.New("PulseAudio client not initialized")
	}

	devType, ok := attributes["device"].(string)
	if !ok {
		return errors.New("Missing 'device' attribute")
	}

	match, ok := attributes["match"].(string)
	if !ok {
		return errors.New("Missing 'match' attribute")
	}

	var (
		err        error
		mute       bool
		notPresent bool
		volume     float64
	)

	switch devType {

	case "input":
		volume, mute, err = pulseClient.GetSinkInputVolume(match)

	case "sink":
		volume, mute, err = pulseClient.GetSinkVolume(match)

	case "source":
		volume, mute, err = pulseClient.GetSourceVolume(match)

	default:
		return errors.Errorf("Unsupported device type: %q", devType)

	}

	if err == errPulseNoSuchDevice {
		notPresent = true
	} else if err != nil {
		return errors.Wrap(err, "Unable to get volume")
	}

	img := newTextOnImageRenderer()

	var (
		text                  = fmt.Sprintf("%.0f%%", volume*100)
		textColor color.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	)

	if notPresent {
		text = "--"
		textColor = color.RGBA{0xff, 0x0, 0x0, 0x0}
	} else if mute {
		text = "M"
		textColor = color.RGBA{0xff, 0x0, 0x0, 0x0}
	}

	// Initialize color
	if rgba, ok := attributes["color"].([]interface{}); ok {
		if len(rgba) != 4 {
			return errors.New("RGBA color definition needs 4 hex values")
		}

		tmpCol := color.RGBA{}

		for cidx, vp := range []*uint8{&tmpCol.R, &tmpCol.G, &tmpCol.B, &tmpCol.A} {
			switch rgba[cidx].(type) {
			case int:
				*vp = uint8(rgba[cidx].(int))
			case float64:
				*vp = uint8(rgba[cidx].(float64))
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

	if err = img.DrawBigText(text, fontsize, border, textColor); err != nil {
		return errors.Wrap(err, "Unable to draw text")
	}

	if err = ctx.Err(); err != nil {
		// Page context was cancelled, do not draw
		return err
	}

	return errors.Wrap(sd.FillImage(idx, img.GetImage()), "Unable to set image")
}

func (d displayElementPulseVolume) NeedsLoop(attributes map[string]interface{}) bool { return true }

func (d *displayElementPulseVolume) StartLoopDisplay(ctx context.Context, idx int, attributes map[string]interface{}) error {
	var interval = time.Second
	if v, ok := attributes["interval"].(int); ok {
		interval = time.Duration(v) * time.Second
	}

	go func() {
		for tick := time.NewTicker(interval); ; <-tick.C {
			if ctx.Err() != nil {
				return
			}

			if err := d.Display(ctx, idx, attributes); err != nil {
				log.WithError(err).Error("Unable to refresh element")
			}
		}
	}()

	return nil
}

func (d *displayElementPulseVolume) StopLoopDisplay() error {
	return nil
}
