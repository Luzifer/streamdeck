package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sstallion/go-hid"

	"github.com/Luzifer/streamdeck/v2"
)

func getAvailableDecks() (out []uint16, err error) {
	if err = hid.Enumerate(streamdeck.VendorElgato, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		if _, ok := streamdeck.DeckToName[info.ProductID]; !ok {
			// Is from Elgato but not a supported StreamDeck
			return nil
		}

		out = append(out, info.ProductID)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("enumerating devices: %w", err)
	}

	return out, nil
}

func listDevices() error {
	av, err := getAvailableDecks()
	if err != nil {
		return fmt.Errorf("getting available decks: %w", err)
	}

	for _, id := range av {
		fmt.Printf("0x%x - %s\n", id, streamdeck.DeckToName[id]) //nolint:forbidigo // printing explicitly requested
	}

	return nil
}

func selectDeckToUse() (uint16, error) {
	if cfg.ProductID != "" {
		// User selected a specific deck to use
		id, err := strconv.ParseUint(strings.TrimPrefix(cfg.ProductID, "0x"), 16, 16)
		if err != nil {
			return 0, fmt.Errorf("parsing given product ID: %w", err)
		}
		return uint16(id), nil
	}

	av, err := getAvailableDecks()
	if err != nil {
		return 0, fmt.Errorf("getting available decks: %w", err)
	}

	if len(av) == 0 {
		return 0, fmt.Errorf("found no supported decks")
	}

	// There is at least one supported deck, use the first one
	return av[0], nil
}
