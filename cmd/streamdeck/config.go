package main

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const defaultLongPressDuration = 500 * time.Millisecond

type config struct {
	AutoReload        bool            `json:"auto_reload" yaml:"auto_reload"`
	CaptionBorder     int             `json:"caption_border" yaml:"caption_border"`
	CaptionColor      [4]int          `json:"caption_color" yaml:"caption_color"`
	CaptionFont       string          `json:"caption_font" yaml:"caption_font"`
	CaptionFontSize   float64         `json:"caption_font_size" yaml:"caption_font_size"`
	CaptionPosition   captionPosition `json:"caption_position" yaml:"caption_position"`
	DefaultBrightness int             `json:"default_brightness" yaml:"default_brightness"`
	DefaultPage       string          `json:"default_page" yaml:"default_page"`
	DisplayOffTime    time.Duration   `json:"display_off_time" yaml:"display_off_time"`
	LongPressDuration time.Duration   `json:"long_press_duration" yaml:"long_press_duration"`
	Pages             map[string]page `json:"pages" yaml:"pages"`
	RenderFont        string          `json:"render_font" yaml:"render_font"`
}

type page struct {
	Keys     map[int]keyDefinition `json:"keys" yaml:"keys"`
	Overlay  string                `json:"overlay" yaml:"overlay"`
	Underlay string                `json:"underlay" yaml:"underlay"`
}

type keyDefinition struct {
	Display dynamicElement   `json:"display" yaml:"display"`
	Actions []dynamicElement `json:"actions" yaml:"actions"`
}

type dynamicElement struct {
	Type       string                 `json:"type" yaml:"type"`
	LongPress  bool                   `json:"long_press" yaml:"long_press"`
	Attributes map[string]interface{} `json:"attributes" yaml:"attributes"`
}

func newConfig() config {
	return config{
		AutoReload:        true,
		LongPressDuration: defaultLongPressDuration,
	}
}

type captionPosition string

const (
	captionPositionBottom = "bottom"
	captionPositionTop    = "top"
)

func loadConfig() error {
	userConfFile, err := os.Open(cfg.Config)
	if err != nil {
		return errors.Wrap(err, "Unable to open config")
	}
	defer userConfFile.Close()

	var (
		decoder  = yaml.NewDecoder(userConfFile)
		tempConf = newConfig()
	)

	decoder.SetStrict(true)
	if err = decoder.Decode(&tempConf); err != nil {
		return errors.Wrap(err, "Unable to parse config")
	}

	applySystemPages(&tempConf)

	userConfig = tempConf

	return nil
}
