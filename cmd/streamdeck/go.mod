module github.com/Luzifer/streamdeck/cmd/streamdeck

go 1.26.0

replace github.com/Luzifer/streamdeck => ../../

require (
	github.com/Luzifer/go_helpers/env v0.5.2
	github.com/Luzifer/rconfig/v2 v2.6.2
	github.com/Luzifer/streamdeck v1.7.1
	github.com/fsnotify/fsnotify v1.10.1
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/sashko/go-uinput v0.0.0-20250718151327-faf003f14a20
	github.com/sirupsen/logrus v1.9.4
	github.com/sstallion/go-hid v0.15.0
	github.com/stretchr/testify v1.11.1
	go.yaml.in/yaml/v3 v3.0.4
	golang.org/x/image v0.40.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/disintegration/imaging v1.6.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/sys v0.26.0 // indirect
	gopkg.in/validator.v2 v2.0.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
