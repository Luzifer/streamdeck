package main

import (
	"context"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/pkg/errors"
)

func init() {
	registerDisplayElement("text", &displayElementText{})
}

type displayElementText struct {
	running bool
}

func (d displayElementText) Display(ctx context.Context, idx int, attributes attributeCollection) error {
	var (
		err         error
		imgRenderer = newTextOnImageRenderer()
	)

	// Initialize background
	if attributes.BackgroundColor != nil {
		if len(attributes.BackgroundColor) != 4 {
			return errors.New("Background color definition needs 4 hex values")
		}

		if err := ctx.Err(); err != nil {
			// Page context was cancelled, do not draw
			return err
		}

		imgRenderer.DrawBackgroundColor(attributes.BackgroundToColor())
	}

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

func (d displayElementText) NeedsLoop(attributes attributeCollection) bool { return false }
