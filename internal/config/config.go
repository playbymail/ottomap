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
	Zoom   float64  `json:"Zoom"`
	Layers Layers_t `json:"Layers"`
}

type Layers_t struct {
	LargeCoords bool `json:"LargeCoords,omitempty"`
	MPCost      bool `json:"MPCost,omitempty"`
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
			},
		},
	}
}
func Load(name string, debug bool) (*Config, error) {
	// create a config with default values for the application
	cfg := Default()
	if sb, err := os.Stat(name); errors.Is(err, os.ErrNotExist) || os.IsNotExist(err) {
		return cfg, nil
	} else if sb.Mode().IsDir() {
		return cfg, ErrIsDirectory
	} else if !sb.Mode().IsRegular() {
		return cfg, ErrIsNotAFile
	}

	var tmp Config
	if data, err := os.ReadFile(name); err != nil {
		return cfg, nil
	} else if err = json.Unmarshal(data, &tmp); err != nil {
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
