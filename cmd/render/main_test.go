// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/domain"
	"github.com/playbymail/ottomap/internal/resources"
	"github.com/playbymail/ottomap/internal/terrain"
	schema "github.com/playbymail/ottomap/internal/tniif"
	"github.com/playbymail/ottomap/internal/wxx"
)

func TestMain(m *testing.M) {
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	os.Exit(m.Run())
}

func TestLoadDocuments(t *testing.T) {
	t.Run("valid document", func(t *testing.T) {
		doc := schema.Document{
			Schema: schema.Version,
			Game:   "0300",
			Turn:   "0904-01",
			Clan:   "0987",
		}
		data, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		tmp := filepath.Join(t.TempDir(), "test.json")
		if err := os.WriteFile(tmp, data, 0o644); err != nil {
			t.Fatal(err)
		}

		docs, err := loadDocuments([]string{tmp})
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 1 {
			t.Fatalf("expected 1 document, got %d", len(docs))
		}
		if docs[0].Doc.Schema != schema.Version {
			t.Errorf("schema: got %q, want %q", docs[0].Doc.Schema, schema.Version)
		}
		if docs[0].Doc.Game != "0300" {
			t.Errorf("game: got %q, want %q", docs[0].Doc.Game, "0300")
		}
		if docs[0].Doc.Turn != "0904-01" {
			t.Errorf("turn: got %q, want %q", docs[0].Doc.Turn, "0904-01")
		}
		if docs[0].File != "test.json" {
			t.Errorf("file: got %q, want %q", docs[0].File, "test.json")
		}
	})

	t.Run("multiple documents", func(t *testing.T) {
		dir := t.TempDir()
		for i, turn := range []string{"0904-01", "0904-02"} {
			doc := schema.Document{
				Schema: schema.Version,
				Game:   "0300",
				Turn:   schema.TurnID(turn),
				Clan:   "0987",
			}
			data, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			name := filepath.Join(dir, "doc"+string(rune('0'+i))+".json")
			if err := os.WriteFile(name, data, 0o644); err != nil {
				t.Fatal(err)
			}
		}
		paths := []string{
			filepath.Join(dir, "doc0.json"),
			filepath.Join(dir, "doc1.json"),
		}
		docs, err := loadDocuments(paths)
		if err != nil {
			t.Fatal(err)
		}
		if len(docs) != 2 {
			t.Fatalf("expected 2 documents, got %d", len(docs))
		}
		if docs[0].Doc.Turn != "0904-01" {
			t.Errorf("doc[0] turn: got %q, want %q", docs[0].Doc.Turn, "0904-01")
		}
		if docs[1].Doc.Turn != "0904-02" {
			t.Errorf("doc[1] turn: got %q, want %q", docs[1].Doc.Turn, "0904-02")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := loadDocuments([]string{filepath.Join(t.TempDir(), "noexist.json")})
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		tmp := filepath.Join(t.TempDir(), "bad.json")
		if err := os.WriteFile(tmp, []byte("{not json}"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := loadDocuments([]string{tmp})
		if err == nil {
			t.Fatal("expected error for invalid json")
		}
	})

	t.Run("directory rejected", func(t *testing.T) {
		_, err := loadDocuments([]string{t.TempDir()})
		if err == nil {
			t.Fatal("expected error for directory")
		}
	})
}

func TestValidateDocuments(t *testing.T) {
	validDoc := func() schema.Document {
		return schema.Document{
			Schema: schema.Version,
			Game:   "0300",
			Turn:   "0904-01",
			Clan:   "0987",
		}
	}
	wrap := func(doc schema.Document) loadedDoc_t {
		return loadedDoc_t{File: "0300.0901-01.0987.json", Doc: doc}
	}
	wrapN := func(name string, doc schema.Document) loadedDoc_t {
		return loadedDoc_t{File: name, Doc: doc}
	}

	t.Run("valid document", func(t *testing.T) {
		errs := validateDocuments([]loadedDoc_t{wrap(validDoc())})
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
	})

	t.Run("invalid schema version", func(t *testing.T) {
		doc := validDoc()
		doc.Schema = "wrong-version"
		errs := validateDocuments([]loadedDoc_t{wrap(doc)})
		if len(errs) == 0 {
			t.Fatal("expected error for invalid schema version")
		}
		found := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "schema") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected schema error, got: %v", errs)
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		doc := schema.Document{Schema: schema.Version, Turn: "0904-01"}
		errs := validateDocuments([]loadedDoc_t{wrap(doc)})
		gameErr, clanErr := false, false
		for _, e := range errs {
			if strings.Contains(e.Error(), "game is required") {
				gameErr = true
			}
			if strings.Contains(e.Error(), "clan is required") {
				clanErr = true
			}
		}
		if !gameErr {
			t.Error("expected game required error")
		}
		if !clanErr {
			t.Error("expected clan required error")
		}
	})

	t.Run("invalid turn format", func(t *testing.T) {
		for _, turn := range []string{"", "0904", "0904/01", "ABCD-01", "0904-AB"} {
			doc := validDoc()
			doc.Turn = schema.TurnID(turn)
			errs := validateDocuments([]loadedDoc_t{wrap(doc)})
			found := false
			for _, e := range errs {
				if strings.Contains(e.Error(), "YYYY-MM") {
					found = true
				}
			}
			if !found {
				t.Errorf("turn %q: expected YYYY-MM format error, got: %v", turn, errs)
			}
		}
	})

	t.Run("invalid coordinate", func(t *testing.T) {
		doc := validDoc()
		doc.Clans = []schema.Clan{{
			ID: "0987",
			Units: []schema.Unit{{
				ID:             "0987",
				EndingLocation: "BADCOORD",
			}},
		}}
		errs := validateDocuments([]loadedDoc_t{wrap(doc)})
		found := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "endingLocation") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected endingLocation error, got: %v", errs)
		}
	})

	t.Run("valid coordinate passes", func(t *testing.T) {
		doc := validDoc()
		doc.Clans = []schema.Clan{{
			ID: "0987",
			Units: []schema.Unit{{
				ID:             "0987",
				EndingLocation: "AA 0101",
			}},
		}}
		errs := validateDocuments([]loadedDoc_t{wrap(doc)})
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
	})

	t.Run("invalid direction in edge", func(t *testing.T) {
		doc := validDoc()
		doc.Clans = []schema.Clan{{
			ID: "0987",
			Units: []schema.Unit{{
				ID:             "0987",
				EndingLocation: "AA 0101",
				Moves: []schema.Moves{{
					ID: "0987",
					Steps: []schema.MoveStep{{
						Intent:         schema.IntentStill,
						EndingLocation: "AA 0101",
						Observation: &schema.Observation{
							Location: "AA 0101",
							Edges:    []schema.Edge{{Dir: "BAD"}},
						},
					}},
				}},
			}},
		}}
		errs := validateDocuments([]loadedDoc_t{wrap(doc)})
		found := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "direction") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected direction error, got: %v", errs)
		}
	})

	t.Run("invalid bearing in compass point", func(t *testing.T) {
		doc := validDoc()
		doc.Clans = []schema.Clan{{
			ID: "0987",
			Units: []schema.Unit{{
				ID:             "0987",
				EndingLocation: "AA 0101",
				Moves: []schema.Moves{{
					ID: "0987",
					Steps: []schema.MoveStep{{
						Intent:         schema.IntentStill,
						EndingLocation: "AA 0101",
						Observation: &schema.Observation{
							Location:      "AA 0101",
							CompassPoints: []schema.CompassPoint{{Bearing: "WRONG"}},
						},
					}},
				}},
			}},
		}}
		errs := validateDocuments([]loadedDoc_t{wrap(doc)})
		found := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "bearing") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected bearing error, got: %v", errs)
		}
	})

	t.Run("mismatched game IDs", func(t *testing.T) {
		doc1 := validDoc()
		doc2 := validDoc()
		doc2.Game = "0301"
		errs := validateDocuments([]loadedDoc_t{
			wrapN("0300.0901-01.0987.json", doc1),
			wrapN("0301.0901-01.0987.json", doc2),
		})
		found := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "does not match") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected game mismatch error, got: %v", errs)
		}
	})

	t.Run("scout step invalid coordinate", func(t *testing.T) {
		doc := validDoc()
		doc.Clans = []schema.Clan{{
			ID: "0987",
			Units: []schema.Unit{{
				ID:             "0987",
				EndingLocation: "AA 0101",
				Scouts: []schema.ScoutRun{{
					ID: "0987s1",
					Steps: []schema.MoveStep{{
						Intent:         schema.IntentAdvance,
						Advance:        schema.DirN,
						EndingLocation: "INVALID",
					}},
				}},
			}},
		}}
		errs := validateDocuments([]loadedDoc_t{wrap(doc)})
		found := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "scout 0987s1") && strings.Contains(e.Error(), "endingLocation") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected scout step coordinate error, got: %v", errs)
		}
	})

	t.Run("empty clan ID", func(t *testing.T) {
		doc := validDoc()
		doc.Clans = []schema.Clan{{ID: ""}}
		errs := validateDocuments([]loadedDoc_t{wrap(doc)})
		found := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "clans[0]") && strings.Contains(e.Error(), "id is required") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected clan id error, got: %v", errs)
		}
	})

	t.Run("empty unit ID", func(t *testing.T) {
		doc := validDoc()
		doc.Clans = []schema.Clan{{
			ID:    "0987",
			Units: []schema.Unit{{ID: "", EndingLocation: "AA 0101"}},
		}}
		errs := validateDocuments([]loadedDoc_t{wrap(doc)})
		found := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "units[0]") && strings.Contains(e.Error(), "id is required") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected unit id error, got: %v", errs)
		}
	})
}

