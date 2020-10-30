package main

func applySystemPages(conf *config) {
	blankPage := page{Keys: make(map[int]keyDefinition)}

	for i := 0; i < sd.NumKeys(); i++ {
		blankPage.Keys[i] = keyDefinition{
			Display: dynamicElement{
				Type: "color",
				Attributes: map[string]interface{}{
					"rgba": []interface{}{0x0, 0x0, 0x0, 0xff},
				},
			},
			Actions: []dynamicElement{
				{
					Type: "page",
					Attributes: map[string]interface{}{
						"relative": 1,
					},
				},
			},
		}
	}

	conf.Pages["@@blank"] = blankPage
}
