package main

import (
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
