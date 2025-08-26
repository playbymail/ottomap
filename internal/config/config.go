// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/playbymail/ottomap/cerrs"
)

// Config allows each player to have their own configuration.
// In a future version, this would go into the shared database.
type Config struct {
	Clan          string          `json:"Clan,omitempty"`
	AllowConfig   bool            `json:"AllowConfig,omitempty"`
	DebugFlags    DebugFlags_t    `json:"DebugFlags"`
	Experimental  Experimental_t  `json:"Experimental"`
	Parser        Parser_t        `json:"Parser"`
	Worldographer Worldographer_t `json:"Worldographer"`
}

type DebugFlags_t struct {
	DumpAllTiles       bool `json:"DumpAllTiles,omitempty"`
	DumpAllTurns       bool `json:"DumpAllTurns,omitempty"`
	DumpBorderCounts   bool `json:"DumpBorderCounts,omitempty"`
	DumpDefaultTileMap bool `json:"DumpDefaultTileMap,omitempty"`
	FleetMovement      bool `json:"FleetMovement,omitempty"`
	IgnoreScouts       bool `json:"IgnoreScouts,omitempty"`
	LogFile            bool `json:"LogFile,omitempty"`
	LogTime            bool `json:"LogTime,omitempty"`
	Maps               bool `json:"Maps,omitempty"`
	Nodes              bool `json:"Nodes,omitempty"`
	Parser             bool `json:"Parser,omitempty"`
	Sections           bool `json:"Sections,omitempty"`
	Steps              bool `json:"Steps,omitempty"`
}

type Experimental_t struct {
	CleanupScoutStill  bool `json:"CleanupScoutStill,omitempty"`
	ReverseWalker      bool `json:"ReverseWalker,omitempty"`
	SplitTrailingUnits bool `json:"SplitTrailingUnits,omitempty"`
	StripCR            bool `json:"StripCR,omitempty"`
}

type Worldographer_t struct {
	Map    Map_t       `json:"Map"`
	Render Render_t    `json:"Render"`
	Mentee []*Mentee_t `json:"Mentee,omitempty"`
	Solo   []*Solo_t   `json:"Solo,omitempty"`
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

type Parser_t struct {
	AcceptLoneDash      bool `json:"AcceptLoneDash,omitempty"`
	CheckObscuredGrids  bool `json:"CheckObscuredGrids,omitempty"`
	QuitOnObscuredGrids bool `json:"QuitOnObscuredGrids,omitempty"`
}

type Render_t struct {
	FordsAsGaps bool `json:"FordsAsGaps,omitempty"`
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

type Mentee_t struct {
	Unit      string `json:"Unit"`  // unit being mentored
	StartTurn string `json:"Start"` // turn id to start mentoring
	StopTurn  string `json:"Stop"`  // turn id to stop mentoring
}

type Solo_t struct {
	Unit      string `json:"Unit"`  // unit to create solo map for
	StartTurn string `json:"Start"` // turn id to start soloing
	StopTurn  string `json:"Stop"`  // turn id to stop soloing
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
	if debug {
		log.Printf("[config] %q: loading configuration...\n", name)
	}
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
		if nice, err := json.MarshalIndent(tmp, "", "  "); err == nil {
			log.Printf("[config] %s\n", nice)
		} else {
			log.Printf("[config] %q: loaded %s\n", name, string(data))
		}
	}
	// validate some stuff
	if tmp.Parser.QuitOnObscuredGrids {
		tmp.Parser.CheckObscuredGrids = true
	}
	for _, v := range tmp.Worldographer.Mentee {
		if v.Unit == "" || strings.TrimSpace(v.Unit) != v.Unit {
			return nil, fmt.Errorf("mentee: invalid unit %q", v.Unit)
		}
		if year, month, err := strToTurnId(v.StartTurn); err != nil {
			return nil, fmt.Errorf("mentee: %s:%s: invalid start date", v.Unit, v.StartTurn)
		} else {
			v.StartTurn = fmt.Sprintf("%04d-%02d", year, month)
		}
		if year, month, err := strToTurnId(v.StopTurn); err != nil {
			return nil, fmt.Errorf("mentee: %s:%s: invalid stop date", v.Unit, v.StopTurn)
		} else {
			v.StopTurn = fmt.Sprintf("%04d-%02d", year, month)
		}
	}
	for _, v := range tmp.Worldographer.Solo {
		if v.Unit == "" || strings.TrimSpace(v.Unit) != v.Unit {
			return nil, fmt.Errorf("solo: invalid unit %q", v.Unit)
		}
		if year, month, err := strToTurnId(v.StartTurn); err != nil {
			return nil, fmt.Errorf("solo: %s:%s: invalid start date", v.Unit, v.StartTurn)
		} else {
			v.StartTurn = fmt.Sprintf("%04d-%02d", year, month)
		}
		if year, month, err := strToTurnId(v.StopTurn); err != nil {
			return nil, fmt.Errorf("solo: %s:%s: invalid stop date", v.Unit, v.StopTurn)
		} else {
			v.StopTurn = fmt.Sprintf("%04d-%02d", year, month)
		}
	}

	// copy over every value from tmp to config that isn't the default (zero) value
	copyNonZeroFields(&tmp, cfg)
	if cfg.Parser.QuitOnObscuredGrids {
		cfg.Parser.CheckObscuredGrids = true
	}
	if len(cfg.Worldographer.Solo) != 0 { // all solo players must use the reverse walker
		cfg.Experimental.ReverseWalker = true
	}

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

func strToTurnId(t string) (year, month int, err error) {
	fields := strings.Split(t, "-")
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("invalid date")
	}
	yyyy, mm, ok := strings.Cut(t, "-")
	if !ok {
		return 0, 0, fmt.Errorf("invalid date")
	} else if year, err = strconv.Atoi(yyyy); err != nil {
		return 0, 0, fmt.Errorf("invalid date")
	} else if month, err = strconv.Atoi(mm); err != nil {
		return 0, 0, fmt.Errorf("invalid date")
	} else if year < 899 || year > 9999 {
		return 0, 0, fmt.Errorf("invalid date")
	} else if month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("invalid date")
	}
	return year, month, nil
}
