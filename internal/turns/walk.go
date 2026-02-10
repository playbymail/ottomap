// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/domain"
	"github.com/playbymail/ottomap/internal/tiles"
)

func Walk(input []*domain.Turn_t, specialNames map[string]*domain.Special_t, originGrid string, quitOnInvalidGrid, warnOnInvalidGrid, warnOnNewSettlement, warnOnTerrainChange, debug bool, reverseWalker bool) (*tiles.Map_t, error) {
	started := time.Now()
	log.Printf("walk: input: %8d turns\n", len(input))

	// last seen is a map containing the last seen location for each unit
	lastSeen := map[domain.UnitId_t]coords.Map{}
	lastSeenAt := map[domain.UnitId_t]coords.WorldMapCoord{}

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
				if coordinates, err := coords.NewWorldMapCoord(unit.FromHex); err != nil {
					log.Printf("walk: %s: %s: %q: %v\n", turn.Id, unit.UnitId, unit.FromHex, err)
					panic(err)
				} else {
					unit.Coordinates, lastSeenAt[unit.UnitId] = coordinates, coordinates
				}
				if location, err := coords.HexToMap(unit.FromHex); err != nil {
					log.Printf("walk: %s: %s: %q: %v\n", turn.Id, unit.UnitId, unit.FromHex, err)
					panic(err)
				} else {
					unit.Location, lastSeen[unit.UnitId] = location, location
				}
				if reverseWalker {
					log.Printf("walk: turn %s unit %-8s goto %-8s follows %-8s %-8s -> %s :: %q\n", turn.Id, unit.UnitId, unit.GoesTo, unit.Follows, unit.FromHex, unit.Location, unit.Coordinates)
				}
			}
		}

		// update the locations of all units that have a prior location
		for _, unit := range turn.SortedMoves {
			if unit.Location.IsZero() {
				unit.Location = lastSeen[unit.UnitId]
				unit.Coordinates = lastSeenAt[unit.UnitId]
			}
		}

		// update the locations of all units created this turn
		// note: if the unit was created at the end of the parent's move, this will be wrong.
		// when that happens, the user must manually update the report to set the correct starting hex.
		for _, unit := range turn.SortedMoves {
			if unit.Location.IsZero() {
				// it should be an error if we can't derive it from the parent's location
				if parent, ok := lastSeenAt[unit.UnitId.Parent()]; !ok {
					log.Printf("%s: %-6s: parent %q: missing\n", unit.TurnId, unit.UnitId, unit.UnitId.Parent())
					log.Fatalf("error: expected unit to have parent\n")
				} else {
					//log.Printf("walk: turn %s unit %-8s goto %-8s follows %-8s %-8s -> %s\n", turn.Id, unit.Id, unit.GoesTo, unit.Follows, unit.FromHex, parent)
					unit.Coordinates, lastSeenAt[unit.UnitId] = parent, parent
				}
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
			//if unit == "0988" {
			//	log.Printf("walk: turn %s unit %-8s goto %-8s follows %-8s %-8s    %s\n", turn.Id, unit, moves.GoesTo, moves.Follows, moves.FromHex, moves.Location.GridString())
			//}
			// if we're missing the location, can we derive it from the previous turn?
			if moves.Location.IsZero() {
				panic("location is zero")
			}

			// set leader coordinates only if this is a follows move
			var leaderCoordinates coords.WorldMapCoord
			if moves.Follows != "" {
				if coordinates, ok := lastSeenAt[moves.Follows]; !ok {
					panic("!!!")
				} else {
					leaderCoordinates = coordinates
				}
			}
			var leader coords.Map // set only if this is a follows move
			if moves.Follows != "" {
				if location, ok := lastSeen[moves.Follows]; !ok {
					panic("!!!")
				} else {
					leader = location
				}
			}

			currentCoordinates := moves.Coordinates
			current := moves.Location

			// step through all the moves this unit makes this turn, tracking the location of the unit after each step
			for _, move := range moves.Moves {
				location, err := Step(turn.Id, move, current, leader, currentCoordinates, leaderCoordinates, worldMap, specialNames, false, warnOnNewSettlement, warnOnTerrainChange, debug)
				if err != nil {
					panic(err)
				}
				if move.Debug.FleetMoves {
					log.Printf("%s: %-6s: %d: step %d: result %q: to %q\n", turn.Id, unit, move.LineNo, move.StepNo, move.Result, location)
				}
				current = location
			}
			// log.Printf("walk: %s: %q: %q\n", moves.TurnId, moves.UnitId, current.GridString())
			if current.Column < 0 || current.Row < 0 {
				log.Printf("walk: %s: %q: %q (%d, %d)\n", moves.TurnId, moves.UnitId, current.GridString(), current.Column, current.Row)
				panic("!")
			}
			if strings.Contains(current.GridString(), "-") {
				log.Printf("walk: %s: %s: %s\n", moves.TurnId, moves.UnitId, current.GridString())
			}

			moves.Location, lastSeen[unit] = current, current

			// the unit's final location has been updated, so we can now send out the scouting parties
			for _, scout := range moves.Scouts {
				currentCoordinates = moves.Coordinates // start the scout with the unit's current coordinates
				// each scout will start in the unit's current location
				current = moves.Location
				// step through all the moves this scout makes this turn, tracking the location of the scout after each step
				for _, move := range scout.Moves {
					location, err := Step(turn.Id, move, current, leader, currentCoordinates, leaderCoordinates, worldMap, specialNames, true, warnOnNewSettlement, warnOnTerrainChange, debug)
					if err != nil {
						panic(err)
					}
					//log.Printf("%s: %-6s: %d: step %d: result %q: to %q\n", turn.Id, unit, move.LineNo, move.StepNo, move.Result, location)
					current = location
				}
			}

			// this is where we should stitch in the scry lines
			for _, scry := range moves.Scries {
				//debug = true
				fromCoordinates := scry.Coordinates // each move will start at the scry's coordinates
				fromHex := scry.Location            // each move will start in the scry's location
				for _, move := range scry.Moves {
					//log.Printf("%s: %-6s: %d: step %d: result %q: to %q\n", turn.Id, unit, move.LineNo, move.StepNo, move.Result, scry.Origin)
					toHex, err := Step(turn.Id, move, fromHex, leader, fromCoordinates, leaderCoordinates, worldMap, specialNames, false, warnOnNewSettlement, warnOnTerrainChange, debug)
					if err != nil {
						panic(err)
					}
					//log.Printf("scry: status: from %q: to %q\n", fromHex.ToHex(), toHex.ToHex())
					fromHex = toHex
				}
				fromHex = scry.Location // each scout will start in the scry's origin
				if scry.Scouts != nil {
					for _, move := range scry.Scouts.Moves {
						//log.Printf("%s: %-6s: %d: step %d: result %q: to %q\n", turn.Id, unit, move.LineNo, move.StepNo, move.Result, scry.Origin)
						toHex, err := Step(turn.Id, move, fromHex, leader, fromCoordinates, leaderCoordinates, worldMap, specialNames, true, warnOnNewSettlement, warnOnTerrainChange, debug)
						if err != nil {
							panic(err)
						}
						//log.Printf("scry: scout: from %q: to %q\n", fromHex.ToHex(), toHex.ToHex())
						fromHex = toHex
					}
				}
			}
		}
	}

	log.Printf("walk: %8d nodes: elapsed %v\n", len(input), time.Since(started))

	return worldMap, nil
}

