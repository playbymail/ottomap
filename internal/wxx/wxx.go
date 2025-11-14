// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"bytes"
	"github.com/playbymail/ottomap/internal/config"
	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/terrain"
)

type WXX struct {
	buffer *bytes.Buffer

	tiles map[coords.Map]*Tile

	// terrainTileSlot maps our terrain type to the name of a Worldographer tile.
	// this list is indexed by Terrain_e, with "Blank" at index 0.
	terrainTileSlot []string
}

func NewWXX(gcfg *config.Config, options ...Option) (*WXX, error) {
	w := &WXX{
		tiles:           map[coords.Map]*Tile{},
		terrainTileSlot: make([]string, terrain.NumberOfTerrainTypes, terrain.NumberOfTerrainTypes),
	}
	// this is a terrible hack - initialize the terrain tile names from the global configuration
	w.terrainTileSlot[terrain.Blank] = gcfg.Worldographer.Map.Terrain.Blank
	w.terrainTileSlot[terrain.HighMountainAlps] = gcfg.Worldographer.Map.Terrain.Alps
	w.terrainTileSlot[terrain.HillsArid] = gcfg.Worldographer.Map.Terrain.AridHills
	w.terrainTileSlot[terrain.FlatArid] = gcfg.Worldographer.Map.Terrain.AridTundra
	w.terrainTileSlot[terrain.FlatBrush] = gcfg.Worldographer.Map.Terrain.BrushFlat
	w.terrainTileSlot[terrain.HillsBrush] = gcfg.Worldographer.Map.Terrain.BrushHills
	w.terrainTileSlot[terrain.HillsConifer] = gcfg.Worldographer.Map.Terrain.ConiferHills
	w.terrainTileSlot[terrain.FlatDeciduous] = gcfg.Worldographer.Map.Terrain.Deciduous
	w.terrainTileSlot[terrain.HillsDeciduous] = gcfg.Worldographer.Map.Terrain.DeciduousHills
	w.terrainTileSlot[terrain.FlatDesert] = gcfg.Worldographer.Map.Terrain.Desert
	w.terrainTileSlot[terrain.HillsGrassy] = gcfg.Worldographer.Map.Terrain.GrassyHills
	w.terrainTileSlot[terrain.HillsGrassyPlateau] = gcfg.Worldographer.Map.Terrain.GrassyHillsPlateau
	w.terrainTileSlot[terrain.HighMountainsSnowy] = gcfg.Worldographer.Map.Terrain.HighSnowyMountains
	w.terrainTileSlot[terrain.FlatJungle] = gcfg.Worldographer.Map.Terrain.Jungle
	w.terrainTileSlot[terrain.HillsJungle] = gcfg.Worldographer.Map.Terrain.JungleHills
	w.terrainTileSlot[terrain.WaterLake] = gcfg.Worldographer.Map.Terrain.Lake
	w.terrainTileSlot[terrain.LowMountainsArid] = gcfg.Worldographer.Map.Terrain.LowAridMountains
	w.terrainTileSlot[terrain.LowMountainsConifer] = gcfg.Worldographer.Map.Terrain.LowConiferMountains
	w.terrainTileSlot[terrain.LowMountainsJungle] = gcfg.Worldographer.Map.Terrain.LowJungleMountains
	w.terrainTileSlot[terrain.LowMountainsSnowy] = gcfg.Worldographer.Map.Terrain.LowSnowyMountains
	w.terrainTileSlot[terrain.LowMountainsVolcanic] = gcfg.Worldographer.Map.Terrain.LowVolcanicMountains
	w.terrainTileSlot[terrain.WaterOcean] = gcfg.Worldographer.Map.Terrain.Ocean
	w.terrainTileSlot[terrain.FlatPolarIce] = gcfg.Worldographer.Map.Terrain.PolarIce
	w.terrainTileSlot[terrain.FlatPrairie] = gcfg.Worldographer.Map.Terrain.Prairie
	w.terrainTileSlot[terrain.FlatPrairiePlateau] = gcfg.Worldographer.Map.Terrain.PrairiePlateau
	w.terrainTileSlot[terrain.HillsRocky] = gcfg.Worldographer.Map.Terrain.RockyHills
	w.terrainTileSlot[terrain.HillsSnowy] = gcfg.Worldographer.Map.Terrain.SnowyHills
	w.terrainTileSlot[terrain.FlatSwamp] = gcfg.Worldographer.Map.Terrain.Swamp
	w.terrainTileSlot[terrain.FlatTundra] = gcfg.Worldographer.Map.Terrain.Tundra
	w.terrainTileSlot[terrain.UnknownJungleSwamp] = gcfg.Worldographer.Map.Terrain.UnknownJungleSwamp
	w.terrainTileSlot[terrain.UnknownLand] = gcfg.Worldographer.Map.Terrain.UnknownLand
	w.terrainTileSlot[terrain.UnknownMountain] = gcfg.Worldographer.Map.Terrain.UnknownMountain
	w.terrainTileSlot[terrain.UnknownWater] = gcfg.Worldographer.Map.Terrain.UnknownWater

	for _, option := range options {
		if err := option(w); err != nil {
			return nil, err
		}
	}

	//// for debugging tiles when we start using tile templates from the user
	//var msgs []string
	//for terrainCode, tileName := range w.terrainTileSlot {
	//	msgs = append(msgs, fmt.Sprintf("%-7s %4d %q", terrain.EnumToString[terrain.Terrain_e(terrainCode)], terrainCode, tileName))
	//}
	//sort.Strings(msgs)
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
