package streamdeck

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"strings"

	"github.com/pkg/errors"
	"github.com/sstallion/go-hid"
)

const (
	deckOriginalV2MaxPacketSize = 1024
	deckOriginalV2HeaderSize    = 8
)

type deckConfigOriginalV2 struct {
	dev *hid.Device

	keyState []EventType
}

func newDeckConfigOriginalV2() *deckConfigOriginalV2 {
	d := &deckConfigOriginalV2{}
	d.keyState = make([]EventType, d.NumKeys())

	return d
}

func (d *deckConfigOriginalV2) SetDevice(dev *hid.Device) { d.dev = dev }

func (d deckConfigOriginalV2) NumKeys() int                     { return 15 }
func (d deckConfigOriginalV2) KeyColumns() int                  { return 5 }
func (d deckConfigOriginalV2) KeyRows() int                     { return 3 }
func (d deckConfigOriginalV2) KeyDirection() keyDirection       { return keyDirectionLTR }
func (d deckConfigOriginalV2) KeyDataOffset() int               { return 4 }
func (d deckConfigOriginalV2) TransformKeyIndex(keyIdx int) int { return keyIdx }

func (d deckConfigOriginalV2) IconSize() int  { return 72 }
func (d deckConfigOriginalV2) IconBytes() int { return d.IconSize() * d.IconSize() * 3 }

func (d deckConfigOriginalV2) Model() uint16 { return StreamDeckOriginalV2 }

func (d deckConfigOriginalV2) FillColor(keyIdx int, col color.RGBA) error {
	img := image.NewRGBA(image.Rect(0, 0, d.IconSize(), d.IconSize()))

	for x := 0; x < d.IconSize(); x++ {
		for y := 0; y < d.IconSize(); y++ {
			img.Set(x, y, col)
		}
	}

	return d.FillImage(keyIdx, img)
}

func (d deckConfigOriginalV2) FillImage(keyIdx int, img image.Image) error {
	buf := new(bytes.Buffer)
	if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 95}); err != nil {
		return errors.Wrap(err, "Unable to encode jpeg")
	}

	var partIndex int16
	for buf.Len() > 0 {
		chunk := make([]byte, deckOriginalV2MaxPacketSize-deckOriginalV2HeaderSize)
		n, err := buf.Read(chunk)
		if err != nil {
			return errors.Wrap(err, "Unable to read image chunk")
		}

		var last uint8
		if n < deckOriginalV2MaxPacketSize-deckOriginalV2HeaderSize {
			last = 1
		}

		tbuf := new(bytes.Buffer)
		tbuf.Write([]byte{0x02, 0x07, byte(keyIdx), last})
		binary.Write(tbuf, binary.LittleEndian, int16(n))
		binary.Write(tbuf, binary.LittleEndian, partIndex)
		tbuf.Write(chunk)

		if _, err = d.dev.Write(tbuf.Bytes()); err != nil {
			return errors.Wrap(err, "Unable to send image chunk")
		}

		partIndex++
	}

	return nil
}

func (d deckConfigOriginalV2) FillPanel(img image.RGBA) error {
	if img.Bounds().Size().X < d.KeyColumns()*d.IconSize() || img.Bounds().Size().Y < d.KeyRows()*d.IconSize() {
		return errors.New("Image is too small")
	}

	for k := 0; k < d.NumKeys(); k++ {
		var (
			ky = k / d.KeyColumns()
			kx = k % d.KeyColumns()
		)

		if err := d.FillImage(k, img.SubImage(image.Rect(kx*d.IconSize(), ky*d.IconSize(), (kx+1)*d.IconSize(), (ky+1)*d.IconSize()))); err != nil {
			return errors.Wrap(err, "Unable to set key image")
		}
	}

	return nil
}

func (d deckConfigOriginalV2) ClearKey(keyIdx int) error {
	return d.FillColor(keyIdx, color.RGBA{0x0, 0x0, 0x0, 0xff})
}

func (d deckConfigOriginalV2) ClearAllKeys() error {
	for i := 0; i < d.NumKeys(); i++ {
		if err := d.ClearKey(i); err != nil {
			return errors.Wrap(err, "Unable to clear key")
		}
	}
	return nil
}

func (d deckConfigOriginalV2) SetBrightness(pct int) error {
	if pct < 0 || pct > 100 {
		return errors.New("Percentage out of bounds")
	}

	_, err := d.dev.SendFeatureReport([]byte{
		0x03, 0x08, byte(pct), 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})

	return errors.Wrap(err, "Unable to send feature report")
}

func (d deckConfigOriginalV2) ResetToLogo() error {
	_, err := d.dev.SendFeatureReport([]byte{
		0x03,
		0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})

	return errors.Wrap(err, "Unable to send feature report")
}

func (d deckConfigOriginalV2) GetFimwareVersion() (string, error) {
	fw := make([]byte, 32)
	fw[0] = 5

	_, err := d.dev.GetFeatureReport(fw)
	if err != nil {
		return "", errors.Wrap(err, "Unable to get feature report")
	}

	return strings.TrimRight(string(fw[6:]), "\x00"), nil
}
