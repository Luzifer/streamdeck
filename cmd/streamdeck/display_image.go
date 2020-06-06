package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/pkg/errors"
)

func init() {
	registerDisplayElement("image", displayElementImage{})
}

type displayElementImage struct{}

func (d displayElementImage) Display(idx int, attributes map[string]interface{}) error {
	filename, ok := attributes["path"].(string)
	if !ok {
		return errors.New("No path attribute specified")
	}

	f, err := os.Open(filename)
	if err != nil {
		return errors.Wrap(err, "Unable to open image")
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return errors.Wrap(err, "Umable to decode image")
	}

	img = autoSizeImage(img, sd.IconSize())

	return errors.Wrap(sd.FillImage(idx, img), "Unable to set image")
}
