// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package terrain

import (
	"encoding/json"
	"fmt"
)

// Terrain_e is an enum for the terrain
type Terrain_e int

const (
	// Blank must be the first enum value or the map will not render
	Blank Terrain_e = iota
	Alps
	AridHills
	AridTundra
	BrushFlat
	BrushHills
	ConiferHills
	Deciduous
	DeciduousHills
	Desert
	GrassyHills
	GrassyHillsPlateau
	HighSnowyMountains
	Jungle
	JungleHills
	Lake
	LowAridMountains
	LowConiferMountains
	LowJungleMountains
	LowSnowyMountains
	LowVolcanicMountains
	Ocean
	PolarIce
	Prairie
	PrairiePlateau
	RockyHills
	SnowyHills
	Swamp
	Tundra
	UnknownJungleSwamp
	UnknownLand
	UnknownMountain
	UnknownWater
)

// NumberOfTerrainTypes must be updated if we add new terrain types
const NumberOfTerrainTypes = int(UnknownWater + 1)

func (e Terrain_e) IsAnyLand() bool {
	return e == Alps || e == AridHills || e == AridTundra ||
		e == BrushFlat || e == BrushHills ||
		e == ConiferHills ||
		e == Deciduous || e == DeciduousHills || e == Desert ||
		e == GrassyHills || e == GrassyHillsPlateau ||
		e == HighSnowyMountains ||
		e == Jungle || e == JungleHills ||
		e == LowAridMountains || e == LowConiferMountains || e == LowJungleMountains || e == LowSnowyMountains || e == LowVolcanicMountains ||
		e == PolarIce || e == Prairie || e == PrairiePlateau ||
		e == RockyHills ||
		e == SnowyHills || e == Swamp ||
		e == Tundra ||
		e == UnknownLand || e == UnknownMountain
}

func (e Terrain_e) IsAnyMountain() bool {
	return e == Alps ||
		e == HighSnowyMountains ||
		e == LowAridMountains ||
		e == LowConiferMountains ||
		e == LowJungleMountains ||
		e == LowSnowyMountains ||
		e == LowVolcanicMountains
}

func (e Terrain_e) IsAnyWater() bool {
	return e == Lake || e == Ocean || e == UnknownWater
}

func (e Terrain_e) IsJungle() bool {
	return e == Jungle || e == JungleHills
}

func (e Terrain_e) IsSwamp() bool {
	return e == Swamp
}

func (e Terrain_e) MPCost() string {
	switch e {
	case Blank:
		return ""
	case Alps:
		return "∞"
	case AridHills:
		return "8"
	case AridTundra:
		return "8"
	case BrushFlat:
		return "8"
	case BrushHills:
		return "8"
	case ConiferHills:
		return "8"
	case Deciduous:
		return "8"
	case DeciduousHills:
		return "8"
	case Desert:
		return "8"
	case GrassyHills:
		return "8"
	case GrassyHillsPlateau:
		return "8"
	case HighSnowyMountains:
		return "8"
	case Jungle:
		return "8"
	case JungleHills:
		return "8"
	case Lake:
		return "8"
	case LowAridMountains:
		return "8"
	case LowConiferMountains:
		return "8"
	case LowJungleMountains:
		return "8"
	case LowSnowyMountains:
		return "8"
	case LowVolcanicMountains:
		return "8"
	case Ocean:
		return ""
	case PolarIce:
		return "8"
	case Prairie:
		return "8"
	case PrairiePlateau:
		return "8"
	case RockyHills:
		return "8"
	case SnowyHills:
		return "8"
	case Swamp:
		return "8"
	case Tundra:
		return "8"
	case UnknownJungleSwamp:
		return ""
	case UnknownLand:
		return ""
	case UnknownMountain:
		return ""
	case UnknownWater:
		return ""
	}
	panic(fmt.Sprintf("assert(terrain != %d)", e))
}

// MarshalJSON implements the json.Marshaler interface.
func (e Terrain_e) MarshalJSON() ([]byte, error) {
	return json.Marshal(EnumToString[e])
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (e *Terrain_e) UnmarshalJSON(data []byte) error {
	var s string
	var ok bool
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	} else if *e, ok = StringToEnum[s]; !ok {
		return fmt.Errorf("invalid Terrain %q", s)
	}
	return nil
}

// String implements the fmt.Stringer interface.
func (e Terrain_e) String() string {
	if str, ok := EnumToString[e]; ok {
		return str
	}
	return fmt.Sprintf("Terrain(%d)", int(e))
}

func StringToTerrain(s string) (Terrain_e, bool) {
	if e, ok := StringToEnum[s]; ok {
		return e, ok
	}
	return Blank, false
}

