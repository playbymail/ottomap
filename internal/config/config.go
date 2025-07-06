// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package config

import (
	"encoding/json"
	"errors"
	"github.com/playbymail/ottomap/cerrs"
	"log"
	"os"
	"reflect"
)

// Config allows each player to have their own configuration.
// In a future version, this would go into the shared database.
type Config struct {
	Clan          string          `json:"Clan,omitempty"`
	AllowConfig   bool            `json:"AllowConfig,omitempty"`
	Experimental  Experimental_t  `json:"Experimental"`
	Worldographer Worldographer_t `json:"Worldographer"`
}

type Experimental_t struct {
	AllowConfig bool `json:"AllowConfig,omitempty"`
}

type Worldographer_t struct {
	Map Map_t `json:"Map"`
}

type Map_t struct {
	Zoom    float64   `json:"Zoom"`
	Layers  Layers_t  `json:"Layers"`
	Terrain Terrain_t `json:"Terrain"`
	Units   Units_t   `json:"Units"`
}

type Layers_t struct {
	LargeCoords bool `json:"LargeCoords,omitempty"`
	MPCost      bool `json:"MPCost,omitempty"`
}

type Terrain_t struct {
	Blank                string `json:"Blank,omitempty"`
	Alps                 string `json:"Alps,omitempty"`
	AridHills            string `json:"AridHills,omitempty"`
	AridTundra           string `json:"AridTundra,omitempty"`
	BrushFlat            string `json:"BrushFlat,omitempty"`
	BrushHills           string `json:"BrushHills,omitempty"`
	ConiferHills         string `json:"ConiferHills,omitempty"`
	Deciduous            string `json:"Deciduous,omitempty"`
	DeciduousHills       string `json:"DeciduousHills,omitempty"`
	Desert               string `json:"Desert,omitempty"`
	GrassyHills          string `json:"GrassyHills,omitempty"`
	GrassyHillsPlateau   string `json:"GrassyHillsPlateau,omitempty"`
	HighSnowyMountains   string `json:"HighSnowyMountains,omitempty"`
	Jungle               string `json:"Jungle,omitempty"`
	JungleHills          string `json:"JungleHills,omitempty"`
	Lake                 string `json:"Lake,omitempty"`
	LowAridMountains     string `json:"LowAridMountains,omitempty"`
	LowConiferMountains  string `json:"LowConiferMountains,omitempty"`
	LowJungleMountains   string `json:"LowJungleMountains,omitempty"`
	LowSnowyMountains    string `json:"LowSnowyMountains,omitempty"`
	LowVolcanicMountains string `json:"LowVolcanicMountains,omitempty"`
	Ocean                string `json:"Ocean,omitempty"`
	PolarIce             string `json:"PolarIce,omitempty"`
	Prairie              string `json:"Prairie,omitempty"`
	PrairiePlateau       string `json:"PrairiePlateau,omitempty"`
	RockyHills           string `json:"RockyHills,omitempty"`
	SnowyHills           string `json:"SnowyHills,omitempty"`
	Swamp                string `json:"Swamp,omitempty"`
	Tundra               string `json:"Tundra,omitempty"`
	UnknownJungleSwamp   string `json:"UnknownJungleSwamp,omitempty"`
	UnknownLand          string `json:"UnknownLand,omitempty"`
	UnknownMountain      string `json:"UnknownMountain,omitempty"`
	UnknownWater         string `json:"UnknownWater,omitempty"`
}

type Units_t struct {
	Default  string `json:"Default,omitempty"`
	Clan     string `json:"Clan,omitempty"`
	Courier  string `json:"Courier,omitempty"`
	Element  string `json:"Element,omitempty"`
	Fleet    string `json:"Fleet,omitempty"`
	Multiple string `json:"Multiple,omitempty"`
	Garrison string `json:"Garrison,omitempty"`
	Tribe    string `json:"Tribe,omitempty"`
}

const (
	ErrIsDirectory = cerrs.Error("is directory")
	ErrIsNotAFile  = cerrs.Error("is not a file")
)

func Default() *Config {
	return &Config{
		Worldographer: Worldographer_t{
			Map: Map_t{
				Zoom: 1.0,
				Terrain: Terrain_t{
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
				},
				Units: Units_t{
					Default:  "Military Ancient Soldier",
					Clan:     "Military Ancient Soldier",
					Courier:  "Military Ancient Soldier",
					Fleet:    "Military Sailship",
					Garrison: "Military Ancient Soldier",
					Multiple: "Military Ancient Soldier",
					Tribe:    "Military Ancient Soldier",
					//Default:  "Military Ancient Soldier",
					//Clan:     "Building Palace",
					//Courier:  "Military Knight",
					//Fleet:    "Military Sailship",
					//Garrison: "Military Camp",
					//Multiple: "Military Ancient Soldier",
					//Tribe:    "Settlement Village",
				},
			},
		},
	}
}
func Load(name string, debug bool) (*Config, error) {
	// create a config with default values for the application
	cfg := Default()
	if sb, err := os.Stat(name); errors.Is(err, os.ErrNotExist) || os.IsNotExist(err) {
		if debug {
			log.Printf("[config] %q: %v\n", name, err)
		}
		return cfg, nil
	} else if sb.Mode().IsDir() {
		return cfg, ErrIsDirectory
	} else if !sb.Mode().IsRegular() {
		return cfg, ErrIsNotAFile
	}

	var tmp Config
	if data, err := os.ReadFile(name); err != nil {
		if debug {
			log.Printf("[config] %q: %v\n", name, err)
		}
		return cfg, nil
	} else if err = json.Unmarshal(data, &tmp); err != nil {
		if debug {
			log.Printf("[config] %q: %v\n", name, err)
		}
		return cfg, nil
	} else if debug {
		log.Printf("[config] %q: loaded %s\n", name, string(data))
	}

	// copy over every value from tmp to config that isn't the default (zero) value
	copyNonZeroFields(&tmp, cfg)

	return cfg, nil
}

// copyNonZeroFields recursively copies non-zero fields from src to dst using reflection
func copyNonZeroFields(src, dst interface{}) {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	// Dereference pointers
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}
	if dstVal.Kind() == reflect.Ptr {
		dstVal = dstVal.Elem()
	}

	// Only work with structs
	if srcVal.Kind() != reflect.Struct || dstVal.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		dstField := dstVal.Field(i)

		// Skip unexported fields
		if !srcField.CanInterface() || !dstField.CanSet() {
			continue
		}

		// Check if source field is zero value
		if srcField.IsZero() {
			continue
		}

		// Handle different field types
		switch srcField.Kind() {
		case reflect.Struct:
			// Recursively copy struct fields
			copyNonZeroFields(srcField.Interface(), dstField.Addr().Interface())
		default:
			// Copy primitive types and other values
			dstField.Set(srcField)
		}
	}
}
