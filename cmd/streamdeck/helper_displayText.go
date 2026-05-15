package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

const (
	displayDPI = 72
)

const (
	textDrawAnchorCenter textDrawAnchor = iota
	textDrawAnchorBottom
	textDrawAnchorTop
)

type (
	textDrawAnchor uint

	textOnImageRenderer struct {
		img draw.Image
	}
)

func newTextOnImageRenderer() *textOnImageRenderer {
	// Create new black image in icon size
	var img draw.Image = image.NewRGBA(image.Rect(0, 0, sd.IconSize(), sd.IconSize()))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0x0, 0x0, 0x0, 0xff}), image.Point{}, draw.Src)

	return &textOnImageRenderer{
		img: img,
	}
}

func (t *textOnImageRenderer) DrawBackground(bgi image.Image) {
	bgi = autoSizeImage(bgi, sd.IconSize())
	draw.Draw(t.img, t.img.Bounds(), bgi, image.Point{}, draw.Src)
}

func (t *textOnImageRenderer) DrawBackgroundColor(col color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, sd.IconSize(), sd.IconSize()))

	for x := 0; x < sd.IconSize(); x++ {
		for y := 0; y < sd.IconSize(); y++ {
			img.Set(x, y, col)
		}
	}

	t.DrawBackground(img)
}

func (t *textOnImageRenderer) DrawBackgroundFromFile(filename string) error {
	bgi, err := t.getImageFromDisk(filename)
	if err != nil {
		return fmt.Errorf("getting image from disk: %w", err)
	}

	t.DrawBackground(bgi)
	return nil
}

func (t *textOnImageRenderer) DrawBigText(text string, fontSizeHint float64, border int, textColor color.Color) error {
	// Render text
	f, err := t.loadFont(userConfig.RenderFont)
	if err != nil {
		return fmt.Errorf("loading font: %w", err)
	}

	c := freetype.NewContext()
	c.SetClip(t.img.Bounds())
	c.SetDPI(displayDPI)
	c.SetDst(t.img)
	c.SetFont(f)
	c.SetHinting(font.HintingNone)

	return t.drawText(c, text, textColor, fontSizeHint, border, textDrawAnchorCenter)
}

func (t *textOnImageRenderer) DrawCaptionText(text string) error {
	fontFile := userConfig.CaptionFont
	if fontFile == "" {
		fontFile = userConfig.RenderFont
	}

	// Render text
	f, err := t.loadFont(fontFile)
	if err != nil {
		return fmt.Errorf("loading font: %w", err)
	}

	var textColor color.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	if cc, err := int4ToRGBA(userConfig.CaptionColor[:]); err == nil && cc.A != 0x0 { //revive:disable-line:add-constant // just a 0 in hex
		textColor = cc
	}

	var anchor textDrawAnchor
	switch userConfig.CaptionPosition {
	case captionPositionBottom, captionPositionEmpty:
		anchor = textDrawAnchorBottom

	case captionPositionTop:
		anchor = textDrawAnchorTop

	default:
		return fmt.Errorf("invalid caption position %q", userConfig.CaptionPosition)
	}

	c := freetype.NewContext()
	c.SetClip(t.img.Bounds())
	c.SetDPI(displayDPI)
	c.SetDst(t.img)
	c.SetFont(f)
	c.SetHinting(font.HintingNone)

	return t.drawText(c, text, textColor, userConfig.CaptionFontSize, userConfig.CaptionBorder, anchor)
}

func (t textOnImageRenderer) GetImage() image.Image { return t.img }

func (*textOnImageRenderer) drawText(c *freetype.Context, text string, textColor color.Color, fontsize float64, border int, anchor textDrawAnchor) error {
	c.SetSrc(image.NewUniform(color.RGBA{0x0, 0x0, 0x0, 0x0})) // Transparent for text size guessing

	textLines := strings.Split(text, "\n")

	for {
		c.SetFontSize(fontsize)

		var maxX fixed.Int26_6
		for _, tl := range textLines {
			ext, err := c.DrawString(tl, freetype.Pt(0, 0))
			if err != nil {
				return fmt.Errorf("measuring text: %w", err)
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
			return fmt.Errorf("measuring text: %w", err)
		}

		c.SetSrc(image.NewUniform(textColor))

		xcenter := (float64(sd.IconSize()-2*border) / 2.0) - (float64(int(float64(ext.X)/64)) / 2.0) + float64(border)
		ylower := yLineTop + int(c.PointToFixed(fontsize)/64)

		if _, err = c.DrawString(tl, freetype.Pt(int(xcenter), ylower)); err != nil {
			return fmt.Errorf("drawing text: %w", err)
		}

		yLineTop += int(c.PointToFixed(fontsize)/64) + 2
	}

	return nil
}

func (textOnImageRenderer) getImageFromDisk(filename string) (image.Image, error) {
	f, err := os.Open(filename) //#nosec:G304 // intended to open image from disk
	if err != nil {
		return nil, fmt.Errorf("opening image: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.WithError(err).Error("closing image from disk (leaked fd)")
		}
	}()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decoding image: %w", err)
	}

	return img, nil
}

func (textOnImageRenderer) loadFont(fontfile string) (parsedFont *truetype.Font, err error) {
	fontRaw, err := os.ReadFile(fontfile) //#nosec:G304 // intended to open font from disk
	if err != nil {
		return nil, fmt.Errorf("reading font file: %w", err)
	}

	if parsedFont, err = truetype.Parse(fontRaw); err != nil {
		return nil, fmt.Errorf("parsing font: %w", err)
	}

	return parsedFont, nil
}
