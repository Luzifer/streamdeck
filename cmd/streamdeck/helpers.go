package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"

	"github.com/nfnt/resize"
	"golang.org/x/image/draw"
)

func autoSizeImage(img image.Image, size int) image.Image {
	if img.Bounds().Max.X == size && img.Bounds().Max.Y == size {
		// Image has perfect size: Nothing to change
		return img
	}

	// Scale down when required
	img = resize.Thumbnail(uint(size), uint(size), img, resize.Lanczos2)
	if img.Bounds().Max.X == size && img.Bounds().Max.Y == size {
		// Image has perfect size now: Nothing to change
		return img
	}

	// Image is too small, need to pad it
	var (
		dimg = image.NewNRGBA(image.Rect(0, 0, size, size))
		pt   = image.Point{
			X: (size - img.Bounds().Max.X) / 2,
			Y: (size - img.Bounds().Max.Y) / 2,
		}
	)
	// Draw black background
	draw.Copy(dimg, image.Point{}, image.NewUniform(color.RGBA{0x0, 0x0, 0x0, 0xff}), dimg.Bounds(), draw.Src, nil)
	// Copy in image
	draw.Copy(dimg, pt, img, img.Bounds(), draw.Src, nil)

	return dimg
}

func int4ToRGBA(parts []int) (c color.RGBA, err error) {
	if len(parts) != 4 { //revive:disable-line:add-constant // single-use count
		return c, fmt.Errorf("color definition needs 4 numbers")
	}

	var rangeErrors []error
	for i, v := range parts {
		if v < 0 || v > 255 { //revive:disable-line:add-constant // single-use boundary
			rangeErrors = append(rangeErrors, fmt.Errorf("color component %d out of bounds 0..255: %d", i, v))
		}
	}
	if err = errors.Join(rangeErrors...); err != nil {
		return c, err
	}

	//#nosec:G115 // all values are guarded
	return color.RGBA{uint8(parts[0]), uint8(parts[1]), uint8(parts[2]), uint8(parts[3])}, nil
}
