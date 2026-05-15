package main

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

func togglePage(page string) (err error) {
	if activePageCtxCancel != nil {
		// Ensure old display events are no longer executed
		activePageCtxCancel()
	}

	// Reset potentially running looped elements
	for _, l := range activeLoops {
		if err := l.StopLoopDisplay(); err != nil {
			return fmt.Errorf("stopping element refresh: %w", err)
		}
	}
	activeLoops = nil

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

		go func(idx int, kd keyDefinition) {
			localCtx := activePageCtx
			keyLogger := logrus.WithFields(logrus.Fields{
				"key":  idx,
				"page": activePageName,
			})

			if err := callDisplayElement(localCtx, idx, kd); err != nil {
				keyLogger.WithError(err).Error("Unable to execute display element")

				if err := callErrorDisplayElement(localCtx, idx); err != nil {
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

func (p page) GetKeyDefinitions(cfg config) map[int]keyDefinition {
	var (
		defMaps []map[int]keyDefinition
		result  = make(map[int]keyDefinition)
	)

	// First process underlay if defined
	if p.Underlay != "" {
		defMaps = append(defMaps, cfg.Pages[p.Underlay].Keys)
	}

	// Process current definition
	defMaps = append(defMaps, p.Keys)

	// Last process overlay if defined
	if p.Overlay != "" {
		defMaps = append(defMaps, cfg.Pages[p.Overlay].Keys)
	}

	// Assemble combination of keys
	for _, pageDef := range defMaps {
		for idx, kd := range pageDef {
			if kd.Display.Type == "" {
				continue
			}

			result[idx] = kd
		}
	}

	return result
}
