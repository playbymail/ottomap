// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domain

import (
	"fmt"
	"strings"

	"github.com/playbymail/ottomap/internal/compass"
	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/edges"
	"github.com/playbymail/ottomap/internal/items"
	"github.com/playbymail/ottomap/internal/resources"
	"github.com/playbymail/ottomap/internal/terrain"
)

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

// Border_t represents details about the hex border.
type Border_t struct {
	Direction direction.Direction_e
	// Edge is set if there is an edge feature like a river or pass
	Edge edges.Edge_e
	// Terrain is set if the neighbor is observable from this hex
	Terrain terrain.Terrain_e
}

func (b *Border_t) String() string {
	if b == nil {
		return "nil"
	}
	return fmt.Sprintf("(%s %s %s)", b.Direction, b.Edge, b.Terrain)
}

// FarHorizon_t represents a far horizon observation from a hex.
type FarHorizon_t struct {
	Point   compass.Point_e
	Terrain terrain.Terrain_e
}

// FoundItem_t represents items discovered by Scouts as they pass through a hex.
// Note: parsed item data is not propagated to the render pipeline (items are
// ignored during move processing). The items and winds packages are retained
// because external packages depend on these types.
type FoundItem_t struct {
	Quantity int
	Item     items.Item_e
}

func (f *FoundItem_t) String() string {
	if f == nil {
		return ""
	}
	return fmt.Sprintf("found(%d-%s)", f.Quantity, f.Item)
}

// Report_t represents the observations made by a unit.
// All reports are relative to the hex that the unit is reporting from.
type Report_t struct {
	UnitId UnitId_t // id of the unit that made the report

	Location      coords.Map
	TurnId        string // turn the report was received
	ScoutedTurnId string // turn the report was received from a scouting party

	// permanent items in this hex
	Terrain terrain.Terrain_e
	Borders []*Border_t

	// transient items in this hex
	Encounters  []*Encounter_t // other units in the hex
	Items       []*FoundItem_t
	Resources   []resources.Resource_e
	Settlements []*Settlement_t
	FarHorizons []*FarHorizon_t

	WasVisited bool // set to true if the location was visited by any unit
	WasScouted bool // set to true if the location was visited by a scouting party or a unit ended the turn here
}

// MergeBorders adds a new border to the list if it's not already in the list
func (r *Report_t) MergeBorders(b *Border_t) bool {
	for _, l := range r.Borders {
		if l.Direction == b.Direction && l.Edge == b.Edge && l.Terrain == b.Terrain {
			return false
		}
	}
	r.Borders = append(r.Borders, b)
	return true
}

// MergeEncounters adds a new encounter to the list if it's not already in the list
func (r *Report_t) MergeEncounters(e *Encounter_t) bool {
	for _, l := range r.Encounters {
		if l.TurnId == e.TurnId && l.UnitId == e.UnitId {
			return false
		}
	}
	r.Encounters = append(r.Encounters, e)
	return true
}

// MergeFarHorizons adds a new far horizon observation to the list if it's not already in the list
func (r *Report_t) MergeFarHorizons(fh FarHorizon_t) bool {
	for _, l := range r.FarHorizons {
		if l.Point == fh.Point && l.Terrain == fh.Terrain {
			return false
		}
	}
	r.FarHorizons = append(r.FarHorizons, &FarHorizon_t{Point: fh.Point, Terrain: fh.Terrain})
	return true
}

// MergeItems adds an item to the list. If it is already in the list, the quantity is updated.
func (r *Report_t) MergeItems(list []*FoundItem_t, f *FoundItem_t) []*FoundItem_t {
	if f == nil {
		return list
	} else if list == nil {
		return []*FoundItem_t{f}
	}
	for _, l := range list {
		if l.Item != f.Item {
			l.Quantity += f.Quantity
			return list
		}
	}
	return append(list, f)
}

// MergeResources adds a new resource to the list if it's not already in the list
func (r *Report_t) MergeResources(rs resources.Resource_e) bool {
	if rs == resources.None {
		return false
	}
	for _, l := range r.Resources {
		if l == rs {
			return false
		}
	}
	r.Resources = append(r.Resources, rs)
	return true
}

// MergeSettlements adds a new settlement to the list if it's not already in the list
func (r *Report_t) MergeSettlements(s *Settlement_t) bool {
	if s == nil {
		return false
	}
	for _, l := range r.Settlements {
		if strings.ToLower(l.Name) == strings.ToLower(s.Name) {
			return false
		}
	}
	r.Settlements = append(r.Settlements, s)
	return true
}
