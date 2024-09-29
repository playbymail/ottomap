// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package app

type Layout struct {
	Title       string
	Heading     string
	CurrentPage struct {
		Dashboard bool
		Maps      bool
		Reports   bool
		Calendar  bool
	}
	Content any
	Footer  Footer
}

type Footer struct {
	Copyright Copyright
	Version   string
}

type Copyright struct {
	Year  int
	Owner string
}
