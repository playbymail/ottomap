// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"strings"
	"testing"

	schema "github.com/playbymail/ottomap/internal/tniif"
)

func TestValidate_InvalidCoords(t *testing.T) {
	t.Parallel()

	validDoc := func() loadedDoc_t {
		return loadedDoc_t{
			File: "0300.0904-01.0987.json",
			Doc: schema.Document{
				Schema: schema.Version,
				Game:   "0300",
				Turn:   "0904-01",
				Clan:   "0987",
			},
		}
	}

	t.Run("invalid endingLocation", func(t *testing.T) {
		t.Parallel()
		ld := validDoc()
		ld.Doc.Clans = []schema.Clan{{
			ID: "0987",
			Units: []schema.Unit{{
				ID:             "0987",
				EndingLocation: "BADCOORD",
			}},
		}}
		errs := validateDocuments([]loadedDoc_t{ld})
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

	t.Run("invalid observation location", func(t *testing.T) {
		t.Parallel()
		ld := validDoc()
		ld.Doc.Clans = []schema.Clan{{
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
							Location: "## 1316",
						},
					}},
				}},
			}},
		}}
		errs := validateDocuments([]loadedDoc_t{ld})
		found := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "location") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected observation location error, got: %v", errs)
		}
	})

	t.Run("invalid edge direction", func(t *testing.T) {
		t.Parallel()
		ld := validDoc()
		ld.Doc.Clans = []schema.Clan{{
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
							Edges:    []schema.Edge{{Dir: "NORTH"}},
						},
					}},
				}},
			}},
		}}
		errs := validateDocuments([]loadedDoc_t{ld})
		found := false
		for _, e := range errs {
			if strings.Contains(e.Error(), "direction") || strings.Contains(e.Error(), "edges") {
				found = true
			}
		}
		if !found {
			t.Errorf("expected edge direction error, got: %v", errs)
		}
	})

	t.Run("valid coordinates pass", func(t *testing.T) {
		t.Parallel()
		ld := validDoc()
		ld.Doc.Clans = []schema.Clan{{
			ID: "0987",
			Units: []schema.Unit{{
				ID:             "0987",
				EndingLocation: "AA 0101",
			}},
		}}
		errs := validateDocuments([]loadedDoc_t{ld})
		if len(errs) != 0 {
			t.Errorf("expected 0 errors, got %d: %v", len(errs), errs)
		}
	})
}

func TestValidate_MismatchedGame(t *testing.T) {
	t.Parallel()

	docs := []loadedDoc_t{
		{File: "0300.0904-01.0987.json", Doc: schema.Document{Schema: schema.Version, Game: "0300", Turn: "0904-01", Clan: "0987"}},
		{File: "0301.0904-01.0138.json", Doc: schema.Document{Schema: schema.Version, Game: "0301", Turn: "0904-01", Clan: "0138"}},
	}
	errs := validateDocuments(docs)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "game") && strings.Contains(e.Error(), "does not match") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected mismatched game error, got: %v", errs)
	}
}

func TestValidate_InvalidSchema(t *testing.T) {
	t.Parallel()

	docs := []loadedDoc_t{
		{File: "0300.0904-01.0987.json", Doc: schema.Document{Schema: "wrong-version", Game: "0300", Turn: "0904-01", Clan: "0987"}},
	}
	errs := validateDocuments(docs)
	found := false
	for _, e := range errs {
		if strings.Contains(e.Error(), "schema") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected schema version error, got: %v", errs)
	}
}
