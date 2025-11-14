// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"fmt"
	"github.com/playbymail/ottomap/internal/terrain"
	"log"
)

// one grid on the Worldographer map is 30 columns wide by 21 rows high.
// one grid on the consolidated map is 30 columns wide by 21 rows high.
const columnsPerGrid, rowsPerGrid = 30, 21

type Grid struct {
	id    string // AA ... ZZ
	tiles [columnsPerGrid][rowsPerGrid]Tile
}

func gridRowColumnToId(row, column int) string {
	return fmt.Sprintf("%c%c", row+'A', column+'A')
}

func (w *WXX) newGrid(id string) *Grid {
	if len(id) != 2 {
		panic(fmt.Sprintf("assert(len(id) == 2)"))
	}

	// create a new grid with blank tiles and a default elevation of 1.
	grid := &Grid{id: id}
	for column := 0; column < columnsPerGrid; column++ {
		for row := 0; row < rowsPerGrid; row++ {
			grid.tiles[column][row].Elevation = 1
		}
	}

	return grid
}

//// createGrid creates the tiles for a single grid on the larger Tribenet map.
//// The caller is responsible for stitching the grids together in the final map.
//func (w *WXX) createGrid(id string, hexes []*Hex, showGridCoords, showGridNumbers bool) (*Grid, error) {
//	// create a new grid with blank tiles and a default elevation of 1.
//	grid := &Grid{id: id}
//	for column := 0; column < columnsPerGrid; column++ {
//		for row := 0; row < rowsPerGrid; row++ {
//			grid.tiles[column][row].Elevation = 1
//		}
//	}
//
//	// convert the grid hexes to tiles
//	for _, hex := range hexes {
//		col, row := hex.Coords.Column-1, hex.Coords.Row-1
//		grid.tiles[col][row].Terrain = hex.Terrain
//		// todo: add the missing terrain types here.
//		switch grid.tiles[col][row].Terrain {
//		case terrain.WaterLake:
//			grid.tiles[col][row].Elevation = -1
//		case terrain.WaterOcean:
//			grid.tiles[col][row].Elevation = -3
//		case terrain.FlatPrairie:
//			grid.tiles[col][row].Elevation = 1_000
//		}
//		grid.tiles[col][row].Features = hex.Features
//		if showGridCoords {
//			grid.tiles[col][row].Features.Coords = fmt.Sprintf("%s %02d%02d", hex.Grid, hex.Coords.Column, hex.Coords.Row)
//		} else if showGridNumbers {
//			grid.tiles[col][row].Features.Coords = fmt.Sprintf("%02d%02d", hex.Coords.Column, hex.Coords.Row)
//		}
//	}
//
//	return grid, nil
//}

func (g *Grid) addCoords() {
	for column := 0; column < columnsPerGrid; column++ {
		for row := 0; row < rowsPerGrid; row++ {
			if g.tiles[column][row].Terrain != terrain.Blank {
				g.tiles[column][row].Features.CoordsLabel = fmt.Sprintf("%s %02d%02d", g.id, column+1, row+1)
			}
		}
	}
}

func (g *Grid) addNumbers() {
	for column := 0; column < columnsPerGrid; column++ {
		for row := 0; row < rowsPerGrid; row++ {
			if g.tiles[column][row].Terrain != terrain.Blank {
				g.tiles[column][row].Features.NumbersLabel = fmt.Sprintf("%02d%02d", column+1, row+1)
			}
		}
	}
}

func (g *Grid) addTile(turnId string, hex *Hex) error {
	column, row := hex.Location.Column, hex.Location.Row

	tile := &g.tiles[column][row]
	tile.updated = turnId

	// does tile already exist in the grid?
	alreadyExists := tile.created != ""

	// if it does, verify that the terrain has not changed
	if alreadyExists && tile.Terrain != hex.Terrain {
		log.Printf("error: turn %q: tile %q\n", turnId, tile.Location.GridString())
		log.Printf("error: turn %q: hex  %q\n", turnId, hex.Location.GridString())
		panic("assert(tile.Terrain == hex.Terrain)")
	}

	// if it doesn't, set up the terrain
	if !alreadyExists {
		tile.created = turnId
		tile.Terrain = hex.Terrain
		tile.Elevation = 1
		switch tile.Terrain {
		case terrain.Blank, terrain.UnknownLand, terrain.UnknownWater:
			tile.Elevation = 0
		case terrain.HighMountainAlps,
			terrain.HillsArid,
			terrain.FlatArid,
			terrain.FlatBrush,
			terrain.HillsBrush,
			terrain.HillsConifer,
			terrain.FlatDeciduous,
			terrain.HillsDeciduous,
			terrain.FlatDesert,
			terrain.HillsGrassy,
			terrain.HillsGrassyPlateau,
			terrain.HighMountainsSnowy,
			terrain.FlatJungle,
			terrain.HillsJungle,
			terrain.LowMountainsArid,
			terrain.LowMountainsConifer,
			terrain.LowMountainsJungle,
			terrain.LowMountainsSnowy,
			terrain.LowMountainsVolcanic,
			terrain.FlatPrairie,
			terrain.FlatPrairiePlateau,
			terrain.HillsRocky,
			terrain.HillsSnowy,
			terrain.FlatTundra:
			tile.Elevation = 1_250
		case terrain.WaterLake:
			tile.Elevation = -1
		case terrain.WaterOcean:
			tile.Elevation = -3
		case terrain.FlatPolarIce:
			tile.Elevation = 10
		case terrain.FlatSwamp:
			tile.Elevation = 1
		default:
			log.Printf("grid: addTile: unknown terrain type %d %q", hex.Terrain, hex.Terrain.String())
			panic(fmt.Sprintf("assert(hex.Terrain != %d)", hex.Terrain))
		}
	}

	tile.WasScouted = tile.WasScouted || hex.WasScouted
	tile.WasVisited = tile.WasVisited || hex.WasVisited
	tile.Features = hex.Features

	return nil
}

func gridIdToRowColumn(id string) (row, column int) {
	if len(id) != 2 {
		log.Printf("error: invalid grid id %q\n", id)
		panic(fmt.Sprintf("assert(len(id) == 3)"))
	}
	row, column = int(id[0]-'A'), int(id[1]-'A')
	if row < 0 || row > 25 {
		log.Printf("error: invalid grid row %q\n", id)
		panic(fmt.Sprintf("assert(0 < row <= 25)"))
	} else if column < 0 || column > 25 {
		log.Printf("error: invalid grid column %q\n", id)
		panic(fmt.Sprintf("assert(0 < column <= 25)"))
	}
	return row, column
}
