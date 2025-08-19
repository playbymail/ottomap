// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package parser

import (
	"fmt"
	"sort"
	"strings"

	"github.com/playbymail/ottomap/internal/compass"
	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/edges"
	"github.com/playbymail/ottomap/internal/items"
	"github.com/playbymail/ottomap/internal/resources"
	"github.com/playbymail/ottomap/internal/results"
	"github.com/playbymail/ottomap/internal/terrain"
)

// These are the types returned from the parser and parsing functions.

// Turn_t represents a single turn identified by year and month.
type Turn_t struct {
	Id    string
	Year  int
	Month int

	// UnitMoves holds the units that moved in this turn
	UnitMoves            map[UnitId_t]*Moves_t
	SortedMoves          []*Moves_t
	MovesSortedByElement []*Moves_t

	// SpecialNames holds the names of the hexes that are special.
	// It's a hack to get around the fact that the parser doesn't know about the hexes.
	// They are added to the map when parsing and are forced to lower case.
	SpecialNames map[string]*Special_t

	Next, Prev *Turn_t
}

func (t *Turn_t) FromMayBeObscured() bool {
	return true
}

func (t *Turn_t) ToMayBeObscured() bool {
	return t.Id <= LastTurnCurrentLocationObscured
}

// TopoSortMoves sorts the moves in the turn in a way that guarantees that units that depend on other units will be sorted last.
func (t *Turn_t) TopoSortMoves() {
	sort.Slice(t.SortedMoves, func(i, j int) bool {
		a, b := t.SortedMoves[i], t.SortedMoves[j]

		// Determine the type of move for a
		aIsGoto, aIsNormal, aIsFollows := a.GoesTo != "", a.GoesTo == "" && a.Follows == "", a.Follows != ""

		// Determine the type of move for b
		bIsGoto, bIsNormal, bIsFollows := b.GoesTo != "", b.GoesTo == "" && b.Follows == "", b.Follows != ""

		// Goto moves sort before normal and follow moves
		if aIsGoto || bIsGoto {
			if !bIsGoto {
				return true
			} else if !aIsGoto {
				return false
			}
			return a.UnitId < b.UnitId
		}

		// Normal moves sort before follow moves
		if aIsNormal || bIsNormal {
			if bIsFollows {
				return true
			} else if aIsFollows {
				return false
			}
			return a.UnitId < b.UnitId
		}

		// Follow moves sort last
		if a.Follows < b.Follows {
			return true
		} else if a.Follows == b.Follows {
			return a.UnitId < b.UnitId
		}
		return false
	})
}

// SortMovesByElement sorts the moves in the turn by element (the unit id).
func (t *Turn_t) SortMovesByElement() {
	sort.Slice(t.MovesSortedByElement, func(i, j int) bool {
		return t.MovesSortedByElement[i].UnitId < t.MovesSortedByElement[j].UnitId
	})
}

// Moves_t represents the results for a unit that moves and reports in a turn.
// There will be one instance of this struct for each turn the unit moves in.
type Moves_t struct {
	TurnId string
	UnitId UnitId_t // unit that is moving

	// all the moves made this turn
	Moves   []*Move_t
	Follows UnitId_t
	GoesTo  string

	// all the scry results for this turn
	Scries []*Scry_t

	// Scouts are optional and move at the end of the turn
	Scouts []*Scout_t

	// FromHex is the hex the unit starts the move in.
	// This could be "N/A" if the unit was created this turn.
	// In that case, we will populate it when we know where the unit started.
	FromHex string

	// ToHex is the hex is unit ends the movement in.
	// This should always be set from the turn report.
	// It might be the same as the FromHex if the unit stays in place or fails to move.
	ToHex string

	Coordinates coords.WorldMapCoord // coordinates of the tile the unit ends the move in
	Location    coords.Map           // Location is the tile the unit ends the move in
}