func TestFlattenEvents(t *testing.T) {
	mkObs := func(loc string) *schema.Observation {
		return &schema.Observation{
			Location:   schema.Coordinates(loc),
			Terrain:    "PR",
			WasVisited: true,
		}
	}
	wrap := func(doc schema.Document) loadedDoc_t {
		return loadedDoc_t{File: "test.json", Doc: doc}
	}

	t.Run("two clans three units and scouts", func(t *testing.T) {
		doc := schema.Document{
			Schema: schema.Version,
			Game:   "0300",
			Turn:   "0904-01",
			Clan:   "0987",
			Clans: []schema.Clan{
				{
					ID: "0987",
					Units: []schema.Unit{
						{
							ID:             "0987",
							EndingLocation: "AA 0101",
							Moves: []schema.Moves{{
								ID: "0987",
								Steps: []schema.MoveStep{
									{Intent: schema.IntentStill, EndingLocation: "AA 0101", Observation: mkObs("AA 0101")},
								},
							}},
							Scouts: []schema.ScoutRun{{
								ID: "0987s1",
								Steps: []schema.MoveStep{
									{Intent: schema.IntentAdvance, Advance: schema.DirN, EndingLocation: "AA 0102", Observation: mkObs("AA 0102")},
								},
							}},
						},
						{
							ID:             "1987e1",
							EndingLocation: "AB 0305",
							Moves: []schema.Moves{{
								ID: "1987e1",
								Steps: []schema.MoveStep{
									{Intent: schema.IntentStill, EndingLocation: "AB 0305", Observation: mkObs("AB 0305")},
								},
							}},
						},
					},
				},
				{
					ID: "0138",
					Units: []schema.Unit{
						{
							ID:             "0138",
							EndingLocation: "BC 1518",
							Moves: []schema.Moves{{
								ID: "0138",
								Steps: []schema.MoveStep{
									{Intent: schema.IntentStill, EndingLocation: "BC 1518", Observation: mkObs("BC 1518")},
								},
							}},
						},
					},
				},
			},
		}

		events, errs := flattenEvents([]loadedDoc_t{{File: "0300.0904-01.0987.json", Doc: doc}})
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if len(events) != 4 {
			t.Fatalf("expected 4 events, got %d", len(events))
		}

		clanCounts := make(map[schema.ClanID]int)
		unitSet := make(map[string]bool)
		for _, ev := range events {
			clanCounts[ev.Clan]++
			unitSet[ev.Unit] = true
			if ev.Turn != "0904-01" {
				t.Errorf("expected turn 0904-01, got %q", ev.Turn)
			}
		}
		if clanCounts["0987"] != 3 {
			t.Errorf("clan 0987: expected 3 events, got %d", clanCounts["0987"])
		}
		if clanCounts["0138"] != 1 {
			t.Errorf("clan 0138: expected 1 event, got %d", clanCounts["0138"])
		}
		for _, u := range []string{"0987", "0987s1", "1987e1", "0138"} {
			if !unitSet[u] {
				t.Errorf("expected unit %q in events", u)
			}
		}
	})

	t.Run("scout always WasScouted", func(t *testing.T) {
		doc := schema.Document{
			Schema: schema.Version,
			Game:   "0300",
			Turn:   "0904-01",
			Clan:   "0987",
			Clans: []schema.Clan{{
				ID: "0987",
				Units: []schema.Unit{{
					ID:             "0987",
					EndingLocation: "AA 0101",
					Scouts: []schema.ScoutRun{{
						ID: "0987s1",
						Steps: []schema.MoveStep{
							{Intent: schema.IntentAdvance, Advance: schema.DirN, EndingLocation: "AA 0102",
								Observation: &schema.Observation{Location: "AA 0102", Terrain: "PR"}},
						},
					}},
				}},
			}},
		}
		events, errs := flattenEvents([]loadedDoc_t{wrap(doc)})
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if !events[0].WasScouted {
			t.Error("expected WasScouted=true for scout event")
		}
	})

	t.Run("correct coordinate conversion", func(t *testing.T) {
		doc := schema.Document{
			Schema: schema.Version,
			Game:   "0300",
			Turn:   "0904-01",
			Clan:   "0987",
			Clans: []schema.Clan{{
				ID: "0987",
				Units: []schema.Unit{{
					ID:             "0987",
					EndingLocation: "AA 0101",
					Moves: []schema.Moves{{
						ID: "0987",
						Steps: []schema.MoveStep{
							{Intent: schema.IntentStill, EndingLocation: "AA 0101", Observation: mkObs("AA 0101")},
						},
					}},
				}},
			}},
		}
		events, errs := flattenEvents([]loadedDoc_t{wrap(doc)})
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		wantLoc, _ := coords.HexToMap("AA 0101")
		if events[0].Loc != wantLoc {
			t.Errorf("loc: got %v, want %v", events[0].Loc, wantLoc)
		}
	})

	t.Run("steps without observation skipped", func(t *testing.T) {
		doc := schema.Document{
			Schema: schema.Version,
			Game:   "0300",
			Turn:   "0904-01",
			Clan:   "0987",
			Clans: []schema.Clan{{
				ID: "0987",
				Units: []schema.Unit{{
					ID:             "0987",
					EndingLocation: "AA 0101",
					Moves: []schema.Moves{{
						ID: "0987",
						Steps: []schema.MoveStep{
							{Intent: schema.IntentAdvance, Advance: schema.DirN, EndingLocation: "AA 0102"},
							{Intent: schema.IntentStill, EndingLocation: "AA 0102", Observation: mkObs("AA 0102")},
						},
					}},
				}},
			}},
		}
		events, errs := flattenEvents([]loadedDoc_t{wrap(doc)})
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if len(events) != 1 {
			t.Fatalf("expected 1 event (nil obs skipped), got %d", len(events))
		}
	})

	t.Run("invalid observation location returns error", func(t *testing.T) {
		doc := schema.Document{
			Schema: schema.Version,
			Game:   "0300",
			Turn:   "0904-01",
			Clan:   "0987",
			Clans: []schema.Clan{{
				ID: "0987",
				Units: []schema.Unit{{
					ID:             "0987",
					EndingLocation: "AA 0101",
					Moves: []schema.Moves{{
						ID: "0987",
						Steps: []schema.MoveStep{
							{Intent: schema.IntentStill, EndingLocation: "AA 0101",
								Observation: &schema.Observation{Location: "BADCOORD", Terrain: "PR"}},
						},
					}},
				}},
			}},
		}
		events, errs := flattenEvents([]loadedDoc_t{wrap(doc)})
		if len(errs) != 1 {
			t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
		}
		if len(events) != 0 {
			t.Errorf("expected 0 events for invalid location, got %d", len(events))
		}
	})

	t.Run("multiple documents", func(t *testing.T) {
		doc1 := schema.Document{
			Schema: schema.Version,
			Game:   "0300",
			Turn:   "0904-01",
			Clan:   "0987",
			Clans: []schema.Clan{{
				ID: "0987",
				Units: []schema.Unit{{
					ID:             "0987",
					EndingLocation: "AA 0101",
					Moves: []schema.Moves{{
						ID: "0987",
						Steps: []schema.MoveStep{
							{Intent: schema.IntentStill, EndingLocation: "AA 0101", Observation: mkObs("AA 0101")},
						},
					}},
				}},
			}},
		}
		doc2 := schema.Document{
			Schema: schema.Version,
			Game:   "0300",
			Turn:   "0904-02",
			Clan:   "0987",
			Clans: []schema.Clan{{
				ID: "0987",
				Units: []schema.Unit{{
					ID:             "0987",
					EndingLocation: "AB 0305",
					Moves: []schema.Moves{{
						ID: "0987",
						Steps: []schema.MoveStep{
							{Intent: schema.IntentStill, EndingLocation: "AB 0305", Observation: mkObs("AB 0305")},
						},
					}},
				}},
			}},
		}
		events, errs := flattenEvents([]loadedDoc_t{wrap(doc1), wrap(doc2)})
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if len(events) != 2 {
			t.Fatalf("expected 2 events, got %d", len(events))
		}
		if events[0].Turn != "0904-01" {
			t.Errorf("event[0] turn: got %q, want %q", events[0].Turn, "0904-01")
		}
		if events[1].Turn != "0904-02" {
			t.Errorf("event[1] turn: got %q, want %q", events[1].Turn, "0904-02")
		}
	})
}

