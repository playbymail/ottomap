// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"bytes"
	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/terrain"
	"sort"
)

type WXX struct {
	buffer *bytes.Buffer

	tiles map[coords.Map]*Tile

	// terrainTileNames maps our terrain type to the name of a Worldographer tile.
	terrainTileNames map[terrain.Terrain_e]string
	// terrainTileSlot maps our terrain type to a slot in the Worldographer tile map.
	terrainTileSlot map[terrain.Terrain_e]int

	// tileNameSlot maps the tile name to a slot in the Worldographer map.
	tileNameSlot map[string]int
	// tileNameList is the list of Worldographer terrain tile names that we output.
	// this list is sorted, with "Blank" at index 0.
	tileNameList []string
}

func NewWXX(options ...Option) (*WXX, error) {
	w := &WXX{
		tiles: map[coords.Map]*Tile{},
	}

	for _, option := range options {
		if err := option(w); err != nil {
			return nil, err
		}
	}

	if w.terrainTileNames == nil {
		w.terrainTileNames = terrain.TileTerrainNames
	}

	// collect the terrain types into a list, sort them, and ensure that "Blank" is the first element.
	w.tileNameSlot = map[string]int{"Blank": 0}
	for k, v := range w.terrainTileNames {
		if k == terrain.Blank {
			continue
		}
		// if we haven't seen this terrain type before, add it to the list
		if _, ok := w.tileNameSlot[v]; !ok {
			w.tileNameSlot[v] = 0
			w.tileNameList = append(w.tileNameList, v)
		}
	}
	sort.Strings(w.tileNameList)
	w.tileNameList = append([]string{"Blank"}, w.tileNameList...)
	for i, v := range w.tileNameList {
		w.tileNameSlot[v] = i
	}
	w.terrainTileSlot = map[terrain.Terrain_e]int{}
	for k, v := range w.terrainTileNames {
		w.terrainTileSlot[k] = w.tileNameSlot[v]
	}

	return w, nil
}

// GetTile returns the tile at the given coordinates.
func (w *WXX) GetTile(location coords.Map) *Tile {
	t, ok := w.tiles[location]
	if !ok {
		panic("tile not defined")
	}
	return t
}

type Option func(*WXX) error
