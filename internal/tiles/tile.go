// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package tiles

import (
	"fmt"
	"log"
	"strings"

	"github.com/playbymail/ottomap/internal/compass"
	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/edges"
	"github.com/playbymail/ottomap/internal/domain"
	"github.com/playbymail/ottomap/internal/resources"
	"github.com/playbymail/ottomap/internal/terrain"
)

// Tile_t represents a tile on the map.
type Tile_t struct {
	// warning: we're changing from "location" to "coordinates" for tiles.
	// this is a breaking change so we're introducing a new field, Coordinates, to help.
	Coordinates coords.WorldMapCoord
	Location    coords.Map
	//Neighbors [direction.NumDirections]*Tile_t

	Visited string // set to the turn the tile was last visited
	Scouted string // set to the turn the tile was last scouted

	// permanent items in this tile
	Terrain terrain.Terrain_e
	Edges   [direction.NumDirections][]edges.Edge_e

	// transient items in this tile
	Encounters  []*domain.Encounter_t // other units in this tile
	Resources   []resources.Resource_e
	Settlements []*domain.Settlement_t
	Special     []*domain.Special_t

	// map of elements that are responsible for this tile.
	// does not work for fleets!
	SourcedBy map[string]bool
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
func (t *Tile_t) MergeReports(turnId string, report *domain.Report_t, worldMap *Map_t, specialNames map[string]*domain.Special_t, scouting, warnOnNewSettlement, warnOnTerrainChange bool) error {
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
	t.MergeTerrain(report.Terrain, warnOnTerrainChange)
	for _, border := range report.Borders {
		t.MergeBorder(report.UnitId, border, worldMap, warnOnTerrainChange)
		t.MergeEdge(border.Direction, border.Edge)
	}
	for _, encounter := range report.Encounters {
		t.MergeEncounter(encounter)
	}
	for _, fh := range report.FarHorizons {
		t.MergeFarHorizon(report.UnitId, fh, worldMap, warnOnTerrainChange)
	}
	for _, item := range report.Items {
		t.MergeItem(item)
	}
	for _, resource := range report.Resources {
		t.MergeResource(resource)
	}
	for _, settlement := range report.Settlements {
		t.MergeSettlement(settlement, specialNames, warnOnNewSettlement)
	}

	return nil
}

// MergeBorder merges a new border into the tile.
func (t *Tile_t) MergeBorder(unitId domain.UnitId_t, border *domain.Border_t, worldMap *Map_t, warnOnTerrainChange bool) {
	if border.Terrain == terrain.Blank {
		return
	}
	// create neighbor with terrain
	worldMapNeighbor := t.Coordinates.Move(border.Direction)
	locationNeighbor := t.Location.Add(border.Direction)
	neighbor := worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	neighbor.MergeTerrain(border.Terrain, warnOnTerrainChange)
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
func (t *Tile_t) MergeEncounter(e *domain.Encounter_t) {
	for _, l := range t.Encounters {
		if l.TurnId == e.TurnId && l.UnitId == e.UnitId {
			return
		}
	}
	t.Encounters = append(t.Encounters, e)
}

// MergeFarHorizon merges the far horizon from two tiles.
func (t *Tile_t) MergeFarHorizon(unitId domain.UnitId_t, fh *domain.FarHorizon_t, worldMap *Map_t, warnOnTerrainChange bool) {
	if fh == nil {
		return
	}
	// find the neighbor that this far horizon report is for
	var neighbor *Tile_t
	switch fh.Point {
	case compass.North:
		worldMapNeighbor := t.Coordinates.Move(direction.North, direction.North)
		locationNeighbor := t.Location.Move(direction.North, direction.North)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	case compass.NorthNorthEast:
		worldMapNeighbor := t.Coordinates.Move(direction.North, direction.NorthEast)
		locationNeighbor := t.Location.Move(direction.North, direction.NorthEast)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	case compass.NorthEast:
		worldMapNeighbor := t.Coordinates.Move(direction.NorthEast, direction.NorthEast)
		locationNeighbor := t.Location.Move(direction.NorthEast, direction.NorthEast)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	case compass.East:
		worldMapNeighbor := t.Coordinates.Move(direction.NorthEast, direction.SouthEast)
		locationNeighbor := t.Location.Move(direction.NorthEast, direction.SouthEast)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	case compass.SouthEast:
		worldMapNeighbor := t.Coordinates.Move(direction.SouthEast, direction.SouthEast)
		locationNeighbor := t.Location.Move(direction.SouthEast, direction.SouthEast)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	case compass.SouthSouthEast:
		worldMapNeighbor := t.Coordinates.Move(direction.South, direction.SouthEast)
		locationNeighbor := t.Location.Move(direction.South, direction.SouthEast)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	case compass.South:
		worldMapNeighbor := t.Coordinates.Move(direction.South, direction.South)
		locationNeighbor := t.Location.Move(direction.South, direction.South)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	case compass.SouthSouthWest:
		worldMapNeighbor := t.Coordinates.Move(direction.South, direction.SouthWest)
		locationNeighbor := t.Location.Move(direction.South, direction.SouthWest)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	case compass.SouthWest:
		worldMapNeighbor := t.Coordinates.Move(direction.SouthWest, direction.SouthWest)
		locationNeighbor := t.Location.Move(direction.SouthWest, direction.SouthWest)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	case compass.West:
		worldMapNeighbor := t.Coordinates.Move(direction.SouthWest, direction.NorthWest)
		locationNeighbor := t.Location.Move(direction.SouthWest, direction.NorthWest)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	case compass.NorthWest:
		worldMapNeighbor := t.Coordinates.Move(direction.NorthWest, direction.NorthWest)
		locationNeighbor := t.Location.Move(direction.NorthWest, direction.NorthWest)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	case compass.NorthNorthWest:
		worldMapNeighbor := t.Coordinates.Move(direction.North, direction.NorthWest)
		locationNeighbor := t.Location.Move(direction.North, direction.NorthWest)
		neighbor = worldMap.FetchTile(unitId, locationNeighbor, worldMapNeighbor)
	default:
		panic(fmt.Sprintf("assert(point != %d)", fh.Point))
	}
	neighbor.MergeTerrain(fh.Terrain, warnOnTerrainChange)
}

// MergeItem merges a new item into the tile.
func (t *Tile_t) MergeItem(f *domain.FoundItem_t) {
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
func (t *Tile_t) MergeSettlement(s *domain.Settlement_t, specialNames map[string]*domain.Special_t, warnOnNewSettlement bool) {
	if s == nil {
		return
	}
	//log.Printf("merge: settlement: testing %q\n", s.Name)
	//if specialNames != nil {
	//	log.Printf("merge: settlement: special names: %d\n", len(specialNames))
	//}
	if special, ok := specialNames[strings.ToLower(s.Name)]; ok {
		//log.Printf("merge: settlement %q: special %q\n", special.Id, special.Name)
		foundId := false
		for _, ss := range t.Special { // loop to prevent adding duplicates
			foundId = ss.Id == special.Id
			if foundId {
				break
			}
		}
		if !foundId {
			t.Special = append(t.Special, special)
			//log.Printf("merge: settlement %q: special %q: added to %q\n", special.Id, special.Name, t.Location)
			return
		}
	}
	for _, l := range t.Settlements {
		if strings.ToLower(l.Name) == strings.ToLower(s.Name) {
			return
		}
	}
	if warnOnNewSettlement {
		log.Printf("merge: settlement: %q\n", s.Name)
	}
	t.Settlements = append(t.Settlements, s)
}

// MergeTerrain if it is not blank and is different
func (t *Tile_t) MergeTerrain(n terrain.Terrain_e, warnOnTerrainChange bool) {
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

	// if the new terrain is unknown land and the existing terrain is any type of land,
	// then we want to keep the existing terrain and not report an error. likewise, if the new terrain is
	// any type of land and the existing terrain is unknown land, we want to use the new
	// terrain and not report an error. this only happens from fleet observations
	if n == terrain.UnknownLand && (t.Terrain.IsAnyLand()) {
		return
	} else if n.IsAnyLand() && t.Terrain == terrain.UnknownLand {
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

	// if the new terrain is unknown water and the existing terrain is any type of water,
	// then we want to keep the existing terrain and not report an error. likewise, if the new terrain is
	// any type of water and the existing terrain is unknown water, we want to use the new
	// terrain and not report an error. this only happens from fleet observations
	if n == terrain.UnknownWater && (t.Terrain.IsAnyWater()) {
		return
	} else if n.IsAnyWater() && t.Terrain == terrain.UnknownWater {
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
	if warnOnTerrainChange {
		log.Printf("%s: terrain changed from %-4q: to %q\n", t.Location.GridString(), t.Terrain, n)
	}

	t.Terrain = n
}

// Source adds an element to the source list for the tile.
func (t *Tile_t) Source(elements ...string) {
	if t.SourcedBy == nil {
		t.SourcedBy = map[string]bool{}
	}
	for _, element := range elements {
		t.SourcedBy[element] = true
	}
}
