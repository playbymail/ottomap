// Copyright (c) 2025 Michael D Henderson. All rights reserved.

// Package walkers implements functions to walk parse trees and
// return map fragments
package walkers

import (
	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/reports"
)

type MapFragment_t struct {
	Coord coords.WorldMapCoord
	Turn  *reports.Turn_t
}
