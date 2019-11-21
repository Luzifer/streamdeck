package main

type config struct {
	DefaultBrightness int             `yaml:"default_brightness"`
	DefaultPage       string          `yaml:"default_page"`
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
