package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/Luzifer/rconfig/v2"
	"github.com/Luzifer/streamdeck"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/sashko/go-uinput"
	log "github.com/sirupsen/logrus"
)

var (
	cfg = struct {
		Config         string `flag:"config,c" vardefault:"config" description:"Configuration with page / key definitions"`
		List           bool   `flag:"list,l" default:"false" description:"List all available StreamDecks"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		ProductID      string `flag:"product-id,p" default:"" description:"Specify StreamDeck to use (use list to find ID), default first found"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	currentBrightness int

	userConfig          config
	activePage          page
	activePageCtx       context.Context
	activePageCtxCancel context.CancelFunc
	activePageName      string
	activeLoops         []refreshingDisplayElement
	pageStack           []string

	sd *streamdeck.Client

	kbd uinput.Keyboard

	version = "dev"
)

func init() {
	cfgDir, _ := os.UserConfigDir()
	rconfig.SetVariableDefaults(map[string]string{
		"config": path.Join(cfgDir, "streamdeck.yaml"),
	})

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
	if cfg.List {
		listAndQuit()
	}

	deck, err := selectDeckToUse()
	if err != nil {
		log.WithError(err).Fatal("Unable to select StreamDeck to use")
	}

	// Initalize control devices
	kbd, err = uinput.CreateKeyboard()
	if err != nil {
		log.WithError(err).Fatal("Unable to create uinput keyboard")
	}
	defer kbd.Close()

	// Initialize device
	sd, err = streamdeck.New(deck)
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

	// Load config
	if err = loadConfig(); err != nil {
		log.WithError(err).Fatal("Unable to load config")
	}

	// Initial setup

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	defer sd.ResetToLogo()

	if err = sd.SetBrightness(userConfig.DefaultBrightness); err != nil {
		log.WithError(err).Fatal("Unable to set brightness")
	}
	currentBrightness = userConfig.DefaultBrightness

	if err = togglePage(userConfig.DefaultPage); err != nil {
		log.WithError(err).Error("Unable to load default page")
	}

	var offTimer *time.Timer = &time.Timer{}
	if userConfig.DisplayOffTime > 0 {
		offTimer = time.NewTimer(userConfig.DisplayOffTime)
	}

	fswatch, err := fsnotify.NewWatcher()
	if err != nil {
		log.WithError(err).Fatal("Unable to create file watcher")
	}

	if userConfig.AutoReload {
		if err = fswatch.Add(cfg.Config); err != nil {
			log.WithError(err).Error("Unable to watch config, auto-reload will not work")
		}
	}

	var (
		actor    *int
		actStart time.Time
	)

	for {
		select {
		case evt := <-sd.Subscribe():
			if userConfig.DisplayOffTime > 0 {
				offTimer.Reset(userConfig.DisplayOffTime)
			}

			if evt.Type == streamdeck.EventTypeDown {
				actor = &evt.Key
				actStart = time.Now()
				continue
			}

			if evt.Key != *actor {
				continue
			}

			kd, ok := activePage.GetKeyDefinitions(userConfig)[*actor]
			if !ok {
				continue
			}

			isLongPress := time.Since(actStart) > userConfig.LongPressDuration

			if err := triggerAction(kd, isLongPress); err != nil {
				log.WithError(err).Error("Unable to execute action")
			}

		case <-offTimer.C:
			if err := togglePage("@@blank"); err != nil {
				log.WithError(err).Error("Unable to toggle to blank page")
			}

		case evt := <-fswatch.Events:
			if evt.Op&fsnotify.Write == fsnotify.Write {
				log.Info("Detected change of config, reloading")

				if err := loadConfig(); err != nil {
					log.WithError(err).Error("Unable to reload config")
					continue
				}

				nextPage := userConfig.DefaultPage
				if _, ok := userConfig.Pages[activePageName]; ok {
					nextPage = activePageName
				}

				if err := togglePage(nextPage); err != nil {
					log.WithError(err).Error("Unable to reload page")
					continue
				}
			}

		case <-sigs:
			return

		}
	}
}

func togglePage(page string) error {
	if activePageCtxCancel != nil {
		// Ensure old display events are no longer executed
		activePageCtxCancel()
	}

	// Reset potentially running looped elements
	for _, l := range activeLoops {
		if err := l.StopLoopDisplay(); err != nil {
			return errors.Wrap(err, "Unable to stop element refresh")
		}
	}
	activeLoops = nil

	activePage = userConfig.Pages[page]
	activePageName = page
	activePageCtx, activePageCtxCancel = context.WithCancel(context.Background())
	sd.ClearAllKeys()

	for idx, kd := range activePage.GetKeyDefinitions(userConfig) {
		if kd.Display.Type == "" {
			continue
		}

		go func(idx int, kd keyDefinition) {
			keyLogger := log.WithFields(log.Fields{
				"key":  idx,
				"page": activePageName,
			})

			if err := callDisplayElement(activePageCtx, idx, kd); err != nil {
				keyLogger.WithError(err).Error("Unable to execute display element")

				if err := callErrorDisplayElement(activePageCtx, idx); err != nil {
					keyLogger.WithError(err).Error("Unable to execute error display element")
				}
			}
		}(idx, kd)
	}

	if len(pageStack) == 0 || pageStack[0] != page {
		pageStack = append([]string{page}, pageStack...)
	}

	if len(pageStack) > 100 {
		pageStack = pageStack[:100]
	}

	return nil
}

func triggerAction(kd keyDefinition, isLongPress bool) error {
	for _, a := range kd.Actions {
		if a.Type == "" {
			// No type on that action: Invalid
			continue
		}

		if isLongPress != a.LongPress {
			// press duration does not match requirement
			continue
		}

		if err := callAction(a); err != nil {
			return err
		}
	}

	return nil
}
