// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package tiles represents the game world as a collection of hex tiles indexed
// by map coordinates. Each tile contains terrain, settlements, resources, edge
// features, and unit encounters. The package merges turn report observations
// into tiles with conflict resolution logic and tracks which units discovered
// each location.
package tiles