func TestSortEvents(t *testing.T) {
	loc, _ := coords.HexToMap("AA 0101")

	mkEvent := func(turn schema.TurnID, clan schema.ClanID, unit string) obsEvent_t {
		return obsEvent_t{Turn: turn, Clan: clan, Unit: unit, Loc: loc}
	}

	t.Run("owning clan sorts last within each turn", func(t *testing.T) {
		events := []obsEvent_t{
			mkEvent("0901-01", "0987", "1987e1"),
			mkEvent("0901-01", "0138", "2138e2"),
			mkEvent("0902-01", "0987", "1987e1"),
			mkEvent("0902-01", "0138", "2138e2"),
		}
		sortEvents(events, "0987")

		want := []struct {
			Turn string
			Clan string
			Unit string
		}{
			{"0901-01", "0138", "2138e2"},
			{"0901-01", "0987", "1987e1"},
			{"0902-01", "0138", "2138e2"},
			{"0902-01", "0987", "1987e1"},
		}
		if len(events) != len(want) {
			t.Fatalf("expected %d events, got %d", len(want), len(events))
		}
		for i, w := range want {
			if string(events[i].Turn) != w.Turn || string(events[i].Clan) != w.Clan || events[i].Unit != w.Unit {
				t.Errorf("event[%d]: got (%s, %s, %s), want (%s, %s, %s)",
					i, events[i].Turn, events[i].Clan, events[i].Unit, w.Turn, w.Clan, w.Unit)
			}
		}
	})

	t.Run("non-owning clans sorted by clan ID", func(t *testing.T) {
		events := []obsEvent_t{
			mkEvent("0901-01", "0500", "0500"),
			mkEvent("0901-01", "0200", "0200"),
			mkEvent("0901-01", "0138", "0138"),
		}
		sortEvents(events, "0987")

		for i := 0; i < len(events)-1; i++ {
			if events[i].Clan > events[i+1].Clan {
				t.Errorf("event[%d] clan %s should come before event[%d] clan %s",
					i, events[i].Clan, i+1, events[i+1].Clan)
			}
		}
	})

	t.Run("units sorted within same turn and clan", func(t *testing.T) {
		events := []obsEvent_t{
			mkEvent("0901-01", "0138", "2138e2"),
			mkEvent("0901-01", "0138", "0138"),
			mkEvent("0901-01", "0138", "1138e1"),
		}
		sortEvents(events, "0987")

		for i := 0; i < len(events)-1; i++ {
			if events[i].Unit > events[i+1].Unit {
				t.Errorf("event[%d] unit %s should come before event[%d] unit %s",
					i, events[i].Unit, i+1, events[i+1].Unit)
			}
		}
	})

	t.Run("stable sort preserves insertion order for equal keys", func(t *testing.T) {
		e1 := mkEvent("0901-01", "0138", "0138")
		e1.WasVisited = true
		e2 := mkEvent("0901-01", "0138", "0138")
		e2.WasScouted = true
		events := []obsEvent_t{e1, e2}
		sortEvents(events, "0987")

		if !events[0].WasVisited {
			t.Error("expected first event to have WasVisited=true (stable order)")
		}
		if !events[1].WasScouted {
			t.Error("expected second event to have WasScouted=true (stable order)")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		var events []obsEvent_t
		sortEvents(events, "0987")
		if len(events) != 0 {
			t.Errorf("expected 0 events, got %d", len(events))
		}
	})
}

func TestMergeTiles(t *testing.T) {
	loc, _ := coords.HexToMap("AA 0101")
	loc2, _ := coords.HexToMap("AB 0305")

	t.Run("later terrain overwrites earlier", func(t *testing.T) {
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "PR",
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "BH",
			}},
		}
		tiles := mergeTiles(events)
		if len(tiles) != 1 {
			t.Fatalf("expected 1 tile, got %d", len(tiles))
		}
		ts := tiles[loc]
		if ts.Terrain != "BH" {
			t.Errorf("terrain: got %q, want %q", ts.Terrain, "BH")
		}
	})

	t.Run("empty terrain does not overwrite", func(t *testing.T) {
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "PR",
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101",
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if ts.Terrain != "PR" {
			t.Errorf("terrain: got %q, want %q", ts.Terrain, "PR")
		}
	})

	t.Run("nil settlements preserves older", func(t *testing.T) {
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location:    "AA 0101",
				Terrain:     "PR",
				Settlements: []schema.Settlement{{Name: "Gondor"}},
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101",
				Terrain:  "PR",
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if len(ts.Settlements) != 1 || ts.Settlements[0].Name != "Gondor" {
			t.Errorf("settlements: expected [Gondor], got %v", ts.Settlements)
		}
	})

	t.Run("empty slice settlements clears older", func(t *testing.T) {
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location:    "AA 0101",
				Terrain:     "PR",
				Settlements: []schema.Settlement{{Name: "Gondor"}},
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location:    "AA 0101",
				Terrain:     "PR",
				Settlements: []schema.Settlement{},
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if len(ts.Settlements) != 0 {
			t.Errorf("settlements: expected empty, got %v", ts.Settlements)
		}
	})

	t.Run("edge merge overwrites only affected direction", func(t *testing.T) {
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
			t.Errorf("DirN feature: got %q, want %q", ts.Edges[schema.DirN].Feature, "Stone Road")
		}
		if ts.Edges[schema.DirSE].Feature != "Pass" {
			t.Errorf("DirSE feature: got %q, want %q", ts.Edges[schema.DirSE].Feature, "Pass")
		}
	})

	t.Run("WasVisited and WasScouted use logical OR", func(t *testing.T) {
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, WasVisited: true, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "PR", WasVisited: true,
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, WasScouted: true, Obs: &schema.Observation{
				Location: "AA 0101", WasScouted: true,
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if !ts.WasVisited {
			t.Error("expected WasVisited=true")
		}
		if !ts.WasScouted {
			t.Error("expected WasScouted=true")
		}
	})

	t.Run("notes always append", func(t *testing.T) {
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "PR",
				Notes: []schema.Note{{Kind: "info", Message: "first"}},
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
		if ts.Notes[0].Message != "first" {
			t.Errorf("notes[0]: got %q, want %q", ts.Notes[0].Message, "first")
		}
		if ts.Notes[1].Message != "second" {
			t.Errorf("notes[1]: got %q, want %q", ts.Notes[1].Message, "second")
		}
	})

	t.Run("multiple tiles kept separate", func(t *testing.T) {
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location: "AA 0101", Terrain: "PR",
			}},
			{Turn: "0901-01", Clan: "0987", Unit: "1987e1", Loc: loc2, Obs: &schema.Observation{
				Location: "AB 0305", Terrain: "D",
			}},
		}
		tiles := mergeTiles(events)
		if len(tiles) != 2 {
			t.Fatalf("expected 2 tiles, got %d", len(tiles))
		}
		if tiles[loc].Terrain != "PR" {
			t.Errorf("tile1 terrain: got %q, want %q", tiles[loc].Terrain, "PR")
		}
		if tiles[loc2].Terrain != "D" {
			t.Errorf("tile2 terrain: got %q, want %q", tiles[loc2].Terrain, "D")
		}
	})

	t.Run("nil obs skipped", func(t *testing.T) {
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: nil},
		}
		tiles := mergeTiles(events)
		if len(tiles) != 0 {
			t.Errorf("expected 0 tiles, got %d", len(tiles))
		}
	})

	t.Run("empty events", func(t *testing.T) {
		tiles := mergeTiles(nil)
		if len(tiles) != 0 {
			t.Errorf("expected 0 tiles, got %d", len(tiles))
		}
	})

	t.Run("resources overwrite when non-nil", func(t *testing.T) {
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location:  "AA 0101",
				Terrain:   "PR",
				Resources: []schema.Resource{"Copper Ore"},
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location:  "AA 0101",
				Resources: []schema.Resource{"Jade", "Iron"},
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if len(ts.Resources) != 2 || ts.Resources[0] != "Jade" || ts.Resources[1] != "Iron" {
			t.Errorf("resources: expected [Jade Iron], got %v", ts.Resources)
		}
	})

	t.Run("encounters overwrite when non-nil", func(t *testing.T) {
		events := []obsEvent_t{
			{Turn: "0901-01", Clan: "0138", Unit: "0138", Loc: loc, Obs: &schema.Observation{
				Location:   "AA 0101",
				Terrain:    "PR",
				Encounters: []schema.Encounter{{Unit: "0138"}},
			}},
			{Turn: "0902-01", Clan: "0987", Unit: "0987", Loc: loc, Obs: &schema.Observation{
				Location:   "AA 0101",
				Encounters: []schema.Encounter{{Unit: "0987"}, {Unit: "1987e1"}},
			}},
		}
		tiles := mergeTiles(events)
		ts := tiles[loc]
		if len(ts.Encounters) != 2 {
			t.Fatalf("encounters: expected 2, got %d", len(ts.Encounters))
		}
		if ts.Encounters[0].Unit != "0987" {
			t.Errorf("encounters[0]: got %q, want %q", ts.Encounters[0].Unit, "0987")
		}
	})
}

