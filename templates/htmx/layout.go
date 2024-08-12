// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

type Layout struct {
	Site     Site
	Banner   Banner
	MainMenu MainMenu
	Sidebar  Sidebar
	Content  any
	Footer   Footer
}

type Site struct {
	Slug  string
	Title string
}

type Banner struct {
	Title string
	Slug  string
}

type MainMenu struct {
	Items    []MenuItem
	Releases Releases
}

type Releases struct {
	DT  Link
	DDs []Link
}
type Sidebar struct {
	LeftMenu  LeftMenu
	RightMenu RightMenu
	Notice    *Notice
}

type LeftMenu struct {
	Items []MenuItem
}

type RightMenu struct {
	Items []MenuItem
}

type Notice struct {
	Label string
	Lines []string
}

type MenuItem struct {
	Current  bool
	Class    string // optional, something like "sidemenu"
	Link     string
	Label    string
	Children []MenuItem
}

type Footer struct {
	Author        string
	CopyrightYear string
}

type Link struct {
	Label  string
	Link   string
	Target string
	Class  string
}
