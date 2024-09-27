// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package trusted

type Page struct {
	LogoGrid LogoGrid
}

type LogoGrid struct {
	Logos []LogoGridDetail
}

type LogoGridDetail struct {
	Src    string
	Alt    string
	Width  int
	Height int
}
