// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tiles

import (
	"log"
	"sort"
	"strings"

	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/domain"
)

// Map_t represents a map of tiles.
type Map_t struct {
	// key is the grid location of the tile
	Tiles map[coords.Map]*Tile_t
}

// NewMap creates a new map.
func NewMap() *Map_t {
	return &Map_t{
		Tiles: map[coords.Map]*Tile_t{},
	}
}

// NewMap creates a new map.
func NewMapBounded(topLeft, bottomRight coords.Map) *Map_t {
	m := &Map_t{
		Tiles: map[coords.Map]*Tile_t{},
	}
	m.Tiles[topLeft] = &Tile_t{}
	m.Tiles[bottomRight] = &Tile_t{}
	return m
}

func (m *Map_t) Bounds() (upperLeft, lowerRight coords.Map) {
	if m.Length() == 0 {
		return coords.Map{}, coords.Map{}
	}

	for _, tile := range m.Tiles {
		if (tile.Visited != "" || tile.Scouted != "") && strings.Contains(tile.Location.GridString(), "-") {
			log.Printf("tile: %s: visited %q: scouted %q\n", tile.Location.GridString(), tile.Visited, tile.Scouted)
		}
		if upperLeft.Column == 0 {
			// assume that we're on the first tile
			upperLeft.Column, upperLeft.Row = tile.Location.Column, tile.Location.Row
			lowerRight.Column, lowerRight.Row = tile.Location.Column, tile.Location.Row
		}
		if tile.Location.Column < upperLeft.Column {
			upperLeft.Column = tile.Location.Column
		}
		if tile.Location.Row < upperLeft.Row {
			upperLeft.Row = tile.Location.Row
		}
		if lowerRight.Column < tile.Location.Column {
			lowerRight.Column = tile.Location.Column
		}
		if lowerRight.Row < tile.Location.Row {
			lowerRight.Row = tile.Location.Row
		}
	}

	return upperLeft, lowerRight
}

func (m *Map_t) Dump() {
	var sortedTiles []*Tile_t
	for _, tile := range m.Tiles {
		sortedTiles = append(sortedTiles, tile)
	}
	sort.Slice(sortedTiles, func(i, j int) bool {
		return sortedTiles[i].Location.GridString() < sortedTiles[j].Location.GridString()
	})
	for _, tile := range sortedTiles {
		tile.Dump()
	}
}

func (m *Map_t) Length() int {
	if m == nil {
		return 0
	}
	return len(m.Tiles)
}

// FetchTile returns the tile at the given location.
// If the tile does not exist, it is created.
func (m *Map_t) FetchTile(unitId domain.UnitId_t, location coords.Map, coordinates coords.WorldMapCoord) *Tile_t {
	if tile, ok := m.Tiles[location]; ok {
		return tile
	}

	// create a new tile to add to the map
	tile := &Tile_t{
		SourcedBy:   map[string]bool{},
		Coordinates: coordinates,
		Location:    location,
	}
	if unitId != "" {
		tile.SourcedBy[string(unitId)] = true
	}

	// todo: index this by the coordinates, not the location!
	// add the tile to the map
	m.Tiles[tile.Location] = tile

	// and return it
	return tile
}

// Solo returns a map of tiles that are sourced by the given elements.
func (m *Map_t) Solo(elements ...string) *Map_t {
	solo := NewMap()
	for _, tile := range m.Tiles {
		for _, element := range elements {
			if tile.SourcedBy[element] {
				// log.Printf("tile: %s: sourced by %q\n", tile.Location.GridString(), element)
				solo.Tiles[tile.Location] = tile
				break
			}
		}
	}
	return solo
}
