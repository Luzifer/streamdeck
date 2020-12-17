package main

import (
	"bytes"
	"encoding/gob"
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
	AttachStderr bool              `json:"attach_stderr" yaml:"attach_stderr"`
	AttachStdout bool              `json:"attach_stdout" yaml:"attach_stdout"`
	Border       *int              `json:"border" yaml:"border"`
	Caption      string            `json:"caption" yaml:"caption"`
	ChangeVolume *float64          `json:"change_volume" yaml:"change_volume"`
	Color        string            `json:"color" yaml:"color"`
	Command      []string          `json:"command" yaml:"command"`
	Delay        time.Duration     `json:"delay" yaml:"delay"`
	Device       string            `json:"device" yaml:"device"`
	Env          map[string]string `json:"env" yaml:"env"`
	FontSize     *float64          `json:"font_size" yaml:"font_size"`
	Image        string            `json:"image" yaml:"image"`
	Interval     time.Duration     `json:"interval" yaml:"interval"`
	KeyCodes     []int             `json:"key_codes" yaml:"key_codes"`
	Keys         []string          `json:"keys" yaml:"keys"`
	Match        string            `json:"match" yaml:"match"`
	ModAlt       bool              `json:"mod_alt" yaml:"mod_alt"`
	ModCtrl      bool              `json:"mod_ctrl" yaml:"mod_ctrl"`
	ModShift     bool              `json:"mod_shift" yaml:"mod_shift"`
	Mute         string            `json:"mute" yaml:"mute"`
	Name         string            `json:"name" yaml:"name"`
	Path         string            `json:"path" yaml:"path"`
	Relative     int               `json:"relative" yaml:"relative"`
	RGBA         []uint8           `json:"rgba" yaml:"rgba"`
	SetVolume    *float64          `json:"set_volume" yaml:"set_volume"`
	Text         string            `json:"text" yaml:"text"`
	URL          string            `json:"url" yaml:"url"`
	Wait         bool              `json:"wait" yaml:"wait"`
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