var (
	// EnumToString helper map for marshalling the enum
	EnumToString = map[Terrain_e]string{
		Blank:                "",
		Alps:                 "ALPS",
		AridHills:            "AH",
		AridTundra:           "AR",
		BrushFlat:            "BF",
		BrushHills:           "BH",
		ConiferHills:         "CH",
		Deciduous:            "D",
		Desert:               "DE",
		DeciduousHills:       "DH",
		GrassyHills:          "GH",
		GrassyHillsPlateau:   "GHP",
		HighSnowyMountains:   "HSM",
		Jungle:               "JG",
		JungleHills:          "JH",
		Lake:                 "L",
		LowAridMountains:     "LAM",
		LowConiferMountains:  "LCM",
		LowJungleMountains:   "LJM",
		LowSnowyMountains:    "LSM",
		LowVolcanicMountains: "LVM",
		Ocean:                "O",
		PolarIce:             "PI",
		Prairie:              "PR",
		PrairiePlateau:       "PPR",
		RockyHills:           "RH",
		SnowyHills:           "SH",
		Swamp:                "SW",
		Tundra:               "TU",
		UnknownJungleSwamp:   "UJS",
		UnknownLand:          "UL",
		UnknownMountain:      "UM",
		UnknownWater:         "UW",
	}
	// StringToEnum is a helper map for unmarshalling the enum
	StringToEnum = map[string]Terrain_e{
		"":     Blank,
		"ALPS": Alps,
		"AH":   AridHills,
		"AR":   AridTundra,
		"BF":   BrushFlat,
		"BH":   BrushHills,
		"CH":   ConiferHills,
		"D":    Deciduous,
		"DH":   DeciduousHills,
		"DE":   Desert,
		"GH":   GrassyHills,
		"GHP":  GrassyHillsPlateau,
		"HSM":  HighSnowyMountains,
		"JG":   Jungle,
		"JH":   JungleHills,
		"L":    Lake,
		"LAM":  LowAridMountains,
		"LCM":  LowConiferMountains,
		"LJM":  LowJungleMountains,
		"LSM":  LowSnowyMountains,
		"LVM":  LowVolcanicMountains,
		"O":    Ocean,
		"PI":   PolarIce,
		"PPR":  PrairiePlateau,
		"PR":   Prairie,
		"RH":   RockyHills,
		"SH":   SnowyHills,
		"SW":   Swamp,
		"TU":   Tundra,
		"UJS":  UnknownJungleSwamp,
		"UL":   UnknownLand,
		"UM":   UnknownMountain,
		"UW":   UnknownWater,
	}
	// TileTerrainNames is the map for tile terrain name matching. the text values
	// are extracted from the Worldographer tileset. they must match exactly.
	// if you're adding to this list, the values are found by hovering over the
	// terrain in the GUI.
	TileTerrainNames = map[Terrain_e]string{
		Blank:                "Blank",
		Alps:                 "Mountains",
		AridHills:            "Hills",
		AridTundra:           "Flat Moss",
		BrushFlat:            "Flat Shrubland",
		BrushHills:           "Hills Shrubland",
		ConiferHills:         "Hills Forest Evergreen",
		Deciduous:            "Flat Forest Deciduous Heavy",
		DeciduousHills:       "Hills Forest Deciduous",
		Desert:               "Flat Desert Sandy",
		GrassyHills:          "Hills Grassland",
		GrassyHillsPlateau:   "Hills Grassy",
		HighSnowyMountains:   "Mountain Snowcapped",
		Jungle:               "Flat Forest Jungle Heavy",
		JungleHills:          "Hills Forest Jungle",
		Lake:                 "Water Shoals",
		LowAridMountains:     "Mountains Dead Forest",
		LowConiferMountains:  "Mountains Forest Evergreen",
		LowJungleMountains:   "Mountain Forest Jungle",
		LowSnowyMountains:    "Mountains Snowcapped",
		LowVolcanicMountains: "Mountain Volcano Dormant",
		Ocean:                "Water Sea",
		PolarIce:             "Mountains Glacier",
		Prairie:              "Flat Grazing Land",
		PrairiePlateau:       "Flat Grassland",
		RockyHills:           "Underdark Broken Lands",
		SnowyHills:           "Flat Snowfields",
		Swamp:                "Flat Swamp",
		Tundra:               "Flat Steppe",
		UnknownJungleSwamp:   "Flat Forest Wetlands",
		UnknownLand:          "Flat Moss",
		UnknownMountain:      "Mountain Forest Mixed",
		UnknownWater:         "Water Reefs",
	}
)
