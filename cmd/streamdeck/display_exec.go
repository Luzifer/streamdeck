package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/pkg/errors"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func init() {
	registerDisplayElement("exec", displayElementCommand{})
}

type displayElementCommand struct{}

func (d displayElementCommand) Display(idx int, attributes map[string]interface{}) error {
	var (
		err error
		img draw.Image = image.NewRGBA(image.Rect(0, 0, sd.IconSize(), sd.IconSize()))
	)

	// Initialize black image
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0x0, 0x0, 0x0, 0xff}), image.ZP, draw.Src)

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

	if filename, ok := attributes["image"].(string); ok {
		bgi, err := d.getImageFromDisk(filename)
		if err != nil {
			return errors.Wrap(err, "Unable to get image from disk")
		}

		draw.Draw(img, img.Bounds(), bgi, image.ZP, draw.Src)
	}

	var textColor color.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	if rgba, ok := attributes["color"].([]interface{}); ok {
		if len(rgba) != 4 {
			return errors.New("RGBA color definition needs 4 hex values")
		}

		textColor = color.RGBA{uint8(rgba[0].(int)), uint8(rgba[1].(int)), uint8(rgba[2].(int)), uint8(rgba[3].(int))}
	}

	var fontsize float64 = 120
	if v, ok := attributes["font_size"].(float64); ok {
		fontsize = v
	}

	var border = 10
	if v, ok := attributes["border"].(int); ok {
		border = v
	}

	var buf = new(bytes.Buffer)

	command := exec.Command(args[0], args[1:]...)
	command.Env = os.Environ()
	command.Stdout = buf

	if err := command.Run(); err != nil {
		return errors.Wrap(err, "Command has exit != 0")
	}

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

	if strings.TrimSpace(buf.String()) != "" {
		if err = d.drawText(c, strings.TrimSpace(buf.String()), textColor, fontsize, border); err != nil {
			return errors.Wrap(err, "Unable to render text")
		}
	}

	return errors.Wrap(sd.FillImage(idx, img), "Unable to set image")
}

func (displayElementCommand) drawText(c *freetype.Context, text string, textColor color.Color, fontsize float64, border int) error {
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

	c.SetSrc(image.NewUniform(textColor)) // TODO: Dynamic

	xcenter := (float64(sd.IconSize()-2*border) / 2.0) - (float64(int(float64(ext.X)/64)) / 2.0) + float64(border)
	ycenter := (float64(sd.IconSize()-2*border) / 2.0) + (float64(c.PointToFixed(fontsize/2.0)/64) / 2.0) + float64(border)

	_, err = c.DrawString(text, freetype.Pt(int(xcenter), int(ycenter)))

	return err
}

func (displayElementCommand) getImageFromDisk(filename string) (image.Image, error) {
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

func (displayElementCommand) loadFont() (*truetype.Font, error) {
	fontRaw, err := ioutil.ReadFile(userConfig.RenderFont)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read font file")
	}

	return truetype.Parse(fontRaw)
}
