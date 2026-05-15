package streamdeck

//revive:disable:add-constant // many numbers with single use or only protocol value

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/sstallion/go-hid"
)

const (
	deckOriginalV2MaxPacketSize = 1024
	deckOriginalV2HeaderSize    = 8
)

type deckConfigOriginalV2 struct {
	dev       *hid.Device
	writeLock sync.Mutex

	keyState []EventType
}

func newDeckConfigOriginalV2() deckConfig {
	d := &deckConfigOriginalV2{}
	d.keyState = make([]EventType, d.NumKeys())

	return d
}

func (d *deckConfigOriginalV2) ClearAllKeys() error {
	for i := 0; i < d.NumKeys(); i++ {
		if err := d.ClearKey(i); err != nil {
			return fmt.Errorf("clearing key: %w", err)
		}
	}
	return nil
}

func (d *deckConfigOriginalV2) ClearKey(keyIdx int) error {
	return d.FillColor(keyIdx, color.RGBA{0x0, 0x0, 0x0, 0xff})
}

func (d *deckConfigOriginalV2) FillColor(keyIdx int, col color.RGBA) error {
	img := image.NewRGBA(image.Rect(0, 0, d.IconSize(), d.IconSize()))

	for x := 0; x < d.IconSize(); x++ {
		for y := 0; y < d.IconSize(); y++ {
			img.Set(x, y, col)
		}
	}

	return d.FillImage(keyIdx, img)
}

func (d *deckConfigOriginalV2) FillImage(keyIdx int, img image.Image) error {
	if keyIdx >= d.NumKeys() || keyIdx < 0 {
		return fmt.Errorf("key index %d out of bounds", keyIdx)
	}

	d.writeLock.Lock()
	defer d.writeLock.Unlock()

	buf := new(bytes.Buffer)

	// We need to rotate the image or it will be presented upside down
	rimg := imaging.Rotate180(img)

	if err := jpeg.Encode(buf, rimg, &jpeg.Options{Quality: 95}); err != nil {
		return fmt.Errorf("encoding jpeg: %w", err)
	}

	var partIndex int16
	for buf.Len() > 0 {
		chunk := make([]byte, deckOriginalV2MaxPacketSize-deckOriginalV2HeaderSize)
		n, err := buf.Read(chunk)
		if err != nil {
			return fmt.Errorf("reading image chunk: %w", err)
		}

		var last uint8
		if n < deckOriginalV2MaxPacketSize-deckOriginalV2HeaderSize || buf.Len() == 0 {
			last = 1
		}

		tbuf := new(bytes.Buffer)
		tbuf.Write([]byte{0x02, 0x07, byte(keyIdx), last})    //#nosec:G115 // keyIdx is guarded to safe values
		_ = binary.Write(tbuf, binary.LittleEndian, int16(n)) //#nosec:G115 // guarded to safe values
		_ = binary.Write(tbuf, binary.LittleEndian, partIndex)
		tbuf.Write(chunk)

		if _, err = d.dev.Write(tbuf.Bytes()); err != nil {
			return fmt.Errorf("sending image chunk: %w", err)
		}

		partIndex++
	}

	return nil
}

func (d *deckConfigOriginalV2) FillPanel(img image.RGBA) error {
	if img.Bounds().Size().X < d.KeyColumns()*d.IconSize() || img.Bounds().Size().Y < d.KeyRows()*d.IconSize() {
		return fmt.Errorf("image is too small")
	}

	for k := 0; k < d.NumKeys(); k++ {
		var (
			ky = k / d.KeyColumns()
			kx = k % d.KeyColumns()
		)

		if err := d.FillImage(k, img.SubImage(image.Rect(kx*d.IconSize(), ky*d.IconSize(), (kx+1)*d.IconSize(), (ky+1)*d.IconSize()))); err != nil {
			return fmt.Errorf("setting key image: %w", err)
		}
	}

	return nil
}

func (d *deckConfigOriginalV2) GetFimwareVersion() (string, error) {
	fw := make([]byte, 32)
	fw[0] = 5

	_, err := d.dev.GetFeatureReport(fw)
	if err != nil {
		return "", fmt.Errorf("getting feature report: %w", err)
	}

	return strings.TrimRight(string(fw[6:]), "\x00"), nil
}

func (d *deckConfigOriginalV2) IconBytes() int { return d.IconSize() * d.IconSize() * 3 }

func (*deckConfigOriginalV2) IconSize() int { return 72 }

func (*deckConfigOriginalV2) KeyColumns() int { return 5 }

func (*deckConfigOriginalV2) KeyDataOffset() int { return 4 }

func (*deckConfigOriginalV2) KeyDirection() keyDirection { return keyDirectionLTR }

func (*deckConfigOriginalV2) KeyRows() int { return 3 }

func (*deckConfigOriginalV2) Model() uint16 { return StreamDeckOriginalV2 }

func (*deckConfigOriginalV2) NumKeys() int { return 15 }

func (d *deckConfigOriginalV2) ResetToLogo() error {
	if _, err := d.dev.SendFeatureReport([]byte{
		0x03,
		0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}); err != nil {
		return fmt.Errorf("sending feature report: %w", err)
	}

	return nil
}

func (d *deckConfigOriginalV2) SetBrightness(pct int) error {
	if pct < 0 || pct > 100 {
		return fmt.Errorf("percentage %d out of bounds 0-100", pct)
	}

	if _, err := d.dev.SendFeatureReport([]byte{
		0x03, 0x08, byte(pct), 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}); err != nil {
		return fmt.Errorf("sending feature report: %w", err)
	}

	return nil
}

func (d *deckConfigOriginalV2) SetDevice(dev *hid.Device) { d.dev = dev }

func (*deckConfigOriginalV2) TransformKeyIndex(keyIdx int) int { return keyIdx }
