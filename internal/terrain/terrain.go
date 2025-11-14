// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package terrain

import (
	"encoding/json"
	"fmt"
)

// reworking terrain to match the 10.5 Movement Costs table.
// Split into 4 categories: Flat, Hills, Low Mountains, and High Mountains
// Starting with Arid because of a bug reported with movement costs.

// Terrain_e is an enum for the terrain
type Terrain_e int

const (
	// Blank must be the first enum value or the map will not render
	Blank Terrain_e = iota
	HighMountainAlps
	HillsArid
	FlatArid
	FlatBrush
	HillsBrush
	HillsConifer
	FlatDeciduous
	HillsDeciduous
	FlatDesert
	HillsGrassy
	HillsGrassyPlateau
	HighMountainsSnowy
	FlatJungle
	HillsJungle
	WaterLake
	LowMountainsArid
	LowMountainsConifer
	LowMountainsJungle
	LowMountainsSnowy
	LowMountainsVolcanic
	WaterOcean
	FlatPolarIce
	FlatPrairie
	FlatPrairiePlateau
	HillsRocky
	HillsSnowy
	FlatSwamp
	FlatTundra
	UnknownJungleSwamp
	UnknownLand
	UnknownMountain
	UnknownWater
)

// NumberOfTerrainTypes must be updated if we add new terrain types
const NumberOfTerrainTypes = int(UnknownWater + 1)

func (e Terrain_e) IsAnyLand() bool {
	return e == HighMountainAlps || e == HillsArid || e == FlatArid ||
		e == FlatBrush || e == HillsBrush ||
		e == HillsConifer ||
		e == FlatDeciduous || e == HillsDeciduous || e == FlatDesert ||
		e == HillsGrassy || e == HillsGrassyPlateau ||
		e == HighMountainsSnowy ||
		e == FlatJungle || e == HillsJungle ||
		e == LowMountainsArid || e == LowMountainsConifer || e == LowMountainsJungle || e == LowMountainsSnowy || e == LowMountainsVolcanic ||
		e == FlatPolarIce || e == FlatPrairie || e == FlatPrairiePlateau ||
		e == HillsRocky ||
		e == HillsSnowy || e == FlatSwamp ||
		e == FlatTundra ||
		e == UnknownLand || e == UnknownMountain
}

func (e Terrain_e) IsAnyMountain() bool {
	return e == HighMountainAlps ||
		e == HighMountainsSnowy ||
		e == LowMountainsArid ||
		e == LowMountainsConifer ||
		e == LowMountainsJungle ||
		e == LowMountainsSnowy ||
		e == LowMountainsVolcanic
}

func (e Terrain_e) IsAnyWater() bool {
	return e == WaterLake || e == WaterOcean || e == UnknownWater
}

func (e Terrain_e) IsJungle() bool {
	return e == FlatJungle || e == HillsJungle
}

func (e Terrain_e) IsSwamp() bool {
	return e == FlatSwamp
}

