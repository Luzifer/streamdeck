package streamdeck

import (
	"bytes"
	"image"
	"image/color"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
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

func newDeckConfigMini() *deckConfigMini {
	d := &deckConfigMini{}
	d.keyState = make([]EventType, d.NumKeys())

	return d
}

func (d *deckConfigMini) SetDevice(dev *hid.Device)        { d.dev = dev }
func (d *deckConfigMini) NumKeys() int                     { return 6 }
func (d *deckConfigMini) KeyColumns() int                  { return 3 }
func (d *deckConfigMini) KeyRows() int                     { return 2 }
func (d *deckConfigMini) KeyDirection() keyDirection       { return keyDirectionLTR }
func (d *deckConfigMini) KeyDataOffset() int               { return 1 }
func (d *deckConfigMini) TransformKeyIndex(keyIdx int) int { return keyIdx }
func (d *deckConfigMini) IconSize() int                    { return 80 }
func (d *deckConfigMini) IconBytes() int                   { return d.IconSize() * d.IconSize() * 3 }
func (d *deckConfigMini) Model() uint16                    { return StreamDeckMini }

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
	d.writeLock.Lock()
	defer d.writeLock.Unlock()

	buf := new(bytes.Buffer)
	rimg := imaging.Transpose(img)
	if err := bmp.Encode(buf, rimg); err != nil {
		return errors.Wrap(err, "Unable to encode bmp")
	}

	var partIndex int16
	for buf.Len() > 0 {
		chunk := make([]byte, deckMiniMaxPacketSize-deckMiniHeaderSize)
		n, err := buf.Read(chunk)
		if err != nil {
			return errors.Wrap(err, "Unable to read image chunk")
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
		header[5] = byte(keyIdx + 1)

		if _, err = d.dev.Write(append(header, chunk...)); err != nil {
			return errors.Wrap(err, "Unable to send image chunk")
		}

		partIndex++
	}

	return nil
}

func (d *deckConfigMini) FillPanel(img image.RGBA) error {
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

func (d *deckConfigMini) ClearKey(keyIdx int) error {
	return d.FillColor(keyIdx, color.RGBA{0x0, 0x0, 0x0, 0xff})
}

func (d *deckConfigMini) ClearAllKeys() error {
	for i := 0; i < d.NumKeys(); i++ {
		if err := d.ClearKey(i); err != nil {
			return errors.Wrap(err, "Unable to clear key")
		}
	}
	return nil
}

func (d *deckConfigMini) SetBrightness(pct int) error {
	if pct < 0 || pct > 100 {
		return errors.New("Percentage out of bounds")
	}

	r := make([]byte, 17)
	r[0] = 0x05
	r[1] = 0x55
	r[2] = 0xaa
	r[3] = 0xd1
	r[4] = 0x01
	r[5] = byte(pct)

	_, err := d.dev.SendFeatureReport(r)

	return errors.Wrap(err, "Unable to send feature report")
}

func (d *deckConfigMini) ResetToLogo() error {
	r := make([]byte, 17)
	r[0] = 0x0b
	r[1] = 0x63

	_, err := d.dev.SendFeatureReport(r)

	return errors.Wrap(err, "Unable to send feature report")
}

func (d *deckConfigMini) GetFimwareVersion() (string, error) {
	fw := make([]byte, 32)
	fw[0] = 4

	_, err := d.dev.GetFeatureReport(fw)
	if err != nil {
		return "", errors.Wrap(err, "Unable to get feature report")
	}

	return string(fw[5:13]), nil
}
