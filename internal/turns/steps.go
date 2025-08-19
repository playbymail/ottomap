// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"fmt"
	"log"
	"strings"

	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/parser"
	"github.com/playbymail/ottomap/internal/results"
	"github.com/playbymail/ottomap/internal/tiles"
)

func errslug(text []byte, width int) string {
	var slug string
	if len(text) > width {
		slug = string(text[:width-3]) + "..."
	} else {
		slug = string(text)
	}
	return strings.ReplaceAll(fmt.Sprintf("%q", slug), "\\\\", "\\")
}

// Step processes a single step from a unit's move.
// It returns the final location of the unit.
func Step(turnId string, move *parser.Move_t, location, leader coords.Map, coordinates, leaderCoordinates coords.WorldMapCoord, worldMap *tiles.Map_t, specialNames map[string]*parser.Special_t, scouting, warnOnNewSettlement, warnOnTerrainChange, debug bool) (coords.Map, error) {
	// return an error if the starting location is obscured.
	if location.IsZero() {
		return location, fmt.Errorf("missing location")
	}

	// the results in the status line are always considered scouted
	if move.Result == results.StatusLine {
		scouting = true
	}

	// fetch the starting tile
	from := worldMap.FetchTile(move.UnitId, location, coordinates)
	if from == nil {
		// this should never happen because FetchTile will create the tile if it is missing.
		panic("missing tile")
	}

	var to *tiles.Tile_t
	var err error

	// order of the tests matters in this if-block
	if move.Still {
		if to, err = stepStill(turnId, move, from, worldMap, scouting, debug); err != nil {
			return location, err
		}
	} else if move.Follows != "" {
		if to, err = stepFollows(turnId, move, from, leader, leaderCoordinates, worldMap, scouting, debug); err != nil {
			return location, err
		}
	} else if move.GoesTo != "" {
		if to, err = stepGoto(turnId, move, from, move.GoesTo, worldMap, scouting, debug); err != nil {
			return location, err
		}
	} else if move.Result == results.Failed {
		if to, err = stepFailed(turnId, move, from, worldMap, scouting, debug); err != nil {
			return location, err
		}
	} else if move.Result == results.StatusLine {
		if to, err = stepStatus(turnId, move, from, worldMap, scouting, debug); err != nil {
			return location, err
		}
	} else if move.Result == results.Succeeded {
		if to, err = stepSucceeded(turnId, move, from, worldMap, scouting, debug); err != nil {
			return location, err
		}
	} else {
		log.Printf("error: unexpected result while parsing movement\n")
		log.Printf("error: turn %q\n", turnId)
		log.Printf("error: input: line %d\n", move.LineNo)
		log.Printf("error: input: text %s\n", errslug(move.Line, 58))
		log.Printf("error: move: step %d\n", move.StepNo)
		log.Printf("error: found result %q\n", move.Result)
		log.Printf("please report this error\n")
		panic(fmt.Sprintf("assert(result != %q)", move.Result))
	}
	if to == nil {
		panic("missing tile")
	}

	err = to.MergeReports(turnId, move.Report, worldMap, specialNames, scouting, warnOnNewSettlement, warnOnTerrainChange)

	// update the input so that the location represents the final location of the unit after the move
	move.Location = to.Location

	return to.Location, err
}

// stepFailed processes a single step from a unit's move.
// It returns the final location of the unit.
func stepFailed(turnId string, move *parser.Move_t, from *tiles.Tile_t, worldMap *tiles.Map_t, scouting, debug bool) (*tiles.Tile_t, error) {
	// stays in the current hex
	return from, nil
}

// stepFollows processes a single step from a unit's move.
// It returns the final location of the unit.
func stepFollows(turnId string, move *parser.Move_t, from *tiles.Tile_t, leader coords.Map, leaderCoordinates coords.WorldMapCoord, worldMap *tiles.Map_t, scouting, debug bool) (*tiles.Tile_t, error) {
	// the new hex will be the leader's location
	to := worldMap.FetchTile(move.UnitId, leader, leaderCoordinates)
	if to == nil {
		// this should never happen because FetchTile will create the tile if it is missing.
		panic("missing tile")
	}
	return to, nil
}

// stepGoto processes a single step from a unit's move.
// It returns the final location of the unit.
func stepGoto(turnId string, move *parser.Move_t, from *tiles.Tile_t, goesTo string, worldMap *tiles.Map_t, scouting, debug bool) (*tiles.Tile_t, error) {
	// unit is going to a specific location, so update the location to that location
	goesToCoordinates, err := coords.NewWorldMapCoord(goesTo)
	if err != nil {
		log.Printf("walk: %s: %d: %s: %v\n", turnId, move.LineNo, move.GoesTo, err)
		panic(err)
	}
	location, err := coords.HexToMap(goesTo)
	if err != nil {
		log.Printf("walk: %s: %d: %s: %v\n", turnId, move.LineNo, move.GoesTo, err)
		panic(err)
	}
	// update current hex based on the destination's location
	to := worldMap.FetchTile(move.UnitId, location, goesToCoordinates)
	if to == nil {
		// this should never happen because FetchTile will create the tile if it is missing.
		log.Printf("error: fetch tile: failed to create tile at %q\n", goesTo)
		panic("missing tile")
	}
	return to, nil
}

// stepStatus processes a single step from a unit's move.
// It returns the final location of the unit.
func stepStatus(turnId string, move *parser.Move_t, from *tiles.Tile_t, worldMap *tiles.Map_t, scouting, debug bool) (*tiles.Tile_t, error) {
	// status line always stays in the current hex
	return from, nil
}

// stepStill processes a single step from a unit's move.
// It returns the final location of the unit.
func stepStill(turnId string, move *parser.Move_t, from *tiles.Tile_t, worldMap *tiles.Map_t, scouting, debug bool) (*tiles.Tile_t, error) {
	// stays in the current hex
	return from, nil
}

// stepSucceeded processes a single step from a unit's move.
// It returns the final location of the unit.
func stepSucceeded(turnId string, move *parser.Move_t, from *tiles.Tile_t, worldMap *tiles.Map_t, scouting, debug bool) (*tiles.Tile_t, error) {
	// update current hex based on the direction
	var movesToCoordinates coords.WorldMapCoord
	movesFromCoordinates, err := coords.NewWorldMapCoord(from.Location.GridString())
	if err == nil && move.Advance != direction.Unknown {
		movesToCoordinates = movesFromCoordinates.MoveReverse(move.Advance)
	}
	to := worldMap.FetchTile(move.UnitId, from.Location.Add(move.Advance), movesToCoordinates)
	if to == nil {
		// this should never happen because FetchTile will create the tile if it is missing.
		panic("missing tile")
	}

	return to, nil
}
