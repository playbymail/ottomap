// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"testing"

	"github.com/playbymail/ottomap/internal/coords"
	schema "github.com/playbymail/ottomap/internal/tniif"
)

func TestMerge_LWW_Terrain(t *testing.T) {
	t.Parallel()

	loc, _ := coords.HexToMap("AA 0101")

	t.Run("later observation overwrites terrain", func(t *testing.T) {
		t.Parallel()
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "PR",
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "BH",
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if ts.Terrain != "BH" {
			t.Errorf("terrain: got %q, want %q", ts.Terrain, "BH")
		}
	})

	t.Run("empty terrain preserves previous", func(t *testing.T) {
		t.Parallel()
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "D",
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101",
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if ts.Terrain != "D" {
			t.Errorf("terrain: got %q, want %q (empty should not overwrite)", ts.Terrain, "D")
		}
	})

	t.Run("three observations last wins", func(t *testing.T) {
		t.Parallel()
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "PR",
			}},
			{Turn: "0902-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "D",
			}},
			{Turn: "0903-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "BH",
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if ts.Terrain != "BH" {
			t.Errorf("terrain: got %q, want %q", ts.Terrain, "BH")
		}
	})
}

func TestMerge_NilVsEmptySlices(t *testing.T) {
	t.Parallel()

	loc, _ := coords.HexToMap("AA 0101")

	t.Run("nil settlements preserves previous", func(t *testing.T) {
		t.Parallel()
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location:    "AA 0101",
				Terrain:     "PR",
				Settlements: []schema.Settlement{{Name: "Gondor"}},
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101",
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if len(ts.Settlements) != 1 || ts.Settlements[0].Name != "Gondor" {
			t.Errorf("settlements: expected [Gondor], got %v", ts.Settlements)
		}
	})

	t.Run("empty slice overwrites previous", func(t *testing.T) {
		t.Parallel()
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location:    "AA 0101",
				Terrain:     "PR",
				Settlements: []schema.Settlement{{Name: "Gondor"}},
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location:    "AA 0101",
				Settlements: []schema.Settlement{},
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if len(ts.Settlements) != 0 {
			t.Errorf("settlements: expected empty, got %v", ts.Settlements)
		}
	})

	t.Run("nil resources preserves previous", func(t *testing.T) {
		t.Parallel()
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location:  "AA 0101",
				Terrain:   "PR",
				Resources: []schema.Resource{"Copper Ore"},
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101",
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if len(ts.Resources) != 1 || ts.Resources[0] != "Copper Ore" {
			t.Errorf("resources: expected [Copper Ore], got %v", ts.Resources)
		}
	})

	t.Run("nil encounters preserves previous", func(t *testing.T) {
		t.Parallel()
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location:   "AA 0101",
				Terrain:    "PR",
				Encounters: []schema.Encounter{{Unit: "0138"}},
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101",
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if len(ts.Encounters) != 1 || ts.Encounters[0].Unit != "0138" {
			t.Errorf("encounters: expected [0138], got %v", ts.Encounters)
		}
	})

	t.Run("notes always append", func(t *testing.T) {
		t.Parallel()
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101",
				Terrain:  "PR",
				Notes:    []schema.Note{{Kind: "info", Message: "first"}},
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101",
				Notes:    []schema.Note{{Kind: "warn", Message: "second"}},
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if len(ts.Notes) != 2 {
			t.Fatalf("notes: expected 2, got %d", len(ts.Notes))
		}
		if ts.Notes[0].Message != "first" || ts.Notes[1].Message != "second" {
			t.Errorf("notes: got [%q, %q], want [first, second]", ts.Notes[0].Message, ts.Notes[1].Message)
		}
	})
}

func TestMerge_EdgePerDirection(t *testing.T) {
	t.Parallel()

	loc, _ := coords.HexToMap("AA 0101")

	t.Run("later edge overwrites only affected direction", func(t *testing.T) {
		t.Parallel()
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101",
				Terrain:  "PR",
				Edges: []schema.Edge{
					{Dir: schema.DirN, Feature: "River"},
					{Dir: schema.DirSE, Feature: "Pass"},
				},
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101",
				Edges: []schema.Edge{
					{Dir: schema.DirN, Feature: "Stone Road"},
				},
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if ts.Edges[schema.DirN].Feature != "Stone Road" {
			t.Errorf("DirN: got %q, want %q", ts.Edges[schema.DirN].Feature, "Stone Road")
		}
		if ts.Edges[schema.DirSE].Feature != "Pass" {
			t.Errorf("DirSE: got %q, want %q (should be preserved)", ts.Edges[schema.DirSE].Feature, "Pass")
		}
	})

	t.Run("all six directions stored independently", func(t *testing.T) {
		t.Parallel()
		edges := []schema.Edge{
			{Dir: schema.DirN, Feature: "River"},
			{Dir: schema.DirNE, Feature: "Ford"},
			{Dir: schema.DirSE, Feature: "Pass"},
			{Dir: schema.DirS, Feature: "Canal"},
			{Dir: schema.DirSW, Feature: "Stone Road"},
			{Dir: schema.DirNW, Feature: "River"},
		}
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101",
				Terrain:  "PR",
				Edges:    edges,
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if len(ts.Edges) != 6 {
			t.Fatalf("edges: expected 6 directions, got %d", len(ts.Edges))
		}
		for _, dir := range schema.AllDirections {
			if _, ok := ts.Edges[dir]; !ok {
				t.Errorf("direction %s missing from edges", dir)
			}
		}
	})
}
