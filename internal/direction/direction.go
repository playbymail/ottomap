// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package direction

import (
	"encoding/json"
	"fmt"
)

// Direction_e is an enum for the direction
type Direction_e int

const (
	Unknown Direction_e = iota
	North
	NorthEast
	SouthEast
	South
	SouthWest
	NorthWest
)
const (
	NumDirections = int(NorthWest) + 1
)

// Directions is a helper for iterating over the directions
var Directions = []Direction_e{
	North,
	NorthEast,
	SouthEast,
	South,
	SouthWest,
	NorthWest,
}

// MarshalJSON implements the json.Marshaler interface.
func (d Direction_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(EnumToString[d])
}

// MarshalText implements the encoding.TextMarshaler interface.
// This is needed for marshalling the enum as map keys.
//
// Note that this is called by the json package, unlike the UnmarshalText function.
func (d Direction_e) MarshalText() (text []byte, err error) {
	return []byte(EnumToString[d]), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Direction_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *d, ok = StringToEnum[s]; !ok {
		return fmt.Errorf("invalid Direction %q", s)
	}
	return nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
// This is needed for unmarshalling the enum as map keys.
//
// Note that this is never called; it just changes the code path in UnmarshalJSON.
func (d Direction_e) UnmarshalText(text []byte) error {
	panic("!")
}

// String implements the fmt.Stringer interface.
func (d Direction_e) String() string {
	if str, ok := EnumToString[d]; ok {
		return str
	}
	return fmt.Sprintf("Direction(%d)", int(d))
}

var (
	// EnumToString is a helper map for marshalling the enum
	EnumToString = map[Direction_e]string{
		Unknown:   "?",
		North:     "N",
		NorthEast: "NE",
		SouthEast: "SE",
		South:     "S",
		SouthWest: "SW",
		NorthWest: "NW",
	}
	// StringToEnum is a helper map for unmarshalling the enum
	StringToEnum = map[string]Direction_e{
		"?":  Unknown,
		"N":  North,
		"NE": NorthEast,
		"SE": SouthEast,
		"S":  South,
		"SW": SouthWest,
		"NW": NorthWest,
	}
)
