// Package renderer contains image rendering helpers for key displays.
package renderer

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

	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/helpers"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/modules/opts"
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

	// TextOnImageRenderer draws backgrounds, text, and captions into key-sized images.
	TextOnImageRenderer struct {
		devs opts.Runtime
		img  draw.Image
	}
)

// NewTextOnImageRenderer creates a renderer for the current StreamDeck key size.
func NewTextOnImageRenderer(devs opts.Runtime) *TextOnImageRenderer {
	// Create new black image in icon size
	var img draw.Image = image.NewRGBA(image.Rect(0, 0, devs.Deck.IconSize(), devs.Deck.IconSize()))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{0x0, 0x0, 0x0, 0xff}), image.Point{}, draw.Src)

	return &TextOnImageRenderer{
		devs: devs,
		img:  img,
	}
}

// DrawBackground draws an image as the key background.
func (t *TextOnImageRenderer) DrawBackground(bgi image.Image) {
	bgi = helpers.AutoSizeImage(bgi, t.devs.Deck.IconSize())
	draw.Draw(t.img, t.img.Bounds(), bgi, image.Point{}, draw.Src)
}

// DrawBackgroundColor fills the key background with a color.
func (t *TextOnImageRenderer) DrawBackgroundColor(col color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, t.devs.Deck.IconSize(), t.devs.Deck.IconSize()))

	for x := 0; x < t.devs.Deck.IconSize(); x++ {
		for y := 0; y < t.devs.Deck.IconSize(); y++ {
			img.Set(x, y, col)
		}
	}

	t.DrawBackground(img)
}

// DrawBackgroundFromFile loads an image file and draws it as the key background.
func (t *TextOnImageRenderer) DrawBackgroundFromFile(filename string) error {
	bgi, err := t.getImageFromDisk(filename)
	if err != nil {
		return fmt.Errorf("getting image from disk: %w", err)
	}

	t.DrawBackground(bgi)
	return nil
}

// DrawBigText draws centered text scaled to fit the key.
func (t *TextOnImageRenderer) DrawBigText(text string, fontSizeHint float64, border int, textColor color.Color) error {
	// Render text
	f, err := t.loadFont(t.devs.Conf.RenderFont)
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

// DrawCaptionText draws caption text using the configured caption style.
func (t *TextOnImageRenderer) DrawCaptionText(text string) error {
	fontFile := t.devs.Conf.CaptionFont
	if fontFile == "" {
		fontFile = t.devs.Conf.RenderFont
	}

	// Render text
	f, err := t.loadFont(fontFile)
	if err != nil {
		return fmt.Errorf("loading font: %w", err)
	}

	var textColor color.Color = color.RGBA{0xff, 0xff, 0xff, 0xff}
	if cc, err := helpers.Int4ToRGBA(t.devs.Conf.CaptionColor[:]); err == nil && cc.A != 0x0 { //revive:disable-line:add-constant // just a 0 in hex
		textColor = cc
	}

	var anchor textDrawAnchor
	switch t.devs.Conf.CaptionPosition {
	case config.CaptionPositionBottom, config.CaptionPositionEmpty:
		anchor = textDrawAnchorBottom

	case config.CaptionPositionTop:
		anchor = textDrawAnchorTop

	default:
		return fmt.Errorf("invalid caption position %q", t.devs.Conf.CaptionPosition)
	}

	c := freetype.NewContext()
	c.SetClip(t.img.Bounds())
	c.SetDPI(displayDPI)
	c.SetDst(t.img)
	c.SetFont(f)
	c.SetHinting(font.HintingNone)

	return t.drawText(c, text, textColor, t.devs.Conf.CaptionFontSize, t.devs.Conf.CaptionBorder, anchor)
}

// GetImage returns the rendered key image.
func (t TextOnImageRenderer) GetImage() image.Image { return t.img }

func (t *TextOnImageRenderer) drawText(c *freetype.Context, text string, textColor color.Color, fontsize float64, border int, anchor textDrawAnchor) error {
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

		if int(float64(maxX)/64) > t.devs.Deck.IconSize()-2*border || (int(c.PointToFixed(fontsize)/64))*len(textLines)+(len(textLines)-1)*2 > t.devs.Deck.IconSize()-2*border {
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
		yLineTop = int(float64(t.devs.Deck.IconSize())/2.0 - float64(yTotal)/2.0)
	case textDrawAnchorBottom:
		yLineTop = t.devs.Deck.IconSize() - yTotal - border
	}

	for _, tl := range textLines {
		c.SetSrc(image.NewUniform(color.RGBA{0x0, 0x0, 0x0, 0x0})) // Transparent for text size guessing
		ext, err := c.DrawString(tl, freetype.Pt(0, 0))
		if err != nil {
			return fmt.Errorf("measuring text: %w", err)
		}

		c.SetSrc(image.NewUniform(textColor))

		xcenter := (float64(t.devs.Deck.IconSize()-2*border) / 2.0) - (float64(int(float64(ext.X)/64)) / 2.0) + float64(border)
		ylower := yLineTop + int(c.PointToFixed(fontsize)/64)

		if _, err = c.DrawString(tl, freetype.Pt(int(xcenter), ylower)); err != nil {
			return fmt.Errorf("drawing text: %w", err)
		}

		yLineTop += int(c.PointToFixed(fontsize)/64) + 2
	}

	return nil
}

func (TextOnImageRenderer) getImageFromDisk(filename string) (image.Image, error) {
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

func (TextOnImageRenderer) loadFont(fontfile string) (parsedFont *truetype.Font, err error) {
	fontRaw, err := os.ReadFile(fontfile) //#nosec:G304 // intended to open font from disk
	if err != nil {
		return nil, fmt.Errorf("reading font file: %w", err)
	}

	if parsedFont, err = truetype.Parse(fontRaw); err != nil {
		return nil, fmt.Errorf("parsing font: %w", err)
	}

	return parsedFont, nil
}