// Move_t represents a single move by a unit.
// The move can be follows, goes to, stay in place, or attempt to advance a direction.
// The move will fail, succeed, or the unit can simply vanish without a trace.
type Move_t struct {
	UnitId UnitId_t // unit that is moving

	// the types of movement that a unit can make.
	Advance direction.Direction_e // set only if the unit is advancing
	Follows UnitId_t              // id of the unit being followed
	GoesTo  string                // hex teleporting to
	Still   bool                  // true if the unit is not moving (garrison) or a status entry

	// Result should be failed, succeeded, or vanished
	Result results.Result_e

	Report *Report_t // all observations made by the unit at the end of this move

	LineNo int
	StepNo int
	Line   []byte

	TurnId     string
	CurrentHex string

	// warning: we're changing from "location" to "coordinates" for tiles.
	// this is a breaking change so we're introducing new fields, FromCoordinates and ToCoordinates, to help.
	FromCoordinates coords.WorldMapCoord // the tile the unit starts the move in
	ToCoordinates   coords.WorldMapCoord // the tile the unit ends the move in

	// Location is the tile the unit ends the move in
	Location coords.Map // soon to be replaced with FromCoordinates and ToCoordinates

	// Debug settings
	Debug struct {
		FleetMoves bool
		PriorMove  *Move_t
		NextMove   *Move_t
	}
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

// mergeFarHorizons adds a new far horizon observation to the list if it's not already in the list
func (r *Report_t) mergeFarHorizons(fh FarHorizon_t) bool {
	for _, l := range r.FarHorizons {
		if l.Point == fh.Point && l.Terrain == fh.Terrain {
			return false
		}
	}
	r.FarHorizons = append(r.FarHorizons, &FarHorizon_t{Point: fh.Point, Terrain: fh.Terrain})
	return true
}

// mergeItems adds an item to the list. If it is already in the list, the quantity is updated.
func (r *Report_t) mergeItems(list []*FoundItem_t, f *FoundItem_t) []*FoundItem_t {
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

// DirectionTerrain_t is the first component returned from a successful step.
type DirectionTerrain_t struct {
	Direction direction.Direction_e
	Terrain   terrain.Terrain_e
}

func (d DirectionTerrain_t) String() string {
	return fmt.Sprintf("%s-%s", d.Direction, d.Terrain)
}

type Encounter_t struct {
	TurnId   string // turn the encounter happened
	UnitId   UnitId_t
	Friendly bool // true if the encounter was friendly
}

// Exhausted_t is returned when a step fails because the unit was exhausted.
type Exhausted_t struct {
	Direction direction.Direction_e
	Terrain   terrain.Terrain_e
}

func (e *Exhausted_t) String() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("x(%s-%s)", e.Direction, e.Terrain)
}

type FarHorizon_t struct {
	Point   compass.Point_e
	Terrain terrain.Terrain_e
}

// FoundItem_t represents items discovered by Scouts as they pass through a hex.
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

type FoundUnit_t struct {
	Id UnitId_t
}

type Patrolled_t struct {
	FoundUnits []FoundUnit_t
}

type Longhouse_t struct {
	Id       string
	Capacity int
}

// MissingEdge_t is returned for "No River Adjacent to Hex"
type MissingEdge_t struct {
	Direction direction.Direction_e
}

type NearHorizon_t struct {
	Point   direction.Direction_e
	Terrain terrain.Terrain_e
}

// Neighbor_t is the terrain in a neighboring hex that the unit from the current hex.
type Neighbor_t struct {
	Direction direction.Direction_e
	Terrain   terrain.Terrain_e
}

func (n *Neighbor_t) String() string {
	if n == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", n.Direction, n.Terrain)
}

// ProhibitedFrom_t is returned when a step fails because the unit is not allowed to enter the terrain.
type ProhibitedFrom_t struct {
	Direction direction.Direction_e
	Terrain   terrain.Terrain_e
}

func (p *ProhibitedFrom_t) String() string {
	if p == nil {
		return ""
	}
	return fmt.Sprintf("p(%s-%s)", p.Direction, p.Terrain)
}

// Scout_t represents a scout sent out by a unit.
type Scout_t struct {
	No     int // usually from 1..8
	TurnId string
	Moves  []*Move_t

	LineNo int
	Line   []byte
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

type Special_t struct {
	TurnId string // turn the special hex was observed
	Id     string // id of the special hex, full name converted to lower case
	Name   string // short name of the special hex (id if name is empty)
}

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