func (e Terrain_e) MPCost() string {
	switch e {
	case Blank:
		return ""
	case FlatArid: // Wagons not allowed
		return "3"
	case FlatBrush:
		return "4"
	case FlatDeciduous:
		return "5"
	case FlatDesert:
		return "5"
	case FlatJungle:
		return "5"
	case FlatPolarIce:
		return "7"
	case FlatPrairie:
		return "3"
	case FlatPrairiePlateau:
		return "3"
	case FlatSwamp: // Wagons not allowed
		return "8W"
	case FlatTundra:
		return "4"
	case HighMountainAlps:
		return "∞P8"
	case HighMountainsSnowy:
		return "∞P8"
	case HillsArid:
		return "5"
	case HillsBrush:
		return "6"
	case HillsConifer:
		return "6"
	case HillsDeciduous:
		return "6"
	case HillsGrassy:
		return "5"
	case HillsGrassyPlateau:
		return "5"
	case HillsJungle: // Wagons not allowed
		return "6W"
	case HillsRocky:
		return "6"
	case HillsSnowy: // Wagons not allowed
		return "7W"
	case LowMountainsArid: // Wagons not allowed
		return "9WP7"
	case LowMountainsConifer: // Wagons not allowed
		return "10WP7"
	case LowMountainsJungle: // Wagons not allowed
		return "10WP7"
	case LowMountainsSnowy: // Wagons not allowed
		return "10WP7"
	case LowMountainsVolcanic: // Wagons not allowed
		return "10WP7"
	case WaterLake:
		return ""
	case WaterOcean:
		return ""
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
		HighMountainAlps:     "ALPS",
		HillsArid:            "AH",
		FlatArid:             "AR",
		FlatBrush:            "BF",
		HillsBrush:           "BH",
		HillsConifer:         "CH",
		FlatDeciduous:        "D",
		FlatDesert:           "DE",
		HillsDeciduous:       "DH",
		HillsGrassy:          "GH",
		HillsGrassyPlateau:   "GHP",
		HighMountainsSnowy:   "HSM",
		FlatJungle:           "JG",
		HillsJungle:          "JH",
		WaterLake:            "L",
		LowMountainsArid:     "LAM",
		LowMountainsConifer:  "LCM",
		LowMountainsJungle:   "LJM",
		LowMountainsSnowy:    "LSM",
		LowMountainsVolcanic: "LVM",
		WaterOcean:           "O",
		FlatPolarIce:         "PI",
		FlatPrairie:          "PR",
		FlatPrairiePlateau:   "PPR",
		HillsRocky:           "RH",
		HillsSnowy:           "SH",
		FlatSwamp:            "SW",
		FlatTundra:           "TU",
		UnknownJungleSwamp:   "UJS",
		UnknownLand:          "UL",
		UnknownMountain:      "UM",
		UnknownWater:         "UW",
	}
	// StringToEnum is a helper map for unmarshalling the enum
	StringToEnum = map[string]Terrain_e{
		"":     Blank,
		"ALPS": HighMountainAlps,
		"AH":   HillsArid,
		"AR":   FlatArid,
		"BF":   FlatBrush,
		"BH":   HillsBrush,
		"CH":   HillsConifer,
		"D":    FlatDeciduous,
		"DH":   HillsDeciduous,
		"DE":   FlatDesert,
		"GH":   HillsGrassy,
		"GHP":  HillsGrassyPlateau,
		"HSM":  HighMountainsSnowy,
		"JG":   FlatJungle,
		"JH":   HillsJungle,
		"L":    WaterLake,
		"LAM":  LowMountainsArid,
		"LCM":  LowMountainsConifer,
		"LJM":  LowMountainsJungle,
		"LSM":  LowMountainsSnowy,
		"LVM":  LowMountainsVolcanic,
		"O":    WaterOcean,
		"PI":   FlatPolarIce,
		"PPR":  FlatPrairiePlateau,
		"PR":   FlatPrairie,
		"RH":   HillsRocky,
		"SH":   HillsSnowy,
		"SW":   FlatSwamp,
		"TU":   FlatTundra,
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
		HighMountainAlps:     "Mountains",
		HillsArid:            "Hills",
		FlatArid:             "Flat Moss",
		FlatBrush:            "Flat Shrubland",
		HillsBrush:           "Hills Shrubland",
		HillsConifer:         "Hills Forest Evergreen",
		FlatDeciduous:        "Flat Forest Deciduous Heavy",
		HillsDeciduous:       "Hills Forest Deciduous",
		FlatDesert:           "Flat Desert Sandy",
		HillsGrassy:          "Hills Grassland",
		HillsGrassyPlateau:   "Hills Grassy",
		HighMountainsSnowy:   "Mountain Snowcapped",
		FlatJungle:           "Flat Forest Jungle Heavy",
		HillsJungle:          "Hills Forest Jungle",
		WaterLake:            "Water Shoals",
		LowMountainsArid:     "Mountains Dead Forest",
		LowMountainsConifer:  "Mountains Forest Evergreen",
		LowMountainsJungle:   "Mountain Forest Jungle",
		LowMountainsSnowy:    "Mountains Snowcapped",
		LowMountainsVolcanic: "Mountain Volcano Dormant",
		WaterOcean:           "Water Sea",
		FlatPolarIce:         "Mountains Glacier",
		FlatPrairie:          "Flat Grazing Land",
		FlatPrairiePlateau:   "Flat Grassland",
		HillsRocky:           "Underdark Broken Lands",
		HillsSnowy:           "Flat Snowfields",
		FlatSwamp:            "Flat Swamp",
		FlatTundra:           "Flat Steppe",
		UnknownJungleSwamp:   "Flat Forest Wetlands",
		UnknownLand:          "Flat Moss",
		UnknownMountain:      "Mountain Forest Mixed",
		UnknownWater:         "Water Reefs",
	}
)
