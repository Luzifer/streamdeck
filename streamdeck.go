package streamdeck

import (
	"image"
	"image/color"

	"github.com/pkg/errors"
	hid "github.com/sstallion/go-hid"
)

const VendorElgato = 0x0fd9

const (
	// Streamdeck Original V2 (0fd9:006d) 15 keys
	StreamDeckOriginalV2 uint16 = 0x006d
	// Stremdeck XL (0fd9:006c) 32 keys
	StreamDeckXL uint16 = 0x006c
	// StreamDeck Mini (0fd9:0063) 6 keys
	StreamDeckMini uint16 = 0x0063
)

var DeckToName = map[uint16]string{
	StreamDeckOriginalV2: "StreamDeck Original V2",
	StreamDeckXL:         "StreamDeck XL",
	StreamDeckMini:       "StreamDeck Mini",
}

// EventType represents the state of a button (Up / Down)
type EventType uint8

const (
	EventTypeUp EventType = iota
	EventTypeDown
)

// Event represents a state change on a button
type Event struct {
	Key  int
	Type EventType
}

// Client manages the connection to the StreamDeck
type Client struct {
	cfg       deckConfig
	dev       *hid.Device
	devType   uint16
	keyStates []EventType

	evts chan Event
}

// New creates a new Client for the given device (see constants for supported types)
func New(devicePID uint16) (*Client, error) {
	dev, err := hid.OpenFirst(VendorElgato, devicePID)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open device")
	}

	cfg := decks[devicePID]
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

// Close closes the underlying HID connection
func (c Client) Close() error { return c.dev.Close() }

// Serial returns the device serial
func (c Client) Serial() (string, error) { return c.dev.GetSerialNbr() }

// FillColor fills a key with a solid color
func (c Client) FillColor(keyIdx int, col color.RGBA) error { return c.cfg.FillColor(keyIdx, col) }

// FillImage fills a key with an image
func (c Client) FillImage(keyIdx int, img image.Image) error { return c.cfg.FillImage(keyIdx, img) }

// FillPanel slices a big image and fills the keys with the parts
func (c Client) FillPanel(img image.RGBA) error { return c.cfg.FillPanel(img) }

// ClearKey fills a key with solid black
func (c Client) ClearKey(keyIdx int) error { return c.cfg.ClearKey(keyIdx) }

// ClearAllKeys fills all keys with solid black
func (c Client) ClearAllKeys() error { return c.cfg.ClearAllKeys() }

// IconSize returns the required icon size for the StreamDeck
func (c Client) IconSize() int { return c.cfg.IconSize() }

// NumKeys returns the number of keys available on the StreamDeck
func (c Client) NumKeys() int { return c.cfg.NumKeys() }

// SetBrightness sets the brightness of the keys (0-100)
func (c Client) SetBrightness(pct int) error { return c.cfg.SetBrightness(pct) }

// ResetToLogo restores the original Elgato StreamDeck logo
func (c Client) ResetToLogo() error { return c.cfg.ResetToLogo() }

// GetFimwareVersion retrieves the firmware version
func (c Client) GetFimwareVersion() (string, error) { return c.cfg.GetFimwareVersion() }

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
