// Package config loads and decodes StreamDeck configuration files.
package config

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v3"

	"github.com/Luzifer/streamdeck/v2"
)

const defaultLongPressDuration = 500 * time.Millisecond

const (
	// CaptionPositionBottom places captions at the bottom of rendered keys.
	CaptionPositionBottom = "bottom"
	// CaptionPositionEmpty uses the default caption position.
	CaptionPositionEmpty = ""
	// CaptionPositionTop places captions at the top of rendered keys.
	CaptionPositionTop = "top"
)

type (
	// CaptionPosition defines where captions are rendered on keys.
	CaptionPosition string

	// DynamicElement describes a typed action or display element and its raw attributes.
	DynamicElement struct {
		Type       string            `json:"type" yaml:"type"`
		LongPress  bool              `json:"long_press" yaml:"long_press"`
		Attributes DynamicAttributes `json:"attributes" yaml:"attributes"`
	}

	// File is the top-level StreamDeck configuration.
	File struct {
		AutoReload        bool            `json:"auto_reload" yaml:"auto_reload"`
		CaptionBorder     int             `json:"caption_border" yaml:"caption_border"`
		CaptionColor      [4]int          `json:"caption_color" yaml:"caption_color"`
		CaptionFont       string          `json:"caption_font" yaml:"caption_font"`
		CaptionFontSize   float64         `json:"caption_font_size" yaml:"caption_font_size"`
		CaptionPosition   CaptionPosition `json:"caption_position" yaml:"caption_position"`
		DefaultBrightness int             `json:"default_brightness" yaml:"default_brightness"`
		DefaultPage       string          `json:"default_page" yaml:"default_page"`
		DisplayOffTime    time.Duration   `json:"display_off_time" yaml:"display_off_time"`
		LongPressDuration time.Duration   `json:"long_press_duration" yaml:"long_press_duration"`
		Pages             map[string]Page `json:"pages" yaml:"pages"`
		RenderFont        string          `json:"render_font" yaml:"render_font"`
	}

	// KeyDefinition defines display and actions for one key.
	KeyDefinition struct {
		Display DynamicElement   `json:"display" yaml:"display"`
		Actions []DynamicElement `json:"actions" yaml:"actions"`
	}

	// Page contains key definitions and optional overlay or underlay references.
	Page struct {
		Keys     map[int]KeyDefinition `json:"keys" yaml:"keys"`
		Overlay  string                `json:"overlay" yaml:"overlay"`
		Underlay string                `json:"underlay" yaml:"underlay"`
	}
)

// Load reads, validates, and expands a configuration file.
func Load(confFile string, deck *streamdeck.Client) (f File, err error) {
	userConfFile, err := os.Open(confFile) //#nosec:G304 // intended to read specified config file
	if err != nil {
		return f, fmt.Errorf("opening config: %w", err)
	}
	defer func() {
		if err := userConfFile.Close(); err != nil {
			logrus.WithError(err).Error("closing config file (leaked fd)")
		}
	}()

	var rawConf yaml.Node

	decoder := yaml.NewDecoder(userConfFile)
	if err = decoder.Decode(&rawConf); err != nil {
		return f, fmt.Errorf("parsing config: %w", err)
	}

	if err = expandEnvVariables(&rawConf); err != nil {
		return f, fmt.Errorf("expanding env variables: %w", err)
	}

	buf := new(bytes.Buffer)
	if err = yaml.NewEncoder(buf).Encode(&rawConf); err != nil {
		return f, fmt.Errorf("encoding expanded config: %w", err)
	}

	decoder = yaml.NewDecoder(buf)
	f = New()

	decoder.KnownFields(true)
	if err = decoder.Decode(&f); err != nil {
		return f, fmt.Errorf("parsing config: %w", err)
	}

	applySystemPages(deck, &f)

	return f, nil
}

// New returns a configuration populated with defaults.
func New() File {
	return File{
		AutoReload:        true,
		LongPressDuration: defaultLongPressDuration,
	}
}
