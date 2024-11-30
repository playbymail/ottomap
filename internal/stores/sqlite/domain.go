// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlite

type Report_t struct {
	ID    int
	Clan  int
	Year  int
	Month int
	Unit  string
	Hash  string
	Lines string
}
