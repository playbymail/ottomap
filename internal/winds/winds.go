// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package winds

import (
	"encoding/json"
	"fmt"
)

type Strength_e int

const (
	Unknown Strength_e = iota
	Calm
	Mild
	Strong
	Gale
)

// MarshalJSON implements the json.Marshaler interface.
func (e Strength_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(EnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Strength_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *e, ok = StringToEnum[s]; !ok {
		return fmt.Errorf("invalid WindStrength %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (e Strength_e) String() string {
	if str, ok := EnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("WindStrength(%d)", int(e))
}

var (
	// EnumToString is a helper map for marshalling the enum
	EnumToString = map[Strength_e]string{
		Unknown: "N/A",
		Calm:    "CALM",
		Mild:    "MILD",
		Strong:  "STRONG",
		Gale:    "GALE",
	}
	// StringToEnum is a helper map for unmarshalling the enum
	StringToEnum = map[string]Strength_e{
		"N/A":    Unknown,
		"CALM":   Calm,
		"MILD":   Mild,
		"STRONG": Strong,
		"GALE":   Gale,
	}
)
