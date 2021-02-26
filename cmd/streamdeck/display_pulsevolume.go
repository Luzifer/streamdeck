// +build linux

package main

import (
	"context"
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func init() {
	registerDisplayElement("pulsevolume", &displayElementPulseVolume{})
}

type displayElementPulseVolume struct{}

func (d displayElementPulseVolume) Display(ctx context.Context, idx int, attributes attributeCollection) error {
	if pulseClient == nil {
		return errors.New("PulseAudio client not initialized")
	}

	if attributes.Device == "" {
		return errors.New("Missing 'device' attribute")
	}

	if attributes.Match == "" {
		return errors.New("Missing 'match' attribute")
	}

	var (
		err        error
		mute       bool
		notPresent bool
		volume     float64
	)

	switch attributes.Device {

	case "input":
		volume, mute, _, _, err = pulseClient.GetSinkInputVolume(attributes.Match)

	case "sink":
		volume, mute, _, _, err = pulseClient.GetSinkVolume(attributes.Match)

	case "source":
		volume, mute, _, _, err = pulseClient.GetSourceVolume(attributes.Match)

	default:
		return errors.Errorf("Unsupported device type: %q", attributes.Device)

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
	if attributes.RGBA != nil {
		if len(attributes.RGBA) != 4 {
			return errors.New("RGBA color definition needs 4 hex values")
		}

		textColor = attributes.RGBAToColor()
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

	if err = img.DrawBigText(text, fontsize, border, textColor); err != nil {
		return errors.Wrap(err, "Unable to draw text")
	}

	if strings.TrimSpace(attributes.Caption) != "" {
		if err = img.DrawCaptionText(strings.TrimSpace(attributes.Caption)); err != nil {
			return errors.Wrap(err, "Unable to render caption")
		}
	}

	if err = ctx.Err(); err != nil {
		// Page context was cancelled, do not draw
		return err
	}

	return errors.Wrap(sd.FillImage(idx, img.GetImage()), "Unable to set image")
}

func (d displayElementPulseVolume) NeedsLoop(attributes attributeCollection) bool { return true }

func (d *displayElementPulseVolume) StartLoopDisplay(ctx context.Context, idx int, attributes attributeCollection) error {
	interval := time.Second
	if attributes.Interval > 100*time.Millisecond {
		interval = attributes.Interval
	} else if attributes.Interval > 0 {
		log.WithFields(log.Fields{
			"tpye":     "pulsevolume",
			"idx":      idx,
			"interval": attributes.Interval,
		}).Warn("Ignoring interval below 100ms")
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
