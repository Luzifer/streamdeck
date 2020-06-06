package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Luzifer/streamdeck"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/sstallion/go-hid"
)

func getAvailableDecks() ([]uint16, error) {
	var out []uint16
	return out, hid.Enumerate(streamdeck.VendorElgato, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		if _, ok := streamdeck.DeckToName[info.ProductID]; !ok {
			// Is from Elgato but not a supported StreamDeck
			return nil
		}

		out = append(out, info.ProductID)
		return nil
	})
}

func listAndQuit() {
	av, err := getAvailableDecks()
	if err != nil {
		log.WithError(err).Fatal("Unable to get available decks")
	}

	for _, id := range av {
		fmt.Printf("0x%x - %s\n", id, streamdeck.DeckToName[id])
	}

	// Quit now as listing is done
	os.Exit(0)
}

func selectDeckToUse() (uint16, error) {
	if cfg.ProductID != "" {
		// User selected a specific deck to use
		id, err := strconv.ParseUint(strings.TrimPrefix(cfg.ProductID, "0x"), 16, 16)
		return uint16(id), errors.Wrap(err, "Unable to parse given product ID")
	}

	av, err := getAvailableDecks()
	if err != nil {
		return 0x0, errors.Wrap(err, "Unable to get available decks")
	}

	if len(av) == 0 {
		return 0x0, errors.Wrap(err, "Found no supported decks")
	}

	// There is at least one supported deck, use the first one
	return av[0], nil
}
