// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package results

import (
	"encoding/json"
	"fmt"
)

type Result_e int

const (
	Unknown Result_e = iota
	Blocked
	ExhaustedMovementPoints
	Failed
	Followed
	Prohibited
	StatusLine
	StayedInPlace
	Succeeded
	Teleported
	Vanished
)

var (
	// EnumToString is a helper map for marshalling the enum
	EnumToString = map[Result_e]string{
		Unknown:                 "?",
		Blocked:                 "Blocked",
		ExhaustedMovementPoints: "Exhausted MPs",
		Failed:                  "Failed",
		Followed:                "Followed",
		Prohibited:              "Prohibited",
		StatusLine:              "Status Line",
		StayedInPlace:           "N/A",
		Succeeded:               "Succeeded",
		Teleported:              "Teleported",
		Vanished:                "Vanished",
	}
	// StringToEnum is a helper map for unmarshalling the enum
	StringToEnum = map[string]Result_e{
		"?":             Unknown,
		"Blocked":       Blocked,
		"Exhausted MPs": ExhaustedMovementPoints,
		"Failed":        Failed,
		"Follows":       Followed,
		"N/A":           StayedInPlace,
		"Prohibited":    Prohibited,
		"Status Line":   StatusLine,
		"Succeeded":     Succeeded,
		"Teleported":    Teleported,
		"Vanished":      Vanished,
	}
)

// MarshalJSON implements the json.Marshaler interface.
func (e Result_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(EnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Result_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *e, ok = StringToEnum[s]; !ok {
		return fmt.Errorf("invalid Result %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (e Result_e) String() string {
	if str, ok := EnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("Result(%d)", int(e))
}
