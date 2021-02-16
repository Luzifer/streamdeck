package main

import (
	"bytes"
	"encoding/gob"
	"image/color"
	"os"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const defaultLongPressDuration = 500 * time.Millisecond

func init() {
	gob.Register(attributeCollection{})
}

type attributeCollection struct {
	AttachStderr bool              `json:"attach_stderr,omitempty" yaml:"attach_stderr,omitempty"`
	AttachStdout bool              `json:"attach_stdout,omitempty" yaml:"attach_stdout,omitempty"`
	Border       *int              `json:"border,omitempty" yaml:"border,omitempty"`
	Caption      string            `json:"caption,omitempty" yaml:"caption,omitempty"`
	ChangeVolume *float64          `json:"change_volume,omitempty" yaml:"change_volume,omitempty"`
	Color        string            `json:"color,omitempty" yaml:"color,omitempty"`
	Command      []string          `json:"command,omitempty" yaml:"command,omitempty"`
	Delay        time.Duration     `json:"delay,omitempty" yaml:"delay,omitempty"`
	Device       string            `json:"device,omitempty" yaml:"device,omitempty"`
	Env          map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	FontSize     *float64          `json:"font_size,omitempty" yaml:"font_size,omitempty"`
	Image        string            `json:"image,omitempty" yaml:"image,omitempty"`
	Interval     time.Duration     `json:"interval,omitempty" yaml:"interval,omitempty"`
	KeyCodes     []int             `json:"key_codes,omitempty" yaml:"key_codes,omitempty"`
	Keys         []string          `json:"keys,omitempty" yaml:"keys,omitempty"`
	Match        string            `json:"match,omitempty" yaml:"match,omitempty"`
	ModAlt       bool              `json:"mod_alt,omitempty" yaml:"mod_alt,omitempty"`
	ModCtrl      bool              `json:"mod_ctrl,omitempty" yaml:"mod_ctrl,omitempty"`
	ModShift     bool              `json:"mod_shift,omitempty" yaml:"mod_shift,omitempty"`
	Mute         string            `json:"mute,omitempty" yaml:"mute,omitempty"`
	Name         string            `json:"name,omitempty" yaml:"name,omitempty"`
	Path         string            `json:"path,omitempty" yaml:"path,omitempty"`
	Relative     int               `json:"relative,omitempty" yaml:"relative,omitempty"`
	RGBA         []int             `json:"rgba,omitempty" yaml:"rgba,omitempty"`
	SetVolume    *float64          `json:"set_volume,omitempty" yaml:"set_volume,omitempty"`
	Text         string            `json:"text,omitempty" yaml:"text,omitempty"`
	URL          string            `json:"url,omitempty" yaml:"url,omitempty"`
	Wait         bool              `json:"wait,omitempty" yaml:"wait,omitempty"`
}

func (a attributeCollection) Clone() attributeCollection {
	var (
		buf = new(bytes.Buffer)
		out attributeCollection
	)

	gob.NewEncoder(buf).Encode(a)
	gob.NewDecoder(buf).Decode(&out)

	return out
}

func (a attributeCollection) RGBAToColor() color.RGBA {
	if len(a.RGBA) != 4 {
		return color.RGBA{}
	}
	return color.RGBA{uint8(a.RGBA[0]), uint8(a.RGBA[1]), uint8(a.RGBA[2]), uint8(a.RGBA[3])}
}

type config struct {
	APIListen         string          `json:"api_listen" yaml:"api_listen"`
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
	Type       string              `json:"type" yaml:"type"`
	LongPress  bool                `json:"long_press" yaml:"long_press"`
	Attributes attributeCollection `json:"attributes" yaml:"attributes"`
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
