// Package text provides text display elements.
package text

import (
	"context"
	"fmt"
	"image/color"
	"strings"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/helpers"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/modules/opts"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/renderer"
)

const defaultTextToImageBorderSize = 10 // Pixel

type (
	// Display renders text on a key.
	Display struct{}

	// Attrs contains configuration for the text display.
	Attrs struct {
		BackgroundColor []int    `json:"background_color,omitempty" yaml:"background_color,omitempty"`
		Image           string   `json:"image,omitempty" yaml:"image,omitempty"`
		RGBA            []int    `json:"rgba,omitempty" yaml:"rgba,omitempty"`
		FontSize        *float64 `json:"font_size,omitempty" yaml:"font_size,omitempty"`
		Border          *int     `json:"border,omitempty" yaml:"border,omitempty"`
		Caption         string   `json:"caption,omitempty" yaml:"caption,omitempty"`
		Text            string   `json:"text,omitempty" yaml:"text,omitempty"`
	}
)

// Display decodes attributes and renders text on the selected key.
func (d Display) Display(ctx context.Context, idx int, devs opts.Runtime, atts config.DynamicAttributes) (err error) {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return fmt.Errorf("decoding attributes: %w", err)
	}

	return d.Render(ctx, idx, devs, attributes)
}

// Render renders already-decoded text attributes on the selected key.
//
//nolint:gocyclo // better to keep it together
func (Display) Render(ctx context.Context, idx int, devs opts.Runtime, attributes Attrs) (err error) {
	imgRenderer := renderer.NewTextOnImageRenderer(devs)

	// Initialize background
	if attributes.BackgroundColor != nil {
		if err := ctx.Err(); err != nil {
			// Page context was cancelled, do not draw
			return fmt.Errorf("page context cancelled: %w", err)
		}

		bgColor, err := helpers.Int4ToRGBA(attributes.BackgroundColor)
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
		if textColor, err = helpers.Int4ToRGBA(attributes.RGBA); err != nil {
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

	if err = ctx.Err(); err != nil {
		// Page context was cancelled, do not draw
		return fmt.Errorf("page context cancelled: %w", err)
	}

	if err = devs.Deck.FillImage(idx, imgRenderer.GetImage()); err != nil {
		return fmt.Errorf("setting image: %w", err)
	}

	return nil
}
