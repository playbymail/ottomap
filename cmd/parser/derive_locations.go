// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"

	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	schema "github.com/playbymail/ottomap/internal/tniif"
)

func DeriveLocations(doc *schema.Document) {
	deriveUnitMoveStepLocations(doc)
	deriveScoutLocations(doc)
	deriveObservationLocations(doc)
	deriveCompassPointLocations(doc)
}

func deriveUnitMoveStepLocations(doc *schema.Document) {
	for ci := range doc.Clans {
		clan := &doc.Clans[ci]
		for ui := range clan.Units {
			unit := &clan.Units[ui]
			if unit.EndingLocation == "" {
				continue
			}
			cur, err := parseCoord(unit.EndingLocation)
			if err != nil {
				doc.Notes = append(doc.Notes, schema.Note{
					Kind:    "warn",
					Message: fmt.Sprintf("unit %s: invalid ending location %q: %v", unit.ID, unit.EndingLocation, err),
				})
				continue
			}
			for mi := range unit.Moves {
				moves := &unit.Moves[mi]
				backwardWalkSteps(moves.Steps, cur)
			}
		}
	}
}

func deriveScoutLocations(doc *schema.Document) {
	for ci := range doc.Clans {
		clan := &doc.Clans[ci]
		for ui := range clan.Units {
			unit := &clan.Units[ui]
			if unit.EndingLocation == "" {
				continue
			}
			cur, err := parseCoord(unit.EndingLocation)
			if err != nil {
				doc.Notes = append(doc.Notes, schema.Note{
					Kind:    "warn",
					Message: fmt.Sprintf("unit %s: invalid ending location %q for scout derivation: %v", unit.ID, unit.EndingLocation, err),
				})
				continue
			}
			for si := range unit.Scouts {
				scout := &unit.Scouts[si]
				scout.StartingLocation = unit.EndingLocation
				forwardWalkSteps(scout.Steps, cur)
			}
		}
	}
}

func deriveObservationLocations(doc *schema.Document) {
	for ci := range doc.Clans {
		clan := &doc.Clans[ci]
		for ui := range clan.Units {
			unit := &clan.Units[ui]
			for mi := range unit.Moves {
				for si := range unit.Moves[mi].Steps {
					step := &unit.Moves[mi].Steps[si]
					if step.Observation != nil {
						step.Observation.Location = step.EndingLocation
					}
				}
			}
			for si := range unit.Scouts {
				for sti := range unit.Scouts[si].Steps {
					step := &unit.Scouts[si].Steps[sti]
					if step.Observation != nil {
						step.Observation.Location = step.EndingLocation
					}
				}
			}
		}
	}
}

func forwardWalkSteps(steps []schema.MoveStep, startLoc coords.Map) {
	cur := startLoc
	for i := range steps {
		step := &steps[i]
		switch {
		case step.Result != schema.ResultSucceeded:
			// failed, vanished, unknown — location doesn't change
		case step.Intent == schema.IntentStill:
			// still — location doesn't change
		case step.Intent == schema.IntentAdvance:
			d, err := schemaToDirection(step.Advance)
			if err != nil {
				panic(fmt.Sprintf("scout step %d: %v", i, err))
			}
			cur = cur.Add(d)
		case step.Intent == schema.IntentFollows || step.Intent == schema.IntentGoesTo:
			panic(fmt.Sprintf("scout step %d: %q is not allowed for scouts", i, step.Intent))
		}
		step.EndingLocation = formatCoord(cur)
	}
}

func backwardWalkSteps(steps []schema.MoveStep, endLoc coords.Map) {
	cur := endLoc
	for i := len(steps) - 1; i >= 0; i-- {
		step := &steps[i]
		step.EndingLocation = formatCoord(cur)

		if i == 0 {
			break
		}

		switch {
		case step.Result != schema.ResultSucceeded:
			// failed, vanished, unknown — location doesn't change
		case step.Intent == schema.IntentStill:
			// still — location doesn't change
		case step.Intent == schema.IntentAdvance:
			d, err := schemaToDirection(step.Advance)
			if err != nil {
				panic(fmt.Sprintf("unit step %d: %v", i, err))
			}
			cur = cur.Add(opposite(d))
		case step.Intent == schema.IntentFollows || step.Intent == schema.IntentGoesTo:
			panic(fmt.Sprintf("unit step %d: %q should never have a prior step", i, step.Intent))
		}
	}
}

