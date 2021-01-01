package main

import (
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const defaultLongPressDuration = 500 * time.Millisecond

type config struct {
	AutoReload        bool            `yaml:"auto_reload"`
	CaptionBorder     int             `yaml:"caption_border"`
	CaptionColor      [4]int          `yaml:"caption_color"`
	CaptionFont       string          `yaml:"caption_font"`
	CaptionFontSize   float64         `yaml:"caption_font_size"`
	CaptionPosition   captionPosition `yaml:"caption_position"`
	DefaultBrightness int             `yaml:"default_brightness"`
	DefaultPage       string          `yaml:"default_page"`
	DisplayOffTime    time.Duration   `yaml:"display_off_time"`
	LongPressDuration time.Duration   `yaml:"long_press_duration"`
	Pages             map[string]page `yaml:"pages"`
	RenderFont        string          `yaml:"render_font"`
}

type page struct {
	Keys     map[int]keyDefinition `yaml:"keys"`
	Overlay  string                `yaml:"overlay"`
	Underlay string                `yaml:"underlay"`
}

type keyDefinition struct {
	Display dynamicElement   `yaml:"display"`
	Actions []dynamicElement `yaml:"actions"`
}

type dynamicElement struct {
	Type       string                 `yaml:"type"`
	LongPress  bool                   `yaml:"long_press"`
	Attributes map[string]interface{} `yaml:"attributes"`
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
