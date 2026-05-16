// Package image provides image display elements.
package image

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/modules/opts"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/renderer"
)

const cacheDirMode = 0o700

type (
	// Display renders an image on a key.
	Display struct{}

	// Attrs contains configuration for the image display.
	Attrs struct {
		Caption string `json:"caption,omitempty" yaml:"caption,omitempty"`
		Path    string `json:"path,omitempty" yaml:"path,omitempty"`
		URL     string `json:"url,omitempty" yaml:"url,omitempty"`
	}
)

// Display renders the configured image on the selected key.
func (d Display) Display(ctx context.Context, idx int, devs opts.Runtime, atts config.DynamicAttributes) error {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return fmt.Errorf("decoding attributes: %w", err)
	}

	filename, err := d.getRenderImageFileName(ctx, attributes)
	if err != nil {
		return err
	}

	imgRenderer := renderer.NewTextOnImageRenderer(devs)

	if err = imgRenderer.DrawBackgroundFromFile(filename); err != nil {
		return fmt.Errorf("drawing background from disk: %w", err)
	}

	if strings.TrimSpace(attributes.Caption) != "" {
		if err = imgRenderer.DrawCaptionText(strings.TrimSpace(attributes.Caption)); err != nil {
			return fmt.Errorf("rendering caption: %w", err)
		}
	}

	if err = ctx.Err(); err != nil {
		// Page context was cancelled, do not draw
		return fmt.Errorf("page context cancelled: %w", err)
	}

	if err = devs.Deck.FillImage(idx, imgRenderer.GetImage()); err != nil {
		return fmt.Errorf("setting image: %w", err)
	}

	return nil
}

func (Display) getCacheFileName(url string) (string, error) {
	ucd, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("getting user cache dir: %w", err)
	}

	cacheDir := path.Join(ucd, "io.luzifer.streamdeck")
	if err = os.MkdirAll(cacheDir, cacheDirMode); err != nil {
		return "", fmt.Errorf("creating cache dir: %w", err)
	}

	return path.Join(cacheDir, fmt.Sprintf("%x", sha256.Sum256([]byte(url)))), nil
}

func (d Display) getRenderImageFileName(ctx context.Context, attributes Attrs) (filename string, err error) {
	if attributes.Path != "" {
		// User supplied a path, rely on that
		return attributes.Path, nil
	}

	if attributes.URL == "" {
		// We have neither an URL nor a filename
		return "", fmt.Errorf("no path or url attribute specified")
	}

	filename, err = d.getCacheFileName(attributes.URL)
	if err != nil {
		return "", fmt.Errorf("getting cache filename for image url: %w", err)
	}

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, attributes.URL, nil)
		if err != nil {
			return "", fmt.Errorf("creating image request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", fmt.Errorf("requesting image url: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				logrus.WithError(err).Error("closing image response (leaked fd)")
			}
		}()

		imgFile, err := os.Create(filename) //#nosec:G304 // safely calculated path
		if err != nil {
			return "", fmt.Errorf("creating cache file: %w", err)
		}

		if _, err = io.Copy(imgFile, resp.Body); err != nil {
			_ = imgFile.Close()
			return "", fmt.Errorf("downloading file: %w", err)
		}

		if err = imgFile.Close(); err != nil {
			return "", fmt.Errorf("closing cache file: %w", err)
		}
	}

	return filename, nil
}
