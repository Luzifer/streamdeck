// Package httpdisplay provides HTTP-backed display elements.
package httpdisplay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/displays/text"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/modules/opts"
)

const minLoopInterval = 100 * time.Millisecond

type (
	// Display renders text attributes produced by an HTTP response.
	Display struct{}

	// Attrs contains configuration for the HTTP display.
	Attrs struct {
		Body         string            `yaml:"body"`
		ExpectStatus int               `yaml:"expect_status"`
		Headers      map[string]string `yaml:"headers"`
		Method       string            `yaml:"method"`
		URL          string            `yaml:"url"`
		Interval     time.Duration     `yaml:"interval"`
		Timeout      time.Duration     `yaml:"timeout"`

		text.Attrs `yaml:",inline"`
	}
)

// Display executes the HTTP request and renders its response as a text display.
//
//nolint:gocyclo // just some default setting
func (Display) Display(ctx context.Context, idx int, devs opts.Runtime, atts config.DynamicAttributes) (err error) {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return fmt.Errorf("decoding attributes: %w", err)
	}

	if attributes.URL == "" {
		return fmt.Errorf("no URL supplied")
	}

	if attributes.Method == "" {
		attributes.Method = http.MethodGet
	}

	if attributes.ExpectStatus == 0 {
		attributes.ExpectStatus = http.StatusOK
	}

	var body io.Reader
	if strings.TrimSpace(attributes.Body) != "" {
		body = strings.NewReader(attributes.Body)
	}

	var (
		reqCtx = ctx
		cancel context.CancelFunc
	)
	if attributes.Timeout > 0 {
		reqCtx, cancel = context.WithTimeout(ctx, attributes.Timeout)
		defer cancel()
	} else if attributes.Interval > minLoopInterval {
		reqCtx, cancel = context.WithTimeout(ctx, attributes.Interval)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(reqCtx, attributes.Method, attributes.URL, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	if attributes.Headers != nil {
		for k, v := range attributes.Headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Error("closing http display body")
		}
	}()

	if resp.StatusCode != attributes.ExpectStatus {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %w", err)
	}

	// Build Text-Render
	forwardAtts := attributes.Attrs

	switch ct := resp.Header.Get("Content-Type"); ct {
	case "application/json":
		if err = json.Unmarshal(raw, &forwardAtts); err != nil {
			return fmt.Errorf("parsing json response: %w", err)
		}

	case "text/plain":
		forwardAtts.Text = strings.TrimSpace(string(raw))

	default:
		return fmt.Errorf("unexpected content type %q", ct)
	}

	return new(text.Display).Render(ctx, idx, devs, forwardAtts) //nolint:wrapcheck // fine for this as that's a normal render module itself
}

// NeedsLoop reports whether the display should refresh periodically.
func (Display) NeedsLoop(atts config.DynamicAttributes) bool {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return false
	}

	if attributes.Interval > 0 && attributes.Interval < minLoopInterval {
		logrus.WithFields(logrus.Fields{
			"type":     "http",
			"interval": attributes.Interval,
		}).Warn("Ignoring interval below 100ms")
	}

	return attributes.Interval > minLoopInterval
}

// StartLoopDisplay starts periodic display refresh until the context is cancelled.
func (d *Display) StartLoopDisplay(ctx context.Context, idx int, devs opts.Runtime, atts config.DynamicAttributes) error {
	attributes, err := config.DecodeAttributes[Attrs](atts)
	if err != nil {
		return fmt.Errorf("decoding attributes: %w", err)
	}

	go func() {
		tick := time.NewTicker(attributes.Interval)
		defer tick.Stop()

		for {
			if err := d.Display(ctx, idx, devs, atts); err != nil {
				if errors.Is(ctx.Err(), context.Canceled) {
					return
				}

				logrus.WithError(err).Error("refreshing element")
			}

			select {
			case <-ctx.Done():
				return

			case <-tick.C:
			}
		}
	}()

	return nil
}
