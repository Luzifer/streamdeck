package main

import "time"

type config struct {
	AutoReload        bool            `yaml:"auto_reload"`
	DefaultBrightness int             `yaml:"default_brightness"`
	DefaultPage       string          `yaml:"default_page"`
	DisplayOffTime    time.Duration   `yaml:"display_off_time"`
	Pages             map[string]page `yaml:"pages"`
	RenderFont        string          `yaml:"render_font"`
}

type page struct {
	Keys []keyDefinition `yaml:"keys"`
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
