// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/parser"
	"github.com/playbymail/ottomap/internal/resources"
	"github.com/playbymail/ottomap/internal/terrain"
)

// Hex is a hex on the Tribenet map.
type Hex struct {
	Location   coords.Map // coordinates from the turn report
	RenderAt   coords.Map // shifted location to render tile at
	Terrain    terrain.Terrain_e
	WasScouted bool
	WasVisited bool
	Features   Features
}

func (h *Hex) Grid() string {
	return h.Location.GridId()
}

// Tile is a hex on the Worldographer map.
type Tile struct {
	created    string     // turn id when the tile was created
	updated    string     // turn id when the tile was updated
	Location   coords.Map // original grid coordinates
	RenderAt   coords.Map // shifted location to render tile at
	Terrain    terrain.Terrain_e
	Elevation  int
	IsIcy      bool
	IsGMOnly   bool
	Resources  Resources
	WasScouted bool
	WasVisited bool
	Features   Features
}

func newTile(location, renderAt coords.Map) *Tile {
	t := &Tile{
		Location: location,
		RenderAt: renderAt,
	}
	return t
}

func (t *Tile) addCoords() {
	if t.Terrain == terrain.Blank {
		return
	}
	t.Features.CoordsLabel = t.Location.GridString()
}

func (t *Tile) addNumbers() {
	if t.Terrain == terrain.Blank {
		return
	}
	t.Features.NumbersLabel = t.Location.GridString()[3:]
}

// Features are things to display on the map
type Features struct {
	Edges struct {
		Ford      []direction.Direction_e
		Pass      []direction.Direction_e
		River     []direction.Direction_e
		StoneRoad []direction.Direction_e
	}

	// set label for either Coords or Numbers, not both
	CoordsLabel  string
	NumbersLabel string

	IsOrigin    bool // true for the clan's origin hex
	Label       *Label
	Encounters  []*parser.Encounter_t // other units in this tile
	Resources   []resources.Resource_e
	Settlements []*parser.Settlement_t // name of settlement
}

type Resources struct {
	Animal int
	Brick  int
	Crops  int
	Gems   int
	Lumber int
	Metals int
	Rock   int
}

// Offset captures the layout.
// Are these one-based or zero-based?
type Offset struct {
	Column int
	Row    int
}
