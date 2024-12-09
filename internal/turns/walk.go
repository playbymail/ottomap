// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/parser"
	"github.com/playbymail/ottomap/internal/tiles"
	"log"
	"strings"
	"time"
)

func Walk(input []*parser.Turn_t, specialNames map[string]*parser.Special_t, originGrid string, quitOnInvalidGrid, warnOnInvalidGrid, warnOnNewSettlement, warnOnTerrainChange, debug bool) (*tiles.Map_t, error) {
	started := time.Now()
	log.Printf("walk: input: %8d turns\n", len(input))

	// last seen is a map containing the last seen location for each unit
	lastSeen := map[parser.UnitId_t]coords.Map{}

	worldMap := tiles.NewMap()
	for _, turn := range input {
		// sanity check, these should always be the same value
		for _, moves := range turn.SortedMoves {
			if moves.GoesTo != "" && moves.ToHex != moves.GoesTo {
				unit := moves.UnitId
				log.Printf("turn %s: unit %-6s: current hex is %q\n", moves.TurnId, unit, moves.ToHex)
				log.Printf("turn %s: unit %-6s: goes to hex is %q\n", moves.TurnId, unit, moves.GoesTo)
				log.Fatalf("error: current hex != goes to hex\n")
			}
		}

		// leap of faith, update the location of all units that have a valid FromHex
		for _, unit := range turn.SortedMoves {
			if !strings.HasPrefix(unit.FromHex, "##") {
				if location, err := coords.HexToMap(unit.FromHex); err != nil {
					log.Printf("walk: %s: %s: %q: %v\n", turn.Id, unit.UnitId, unit.FromHex, err)
					panic(err)
				} else {
					unit.Location, lastSeen[unit.UnitId] = location, location
					//log.Printf("walk: turn %s unit %-8s goto %-8s follows %-8s %-8s -> %s\n", turn.Id, unit.Id, unit.GoesTo, unit.Follows, unit.FromHex, unit.Location)
				}
			}
		}

		// update the locations of all units that have a prior location
		for _, unit := range turn.SortedMoves {
			if unit.Location.IsZero() {
				unit.Location = lastSeen[unit.UnitId]
			}
		}

		// update the locations of all units created this turn
		// note: if the unit was created at the end of the parent's move, this will be wrong.
		// when that happens, the user must manually update the report to set the correct starting hex.
		for _, unit := range turn.SortedMoves {
			if unit.Location.IsZero() {
				// it should be an error if we can't derive it from the parent's location
				if parent, ok := lastSeen[unit.UnitId.Parent()]; !ok {
					log.Printf("%s: %-6s: parent %q: missing\n", unit.TurnId, unit.UnitId, unit.UnitId.Parent())
					log.Fatalf("error: expected unit to have parent\n")
				} else {
					//log.Printf("walk: turn %s unit %-8s goto %-8s follows %-8s %-8s -> %s\n", turn.Id, unit.Id, unit.GoesTo, unit.Follows, unit.FromHex, parent)
					unit.Location, lastSeen[unit.UnitId] = parent, parent
				}
			}
		}

		turn.TopoSortMoves()

		// walk the moves for all the units in this turn
		for _, moves := range turn.SortedMoves {
			unit := moves.UnitId
			//log.Printf("walk: turn %s unit %-8s goto %-8s follows %-8s %-8s    %s\n", turn.Id, unit, moves.GoesTo, moves.Follows, moves.FromHex, moves.Location.GridString())
			// if we're missing the location, can we derive it from the previous turn?
			if moves.Location.IsZero() {
				panic("location is zero")
			}

			var leader coords.Map // set only if this is a follows move
			if moves.Follows != "" {
				if location, ok := lastSeen[moves.Follows]; !ok {
					panic("!!!")
				} else {
					leader = location
				}
			}

			current := moves.Location

			// step through all the moves this unit makes this turn, tracking the location of the unit after each step
			for _, move := range moves.Moves {
				location, err := Step(turn.Id, move, current, leader, worldMap, specialNames, false, warnOnNewSettlement, warnOnTerrainChange, debug)
				if err != nil {
					panic(err)
				}
				if move.Debug.FleetMoves {
					log.Printf("%s: %-6s: %d: step %d: result %q: to %q\n", turn.Id, unit, move.LineNo, move.StepNo, move.Result, location)
				}
				current = location
			}
			if strings.Contains(current.GridString(), "-") {
				log.Printf("walk: %s: %s: %s\n", moves.TurnId, moves.UnitId, current.GridString())
			}

			moves.Location, lastSeen[unit] = current, current

			// the unit's final location has been updated, so we can now send out the scouting parties
			for _, scout := range moves.Scouts {
				// each scout will start in the unit's current location
				current = moves.Location
				// step through all the moves this scout makes this turn, tracking the location of the scout after each step
				for _, move := range scout.Moves {
					location, err := Step(turn.Id, move, current, leader, worldMap, specialNames, true, warnOnNewSettlement, warnOnTerrainChange, debug)
					if err != nil {
						panic(err)
					}
					//log.Printf("%s: %-6s: %d: step %d: result %q: to %q\n", turn.Id, unit, move.LineNo, move.StepNo, move.Result, location)
					current = location
				}
			}
		}
	}

	log.Printf("walk: %8d nodes: elapsed %v\n", len(input), time.Since(started))

	return worldMap, nil
}
