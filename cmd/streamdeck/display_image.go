package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/pkg/errors"
)

func init() {
	registerDisplayElement("image", displayElementImage{})
}

type displayElementImage struct{}

func (d displayElementImage) Display(ctx context.Context, idx int, attributes attributeCollection) error {
	filename := attributes.Path
	if filename == "" {
		if attributes.URL == "" {
			return errors.New("No path or url attribute specified")
		}

		var err error
		filename, err = d.getCacheFileName(attributes.URL)
		if err != nil {
			return errors.Wrap(err, "Unable to get cache filename for image url")
		}

		if _, err := os.Stat(filename); os.IsNotExist(err) {
			resp, err := http.Get(attributes.URL)
			if err != nil {
				return errors.Wrap(err, "Unable to request image url")
			}
			defer resp.Body.Close()

			imgFile, err := os.Create(filename)
			if err != nil {
				return errors.Wrap(err, "Unable to create cache file")
			}

			if _, err = io.Copy(imgFile, resp.Body); err != nil {
				imgFile.Close()
				return errors.Wrap(err, "Unable to download file")
			}

			imgFile.Close()
		}
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

	if err := ctx.Err(); err != nil {
		// Page context was cancelled, do not draw
		return err
	}

	return errors.Wrap(sd.FillImage(idx, img), "Unable to set image")
}

func (d displayElementImage) getCacheFileName(url string) (string, error) {
	ucd, err := os.UserCacheDir()
	if err != nil {
		return "", errors.Wrap(err, "Unable to get user cache dir")
	}

	cacheDir := path.Join(ucd, "io.luzifer.streamdeck")
	if err = os.MkdirAll(cacheDir, 0o750); err != nil {
		return "", errors.Wrap(err, "Unable to create cache dir")
	}

	return path.Join(cacheDir, fmt.Sprintf("%x", sha256.Sum256([]byte(url)))), nil
}
