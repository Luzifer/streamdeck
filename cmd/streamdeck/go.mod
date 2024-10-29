module github.com/Luzifer/streamdeck/cmd/streamdeck

go 1.22

toolchain go1.23.2

replace github.com/Luzifer/streamdeck => ../../

require (
	github.com/Luzifer/go_helpers/v2 v2.25.0
	github.com/Luzifer/rconfig/v2 v2.5.2
	github.com/Luzifer/streamdeck v1.7.1
	github.com/fsnotify/fsnotify v1.7.0
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/jfreymuth/pulse v0.1.1
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/pkg/errors v0.9.1
	github.com/sashko/go-uinput v0.0.0-20200718185411-c753d6644126
	github.com/sirupsen/logrus v1.9.3
	github.com/sstallion/go-hid v0.14.1
	golang.org/x/image v0.21.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.26.0 // indirect
	gopkg.in/validator.v2 v2.0.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
