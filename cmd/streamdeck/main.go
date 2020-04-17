package main

import (
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
	"gopkg.in/yaml.v2"
)

var (
	cfg = struct {
		Config         string `flag:"config,c" vardefault:"config" description:"Configuration with page / key definitions"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	currentBrightness int

	userConfig     config
	activePage     page
	activePageName string
	activeLoops    []refreshingDisplayElement

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

func loadConfig() error {
	userConfFile, err := os.Open(cfg.Config)
	if err != nil {
		return errors.Wrap(err, "Unable to open config")
	}
	defer userConfFile.Close()

	var tempConf config
	if err = yaml.NewDecoder(userConfFile).Decode(&tempConf); err != nil {
		return errors.Wrap(err, "Unable to parse config")
	}

	applySystemPages(&tempConf)

	userConfig = tempConf

	return nil
}

func main() {
	var err error

	// Initalize control devices
	kbd, err = uinput.CreateKeyboard()
	if err != nil {
		log.WithError(err).Fatal("Unable to create uinput keyboard")
	}
	defer kbd.Close()

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

	for {
		select {
		case evt := <-sd.Subscribe():
			if userConfig.DisplayOffTime > 0 {
				offTimer.Reset(userConfig.DisplayOffTime)
			}

			if evt.Key >= len(activePage.Keys) {
				continue
			}

			kd := activePage.Keys[evt.Key]
			if kd.On == "down" && evt.Type == streamdeck.EventTypeDown || (kd.On == "up" || kd.On == "") && evt.Type == streamdeck.EventTypeUp || kd.On == "both" {
				if err := triggerAction(kd); err != nil {
					log.WithError(err).Error("Unable to execute action")
				}
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

				var nextPage = userConfig.DefaultPage
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
	// Reset potentially running looped elements
	for _, l := range activeLoops {
		if err := l.StopLoopDisplay(); err != nil {
			return errors.Wrap(err, "Unable to stop element refresh")
		}
	}
	activeLoops = nil

	activePage = userConfig.Pages[page]
	activePageName = page
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
	for _, a := range kd.Actions {
		if a.Type != "" {
			if err := callAction(a); err != nil {
				return err
			}
		}
	}

	return nil
}