// Step_t is one step of a walk.
type Step_t struct {
	ReportId string          // report that contains the step
	TurnId   string          // turn the step was taken
	UnitId   domain.UnitId_t // unit that took the step

	// From and To are the hexes that the step started and ended up in.
	// If the unit didn't move, then both will have the same value.
	From    coords.WorldMapCoord
	Advance direction.Direction_e // set to Unknown if the unit didn't move
	To      coords.WorldMapCoord

	// Status is everything that the unit sees in (or from) this hex.
	// Scouts will populate this every step; other units will populate only on the very last step of a move.
	Units  []string // other units in this hex
	Errors []error  // all errors from processing this step

	PrevStep, NextStep *Step_t // links to previous and next steps
}

// WalkTurn walks through the input in reverse (starting from the last step and working backwards through the move).
// It returns a slice containing all the steps taken in the turn.
//
// The steps for a unit may not be complete. In general, if we find an error we will stop walking the unit.
func WalkTurn(input *domain.Turn_t, specialNames map[string]*domain.Special_t, debug bool) ([]*Step_t, error) {
	started := time.Now()
	log.Printf("%s: walk: units %6d\n", input.Id, len(input.MovesSortedByElement))

	var steps []*Step_t
	for _, moves := range input.MovesSortedByElement {
		results, err := walkUnit(moves)
		if err != nil {
			log.Printf("error %q\n", err)
		} else if results != nil {
			steps = append(steps, results...)
		}
	}

	log.Printf("%s: walk: units %6d: steps %8d: elapsed %v\n", input.Id, len(input.MovesSortedByElement), len(steps), time.Since(started))
	if len(steps) == 0 {
		return nil, fmt.Errorf("no steps")
	}
	return steps, nil
}

