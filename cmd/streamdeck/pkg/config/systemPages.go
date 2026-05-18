package config

import (
	"github.com/Luzifer/streamdeck/v2"
)

func applySystemPages(deck *streamdeck.Client, conf *File) {
	blankPage := Page{Keys: make(map[int]KeyDefinition)}

	displayConf, _ := EncodeAttributes(struct {
		RGBA []int `json:"rgba,omitempty" yaml:"rgba,omitempty"`
	}{
		//revive:disable-next-line:add-constant // color definition
		RGBA: []int{0x0, 0x0, 0x0, 0xff},
	})

	actionConf, _ := EncodeAttributes(struct {
		Relative int
	}{
		Relative: 1,
	})

	for i := 0; i < deck.NumKeys(); i++ {
		blankPage.Keys[i] = KeyDefinition{
			Display: DynamicElement{
				Type:       "color",
				Attributes: displayConf,
			},
			Actions: []DynamicElement{
				{
					Type:       "page",
					Attributes: actionConf,
				},
			},
		}
	}

	conf.Pages["@@blank"] = blankPage
}
