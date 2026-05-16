package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/modules"
)

func togglePage(page string) (err error) {
	if activePageCtxCancel != nil {
		// Ensure old display events are no longer executed
		activePageCtxCancel()
	}

	activePage = userConfig.Pages[page]
	activePageName = page
	activePageCtx, activePageCtxCancel = context.WithCancel(context.Background())
	if err = sd.ClearAllKeys(); err != nil {
		return fmt.Errorf("clearing keys: %w", err)
	}

	for idx, kd := range activePage.GetKeyDefinitions(userConfig) {
		if kd.Display.Type == "" {
			continue
		}

		go func(idx int, kd config.KeyDefinition) {
			localCtx := activePageCtx
			keyLogger := logrus.WithFields(logrus.Fields{
				"key":  idx,
				"page": activePageName,
			})

			if err := modules.CallDisplayElement(localCtx, idx, moduleRuntime(), kd); err != nil {
				keyLogger.WithError(err).Error("Unable to execute display element")

				if err := modules.CallErrorDisplayElement(localCtx, idx, moduleRuntime()); err != nil {
					keyLogger.WithError(err).Error("Unable to execute error display element")
				}
			}
		}(idx, kd)
	}

	if len(pageStack) == 0 || pageStack[0] != page {
		pageStack = append([]string{page}, pageStack...)
	}

	if len(pageStack) > maxPageStackSize {
		pageStack = pageStack[:maxPageStackSize]
	}

	return nil
}

func toggleRelativePage(rel int) (err error) {
	if rel >= len(pageStack) {
		return fmt.Errorf("relative page %d out of range", rel)
	}

	nextPage := pageStack[rel]
	pageStack = pageStack[rel+1:]

	if err = togglePage(nextPage); err != nil {
		return fmt.Errorf("switching relative page: %w", err)
	}

	return nil
}
