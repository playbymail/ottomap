// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domain

// UnitId_t is a string type representing a unit identifier.
// Used by every package in the pipeline.
type UnitId_t string

func (u UnitId_t) InClan(clan UnitId_t) bool {
	if len(u) != 4 {
		return u.Parent().Parent() == clan
	}
	return u.Parent() == clan
}

func (u UnitId_t) IsFleet() bool {
	return len(u) == 6 && u[4] == 'f'
}

func (u UnitId_t) Parent() UnitId_t {
	if len(u) == 4 {
		return "0" + u[1:]
	}
	return u[:4]
}

func (u UnitId_t) String() string {
	return string(u)
}

// Encounter_t represents an encounter with another unit in a hex.
type Encounter_t struct {
	TurnId   string // turn the encounter happened
	UnitId   UnitId_t
	Friendly bool // true if the encounter was friendly
}

// Settlement_t is a settlement that the unit sees in the current hex.
type Settlement_t struct {
	TurnId string // turn the settlement was observed
	Name   string
}

func (s *Settlement_t) String() string {
	if s == nil {
		return ""
	}
	return s.Name
}

// Special_t represents a special hex observation.
type Special_t struct {
	TurnId string // turn the special hex was observed
	Id     string // id of the special hex, full name converted to lower case
	Name   string // short name of the special hex (id if name is empty)
}
