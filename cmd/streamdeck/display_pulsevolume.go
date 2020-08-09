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

	sinkMatch, sinkOK := attributes["sink"].(string)
	sinkOK = sinkOK && sinkMatch != ""

	sinkInputMatch, sinkInputOK := attributes["sink_input"].(string)
	sinkInputOK = sinkInputOK && sinkInputMatch != ""

	var (
		err    error
		volume float64
	)

	switch {

	case (sinkInputOK && sinkOK) || (!sinkInputOK && !sinkOK):
		return errors.New("Exactly one of 'sink' and 'sink_input' must be specified")

	case sinkInputOK:
		volume, err = pulseClient.GetSinkInputVolume(sinkInputMatch)

	case sinkOK:
		volume, err = pulseClient.GetSinkVolume(sinkMatch)

	}

	if err != nil {
		return errors.Wrap(err, "Unable to get volume")
	}

	img := newTextOnImageRenderer()

	// Initialize color
	var textColor color.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
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

	if err = img.DrawBigText(fmt.Sprintf("%.0f%%", volume*100), fontsize, border, textColor); err != nil {
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
