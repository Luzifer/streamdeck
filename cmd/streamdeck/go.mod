module github.com/Luzifer/streamdeck/cmd/streamdeck

go 1.13

replace github.com/Luzifer/streamdeck => ../../

require (
	github.com/Luzifer/rconfig/v2 v2.2.1
	github.com/Luzifer/streamdeck v0.0.0-20191122001547-bfc4857847cb
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0
	github.com/pkg/errors v0.8.1
	github.com/sashko/go-uinput v0.0.0-20180923134002-15fcac7aa54a
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/image v0.0.0-20191009234506-e7c1f5e7dbb8
	gopkg.in/yaml.v2 v2.2.7
)
