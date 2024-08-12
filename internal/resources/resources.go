// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package resources

import (
	"encoding/json"
	"fmt"
)

// Resource_e is an enum for resources
type Resource_e int

const (
	None Resource_e = iota
	Coal
	CopperOre
	Diamond
	Frankincense
	Gold
	IronOre
	Jade
	Kaolin
	LeadOre
	Limestone
	NickelOre
	Pearls
	Pyrite
	Rubies
	Salt
	Silver
	Sulphur
	TinOre
	VanadiumOre
	ZincOre
)

var (
	// EnumToString is a helper map for marshalling the enum
	EnumToString = map[Resource_e]string{
		None:         "",
		Coal:         "Coal",
		CopperOre:    "Copper Ore",
		Diamond:      "Diamond",
		Frankincense: "Frankincense",
		Gold:         "Gold",
		IronOre:      "Iron Ore",
		Jade:         "Jade",
		Kaolin:       "Kaolin",
		LeadOre:      "Lead Ore",
		Limestone:    "Limestone",
		NickelOre:    "Nickel Ore",
		Pearls:       "Pearls",
		Pyrite:       "Pyrite",
		Rubies:       "Rubies",
		Salt:         "Salt",
		Silver:       "Silver",
		Sulphur:      "Sulphur",
		TinOre:       "Tin Ore",
		VanadiumOre:  "Vanadium Ore",
		ZincOre:      "Zinc Ore",
	}
	// StringToEnum is a helper map for unmarshalling the enum
	StringToEnum = map[string]Resource_e{
		"":             None,
		"Coal":         Coal,
		"Copper Ore":   CopperOre,
		"Diamond":      Diamond,
		"Frankincense": Frankincense,
		"Gold":         Gold,
		"Iron Ore":     IronOre,
		"Jade":         Jade,
		"Kaolin":       Kaolin,
		"Lead Ore":     LeadOre,
		"Limestone":    Limestone,
		"Nickel Ore":   NickelOre,
		"Pearls":       Pearls,
		"Pyrite":       Pyrite,
		"Rubies":       Rubies,
		"Salt":         Salt,
		"Silver":       Silver,
		"Sulphur":      Sulphur,
		"Tin Ore":      TinOre,
		"Vanadium Ore": VanadiumOre,
		"Zinc Ore":     ZincOre,
	}
)

// MarshalJSON implements the json.Marshaler interface.
func (e Resource_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(EnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Resource_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *e, ok = StringToEnum[s]; !ok {
		return fmt.Errorf("invalid Resource %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (e Resource_e) String() string {
	if str, ok := EnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("Resource(%d)", int(e))
}
