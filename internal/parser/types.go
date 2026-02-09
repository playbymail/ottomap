// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package parser

import (
	"fmt"
	"sort"

	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/domain"
	"github.com/playbymail/ottomap/internal/results"
	"github.com/playbymail/ottomap/internal/terrain"
)

// Type aliases for backward compatibility during migration.
// These allow downstream packages to continue using parser.UnitId_t, etc.
// without changes until they are migrated to import domain directly.
type UnitId_t = domain.UnitId_t
type Encounter_t = domain.Encounter_t
type Settlement_t = domain.Settlement_t
type Special_t = domain.Special_t
type Border_t = domain.Border_t
type FarHorizon_t = domain.FarHorizon_t
type FoundItem_t = domain.FoundItem_t
type Report_t = domain.Report_t

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


// DirectionTerrain_t is the first component returned from a successful step.
type DirectionTerrain_t struct {
	Direction direction.Direction_e
	Terrain   terrain.Terrain_e
}

func (d DirectionTerrain_t) String() string {
	return fmt.Sprintf("%s-%s", d.Direction, d.Terrain)
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
