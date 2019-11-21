package main

type config struct {
	DefaultBrightness int             `yaml:"default_brightness"`
	DefaultPage       string          `yaml:"default_page"`
	Pages             map[string]page `yaml:"pages"`
}

type page struct {
	Keys []keyDefinition `yaml:"keys"`
}

type keyDefinition struct {
	Display dynamicElement `yaml:"display"`
	Action  dynamicElement `yaml:"action"`
	On      string         `yaml:"on"`
}

type dynamicElement struct {
	Type       string                 `yaml:"type"`
	Attributes map[string]interface{} `yaml:"attributes"`
}
