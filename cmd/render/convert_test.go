// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"testing"

	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/terrain"
	schema "github.com/playbymail/ottomap/internal/tniif"
	"github.com/playbymail/ottomap/internal/wxx"
)

func TestConvert_TerrainMapping(t *testing.T) {
	t.Parallel()

	loc, _ := coords.HexToMap("AA 0101")
	noOffset := coords.Map{}

	terrainCases := []struct {
		code string
		want terrain.Terrain_e
	}{
		{"PR", terrain.FlatPrairie},
		{"BH", terrain.HillsBrush},
		{"D", terrain.FlatDeciduous},
		{"DE", terrain.FlatDesert},
		{"SW", terrain.FlatSwamp},
		{"O", terrain.WaterOcean},
		{"L", terrain.WaterLake},
		{"PI", terrain.FlatPolarIce},
	}

	for _, tc := range terrainCases {
		t.Run(tc.code, func(t *testing.T) {
			t.Parallel()
			ts := &tileState_t{
				Loc:     loc,
				Terrain: schema.Terrain(tc.code),
				Edges:   make(map[schema.Direction]schema.Edge),
			}
			hex, errs := convertTileToHex(ts, noOffset, "0987")
			if len(errs) != 0 {
				t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
			}
			if hex.Terrain != tc.want {
				t.Errorf("terrain %q: got %v, want %v", tc.code, hex.Terrain, tc.want)
			}
		})
	}

	t.Run("unknown terrain produces error", func(t *testing.T) {
		t.Parallel()
		ts := &tileState_t{
			Loc:     loc,
			Terrain: "XYZZY",
			Edges:   make(map[schema.Direction]schema.Edge),
		}
		_, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) == 0 {
			t.Fatal("expected error for unknown terrain")
		}
	})

	t.Run("empty terrain no error", func(t *testing.T) {
		t.Parallel()
		ts := &tileState_t{
			Loc:   loc,
			Edges: make(map[schema.Direction]schema.Edge),
		}
		hex, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if hex.Terrain != terrain.Blank {
			t.Errorf("terrain: got %v, want Blank", hex.Terrain)
		}
	})
}

func TestConvert_EdgeFeatureMapping(t *testing.T) {
	t.Parallel()

	loc, _ := coords.HexToMap("AA 0101")
	noOffset := coords.Map{}

	edgeCases := []struct {
		feature string
		dir     schema.Direction
		check   func(t *testing.T, hex *wxx.Hex)
	}{
		{"Canal", schema.DirN, func(t *testing.T, hex *wxx.Hex) {
			if len(hex.Features.Edges.Canal) != 1 || hex.Features.Edges.Canal[0] != direction.North {
				t.Errorf("canal: got %v, want [N]", hex.Features.Edges.Canal)
			}
		}},
		{"Ford", schema.DirNE, func(t *testing.T, hex *wxx.Hex) {
			if len(hex.Features.Edges.Ford) != 1 || hex.Features.Edges.Ford[0] != direction.NorthEast {
				t.Errorf("ford: got %v, want [NE]", hex.Features.Edges.Ford)
			}
		}},
		{"Pass", schema.DirSE, func(t *testing.T, hex *wxx.Hex) {
			if len(hex.Features.Edges.Pass) != 1 || hex.Features.Edges.Pass[0] != direction.SouthEast {
				t.Errorf("pass: got %v, want [SE]", hex.Features.Edges.Pass)
			}
		}},
		{"River", schema.DirS, func(t *testing.T, hex *wxx.Hex) {
			if len(hex.Features.Edges.River) != 1 || hex.Features.Edges.River[0] != direction.South {
				t.Errorf("river: got %v, want [S]", hex.Features.Edges.River)
			}
		}},
		{"Stone Road", schema.DirSW, func(t *testing.T, hex *wxx.Hex) {
			if len(hex.Features.Edges.StoneRoad) != 1 || hex.Features.Edges.StoneRoad[0] != direction.SouthWest {
				t.Errorf("stone road: got %v, want [SW]", hex.Features.Edges.StoneRoad)
			}
		}},
	}

	for _, tc := range edgeCases {
		t.Run(tc.feature, func(t *testing.T) {
			t.Parallel()
			ts := &tileState_t{
				Loc:     loc,
				Terrain: "PR",
				Edges: map[schema.Direction]schema.Edge{
					tc.dir: {Dir: tc.dir, Feature: schema.Feature(tc.feature)},
				},
			}
			hex, errs := convertTileToHex(ts, noOffset, "0987")
			if len(errs) != 0 {
				t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
			}
			tc.check(t, hex)
		})
	}

	t.Run("unknown feature produces error", func(t *testing.T) {
		t.Parallel()
		ts := &tileState_t{
			Loc:     loc,
			Terrain: "PR",
			Edges: map[schema.Direction]schema.Edge{
				schema.DirN: {Dir: schema.DirN, Feature: "Stine Road"},
			},
		}
		_, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) == 0 {
			t.Fatal("expected error for unknown edge feature")
		}
	})

	t.Run("edge with no feature skipped", func(t *testing.T) {
		t.Parallel()
		ts := &tileState_t{
			Loc:     loc,
			Terrain: "PR",
			Edges: map[schema.Direction]schema.Edge{
				schema.DirN: {Dir: schema.DirN, NeighborTerrain: "D"},
			},
		}
		hex, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		total := len(hex.Features.Edges.Canal) + len(hex.Features.Edges.Ford) +
			len(hex.Features.Edges.Pass) + len(hex.Features.Edges.River) +
			len(hex.Features.Edges.StoneRoad)
		if total != 0 {
			t.Errorf("expected 0 edge features, got %d", total)
		}
	})
}
