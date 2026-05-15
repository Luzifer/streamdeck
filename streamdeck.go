// Package streamdeck contains a library to communicate with StreamDeck
// devices through their USB-HID interface
package streamdeck

import (
	"fmt"
	"image"
	"image/color"

	hid "github.com/sstallion/go-hid"
)

// VendorElgato is the commonly used vendor ID by Elgato
const VendorElgato = 0x0fd9

// Collection of supported StreamDecks
const (
	// Streamdeck Original V2 (0fd9:006d) 15 keys
	StreamDeckOriginalV2 uint16 = 0x006d
	// Stremdeck XL (0fd9:006c) 32 keys
	StreamDeckXL uint16 = 0x006c
	// StreamDeck Mini (0fd9:0063) 6 keys
	StreamDeckMini uint16 = 0x0063
	// StreamDeck Mini V2 (0fd9:0090) 6 keys
	StreamDeckMiniV2 uint16 = 0x0090
)

// Collection of supported EventType from keys
const (
	EventTypeUp EventType = iota
	EventTypeDown
)

type (
	// Client manages the connection to the StreamDeck
	Client struct {
		cfg       deckConfig
		dev       *hid.Device
		devType   uint16
		keyStates []EventType

		evts chan Event
	}

	// EventType represents the state of a button (Up / Down)
	EventType uint8

	// Event represents a state change on a button
	Event struct {
		Key  int
		Type EventType
	}
)

// DeckToName contains a listing of device market-names for Streamdecks
var DeckToName = map[uint16]string{
	StreamDeckOriginalV2: "StreamDeck Original V2",
	StreamDeckXL:         "StreamDeck XL",
	StreamDeckMini:       "StreamDeck Mini",
	StreamDeckMiniV2:     "StreamDeck Mini V2",
}

var decks = map[uint16]deckConfigCreateFunc{
	StreamDeckOriginalV2: newDeckConfigOriginalV2,
	StreamDeckXL:         newDeckConfigXL,
	StreamDeckMini:       newDeckConfigMini,
	StreamDeckMiniV2:     newDeckConfigMini,
}

// New creates a new Client for the given device (see constants for supported types)
func New(devicePID uint16) (*Client, error) {
	dev, err := hid.OpenFirst(VendorElgato, devicePID)
	if err != nil {
		return nil, fmt.Errorf("opening device: %w", err)
	}

	cfg := decks[devicePID]()
	cfg.SetDevice(dev)

	client := &Client{
		cfg:       cfg,
		dev:       dev,
		devType:   devicePID,
		keyStates: make([]EventType, cfg.NumKeys()),

		evts: make(chan Event, 100),
	}

	go client.read()

	return client, nil
}

// ClearAllKeys fills all keys with solid black
func (c Client) ClearAllKeys() error { return c.cfg.ClearAllKeys() } //nolint:wrapcheck // wraps internal interface

// ClearKey fills a key with solid black
func (c Client) ClearKey(keyIdx int) error { return c.cfg.ClearKey(keyIdx) } //nolint:wrapcheck // wraps internal interface

// Close closes the underlying HID connection
func (c Client) Close() error { return c.dev.Close() } //nolint:wrapcheck // wraps internal interface

// FillColor fills a key with a solid color
func (c Client) FillColor(keyIdx int, col color.RGBA) error { return c.cfg.FillColor(keyIdx, col) } //nolint:wrapcheck // wraps internal interface

// FillImage fills a key with an image
func (c Client) FillImage(keyIdx int, img image.Image) error { return c.cfg.FillImage(keyIdx, img) } //nolint:wrapcheck // wraps internal interface

// FillPanel slices a big image and fills the keys with the parts
func (c Client) FillPanel(img image.RGBA) error { return c.cfg.FillPanel(img) } //nolint:wrapcheck // wraps internal interface

// GetFimwareVersion retrieves the firmware version
func (c Client) GetFimwareVersion() (string, error) { return c.cfg.GetFimwareVersion() } //nolint:wrapcheck // wraps internal interface

// IconSize returns the required icon size for the StreamDeck
func (c Client) IconSize() int { return c.cfg.IconSize() }

// NumKeys returns the number of keys available on the StreamDeck
func (c Client) NumKeys() int { return c.cfg.NumKeys() }

// ResetToLogo restores the original Elgato StreamDeck logo
func (c Client) ResetToLogo() error { return c.cfg.ResetToLogo() } //nolint:wrapcheck // wraps internal interface

// Serial returns the device serial
func (c Client) Serial() (string, error) { return c.dev.GetSerialNbr() } //nolint:wrapcheck // wraps internal interface

// SetBrightness sets the brightness of the keys (0-100)
func (c Client) SetBrightness(pct int) error { return c.cfg.SetBrightness(pct) } //nolint:wrapcheck // wraps internal interface

// Subscribe returns a channel to listen for incoming events
func (c Client) Subscribe() <-chan Event { return c.evts }

func (c Client) emit(key int, t EventType) { c.evts <- Event{Key: key, Type: t} }

func (c *Client) read() {
	for {
		buf := make([]byte, 1024)
		_, err := c.dev.Read(buf)
		if err != nil {
			continue
		}

		for k := 0; k < c.cfg.NumKeys(); k++ {
			newState := EventType(buf[k+c.cfg.KeyDataOffset()])
			if c.keyStates[k] != newState {
				c.emit(k, newState)
				c.keyStates[k] = newState
			}
		}
	}
}
