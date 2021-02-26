package main

import (
	"context"
	"github.com/pkg/errors"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

func init() {
	registerDisplayElement("text", &displayElementText{})
}

type displayElementText struct {
	running bool
}

func (d displayElementText) Display(ctx context.Context, idx int, attributes map[string]interface{}) error {
	var (
		err         error
		imgRenderer = newTextOnImageRenderer()
	)

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

	border := 10
	if v, ok := attributes["border"].(int); ok {
		border = v
	}

	if strings.TrimSpace(attributes["text"].(string)) != "" {
		if err = imgRenderer.DrawBigText(strings.TrimSpace(attributes["text"].(string)), fontsize, border, textColor); err != nil {
			return errors.Wrap(err, "Unable to render text")
		}
	}

	if caption, ok := attributes["caption"].(string); ok && strings.TrimSpace(caption) != "" {
		if err = imgRenderer.DrawCaptionText(strings.TrimSpace(caption)); err != nil {
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

func (d displayElementText) NeedsLoop(attributes map[string]interface{}) bool { return false }
