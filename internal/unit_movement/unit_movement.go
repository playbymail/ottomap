// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package unit_movement

import (
	"encoding/json"
	"fmt"
)

type Type_e int

const (
	Unknown Type_e = iota
	Fleet
	Follows
	GoesTo
	Scouts
	Status
	Still
	Tribe
)

// MarshalJSON implements the json.Marshaler interface.
func (e Type_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(EnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Type_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *e, ok = StringToEnum[s]; !ok {
		return fmt.Errorf("invalid UnitMovement %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (e Type_e) String() string {
	if str, ok := EnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("UnitMovement(%d)", int(e))
}

var (
	// EnumToString is a helper map for marshalling the enum
	EnumToString = map[Type_e]string{
		Unknown: "N/A",
		Fleet:   "Fleet",
		Follows: "Follows",
		GoesTo:  "GoesTo",
		Scouts:  "Scout",
		Status:  "Status",
		Still:   "Still",
		Tribe:   "Tribe",
	}
	// StringToEnum is a helper map for unmarshalling the enum
	StringToEnum = map[string]Type_e{
		"N/A":     Unknown,
		"Fleet":   Fleet,
		"Follows": Follows,
		"GoesTo":  GoesTo,
		"Scout":   Scouts,
		"Status":  Status,
		"Still":   Still,
		"Tribe":   Tribe,
	}
)
