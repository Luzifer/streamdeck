package main

import (
	"context"
	"fmt"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"strings"
)

const defaultTextToImageBorderSize = 10 // Pixel

type displayElementText struct {
	running bool
}

func init() {
	registerDisplayElement("text", &displayElementText{})
}

//nolint:gocyclo // single rendering flow
func (d displayElementText) Display(ctx context.Context, idx int, attributes attributeCollection) (err error) {
	imgRenderer := newTextOnImageRenderer()

	// Initialize background
	if attributes.BackgroundColor != nil {
		if err := ctx.Err(); err != nil {
			// Page context was cancelled, do not draw
			return fmt.Errorf("page context cancelled: %w", err)
		}

		bgColor, err := int4ToRGBA(attributes.BackgroundColor)
		if err != nil {
			return fmt.Errorf("invalid 'background_color' color definition: %w", err)
		}

		imgRenderer.DrawBackgroundColor(bgColor)
	}

	if attributes.Image != "" {
		if err = imgRenderer.DrawBackgroundFromFile(attributes.Image); err != nil {
			return fmt.Errorf("drawing background from disk: %w", err)
		}
	}

	// Initialize color
	var textColor color.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	if attributes.RGBA != nil {
		if textColor, err = int4ToRGBA(attributes.RGBA); err != nil {
			return fmt.Errorf("invalid 'rgba' color definition: %w", err)
		}
	}

	// Initialize fontsize
	var fontsize float64 = 120
	if attributes.FontSize != nil {
		fontsize = *attributes.FontSize
	}

	border := defaultTextToImageBorderSize
	if attributes.Border != nil {
		border = *attributes.Border
	}

	if strings.TrimSpace(attributes.Text) != "" {
		if err = imgRenderer.DrawBigText(strings.TrimSpace(attributes.Text), fontsize, border, textColor); err != nil {
			return fmt.Errorf("rendering text: %w", err)
		}
	}

	if strings.TrimSpace(attributes.Caption) != "" {
		if err = imgRenderer.DrawCaptionText(strings.TrimSpace(attributes.Caption)); err != nil {
			return fmt.Errorf("rendering caption: %w", err)
		}
	}

	if !d.running && d.NeedsLoop(attributes) {
		return nil
	}

	if err = ctx.Err(); err != nil {
		// Page context was cancelled, do not draw
		return fmt.Errorf("page context cancelled: %w", err)
	}

	if err = sd.FillImage(idx, imgRenderer.GetImage()); err != nil {
		return fmt.Errorf("setting image: %w", err)
	}

	return nil
}

func (displayElementText) NeedsLoop(_ attributeCollection) bool { return false }
