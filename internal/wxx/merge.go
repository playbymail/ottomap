// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"fmt"
	"github.com/playbymail/ottomap/internal/terrain"
	"log"
)

// MergeHex merges the hex into the consolidated map, creating new grids and tiles as necessary.
// It returns the first error encountered merging the new hex.
func (w *WXX) MergeHex(hex *Hex) error {
	// create a new tile if necessary
	t, ok := w.tiles[hex.Location]
	if !ok {
		//log.Printf("wxx: merge: creating tile %s\n", hex.Location.GridString())
		t = newTile(hex.Location, hex.RenderAt)

		// set up the terrain
		t.Terrain = hex.Terrain
		t.Elevation = 1
		switch t.Terrain {
		case terrain.Blank, terrain.UnknownLand, terrain.UnknownWater:
			t.Elevation = 0
		case terrain.Alps,
			terrain.AridHills,
			terrain.AridTundra,
			terrain.BrushFlat,
			terrain.BrushHills,
			terrain.ConiferHills,
			terrain.Deciduous,
			terrain.DeciduousHills,
			terrain.Desert,
			terrain.GrassyHills,
			terrain.GrassyHillsPlateau,
			terrain.HighSnowyMountains,
			terrain.Jungle,
			terrain.JungleHills,
			terrain.LowAridMountains,
			terrain.LowConiferMountains,
			terrain.LowJungleMountains,
			terrain.LowSnowyMountains,
			terrain.LowVolcanicMountains,
			terrain.Prairie,
			terrain.PrairiePlateau,
			terrain.RockyHills,
			terrain.SnowyHills,
			terrain.Tundra:
			t.Elevation = 1_250
		case terrain.Lake:
			t.Elevation = -1
		case terrain.Ocean:
			t.Elevation = -3
		case terrain.PolarIce:
			t.Elevation = 10
		case terrain.Swamp:
			t.Elevation = 1
		default:
			log.Printf("grid: addTile: unknown terrain type %d %q", hex.Terrain, hex.Terrain.String())
			panic(fmt.Sprintf("assert(hex.Terrain != %d)", hex.Terrain))
		}

		w.tiles[hex.Location] = t
	}

	// verify that the terrain has not changed
	if t.Terrain != hex.Terrain {
		log.Printf("error: turn %q: tile %q\n", "?", t.Location.GridString())
		log.Printf("error: turn %q: hex  %q\n", "?", hex.Location.GridString())
		panic("assert(tile.Terrain == hex.Terrain)")
	}

	t.WasScouted = t.WasScouted || hex.WasScouted
	t.WasVisited = t.WasVisited || hex.WasVisited
	t.Features = hex.Features

	return nil
}
