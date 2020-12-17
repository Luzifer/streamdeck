package main

func applySystemPages(conf *config) {
	blankPage := page{Keys: make(map[int]keyDefinition)}

	for i := 0; i < sd.NumKeys(); i++ {
		blankPage.Keys[i] = keyDefinition{
			Display: dynamicElement{
				Type: "color",
				Attributes: attributeCollection{
					RGBA: []uint8{0x0, 0x0, 0x0, 0xff},
				},
			},
			Actions: []dynamicElement{
				{
					Type: "page",
					Attributes: attributeCollection{
						Relative: 1,
					},
				},
			},
		}
	}

	conf.Pages["@@blank"] = blankPage
}
