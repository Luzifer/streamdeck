package main

import (
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/pkg/errors"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type textDrawAnchor uint

const (
	textDrawAnchorCenter textDrawAnchor = iota
	textDrawAnchorBottom
	textDrawAnchorTop
)

type textOnImageRenderer struct {
	img draw.Image
}

func newTextOnImageRenderer() *textOnImageRenderer {
	// Create new black image in icon size
	var img draw.Image = image.NewRGBA(image.Rect(0, 0, sd.IconSize(), sd.IconSize()))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0x0, 0x0, 0x0, 0xff}), image.ZP, draw.Src)

	return &textOnImageRenderer{
		img: img,
	}
}

func (t *textOnImageRenderer) DrawBackground(bgi image.Image) {
	bgi = autoSizeImage(bgi, sd.IconSize())
	draw.Draw(t.img, t.img.Bounds(), bgi, image.ZP, draw.Src)
}

func (t *textOnImageRenderer) DrawBackgroundFromFile(filename string) error {
	bgi, err := t.getImageFromDisk(filename)
	if err != nil {
		return errors.Wrap(err, "Unable to get image from disk")
	}

	t.DrawBackground(bgi)
	return nil
}

func (t *textOnImageRenderer) DrawBigText(text string, fontSizeHint float64, border int, textColor color.Color) error {
	// Render text
	f, err := t.loadFont(userConfig.RenderFont)
	if err != nil {
		return errors.Wrap(err, "Unable to load font")
	}

	c := freetype.NewContext()
	c.SetClip(t.img.Bounds())
	c.SetDPI(72)
	c.SetDst(t.img)
	c.SetFont(f)
	c.SetHinting(font.HintingNone)

	return t.drawText(c, text, textColor, fontSizeHint, border, textDrawAnchorCenter)
}

func (t *textOnImageRenderer) DrawCaptionText(text string) error {
	// Render text
	f, err := t.loadFont(userConfig.CaptionFont)
	if err != nil {
		return errors.Wrap(err, "Unable to load font")
	}

	var textColor color.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	if userConfig.CaptionColor[3] != 0x0 {
		textColor = color.RGBA{
			uint8(userConfig.CaptionColor[0]),
			uint8(userConfig.CaptionColor[1]),
			uint8(userConfig.CaptionColor[2]),
			uint8(userConfig.CaptionColor[3]),
		}
	}

	var anchor = textDrawAnchorBottom
	if userConfig.CaptionPosition == "top" {
		anchor = textDrawAnchorTop
	}

	c := freetype.NewContext()
	c.SetClip(t.img.Bounds())
	c.SetDPI(72)
	c.SetDst(t.img)
	c.SetFont(f)
	c.SetHinting(font.HintingNone)

	return t.drawText(c, text, textColor, userConfig.CaptionFontSize, userConfig.CaptionBorder, anchor)
}

func (t textOnImageRenderer) GetImage() image.Image { return t.img }

func (t *textOnImageRenderer) drawText(c *freetype.Context, text string, textColor color.Color, fontsize float64, border int, anchor textDrawAnchor) error {
	c.SetSrc(image.NewUniform(color.RGBA{0x0, 0x0, 0x0, 0x0})) // Transparent for text size guessing

	textLines := strings.Split(text, "\n")

	for {
		c.SetFontSize(fontsize)

		var maxX fixed.Int26_6
		for _, tl := range textLines {
			ext, err := c.DrawString(tl, freetype.Pt(0, 0))
			if err != nil {
				return errors.Wrap(err, "Unable to measure text")
			}
			if ext.X > maxX {
				maxX = ext.X
			}
		}

		if int(float64(maxX)/64) > sd.IconSize()-2*border || (int(c.PointToFixed(fontsize)/64))*len(textLines)+(len(textLines)-1)*2 > sd.IconSize()-2*border {
			fontsize -= 2
			continue
		}

		break
	}

	var (
		yTotal   = (int(c.PointToFixed(fontsize)/64))*len(textLines) + (len(textLines)-1)*2
		yLineTop int
	)

	switch anchor {
	case textDrawAnchorTop:
		yLineTop = border
	case textDrawAnchorCenter:
		yLineTop = int(float64(sd.IconSize())/2.0 - float64(yTotal)/2.0)
	case textDrawAnchorBottom:
		yLineTop = sd.IconSize() - yTotal - border
	}

	for _, tl := range textLines {
		c.SetSrc(image.NewUniform(color.RGBA{0x0, 0x0, 0x0, 0x0})) // Transparent for text size guessing
		ext, err := c.DrawString(tl, freetype.Pt(0, 0))
		if err != nil {
			return errors.Wrap(err, "Unable to measure text")
		}

		c.SetSrc(image.NewUniform(textColor))

		xcenter := (float64(sd.IconSize()-2*border) / 2.0) - (float64(int(float64(ext.X)/64)) / 2.0) + float64(border)
		ylower := yLineTop + int(c.PointToFixed(fontsize)/64)

		if _, err = c.DrawString(tl, freetype.Pt(int(xcenter), int(ylower))); err != nil {
			return errors.Wrap(err, "Unable to draw text")
		}

		yLineTop += int(c.PointToFixed(fontsize)/64) + 2
	}

	return nil
}

func (textOnImageRenderer) getImageFromDisk(filename string) (image.Image, error) {
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

func (textOnImageRenderer) loadFont(fontfile string) (*truetype.Font, error) {
	fontRaw, err := ioutil.ReadFile(fontfile)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to read font file")
	}

	return truetype.Parse(fontRaw)
}
