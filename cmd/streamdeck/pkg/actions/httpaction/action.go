// Package httpaction provides HTTP request actions.
package httpaction

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/modules/opts"
	"github.com/sirupsen/logrus"
)

const defaultRequestTimeout = 30 * time.Second

type (
	// Action executes a configured HTTP request.
	Action struct{}

	// Attrs contains configuration for the HTTP action.
	Attrs struct {
		Body         string            `yaml:"body"`
		ExpectStatus int               `yaml:"expect_status"`
		Headers      map[string]string `yaml:"headers"`
		Method       string            `yaml:"method"`
		URL          string            `yaml:"url"`
		Timeout      time.Duration     `yaml:"timeout"`
	}
)

// Execute runs the configured HTTP request.
func (Action) Execute(_ opts.Runtime, atts config.DynamicAttributes) (err error) {
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
		reqCtx = context.Background()
		cancel context.CancelFunc
	)
	if attributes.Timeout > 0 {
		reqCtx, cancel = context.WithTimeout(reqCtx, attributes.Timeout)
		defer cancel()
	} else {
		reqCtx, cancel = context.WithTimeout(reqCtx, defaultRequestTimeout)
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
			logrus.WithError(err).Error("closing http action body")
		}
	}()

	if resp.StatusCode != attributes.ExpectStatus {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}
