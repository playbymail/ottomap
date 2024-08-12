// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package compass

import (
	"encoding/json"
	"fmt"
)

// Point_e is an enum for the points two hexes away
type Point_e int

const (
	Unknown Point_e = iota
	North
	NorthNorthEast
	NorthEast
	East
	SouthEast
	SouthSouthEast
	South
	SouthSouthWest
	SouthWest
	West
	NorthWest
	NorthNorthWest
)

// MarshalJSON implements the json.Marshaler interface.
func (p Point_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(EnumToString[p])
}

// MarshalText implements the encoding.TextMarshaler interface.
// This is needed for marshalling the enum as map keys.
//
// Note that this is called by the json package, unlike the UnmarshalText function.
func (p Point_e) MarshalText() (text []byte, err error) {
	return []byte(EnumToString[p]), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (p *Point_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *p, ok = StringToEnum[s]; !ok {
		return fmt.Errorf("invalid CompassPoint %q", s)
	}
	return nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// This is needed for unmarshalling the enum as map keys.
//
// Note that this is never called; it just changes the code path in UnmarshalJSON.
func (p Point_e) UnmarshalText(text []byte) error {
	panic("!")
}

// String implements the fmt.Stringer interface.
func (p Point_e) String() string {
	if str, ok := EnumToString[p]; ok {
		return str
	}
	return fmt.Sprintf("Compass(%d)", int(p))
}

var (
	// EnumToString is a helper map for marshalling the enum
	EnumToString = map[Point_e]string{
		Unknown:        "?",
		North:          "North",
		NorthNorthEast: "NorthNorthEast",
		NorthEast:      "NorthEast",
		East:           "East",
		SouthEast:      "SouthEast",
		SouthSouthEast: "SouthSouthEast",
		South:          "South",
		SouthSouthWest: "SouthSouthWest",
		SouthWest:      "SouthWest",
		West:           "West",
		NorthWest:      "NorthWest",
		NorthNorthWest: "NorthNorthWest",
	}
	// StringToEnum is a helper map for unmarshalling the enum
	StringToEnum = map[string]Point_e{
		"?":              Unknown,
		"North":          North,
		"NorthNorthEast": NorthNorthEast,
		"NorthEast":      NorthEast,
		"East":           East,
		"SouthEast":      SouthEast,
		"SouthSouthEast": SouthSouthEast,
		"South":          South,
		"SouthSouthWest": SouthSouthWest,
		"SouthWest":      SouthWest,
		"West":           West,
		"NorthWest":      NorthWest,
		"NorthNorthWest": NorthNorthWest,
	}
)
