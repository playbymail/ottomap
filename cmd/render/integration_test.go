// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// NOTE: golden test files are not created in this sprint. Future sprints
// will import a new package that does a better job of reading and writing
// Worldographer data files. At that point, golden tests should compare the
// full WXX output byte-for-byte against a known-good reference file.

package main

import (
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/playbymail/ottomap/internal/config"
	schema "github.com/playbymail/ottomap/internal/tniif"
	"github.com/playbymail/ottomap/internal/wxx"
)

func TestIntegration_RenderSmallDoc(t *testing.T) {
	t.Parallel()

	doc1 := schema.Document{
		Schema: schema.Version,
		Game:   "0300",
		Turn:   "0901-01",
		Clan:   "0987",
		Clans: []schema.Clan{{
			ID: "0987",
			Units: []schema.Unit{{
				ID:             "0987",
				EndingLocation: "AA 0505",
				Moves: []schema.Moves{{
					ID: "0987",
					Steps: []schema.MoveStep{
						{
							Intent:         schema.IntentStill,
							EndingLocation: "AA 0505",
							Observation: &schema.Observation{
								Location:   "AA 0505",
								Terrain:    "PR",
								WasVisited: true,
								Edges: []schema.Edge{
									{Dir: schema.DirN, Feature: "River"},
								},
								Settlements: []schema.Settlement{{Name: "Gondor"}},
								Encounters:  []schema.Encounter{{Unit: "0138"}},
							},
						},
					},
				}},
			}},
		}},
	}

	doc2 := schema.Document{
		Schema: schema.Version,
		Game:   "0300",
		Turn:   "0902-01",
		Clan:   "0987",
		Clans: []schema.Clan{{
			ID: "0987",
			Units: []schema.Unit{{
				ID:             "0987",
				EndingLocation: "AA 0606",
				Moves: []schema.Moves{{
					ID: "0987",
					Steps: []schema.MoveStep{
						{
							Intent:         schema.IntentAdvance,
							Advance:        schema.DirNE,
							EndingLocation: "AA 0606",
							Observation: &schema.Observation{
								Location:   "AA 0606",
								Terrain:    "BH",
								WasVisited: true,
								WasScouted: true,
							},
						},
					},
				}},
			}},
		}},
	}

	loaded := []loadedDoc_t{
		{File: "0300.0901-01.0987.json", Doc: doc1},
		{File: "0300.0902-01.0987.json", Doc: doc2},
	}

	if errs := validateDocuments(loaded); len(errs) > 0 {
		t.Fatalf("validate failed: %v", errs)
	}

	events, flattenErrs := flattenEvents(loaded)
	if len(flattenErrs) > 0 {
		t.Fatalf("flatten failed: %v", flattenErrs)
	}

	owningClan := schema.ClanID("0987")
	sortEvents(events, owningClan)

	tiles := mergeTiles(events)
	if len(tiles) == 0 {
		t.Fatal("expected tiles from merge, got 0")
	}
	if len(tiles) != 2 {
		t.Errorf("expected 2 tiles, got %d", len(tiles))
	}

	upperLeft, lowerRight, offset := computeBoundsAndOffset(tiles)

	var hexes []*wxx.Hex
	for _, ts := range tiles {
		hex, errs := convertTileToHex(ts, offset, owningClan)
		if len(errs) > 0 {
			t.Fatalf("convert errors: %v", errs)
		}
		if hex != nil {
			hexes = append(hexes, hex)
		}
	}

	specials := collectSpecialHexes(loaded)
	applySpecialHexes(hexes, specials)

	if len(hexes) != 2 {
		t.Fatalf("expected 2 hexes, got %d", len(hexes))
	}

	gcfg := config.Default()
	w, err := wxx.NewWXX(gcfg)
	if err != nil {
		t.Fatalf("NewWXX: %v", err)
	}

	for _, hex := range hexes {
		if err := w.MergeHex(hex); err != nil {
			t.Fatalf("MergeHex %s: %v", hex.Location.GridString(), err)
		}
	}

	// wxx.Create writes directly to the filesystem via os.WriteFile,
	// so we use t.TempDir() which is automatically cleaned up.
	// Future sprints may refactor wxx.Create to accept an io.Writer
	// or an Afero filesystem for full in-memory testing.
	outputPath := filepath.Join(t.TempDir(), "test-output.wxx")

	var maxTurn schema.TurnID
	for _, ld := range loaded {
		if ld.Doc.Turn > maxTurn {
			maxTurn = ld.Doc.Turn
		}
	}

	renderCfg := wxx.RenderConfig{
		Version: version,
	}
	renderCfg.Meta.IncludeMeta = true
	renderCfg.Meta.IncludeOrigin = true

	if err := w.Create(outputPath, string(maxTurn), upperLeft, lowerRight, renderCfg, gcfg); err != nil {
		t.Fatalf("Create: %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("output file is empty")
	}

	// WXX files are gzip-compressed; verify the output is valid gzip
	f, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("opening output: %v", err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("output is not valid gzip: %v", err)
	}
	defer gr.Close()

	buf := make([]byte, 1024)
	n, err := gr.Read(buf)
	if n == 0 && err != nil {
		t.Fatalf("could not read gzip content: %v", err)
	}
}
