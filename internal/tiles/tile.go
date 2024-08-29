// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tiles

import (
	"fmt"
	"github.com/playbymail/ottomap/internal/compass"
	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/edges"
	"github.com/playbymail/ottomap/internal/parser"
	"github.com/playbymail/ottomap/internal/resources"
	"github.com/playbymail/ottomap/internal/terrain"
	"log"
	"strings"
)

// Tile_t represents a tile on the map.
type Tile_t struct {
	Location coords.Map
	//Neighbors [direction.NumDirections]*Tile_t

	Visited string // set to the turn the tile was last visited
	Scouted string // set to the turn the tile was last scouted

	// permanent items in this tile
	Terrain terrain.Terrain_e
	Edges   [direction.NumDirections][]edges.Edge_e

	// transient items in this tile
	Encounters  []*parser.Encounter_t // other units in this tile
	Resources   []resources.Resource_e
	Settlements []*parser.Settlement_t
}

func (t *Tile_t) Dump() {
	var resource string
	for _, r := range t.Resources {
		resource = r.String()
	}
	var settlement string
	for _, s := range t.Settlements {
		settlement = s.Name
	}
	if resource == "" && settlement == "" {
		return
	}
	log.Printf("tile: %s %-3s %-7s %-7s . %-8s . %s\n", t.Location.GridString(), t.Terrain, t.Visited, t.Scouted, resource, settlement)
}

// MergeReports merges the reports from two tiles.
func (t *Tile_t) MergeReports(turnId string, report *parser.Report_t, worldMap *Map_t, scouting bool) error {
	// update flags for visited and scouted.
	// panic if the input is not sorted by turn.
	if !(t.Visited <= turnId) {
		panic(fmt.Sprintf("assert(%q <= %q)", t.Visited, turnId))
	}
	t.Visited = turnId
	if scouting {
		t.Scouted = turnId
	}

	// merge the reports from this move into the tile
	t.MergeTerrain(report.Terrain)
	for _, border := range report.Borders {
		t.MergeBorder(border, worldMap)
		t.MergeEdge(border.Direction, border.Edge)
	}
	for _, encounter := range report.Encounters {
		t.MergeEncounter(encounter)
	}
	for _, fh := range report.FarHorizons {
		t.MergeFarHorizon(fh, worldMap)
	}
	for _, item := range report.Items {
		t.MergeItem(item)
	}
	for _, resource := range report.Resources {
		t.MergeResource(resource)
	}
	for _, settlement := range report.Settlements {
		t.MergeSettlement(settlement)
	}

	return nil
}

// MergeBorder merges a new border into the tile.
func (t *Tile_t) MergeBorder(border *parser.Border_t, worldMap *Map_t) {
	if border.Terrain == terrain.Blank {
		return
	}
	// create neighbor with terrain
	neighbor := worldMap.FetchTile(t.Location.Add(border.Direction))
	neighbor.MergeTerrain(border.Terrain)
}

// MergeEdge merges a new edge into the tile.
func (t *Tile_t) MergeEdge(d direction.Direction_e, e edges.Edge_e) {
	if e == edges.None {
		return
	}
	for _, l := range t.Edges[d] {
		if l == e {
			return
		}
	}
	t.Edges[d] = append(t.Edges[d], e)
}

// MergeEncounter merges a new encounter into the tile.
func (t *Tile_t) MergeEncounter(e *parser.Encounter_t) {
	for _, l := range t.Encounters {
		if l.TurnId == e.TurnId && l.UnitId == e.UnitId {
			return
		}
	}
	t.Encounters = append(t.Encounters, e)
}

// MergeFarHorizon merges the far horizon from two tiles.
func (t *Tile_t) MergeFarHorizon(fh *parser.FarHorizon_t, worldMap *Map_t) {
	if fh == nil {
		return
	}
	// find the neighbor that this far horizon report is for
	var neighbor *Tile_t
	switch fh.Point {
	case compass.North:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.North, direction.North))
	case compass.NorthNorthEast:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.North, direction.NorthEast))
	case compass.NorthEast:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.NorthEast, direction.NorthEast))
	case compass.East:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.NorthEast, direction.SouthEast))
	case compass.SouthEast:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.SouthEast, direction.SouthEast))
	case compass.SouthSouthEast:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.South, direction.SouthEast))
	case compass.South:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.South, direction.South))
	case compass.SouthSouthWest:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.South, direction.SouthWest))
	case compass.SouthWest:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.SouthWest, direction.SouthWest))
	case compass.West:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.SouthWest, direction.NorthWest))
	case compass.NorthWest:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.NorthWest, direction.NorthWest))
	case compass.NorthNorthWest:
		neighbor = worldMap.FetchTile(t.Location.Move(direction.North, direction.NorthWest))
	default:
		panic(fmt.Sprintf("assert(point != %d)", fh.Point))
	}
	neighbor.MergeTerrain(fh.Terrain)
}

// MergeItem merges a new item into the tile.
func (t *Tile_t) MergeItem(f *parser.FoundItem_t) {
	// this is currently a no-op since we don't track items
}

// MergeResource merges a new resource into the tile.
func (t *Tile_t) MergeResource(r resources.Resource_e) {
	if r == resources.None {
		return
	}
	for _, l := range t.Resources {
		if l == r {
			return
		}
	}
	t.Resources = append(t.Resources, r)
}

// MergeSettlement merges a new settlement into the tile.
func (t *Tile_t) MergeSettlement(s *parser.Settlement_t) {
	if s == nil {
		return
	}
	log.Printf("merge: settlement: %q\n", s.Name)
	for _, l := range t.Settlements {
		if strings.ToLower(l.Name) == strings.ToLower(s.Name) {
			return
		}
	}
	t.Settlements = append(t.Settlements, s)
}

// MergeTerrain if it is not blank and is different
func (t *Tile_t) MergeTerrain(n terrain.Terrain_e) {
	// ignore the new terrain if it is blank or the same as the existing terrain
	if n == terrain.Blank || n == t.Terrain {
		return
	}
	// always accept if the current terrain is blank
	if t.Terrain == terrain.Blank {
		t.Terrain = n
		return
	}

	// if the new terrain is unknown jungle/swamp and the existing terrain is any type of jungle or swamp,
	// then we want to keep the existing terrain and not report an error. likewise, if the new terrain is
	// any type of jungle or swamp and the existing terrain is unknown jungle/swamp, we want to use the new
	// terrain and not report an error
	if n == terrain.UnknownJungleSwamp && (t.Terrain.IsJungle() || t.Terrain.IsSwamp()) {
		return
	} else if (n.IsJungle() || n.IsSwamp()) && t.Terrain == terrain.UnknownJungleSwamp {
		t.Terrain = n
		return
	}

	// if the new terrain is unknown mountain and the existing terrain is any type of mountain,
	// then we want to keep the existing terrain and not report an error. likewise, if the new terrain is
	// any type of mountain and the existing terrain is unknown mountain, we want to use the new
	// terrain and not report an error
	if n == terrain.UnknownMountain && t.Terrain.IsAnyMountain() {
		return
	} else if n.IsAnyMountain() && t.Terrain == terrain.UnknownMountain {
		t.Terrain = n
		return
	}

	// at this point, we know that t.Terrain != terrain.Blank.
	// we want to make sure that we don't overwrite the terrain with a fleet observation.
	isFleetObservation := n == terrain.UnknownLand || n == terrain.UnknownWater
	if isFleetObservation {
		return
	}

	// log any deltas
	log.Printf("useless tidbit: old terrain %-4q new terrain %q\n", t.Terrain, n)

	t.Terrain = n
}