func deriveCompassPointLocations(doc *schema.Document) {
	for ci := range doc.Clans {
		clan := &doc.Clans[ci]
		for ui := range clan.Units {
			unit := &clan.Units[ui]
			for mi := range unit.Moves {
				for si := range unit.Moves[mi].Steps {
					step := &unit.Moves[mi].Steps[si]
					if step.Observation != nil {
						deriveCompassPoints(step.Observation, doc)
					}
				}
			}
			for si := range unit.Scouts {
				for sti := range unit.Scouts[si].Steps {
					step := &unit.Scouts[si].Steps[sti]
					if step.Observation != nil {
						deriveCompassPoints(step.Observation, doc)
					}
				}
			}
		}
	}
}

func deriveCompassPoints(obs *schema.Observation, doc *schema.Document) {
	if len(obs.CompassPoints) == 0 || obs.Location == "" {
		return
	}
	m, err := parseCoord(obs.Location)
	if err != nil {
		return
	}
	for i := range obs.CompassPoints {
		cp := &obs.CompassPoints[i]
		dirs, ok := bearingToDirections[cp.Bearing]
		if !ok {
			doc.Notes = append(doc.Notes, schema.Note{
				Kind:    "warn",
				Message: fmt.Sprintf("compass point: unknown bearing %q at %s", cp.Bearing, obs.Location),
			})
			continue
		}
		loc := m.Move(dirs[0], dirs[1])
		if !validMap(loc) {
			doc.Notes = append(doc.Notes, schema.Note{
				Kind:    "warn",
				Message: fmt.Sprintf("compass point: bearing %q from %s produces out-of-bounds coordinate", cp.Bearing, obs.Location),
			})
			continue
		}
		cp.Location = formatCoord(loc)
	}
}

func validMap(m coords.Map) bool {
	return 0 <= m.Column && m.Column < 780 && 0 <= m.Row && m.Row < 546
}

// WARNING: These keys use long-form bearing names ("North", "NorthNorthEast", etc.)
// because convertBearing currently uses compass.EnumToString. When the parser is
// updated to emit short codes ("N", "NNE", etc.), this map must be updated to match.
var bearingToDirections = map[schema.Bearing][2]direction.Direction_e{
	"North":          {direction.North, direction.North},
	"NorthNorthEast": {direction.North, direction.NorthEast},
	"NorthEast":      {direction.NorthEast, direction.NorthEast},
	"East":           {direction.NorthEast, direction.SouthEast},
	"SouthEast":      {direction.SouthEast, direction.SouthEast},
	"SouthSouthEast": {direction.South, direction.SouthEast},
	"South":          {direction.South, direction.South},
	"SouthSouthWest": {direction.South, direction.SouthWest},
	"SouthWest":      {direction.SouthWest, direction.SouthWest},
	"West":           {direction.SouthWest, direction.NorthWest},
	"NorthWest":      {direction.NorthWest, direction.NorthWest},
	"NorthNorthWest": {direction.North, direction.NorthWest},
}

func parseCoord(c schema.Coordinates) (coords.Map, error) {
	return coords.HexToMap(string(c))
}

func formatCoord(m coords.Map) schema.Coordinates {
	return schema.Coordinates(m.ToHex())
}

func schemaToDirection(d schema.Direction) (direction.Direction_e, error) {
	if dir, ok := direction.StringToEnum[string(d)]; ok {
		return dir, nil
	}
	return direction.Unknown, fmt.Errorf("invalid direction %q", d)
}

func opposite(d direction.Direction_e) direction.Direction_e {
	switch d {
	case direction.North:
		return direction.South
	case direction.NorthEast:
		return direction.SouthWest
	case direction.SouthEast:
		return direction.NorthWest
	case direction.South:
		return direction.North
	case direction.SouthWest:
		return direction.NorthEast
	case direction.NorthWest:
		return direction.SouthEast
	default:
		panic(fmt.Sprintf("invalid direction %d", d))
	}
}
