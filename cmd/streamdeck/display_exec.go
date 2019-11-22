package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func init() {
	registerDisplayElement("exec", &displayElementExec{})
}

type displayElementExec struct {
	running bool
}

func (d displayElementExec) Display(idx int, attributes map[string]interface{}) error {
	var (
		err error
		img draw.Image = image.NewRGBA(image.Rect(0, 0, sd.IconSize(), sd.IconSize()))
	)

	// Initialize black image
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0x0, 0x0, 0x0, 0xff}), image.ZP, draw.Src)

	// Initialize command
	cmd, ok := attributes["command"].([]interface{})
	if !ok {
		return errors.New("No command supplied")
	}

	var args []string
	for _, c := range cmd {
		if v, ok := c.(string); ok {
			args = append(args, v)
			continue
		}
		return errors.New("Command conatins non-string argument")
	}

	// Execute command and parse it
	var buf = new(bytes.Buffer)

	command := exec.Command(args[0], args[1:]...)
	command.Env = os.Environ()
	command.Stdout = buf

	if err := command.Run(); err != nil {
		return errors.Wrap(err, "Command has exit != 0")
	}

	attributes["text"] = strings.TrimSpace(buf.String())

	tmpAttrs := map[string]interface{}{}
	if err = json.Unmarshal(buf.Bytes(), &tmpAttrs); err == nil {
		for k, v := range tmpAttrs {
			attributes[k] = v
		}
	}

	// Initialize background
	if filename, ok := attributes["image"].(string); ok {
		bgi, err := d.getImageFromDisk(filename)
		if err != nil {
			return errors.Wrap(err, "Unable to get image from disk")
		}

		draw.Draw(img, img.Bounds(), bgi, image.ZP, draw.Src)
	}

	// Initialize color
	var textColor color.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	if rgba, ok := attributes["color"].([]interface{}); ok {
		if len(rgba) != 4 {
			return errors.New("RGBA color definition needs 4 hex values")
		}

		tmpCol := color.RGBA{}

		for idx, vp := range []*uint8{&tmpCol.R, &tmpCol.G, &tmpCol.B, &tmpCol.A} {
			switch rgba[idx].(type) {
			case int:
				*vp = uint8(rgba[idx].(int))
			case float64:
				*vp = uint8(rgba[idx].(float64))
			}
		}

		textColor = tmpCol
	}

	// Initialize fontsize
	var fontsize float64 = 120
	if v, ok := attributes["font_size"].(float64); ok {
		fontsize = v
	}

	var border = 10
	if v, ok := attributes["border"].(int); ok {
		border = v
	}

	// Render text
	f, err := d.loadFont()
	if err != nil {
		return errors.Wrap(err, "Unable to load font")
	}

	c := freetype.NewContext()
	c.SetClip(img.Bounds())
	c.SetDPI(72)
	c.SetDst(img)
	c.SetFont(f)
	c.SetHinting(font.HintingNone)

	if strings.TrimSpace(attributes["text"].(string)) != "" {
		if err = d.drawText(c, strings.TrimSpace(attributes["text"].(string)), textColor, fontsize, border); err != nil {
			return errors.Wrap(err, "Unable to render text")
		}
	}

	return errors.Wrap(sd.FillImage(idx, img), "Unable to set image")
}

func (d displayElementExec) NeedsLoop(attributes map[string]interface{}) bool {
	if v, ok := attributes["interval"].(int); ok {
		return v > 0
	}

	return false
}

func (d *displayElementExec) StartLoopDisplay(idx int, attributes map[string]interface{}) error {
	d.running = true

	var interval = 5 * time.Second
	if v, ok := attributes["interval"].(int); ok {
		interval = time.Duration(v) * time.Second
	}

	go func() {
		for tick := time.NewTicker(interval); ; <-tick.C {
			if !d.running {
				return
			}

			if err := d.Display(idx, attributes); err != nil {
				log.WithError(err).Error("Unable to refresh element")
			}
		}
	}()

	return nil
}

func (d *displayElementExec) StopLoopDisplay() error {
	d.running = false
	return nil
}

func (displayElementExec) drawText(c *freetype.Context, text string, textColor color.Color, fontsize float64, border int) error {
	c.SetSrc(image.NewUniform(color.RGBA{0x0, 0x0, 0x0, 0x0})) // Transparent for text size guessing

	var (
		err error
		ext fixed.Point26_6
	)
	for {
		c.SetFontSize(fontsize)

		ext, err = c.DrawString(text, freetype.Pt(0, 0))
		if err != nil {
			return errors.Wrap(err, "Unable to measure text")
		}

		if int(float64(ext.X)/64) > sd.IconSize()-2*border || int(c.PointToFixed(fontsize/2.0)/64) > sd.IconSize()-2*border {
			fontsize -= 2
			continue
		}

		break
	}

	c.SetSrc(image.NewUniform(textColor))

	xcenter := (float64(sd.IconSize()-2*border) / 2.0) - (float64(int(float64(ext.X)/64)) / 2.0) + float64(border)
	ycenter := (float64(sd.IconSize()-2*border) / 2.0) + (float64(c.PointToFixed(fontsize/2.0)/64) / 2.0) + float64(border)

	_, err = c.DrawString(text, freetype.Pt(int(xcenter), int(ycenter)))

	return err
}

func (displayElementExec) getImageFromDisk(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open image")
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, errors.Wrap(err, "Umable to decode image")
	}

	return img, nil
}

func (displayElementExec) loadFont() (*truetype.Font, error) {
	fontRaw, err := ioutil.ReadFile(userConfig.RenderFont)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read font file")
	}

	return truetype.Parse(fontRaw)
}
