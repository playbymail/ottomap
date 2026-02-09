// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package parser

import (
	"fmt"

	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/domain"
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
type Turn_t = domain.Turn_t
type Moves_t = domain.Moves_t
type Move_t = domain.Move_t
type Scout_t = domain.Scout_t
type Scry_t = domain.Scry_t

// Constant alias for backward compatibility during migration.
const LastTurnCurrentLocationObscured = domain.LastTurnCurrentLocationObscured

// These are the parser-only types not shared with other packages.

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
