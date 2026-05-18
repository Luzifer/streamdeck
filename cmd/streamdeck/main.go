// StreamDeck Command Utility
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
	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/config"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/v2/pkg/modules"
	"github.com/fsnotify/fsnotify"
	"github.com/sashko/go-uinput"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/streamdeck/v2"
)

const maxPageStackSize = 100

var (
	cfg = struct {
		Config         string `flag:"config,c" vardefault:"config" description:"Configuration with page / key definitions"`
		List           bool   `flag:"list,l" default:"false" description:"List all available StreamDecks"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		ProductID      string `flag:"product-id,p" default:"" description:"Specify StreamDeck to use (use list to find ID), default first found"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	userConfig          config.File
	activePage          config.Page
	activePageCtx       context.Context
	activePageCtxCancel context.CancelFunc
	activePageName      string
	pageStack           []string

	sd *streamdeck.Client

	kbd uinput.Keyboard

	version = "dev"
)

func initApp() (err error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("getting user config-dir: %w", err)
	}

	rconfig.SetVariableDefaults(map[string]string{
		"config": path.Join(cfgDir, "streamdeck.yaml"),
	})

	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		return fmt.Errorf("parsing CLI options: %w", err)
	}

	l, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("parsing log-level: %w", err)
	}
	logrus.SetLevel(l)

	return nil
}

//nolint:gocognit,gocyclo,funlen // ignore for now, fix later™
func main() {
	var err error
	if err = initApp(); err != nil {
		logrus.WithError(err).Fatal("initializing app")
	}

	if cfg.VersionAndExit {
		fmt.Printf("streamdeck %s\n", version) //nolint:forbidigo // printing version to stdout is fine
		os.Exit(0)
	}

	if cfg.List {
		if err = listDevices(); err != nil {
			logrus.WithError(err).Fatal("listing devices")
		}
		os.Exit(0)
	}

	deck, err := selectDeckToUse()
	if err != nil {
		logrus.WithError(err).Fatal("Unable to select StreamDeck to use")
	}

	// Initialize control devices
	kbd, err = uinput.CreateKeyboard()
	if err != nil {
		logrus.WithError(err).Fatal("Unable to create uinput keyboard")
	}
	defer kbd.Close() //nolint:errcheck // closed either way by process exit

	// Initialize device
	sd, err = streamdeck.New(deck)
	if err != nil {
		logrus.WithError(err).Fatal("Unable to open StreamDeck connection")
	}
	defer sd.Close() //nolint:errcheck // closed either way by process exit

	serial, err := sd.Serial()
	if err != nil {
		logrus.WithError(err).Fatal("Unable to read serial")
	}

	firmware, err := sd.GetFimwareVersion()
	if err != nil {
		logrus.WithError(err).Fatal("Unable to read firmware")
	}

	logrus.WithFields(logrus.Fields{
		"firmware": firmware,
		"serial":   serial,
	}).Info("Found StreamDeck")

	// Load config
	if userConfig, err = config.Load(cfg.Config, sd); err != nil {
		logrus.WithError(err).Fatal("loading config")
	}

	// Initial setup

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)

	defer func() {
		if err := sd.ResetToLogo(); err != nil {
			logrus.WithError(err).Error("resetting to logo")
		}
	}()

	if err = sd.SetBrightness(userConfig.DefaultBrightness); err != nil {
		logrus.WithError(err).Fatal("Unable to set brightness")
	}

	if err = togglePage(userConfig.DefaultPage); err != nil {
		logrus.WithError(err).Error("Unable to load default page")
	}

	offTimer := &time.Timer{}
	if userConfig.DisplayOffTime > 0 {
		offTimer = time.NewTimer(userConfig.DisplayOffTime)
	}

	fswatch, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.WithError(err).Fatal("Unable to create file watcher")
	}

	if userConfig.AutoReload {
		if err = fswatch.Add(cfg.Config); err != nil {
			logrus.WithError(err).Error("Unable to watch config, auto-reload will not work")
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
				logrus.WithError(err).Error("Unable to execute action")
			}

		case <-offTimer.C:
			if err := togglePage("@@blank"); err != nil {
				logrus.WithError(err).Error("Unable to toggle to blank page")
			}

		case evt := <-fswatch.Events:
			if evt.Op&fsnotify.Write == fsnotify.Write {
				logrus.Info("Detected change of config, reloading")

				if err = reloadConfig(); err != nil {
					logrus.WithError(err).Error("reloading config")
					continue
				}
			}

		case <-sigs:
			return
		}
	}
}

func reloadConfig() (err error) {
	var tmpConfig config.File

	if tmpConfig, err = config.Load(cfg.Config, sd); err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	userConfig = tmpConfig

	nextPage := userConfig.DefaultPage
	if _, ok := userConfig.Pages[activePageName]; ok {
		nextPage = activePageName
	}

	if err := togglePage(nextPage); err != nil {
		return fmt.Errorf("reloading page: %w", err)
	}

	return nil
}

//revive:disable-next-line:flag-parameter // does not switch behavior, just denotes whether key was pressed long
func triggerAction(kd config.KeyDefinition, isLongPress bool) error {
	for _, a := range kd.Actions {
		if a.Type == "" {
			// No type on that action: Invalid
			continue
		}

		if isLongPress != a.LongPress {
			// press duration does not match requirement
			continue
		}

		if err := modules.CallAction(moduleRuntime(), a); err != nil {
			return fmt.Errorf("calling action: %w", err)
		}
	}

	return nil
}
