package main

func applySystemPages(conf *config) {
	blankKey := keyDefinition{
		Display: dynamicElement{Type: "color", Attributes: map[string]interface{}{"rgba": []interface{}{0x0, 0x0, 0x0, 0xff}}},
		Actions: []dynamicElement{{Type: "page", Attributes: map[string]interface{}{"name": conf.DefaultPage}}},
	}

	blankPage := page{}
	for len(blankPage.Keys) < sd.NumKeys() {
		blankPage.Keys = append(blankPage.Keys, blankKey)
	}
	conf.Pages["@@blank"] = blankPage
}
