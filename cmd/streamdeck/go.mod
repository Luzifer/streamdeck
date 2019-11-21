module github.com/Luzifer/streamdeck/cmd/streamdeck

go 1.13

replace github.com/Luzifer/streamdeck => ../../

require (
	github.com/Luzifer/rconfig/v2 v2.2.1
	github.com/Luzifer/streamdeck v0.0.0-20191120013804-40fbd11001d0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.0.0-20191120155948-bd437916bb0e // indirect
	gopkg.in/validator.v2 v2.0.0-20191107172027-c3144fdedc21 // indirect
	gopkg.in/yaml.v2 v2.2.7
)
