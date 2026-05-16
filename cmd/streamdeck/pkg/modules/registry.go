package modules

import (
	execaction "github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/actions/exec"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/actions/keypress"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/actions/page"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/actions/reload"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/actions/toggledisplay"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/displays/color"
	execdisplay "github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/displays/exec"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/displays/image"
	"github.com/Luzifer/streamdeck/cmd/streamdeck/pkg/displays/text"
)

func init() {
	registerAction("exec", execaction.Action{})
	registerAction("key_press", keypress.Action{})
	registerAction("page", page.Action{})
	registerAction("reload_config", reload.Action{})
	registerAction("toggle_display", toggledisplay.Action{})

	registerDisplayElement("color", color.Display{})
	registerDisplayElement("exec", &execdisplay.Display{})
	registerDisplayElement("text", text.Display{})
	registerDisplayElement("image", image.Display{})
}
