package main

import (
	"fmt"
	"os"

	"github.com/Luzifer/rconfig/v2"
	"github.com/Luzifer/streamdeck"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	cfg = struct {
		Config         string `flag:"config,c" default:"config.yml" description:"Configuration with page / key definitions"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	currentBrightness int

	userConfig config
	activePage page

	sd *streamdeck.Client

	version = "dev"
)

func init() {
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("streamdeck %s\n", version)
		os.Exit(0)
	}

	if l, err := log.ParseLevel(cfg.LogLevel); err != nil {
		log.WithError(err).Fatal("Unable to parse log level")
	} else {
		log.SetLevel(l)
	}
}

func main() {
	// Load config
	userConfFile, err := os.Open(cfg.Config)
	if err != nil {
		log.WithError(err).Fatal("Unable to open config")
	}

	if err = yaml.NewDecoder(userConfFile).Decode(&userConfig); err != nil {
		log.WithError(err).Fatal("Unable to parse config")
	}

	userConfFile.Close()

	// Initialize device
	sd, err = streamdeck.New(streamdeck.StreamDeckOriginalV2)
	if err != nil {
		log.WithError(err).Fatal("Unable to open StreamDeck connection")
	}
	defer sd.Close()

	serial, err := sd.Serial()
	if err != nil {
		log.WithError(err).Fatal("Unable to read serial")
	}

	firmware, err := sd.GetFimwareVersion()
	if err != nil {
		log.WithError(err).Fatal("Unable to read firmware")
	}

	log.WithFields(log.Fields{
		"firmware": firmware,
		"serial":   serial,
	}).Info("Found StreamDeck")

	// Initial setup

	if err = sd.SetBrightness(userConfig.DefaultBrightness); err != nil {
		log.WithError(err).Fatal("Unable to set brightness")
	}
	currentBrightness = userConfig.DefaultBrightness

	if err = togglePage(userConfig.DefaultPage); err != nil {
		log.WithError(err).Error("Unable to load default page")
	}

	for {
		select {
		case evt := <-sd.Subscribe():
			if evt.Key >= len(activePage.Keys) {
				continue
			}

			kd := activePage.Keys[evt.Key]
			if kd.On == "down" && evt.Type == streamdeck.EventTypeDown || kd.On == "up" && evt.Type == streamdeck.EventTypeUp || kd.On == "both" {
				if err := triggerAction(kd); err != nil {
					log.WithError(err).Error("Unable to execute action")
				}
			}

		}
	}
}

func togglePage(page string) error {
	activePage = userConfig.Pages[page]
	sd.ClearAllKeys()

	for idx, kd := range activePage.Keys {
		if kd.Display.Type != "" {
			if err := callDisplayElement(idx, kd); err != nil {
				return errors.Wrapf(err, "Unable to execute display element on key %d", idx)
			}
		}
	}

	return nil
}

func triggerAction(kd keyDefinition) error {
	if kd.Action.Type != "" {
		return errors.Wrap(callAction(kd), "Unable to execute action")
	}

	return nil
}
