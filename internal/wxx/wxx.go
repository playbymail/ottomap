// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"bytes"
	"fmt"
	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/terrain"
	"sort"
)

type WXX struct {
	buffer *bytes.Buffer

	tiles map[coords.Map]*Tile

	// terrainTileName maps our terrain type to the name of a Worldographer tile.
	terrainTileName map[terrain.Terrain_e]string

	// terrainTileSlot maps our terrain type to a slot in the Worldographer tile map.
	terrainTileSlot map[terrain.Terrain_e]int

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

	if w.terrainTileName == nil {
		w.terrainTileName = map[terrain.Terrain_e]string{}
		for k, v := range terrain.TileTerrainNames {
			w.terrainTileName[k] = v
		}
	}

	// set missing terrain names to "Floor Mexican Blue".
	// this is a hack to ensure that the Worldographer tile map is complete.
	// the choice is arbitrary but very noticeable.
	for k := range terrain.TileTerrainNames {
		if _, ok := w.terrainTileName[k]; !ok {
			w.terrainTileName[k] = "Floor Mexican Blue"
		}
	}

	// initialize tileNameList. this is the list of Worldographer terrain tile names that we output
	// to the map file, ordered by slot. The "Blank" tile must always be the first element in the list.
	if len(w.terrainTileName) != 0 {
		// names is a map of every terrain name. we must ensure that it includes "Blank".
		names := map[string]bool{"Blank": true}
		for _, v := range w.terrainTileName {
			names[v] = true
		}
		// copy the names into the list so that we can sort them and determine the slot number.
		for k := range names {
			w.tileNameList = append(w.tileNameList, k)
		}
		// sort the list, ensuring that "Blank" is the first element.
		sort.Slice(w.tileNameList, func(i, j int) bool {
			if w.tileNameList[i] == "Blank" {
				return true
			}
			return w.tileNameList[i] < w.tileNameList[j]
		})
		// tileNameList is now initialized and contains the list of Worldographer terrain tile names.
		// the location of the name in the list is the slot number we will use to render the tile.
	}
	if len(w.tileNameList) == 0 || w.tileNameList[0] != "Blank" {
		panic("assert(tileSlot[0] == Blank)")
	}
	//for slot, tileName := range w.tileNameList {
	//	log.Printf("debug: tile %-30q maps to slot %3d", tileName, slot)
	//}

	// initialize tileNameSlot. this is the map from the tile name to the slot number.
	tileNameSlot := map[string]int{}
	for n, v := range w.tileNameList {
		tileNameSlot[v] = n
	}

	// initialize terrainTileSlot. this is the map from the terrain type to the slot number.
	w.terrainTileSlot = map[terrain.Terrain_e]int{}
	for k, v := range w.terrainTileName {
		w.terrainTileSlot[k] = tileNameSlot[v]
	}
	var msgs []string
	for terrainCode, slot := range w.terrainTileSlot {
		tileName := w.tileNameList[slot]
		msgs = append(msgs, fmt.Sprintf("%-7s %-32s %4d", terrainCode, tileName, slot))
	}
	sort.Strings(msgs)
	// for debugging tiles when we start using tile templates from the user
	//log.Printf("terrain tile____________________________ slot\n")
	//for _, msg := range msgs {
	//	log.Printf("%s\n", msg)
	//}

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
