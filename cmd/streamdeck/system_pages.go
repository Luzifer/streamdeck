package main

func applySystemPages(conf *config) {
	blankPage := page{}
	conf.Pages["@@blank"] = blankPage
}
