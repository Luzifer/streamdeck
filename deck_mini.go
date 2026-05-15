package streamdeck

//revive:disable:add-constant // many numbers with single use or only protocol value

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/sstallion/go-hid"
	"golang.org/x/image/bmp"
)

const (
	deckMiniMaxPacketSize = 1024
	deckMiniHeaderSize    = 16
)

type deckConfigMini struct {
	dev       *hid.Device
	writeLock sync.Mutex

	keyState []EventType
}

func newDeckConfigMini() deckConfig {
	d := &deckConfigMini{}
	d.keyState = make([]EventType, d.NumKeys())

	return d
}

func (d *deckConfigMini) ClearAllKeys() error {
	for i := 0; i < d.NumKeys(); i++ {
		if err := d.ClearKey(i); err != nil {
			return fmt.Errorf("clearing key: %w", err)
		}
	}

	return nil
}

func (d *deckConfigMini) ClearKey(keyIdx int) error {
	return d.FillColor(keyIdx, color.RGBA{0x0, 0x0, 0x0, 0xff})
}

func (d *deckConfigMini) FillColor(keyIdx int, col color.RGBA) error {
	img := image.NewRGBA(image.Rect(0, 0, d.IconSize(), d.IconSize()))

	for x := 0; x < d.IconSize(); x++ {
		for y := 0; y < d.IconSize(); y++ {
			img.Set(x, y, col)
		}
	}

	return d.FillImage(keyIdx, img)
}

func (d *deckConfigMini) FillImage(keyIdx int, img image.Image) error {
	if keyIdx >= d.NumKeys() || keyIdx < 0 {
		return fmt.Errorf("key index %d out of bounds", keyIdx)
	}

	d.writeLock.Lock()
	defer d.writeLock.Unlock()

	buf := new(bytes.Buffer)
	rimg := imaging.Transpose(img)
	if err := bmp.Encode(buf, rimg); err != nil {
		return fmt.Errorf("encoding bmp: %w", err)
	}

	var partIndex int16
	for buf.Len() > 0 {
		chunk := make([]byte, deckMiniMaxPacketSize-deckMiniHeaderSize)
		n, err := buf.Read(chunk)
		if err != nil {
			return fmt.Errorf("reading image chunk: %w", err)
		}

		var last uint8
		if n < deckMiniMaxPacketSize-deckMiniHeaderSize || buf.Len() == 0 {
			last = 1
		}

		header := make([]byte, deckMiniHeaderSize)
		header[0] = 0x02
		header[1] = 0x01
		header[2] = byte(partIndex)
		header[4] = last
		header[5] = byte(keyIdx + 1) //#nosec:G115 // keyIdx is guarded to safe values

		if _, err = d.dev.Write(append(header, chunk...)); err != nil {
			return fmt.Errorf("sending image chunk: %w", err)
		}

		partIndex++
	}

	return nil
}

func (d *deckConfigMini) FillPanel(img image.RGBA) error {
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

func (d *deckConfigMini) GetFimwareVersion() (string, error) {
	fw := make([]byte, 32)
	fw[0] = 4

	if _, err := d.dev.GetFeatureReport(fw); err != nil {
		return "", fmt.Errorf("getting feature report: %w", err)
	}

	return string(fw[5:13]), nil
}

func (d *deckConfigMini) IconBytes() int { return d.IconSize() * d.IconSize() * 3 }

func (*deckConfigMini) IconSize() int { return 80 }

func (*deckConfigMini) KeyColumns() int { return 3 }

func (*deckConfigMini) KeyDataOffset() int { return 1 }

func (*deckConfigMini) KeyDirection() keyDirection { return keyDirectionLTR }

func (*deckConfigMini) KeyRows() int { return 2 }

func (*deckConfigMini) Model() uint16 { return StreamDeckMini }

func (*deckConfigMini) NumKeys() int { return 6 }

func (d *deckConfigMini) ResetToLogo() error {
	r := make([]byte, 17)
	r[0] = 0x0b
	r[1] = 0x63

	if _, err := d.dev.SendFeatureReport(r); err != nil {
		return fmt.Errorf("sending feature report: %w", err)
	}

	return nil
}

func (d *deckConfigMini) SetBrightness(pct int) error {
	if pct < 0 || pct > 100 {
		return fmt.Errorf("percentage %d out of bounds 0-100", pct)
	}

	r := make([]byte, 17)
	r[0] = 0x05
	r[1] = 0x55
	r[2] = 0xaa
	r[3] = 0xd1
	r[4] = 0x01
	r[5] = byte(pct)

	if _, err := d.dev.SendFeatureReport(r); err != nil {
		return fmt.Errorf("sending feature report: %w", err)
	}

	return nil
}

func (d *deckConfigMini) SetDevice(dev *hid.Device) { d.dev = dev }

func (*deckConfigMini) TransformKeyIndex(keyIdx int) int { return keyIdx }
