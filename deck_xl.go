package streamdeck

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	"github.com/sstallion/go-hid"
)

const (
	deckXLMaxPacketSize = 1024
	deckXLHeaderSize    = 8
)

type deckConfigXL struct {
	dev       *hid.Device
	writeLock sync.Mutex

	keyState []EventType
}

func newDeckConfigXL() *deckConfigXL {
	d := &deckConfigXL{}
	d.keyState = make([]EventType, d.NumKeys())

	return d
}

func (d *deckConfigXL) SetDevice(dev *hid.Device) { d.dev = dev }

func (d *deckConfigXL) NumKeys() int                     { return 32 }
func (d *deckConfigXL) KeyColumns() int                  { return 8 }
func (d *deckConfigXL) KeyRows() int                     { return 4 }
func (d *deckConfigXL) KeyDirection() keyDirection       { return keyDirectionLTR }
func (d *deckConfigXL) KeyDataOffset() int               { return 4 }
func (d *deckConfigXL) TransformKeyIndex(keyIdx int) int { return keyIdx }

func (d *deckConfigXL) IconSize() int  { return 96 }
func (d *deckConfigXL) IconBytes() int { return d.IconSize() * d.IconSize() * 3 }

func (d *deckConfigXL) Model() uint16 { return StreamDeckXL }

func (d *deckConfigXL) FillColor(keyIdx int, col color.RGBA) error {
	img := image.NewRGBA(image.Rect(0, 0, d.IconSize(), d.IconSize()))

	for x := 0; x < d.IconSize(); x++ {
		for y := 0; y < d.IconSize(); y++ {
			img.Set(x, y, col)
		}
	}

	return d.FillImage(keyIdx, img)
}

func (d *deckConfigXL) FillImage(keyIdx int, img image.Image) error {
	d.writeLock.Lock()
	defer d.writeLock.Unlock()

	buf := new(bytes.Buffer)

	// We need to rotate the image or it will be presented upside down
	rimg := imaging.Rotate180(img)

	if err := jpeg.Encode(buf, rimg, &jpeg.Options{Quality: 95}); err != nil {
		return errors.Wrap(err, "Unable to encode jpeg")
	}

	var partIndex int16
	for buf.Len() > 0 {
		chunk := make([]byte, deckXLMaxPacketSize-deckXLHeaderSize)
		n, err := buf.Read(chunk)
		if err != nil {
			return errors.Wrap(err, "Unable to read image chunk")
		}

		var last uint8
		if n < deckXLMaxPacketSize-deckXLHeaderSize {
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

func (d *deckConfigXL) FillPanel(img image.RGBA) error {
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

func (d *deckConfigXL) ClearKey(keyIdx int) error {
	return d.FillColor(keyIdx, color.RGBA{0x0, 0x0, 0x0, 0xff})
}

func (d *deckConfigXL) ClearAllKeys() error {
	for i := 0; i < d.NumKeys(); i++ {
		if err := d.ClearKey(i); err != nil {
			return errors.Wrap(err, "Unable to clear key")
		}
	}
	return nil
}

func (d *deckConfigXL) SetBrightness(pct int) error {
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

func (d *deckConfigXL) ResetToLogo() error {
	_, err := d.dev.SendFeatureReport([]byte{
		0x03,
		0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})

	return errors.Wrap(err, "Unable to send feature report")
}

func (d *deckConfigXL) GetFimwareVersion() (string, error) {
	fw := make([]byte, 32)
	fw[0] = 5

	_, err := d.dev.GetFeatureReport(fw)
	if err != nil {
		return "", errors.Wrap(err, "Unable to get feature report")
	}

	return strings.TrimRight(string(fw[6:]), "\x00"), nil
}
