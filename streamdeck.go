package streamdeck

import (
	"image"
	"image/color"

	"github.com/pkg/errors"
	hid "github.com/sstallion/go-hid"
)

const vendorElgato = 0x0fd9

const (
	StreamDeckOriginalV2 uint16 = 0x006d
)

type EventType uint8

const (
	EventTypeUp EventType = iota
	EventTypeDown
)

type Event struct {
	Key  int
	Type EventType
}

type Client struct {
	cfg       deckConfig
	dev       *hid.Device
	devType   uint16
	keyStates []EventType

	evts chan Event
}

func New(devicePID uint16) (*Client, error) {
	dev, err := hid.OpenFirst(vendorElgato, devicePID)
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

func (c Client) Close() error            { return c.dev.Close() }
func (c Client) Serial() (string, error) { return c.dev.GetSerialNbr() }

func (c Client) FillColor(keyIdx int, col color.RGBA) error  { return c.cfg.FillColor(keyIdx, col) }
func (c Client) FillImage(keyIdx int, img image.Image) error { return c.cfg.FillImage(keyIdx, img) }
func (c Client) FillPanel(img image.RGBA) error              { return c.cfg.FillPanel(img) }

func (c Client) ClearKey(keyIdx int) error { return c.cfg.ClearKey(keyIdx) }
func (c Client) ClearAllKeys() error       { return c.cfg.ClearAllKeys() }

func (c Client) SetBrightness(pct int) error { return c.cfg.SetBrightness(pct) }

func (c Client) ResetToLogo() error { return c.cfg.ResetToLogo() }

func (c Client) GetFimwareVersion() (string, error) { return c.cfg.GetFimwareVersion() }

func (c Client) Subscribe() <-chan Event   { return c.evts }
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