func walkUnit(moves *domain.Moves_t) ([]*Step_t, error) {
	log.Printf("%s: walk: unit %-8s: from  %q  to  %q\n", moves.TurnId, moves.UnitId, moves.FromHex, moves.ToHex)
	from, err := coords.NewWorldMapCoord(moves.FromHex)
	if err != nil {
		return nil, err
	}
	to, err := coords.NewWorldMapCoord(moves.ToHex)
	if err != nil {
		return nil, err
	}
	log.Printf("%s: walk: unit %-8s: start %q: end %q\n", moves.TurnId, moves.UnitId, from, to)
	var steps []*Step_t
	for i := len(moves.Moves) - 1; i >= 0; i-- {
		move := moves.Moves[i]
		// did the unit move? they may have stayed still, been blocked by terrain, or exhausted their movement allowance
		if move.Advance == direction.Unknown {
			//log.Printf("%s: walk: unit %-8s: %4d: did not advance\n", moves.TurnId, moves.UnitId, i+1)
			from = to
		} else {
			from = to.MoveReverse(move.Advance)
		}
		//log.Printf("%s: walk: unit %-8s: %4d: %q -> %-2s -> %q\n", moves.TurnId, moves.UnitId, i+1, from, move.Advance, to)
		step := &Step_t{
			ReportId: "?",
			TurnId:   moves.TurnId,
			UnitId:   moves.UnitId,
			From:     from,
			Advance:  move.Advance,
			To:       to,
		}
		steps = append(steps, step)
		to = from
	}
	// put the steps into the right order and then link them. not sure that we need the links, but...
	slices.Reverse(steps)
	var prev *Step_t
	for _, cur := range steps {
		cur.PrevStep = prev
		if prev != nil {
			prev.NextStep = cur
		}
		prev = cur
	}
	for i, step := range steps {
		log.Printf("%s: walk: unit %-8s: %4d: %q -> %-2s -> %q\n", step.TurnId, moves.UnitId, i+1, step.From, step.Advance, step.To)
	}
	for _, scout := range moves.Scouts {
		if scout != nil {
			results, err := walkScout(domain.UnitId_t(fmt.Sprintf("%ss%d", moves.UnitId, scout.No)), from, scout)
			if err != nil {
				log.Printf("error %q\n", err)
			} else if results != nil {
				steps = append(steps, results...)
			}
		}
	}
	return steps, nil
}

func walkScout(scoutId domain.UnitId_t, from coords.WorldMapCoord, scout *domain.Scout_t) ([]*Step_t, error) {
	log.Printf("%s: walk: unit %-8s: start %q\n", scout.TurnId, scoutId, from)
	var steps []*Step_t
	for i, move := range scout.Moves {
		// did the unit move? they may have stayed still, been blocked by terrain, or exhausted their movement allowance
		var to coords.WorldMapCoord
		if move.Advance == direction.Unknown {
			// log.Printf("%s: walk: unit %-8s: %3d: did not advance\n", moves.TurnId, moves.UnitId, i+1)
			to = from
		} else {
			to = from.Move(move.Advance)
		}
		log.Printf("%s: walk: unit %-8s: %4d: %q -> %-2s -> %q\n", scout.TurnId, scoutId, i+1, from, move.Advance, to)
		step := &Step_t{
			ReportId: "?",
			TurnId:   scout.TurnId,
			UnitId:   scoutId,
			From:     from,
			To:       to,
		}
		steps = append(steps, step)
		from = to
	}
	// not sure that we need the links, but...
	var prev *Step_t
	for _, cur := range steps {
		cur.PrevStep = prev
		if prev != nil {
			prev.NextStep = cur
		}
		prev = cur
	}
	return steps, nil
}
