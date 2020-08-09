package main

import (
	"time"
)

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
	Pages             map[string]page `yaml:"pages"`
	RenderFont        string          `yaml:"render_font"`
}

type page struct {
	Keys map[int]keyDefinition `yaml:"keys"`
}

type keyDefinition struct {
	Display dynamicElement   `yaml:"display"`
	Actions []dynamicElement `yaml:"actions"`
	On      string           `yaml:"on"`
}

type dynamicElement struct {
	Type       string                 `yaml:"type"`
	Attributes map[string]interface{} `yaml:"attributes"`
}

func newConfig() config {
	return config{
		AutoReload: true,
	}
}

type captionPosition string

const (
	captionPositionBottom = "bottom"
	captionPositionTop    = "top"
)
