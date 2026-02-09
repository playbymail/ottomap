// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package coords implements hexagonal coordinate systems for the TribeNet
// world map. It supports grid coordinates (AA-ZZ with 30x21 sub-grids),
// absolute map coordinates, and cube coordinates for hex geometry. It converts
// between all representations and computes directional movement for both even
// and odd columns on a flat-top hex grid.
package coords