func TestConvertTileToHex(t *testing.T) {
	loc, _ := coords.HexToMap("AA 0101")
	noOffset := coords.Map{}

	t.Run("terrain maps to enum", func(t *testing.T) {
		ts := &tileState_t{
			Loc:     loc,
			Terrain: "PR",
			Edges:   make(map[schema.Direction]schema.Edge),
		}
		hex, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if hex.Terrain != terrain.FlatPrairie {
			t.Errorf("terrain: got %v, want %v", hex.Terrain, terrain.FlatPrairie)
		}
	})

	t.Run("unknown terrain produces error", func(t *testing.T) {
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

	t.Run("river edge on NE", func(t *testing.T) {
		ts := &tileState_t{
			Loc:     loc,
			Terrain: "PR",
			Edges: map[schema.Direction]schema.Edge{
				schema.DirNE: {Dir: schema.DirNE, Feature: "River"},
			},
		}
		hex, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if len(hex.Features.Edges.River) != 1 {
			t.Fatalf("river edges: expected 1, got %d", len(hex.Features.Edges.River))
		}
		if hex.Features.Edges.River[0] != direction.NorthEast {
			t.Errorf("river direction: got %v, want %v", hex.Features.Edges.River[0], direction.NorthEast)
		}
	})

	t.Run("multiple edge types", func(t *testing.T) {
		ts := &tileState_t{
			Loc:     loc,
			Terrain: "BH",
			Edges: map[schema.Direction]schema.Edge{
				schema.DirN:  {Dir: schema.DirN, Feature: "Stone Road"},
				schema.DirSE: {Dir: schema.DirSE, Feature: "Pass"},
				schema.DirS:  {Dir: schema.DirS, Feature: "River"},
			},
		}
		hex, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if len(hex.Features.Edges.StoneRoad) != 1 || hex.Features.Edges.StoneRoad[0] != direction.North {
			t.Errorf("stone road: got %v, want [N]", hex.Features.Edges.StoneRoad)
		}
		if len(hex.Features.Edges.Pass) != 1 || hex.Features.Edges.Pass[0] != direction.SouthEast {
			t.Errorf("pass: got %v, want [SE]", hex.Features.Edges.Pass)
		}
		if len(hex.Features.Edges.River) != 1 || hex.Features.Edges.River[0] != direction.South {
			t.Errorf("river: got %v, want [S]", hex.Features.Edges.River)
		}
	})

	t.Run("edge with no feature skipped", func(t *testing.T) {
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

	t.Run("unknown edge feature produces error", func(t *testing.T) {
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

	t.Run("resource string maps to enum", func(t *testing.T) {
		ts := &tileState_t{
			Loc:       loc,
			Terrain:   "PR",
			Edges:     make(map[schema.Direction]schema.Edge),
			Resources: []schema.Resource{"Copper Ore", "Jade"},
		}
		hex, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if len(hex.Features.Resources) != 2 {
			t.Fatalf("resources: expected 2, got %d", len(hex.Features.Resources))
		}
		found := map[resources.Resource_e]bool{}
		for _, r := range hex.Features.Resources {
			found[r] = true
		}
		if !found[resources.CopperOre] {
			t.Error("expected Copper Ore resource")
		}
		if !found[resources.Jade] {
			t.Error("expected Jade resource")
		}
	})

	t.Run("unknown resource produces error", func(t *testing.T) {
		ts := &tileState_t{
			Loc:       loc,
			Terrain:   "PR",
			Edges:     make(map[schema.Direction]schema.Edge),
			Resources: []schema.Resource{"Unobtanium"},
		}
		_, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) == 0 {
			t.Fatal("expected error for unknown resource")
		}
	})

	t.Run("settlements converted", func(t *testing.T) {
		ts := &tileState_t{
			Loc:         loc,
			Terrain:     "PR",
			Edges:       make(map[schema.Direction]schema.Edge),
			Settlements: []schema.Settlement{{Name: "Gondor"}},
		}
		hex, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if len(hex.Features.Settlements) != 1 {
			t.Fatalf("settlements: expected 1, got %d", len(hex.Features.Settlements))
		}
		if hex.Features.Settlements[0].Name != "Gondor" {
			t.Errorf("settlement name: got %q, want %q", hex.Features.Settlements[0].Name, "Gondor")
		}
	})

	t.Run("encounters mark friendly correctly", func(t *testing.T) {
		ts := &tileState_t{
			Loc:     loc,
			Terrain: "PR",
			Edges:   make(map[schema.Direction]schema.Edge),
			Encounters: []schema.Encounter{
				{Unit: "0987"},
				{Unit: "1987e1"},
				{Unit: "0138"},
			},
		}
		hex, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if len(hex.Features.Encounters) != 3 {
			t.Fatalf("encounters: expected 3, got %d", len(hex.Features.Encounters))
		}
		for _, enc := range hex.Features.Encounters {
			switch enc.UnitId {
			case "0987":
				if !enc.Friendly {
					t.Errorf("unit 0987 should be friendly to clan 0987")
				}
			case "1987e1":
				if !enc.Friendly {
					t.Errorf("unit 1987e1 should be friendly to clan 0987")
				}
			case "0138":
				if enc.Friendly {
					t.Errorf("unit 0138 should not be friendly to clan 0987")
				}
			}
		}
	})

	t.Run("render offset applied", func(t *testing.T) {
		ts := &tileState_t{
			Loc:     loc,
			Terrain: "PR",
			Edges:   make(map[schema.Direction]schema.Edge),
		}
		offset := coords.Map{Column: 5, Row: 3}
		hex, errs := convertTileToHex(ts, offset, "0987")
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		expectedCol := ts.Loc.Column - 5
		expectedRow := ts.Loc.Row - 3
		if hex.RenderAt.Column != expectedCol || hex.RenderAt.Row != expectedRow {
			t.Errorf("renderAt: got (%d,%d), want (%d,%d)", hex.RenderAt.Column, hex.RenderAt.Row, expectedCol, expectedRow)
		}
	})

	t.Run("WasVisited and WasScouted copied", func(t *testing.T) {
		ts := &tileState_t{
			Loc:        loc,
			Terrain:    "PR",
			Edges:      make(map[schema.Direction]schema.Edge),
			WasVisited: true,
			WasScouted: true,
		}
		hex, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if !hex.WasVisited {
			t.Error("expected WasVisited=true")
		}
		if !hex.WasScouted {
			t.Error("expected WasScouted=true")
		}
	})

	t.Run("empty terrain no error", func(t *testing.T) {
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

	t.Run("all edge feature types", func(t *testing.T) {
		ts := &tileState_t{
			Loc:     loc,
			Terrain: "PR",
			Edges: map[schema.Direction]schema.Edge{
				schema.DirN:  {Dir: schema.DirN, Feature: "Canal"},
				schema.DirNE: {Dir: schema.DirNE, Feature: "Ford"},
				schema.DirSE: {Dir: schema.DirSE, Feature: "Pass"},
				schema.DirS:  {Dir: schema.DirS, Feature: "River"},
				schema.DirSW: {Dir: schema.DirSW, Feature: "Stone Road"},
			},
		}
		hex, errs := convertTileToHex(ts, noOffset, "0987")
		if len(errs) != 0 {
			t.Fatalf("expected 0 errors, got %d: %v", len(errs), errs)
		}
		if len(hex.Features.Edges.Canal) != 1 || hex.Features.Edges.Canal[0] != direction.North {
			t.Errorf("canal: got %v, want [N]", hex.Features.Edges.Canal)
		}
		if len(hex.Features.Edges.Ford) != 1 || hex.Features.Edges.Ford[0] != direction.NorthEast {
			t.Errorf("ford: got %v, want [NE]", hex.Features.Edges.Ford)
		}
		if len(hex.Features.Edges.Pass) != 1 || hex.Features.Edges.Pass[0] != direction.SouthEast {
			t.Errorf("pass: got %v, want [SE]", hex.Features.Edges.Pass)
		}
		if len(hex.Features.Edges.River) != 1 || hex.Features.Edges.River[0] != direction.South {
			t.Errorf("river: got %v, want [S]", hex.Features.Edges.River)
		}
		if len(hex.Features.Edges.StoneRoad) != 1 || hex.Features.Edges.StoneRoad[0] != direction.SouthWest {
			t.Errorf("stone road: got %v, want [SW]", hex.Features.Edges.StoneRoad)
		}
	})
}

func TestComputeBoundsAndOffset(t *testing.T) {
	mkTiles := func(hexes ...string) map[coords.Map]*tileState_t {
		tiles := make(map[coords.Map]*tileState_t)
		for _, h := range hexes {
			loc, err := coords.HexToMap(h)
			if err != nil {
				t.Fatalf("bad hex %q: %v", h, err)
			}
			tiles[loc] = &tileState_t{Loc: loc}
		}
		return tiles
	}

	mustMap := func(hex string) coords.Map {
		loc, err := coords.HexToMap(hex)
		if err != nil {
			t.Fatalf("bad hex %q: %v", hex, err)
		}
		return loc
	}

	t.Run("empty tiles", func(t *testing.T) {
		ul, lr, off := computeBoundsAndOffset(nil)
		if ul != (coords.Map{}) || lr != (coords.Map{}) || off != (coords.Map{}) {
			t.Errorf("expected all zero, got ul=%v lr=%v off=%v", ul, lr, off)
		}
	})

	t.Run("single tile near origin", func(t *testing.T) {
		tiles := mkTiles("AA 0101")
		ul, lr, off := computeBoundsAndOffset(tiles)
		wantLoc := mustMap("AA 0101")
		if ul != wantLoc {
			t.Errorf("upperLeft: got %v, want %v", ul, wantLoc)
		}
		if lr != wantLoc {
			t.Errorf("lowerRight: got %v, want %v", lr, wantLoc)
		}
		if off != (coords.Map{}) {
			t.Errorf("offset: got %v, want (0,0) (near origin, no shift)", off)
		}
	})

	t.Run("bounds computed correctly", func(t *testing.T) {
		tiles := mkTiles("AA 0505", "AA 1015", "AA 0110")
		ul, lr, _ := computeBoundsAndOffset(tiles)
		wantUL := mustMap("AA 0105")
		wantLR := mustMap("AA 1015")
		if ul != wantUL {
			t.Errorf("upperLeft: got %s, want %s", ul.GridString(), wantUL.GridString())
		}
		if lr != wantLR {
			t.Errorf("lowerRight: got %s, want %s", lr.GridString(), wantLR.GridString())
		}
	})

	t.Run("offset with border subtraction", func(t *testing.T) {
		tiles := mkTiles("AA 1515")
		_, _, off := computeBoundsAndOffset(tiles)
		if off.Column != 10 {
			t.Errorf("offset.Column: got %d, want 10 (14-4=10, even)", off.Column)
		}
		if off.Row != 10 {
			t.Errorf("offset.Row: got %d, want 10 (14-4=10, even)", off.Row)
		}
	})

	t.Run("odd column offset made even", func(t *testing.T) {
		tiles := mkTiles("AA 1415")
		_, _, off := computeBoundsAndOffset(tiles)
		if off.Column%2 != 0 {
			t.Errorf("offset.Column %d should be even", off.Column)
		}
		if off.Column != 8 {
			t.Errorf("offset.Column: got %d, want 8 (13-4=9, made even=8)", off.Column)
		}
	})

	t.Run("odd row offset made even", func(t *testing.T) {
		tiles := mkTiles("AA 1514")
		_, _, off := computeBoundsAndOffset(tiles)
		if off.Row%2 != 0 {
			t.Errorf("offset.Row %d should be even", off.Row)
		}
		if off.Row != 8 {
			t.Errorf("offset.Row: got %d, want 8 (13-4=9, made even=8)", off.Row)
		}
	})

	t.Run("no shift when near origin", func(t *testing.T) {
		tiles := mkTiles("AA 0303")
		_, _, off := computeBoundsAndOffset(tiles)
		if off.Column != 0 {
			t.Errorf("offset.Column: got %d, want 0 (column <= borderWidth)", off.Column)
		}
		if off.Row != 0 {
			t.Errorf("offset.Row: got %d, want 0 (row <= borderHeight)", off.Row)
		}
	})

	t.Run("offset does not cross grid boundary", func(t *testing.T) {
		tiles := mkTiles("AB 0115")
		_, _, off := computeBoundsAndOffset(tiles)
		minColInGrid := (mustMap("AB 0115").Column / 30) * 30
		if off.Column < minColInGrid {
			t.Errorf("offset.Column %d crossed grid boundary at %d", off.Column, minColInGrid)
		}
	})

	t.Run("sprint plan examples", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			hex    string
			wantUL string
		}{
			{"AA 0101 stays", "AA 0101", "AA 0101"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				tiles := mkTiles(tc.hex)
				ul, _, _ := computeBoundsAndOffset(tiles)
				want := mustMap(tc.wantUL)
				if ul != want {
					t.Errorf("upperLeft: got %s, want %s", ul.GridString(), want.GridString())
				}
			})
		}
	})

	t.Run("multiple tiles span", func(t *testing.T) {
		tiles := mkTiles("AA 1010", "AA 2020")
		ul, lr, off := computeBoundsAndOffset(tiles)
		wantUL := mustMap("AA 1010")
		wantLR := mustMap("AA 2020")
		if ul != wantUL {
			t.Errorf("upperLeft: got %s, want %s", ul.GridString(), wantUL.GridString())
		}
		if lr != wantLR {
			t.Errorf("lowerRight: got %s, want %s", lr.GridString(), wantLR.GridString())
		}
		if off.Column%2 != 0 {
			t.Errorf("offset.Column %d should be even", off.Column)
		}
		if off.Row%2 != 0 {
			t.Errorf("offset.Row %d should be even", off.Row)
		}
	})
}

func TestCollectSpecialHexes(t *testing.T) {
	t.Run("empty docs", func(t *testing.T) {
		specials := collectSpecialHexes(nil)
		if len(specials) != 0 {
			t.Errorf("expected 0 specials, got %d", len(specials))
		}
	})

	t.Run("single special hex", func(t *testing.T) {
		docs := []loadedDoc_t{{File: "test.json", Doc: schema.Document{
			Schema:       schema.Version,
			Game:         "0300",
			Turn:         "0904-01",
			Clan:         "0987",
			SpecialHexes: []schema.SpecialHex{{Name: "Foo"}},
		}}}
		specials := collectSpecialHexes(docs)
		if len(specials) != 1 {
			t.Fatalf("expected 1 special, got %d", len(specials))
		}
		sp, ok := specials["foo"]
		if !ok {
			t.Fatal("expected key 'foo'")
		}
		if sp.Name != "Foo" {
			t.Errorf("name: got %q, want %q", sp.Name, "Foo")
		}
		if sp.Id != "foo" {
			t.Errorf("id: got %q, want %q", sp.Id, "foo")
		}
	})

	t.Run("deduplicates across documents", func(t *testing.T) {
		docs := []loadedDoc_t{
			{File: "test1.json", Doc: schema.Document{
				Schema:       schema.Version,
				Game:         "0300",
				Turn:         "0904-01",
				Clan:         "0987",
				SpecialHexes: []schema.SpecialHex{{Name: "Foo"}},
			}},
			{File: "test2.json", Doc: schema.Document{
				Schema:       schema.Version,
				Game:         "0300",
				Turn:         "0904-02",
				Clan:         "0987",
				SpecialHexes: []schema.SpecialHex{{Name: "FOO"}, {Name: "Bar"}},
			}},
		}
		specials := collectSpecialHexes(docs)
		if len(specials) != 2 {
			t.Fatalf("expected 2 specials, got %d", len(specials))
		}
		if specials["foo"].Name != "Foo" {
			t.Errorf("first-seen name should be preserved: got %q, want %q", specials["foo"].Name, "Foo")
		}
		if _, ok := specials["bar"]; !ok {
			t.Error("expected key 'bar'")
		}
	})
}

func TestApplySpecialHexes(t *testing.T) {
	t.Run("settlement promoted to special", func(t *testing.T) {
		hex := &wxx.Hex{
			Features: wxx.Features{
				Settlements: []*domain.Settlement_t{{Name: "Foo"}},
			},
		}
		specials := map[string]*domain.Special_t{
			"foo": {Id: "foo", Name: "Foo"},
		}
		applySpecialHexes([]*wxx.Hex{hex}, specials)
		if len(hex.Features.Settlements) != 0 {
			t.Errorf("settlements: expected 0, got %d", len(hex.Features.Settlements))
		}
		if len(hex.Features.Special) != 1 {
			t.Fatalf("special: expected 1, got %d", len(hex.Features.Special))
		}
		if hex.Features.Special[0].Name != "Foo" {
			t.Errorf("special name: got %q, want %q", hex.Features.Special[0].Name, "Foo")
		}
	})

	t.Run("non-special settlement kept", func(t *testing.T) {
		hex := &wxx.Hex{
			Features: wxx.Features{
				Settlements: []*domain.Settlement_t{{Name: "Gondor"}, {Name: "Foo"}},
			},
		}
		specials := map[string]*domain.Special_t{
			"foo": {Id: "foo", Name: "Foo"},
		}
		applySpecialHexes([]*wxx.Hex{hex}, specials)
		if len(hex.Features.Settlements) != 1 {
			t.Fatalf("settlements: expected 1, got %d", len(hex.Features.Settlements))
		}
		if hex.Features.Settlements[0].Name != "Gondor" {
			t.Errorf("settlement: got %q, want %q", hex.Features.Settlements[0].Name, "Gondor")
		}
		if len(hex.Features.Special) != 1 {
			t.Fatalf("special: expected 1, got %d", len(hex.Features.Special))
		}
	})

	t.Run("case insensitive match", func(t *testing.T) {
		hex := &wxx.Hex{
			Features: wxx.Features{
				Settlements: []*domain.Settlement_t{{Name: "FOO"}},
			},
		}
		specials := map[string]*domain.Special_t{
			"foo": {Id: "foo", Name: "Foo"},
		}
		applySpecialHexes([]*wxx.Hex{hex}, specials)
		if len(hex.Features.Settlements) != 0 {
			t.Errorf("settlements: expected 0, got %d", len(hex.Features.Settlements))
		}
		if len(hex.Features.Special) != 1 {
			t.Fatalf("special: expected 1, got %d", len(hex.Features.Special))
		}
	})

	t.Run("no duplicates in special", func(t *testing.T) {
		hex := &wxx.Hex{
			Features: wxx.Features{
				Settlements: []*domain.Settlement_t{{Name: "Foo"}, {Name: "Foo"}},
			},
		}
		specials := map[string]*domain.Special_t{
			"foo": {Id: "foo", Name: "Foo"},
		}
		applySpecialHexes([]*wxx.Hex{hex}, specials)
		if len(hex.Features.Settlements) != 0 {
			t.Errorf("settlements: expected 0, got %d", len(hex.Features.Settlements))
		}
		if len(hex.Features.Special) != 1 {
			t.Errorf("special: expected 1, got %d (duplicates not prevented)", len(hex.Features.Special))
		}
	})

	t.Run("empty specials map is no-op", func(t *testing.T) {
		hex := &wxx.Hex{
			Features: wxx.Features{
				Settlements: []*domain.Settlement_t{{Name: "Gondor"}},
			},
		}
		applySpecialHexes([]*wxx.Hex{hex}, nil)
		if len(hex.Features.Settlements) != 1 {
			t.Errorf("settlements: expected 1, got %d", len(hex.Features.Settlements))
		}
		if len(hex.Features.Special) != 0 {
			t.Errorf("special: expected 0, got %d", len(hex.Features.Special))
		}
	})

	t.Run("multiple hexes processed", func(t *testing.T) {
		hex1 := &wxx.Hex{
			Features: wxx.Features{
				Settlements: []*domain.Settlement_t{{Name: "Foo"}},
			},
		}
		hex2 := &wxx.Hex{
			Features: wxx.Features{
				Settlements: []*domain.Settlement_t{{Name: "Bar"}, {Name: "Foo"}},
			},
		}
		specials := map[string]*domain.Special_t{
			"foo": {Id: "foo", Name: "Foo"},
		}
		applySpecialHexes([]*wxx.Hex{hex1, hex2}, specials)
		if len(hex1.Features.Special) != 1 {
			t.Errorf("hex1 special: expected 1, got %d", len(hex1.Features.Special))
		}
		if len(hex2.Features.Settlements) != 1 || hex2.Features.Settlements[0].Name != "Bar" {
			t.Errorf("hex2 settlements: expected [Bar], got %v", hex2.Features.Settlements)
		}
		if len(hex2.Features.Special) != 1 {
			t.Errorf("hex2 special: expected 1, got %d", len(hex2.Features.Special))
		}
	})
}
