package streamdeck

import (
	"image"
	"image/color"

	"github.com/sstallion/go-hid"
)

type keyDirection uint

const (
	keyDirectionLTR keyDirection = iota
	keyDirectionRTL
)

type deckConfig interface {
	SetDevice(dev *hid.Device)

	NumKeys() int
	KeyColumns() int
	KeyRows() int
	KeyDirection() keyDirection
	KeyDataOffset() int

	TransformKeyIndex(keyIdx int) int

	IconSize() int
	IconBytes() int

	Model() uint16

	FillColor(keyIdx int, col color.RGBA) error
	FillImage(keyIdx int, img image.Image) error
	FillPanel(img image.RGBA) error

	ClearKey(keyIdx int) error
	ClearAllKeys() error

	SetBrightness(pct int) error

	ResetToLogo() error

	GetFimwareVersion() (string, error)
}

var decks = map[uint16]deckConfig{
	StreamDeckOriginalV2: &deckConfigOriginalV2{},
}
