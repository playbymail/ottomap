// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package units

import (
	"encoding/json"
	"fmt"
)

// Type_e is an enum for the type of unit.
// Having Tribe as a unit type makes the unit code easier to understand.
type Type_e int

const (
	Unknown Type_e = iota
	Clan
	Tribe
	Courier
	Element
	Fleet
	Garrison
)

// MarshalJSON implements the json.Marshaler interface.
func (t Type_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(EnumToString[t])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Type_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *t, ok = StringToEnum[s]; !ok {
		return fmt.Errorf("invalid UnitType %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (t Type_e) String() string {
	if str, ok := EnumToString[t]; ok {
		return str
	}
	return fmt.Sprintf("UnitType(%d)", int(t))
}

var (
	// EnumToString is a helper map for marshalling the enum
	EnumToString = map[Type_e]string{
		Unknown:  "Unknown",
		Tribe:    "Tribe",
		Courier:  "Courier",
		Element:  "Element",
		Fleet:    "Fleet",
		Garrison: "Garrison",
	}
	// StringToEnum is a helper map for unmarshalling the enum
	StringToEnum = map[string]Type_e{
		"Unknown":  Unknown,
		"Tribe":    Tribe,
		"Courier":  Courier,
		"Element":  Element,
		"Fleet":    Fleet,
		"Garrison": Garrison,
	}
)
