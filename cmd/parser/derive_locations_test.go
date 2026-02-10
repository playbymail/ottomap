// Copyright (c) 2026 Michael D Henderson. All rights reserved.

package main

import (
	"testing"

	schema "github.com/playbymail/ottomap/internal/tniif"
)

func TestBackwardWalkSteps(t *testing.T) {
	tests := []struct {
		name           string
		endingLocation schema.Coordinates
		steps          []schema.MoveStep
		wantLocations  []schema.Coordinates
	}{
		{
			name:           "simple advance chain — 3 steps all advancing N, all succeeded",
			endingLocation: "MM 1510",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirN, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirN, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirN, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"MM 1512", "MM 1511", "MM 1510"},
		},
		{
			name:           "mixed results — advance NE succeeded, advance SE failed, still",
			endingLocation: "MM 1510",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirNE, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirSE, Result: schema.ResultFailed},
				{Intent: schema.IntentStill, Still: true, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"MM 1510", "MM 1510", "MM 1510"},
		},
		{
			name:           "single still step",
			endingLocation: "GE 1009",
			steps: []schema.MoveStep{
				{Intent: schema.IntentStill, Still: true, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"GE 1009"},
		},
		{
			name:           "east border crossing — NE across grid boundary",
			endingLocation: "AB 0110",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirN, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirNE, Result: schema.ResultSucceeded},
				{Intent: schema.IntentStill, Still: true, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"AA 3010", "AB 0110", "AB 0110"},
		},
		{
			name:           "west border crossing — NW across grid boundary",
			endingLocation: "AA 3010",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirN, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirNW, Result: schema.ResultSucceeded},
				{Intent: schema.IntentStill, Still: true, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"AB 0111", "AA 3010", "AA 3010"},
		},
		{
			name:           "north border crossing — N across grid boundary",
			endingLocation: "AA 1521",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirN, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirN, Result: schema.ResultSucceeded},
				{Intent: schema.IntentStill, Still: true, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"BA 1501", "AA 1521", "AA 1521"},
		},
		{
			name:           "south border crossing — S across grid boundary",
			endingLocation: "BA 1501",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirS, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirS, Result: schema.ResultSucceeded},
				{Intent: schema.IntentStill, Still: true, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"AA 1521", "BA 1501", "BA 1501"},
		},
		{
			name:           "corner crossing NE — NE across both grid boundaries",
			endingLocation: "AB 0121",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirN, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirNE, Result: schema.ResultSucceeded},
				{Intent: schema.IntentStill, Still: true, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"AA 3021", "AB 0121", "AB 0121"},
		},
		{
			name:           "corner crossing SW — SW across both grid boundaries",
			endingLocation: "AA 3021",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirN, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirSW, Result: schema.ResultSucceeded},
				{Intent: schema.IntentStill, Still: true, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"AB 0121", "AA 3021", "AA 3021"},
		},
		{
			name:           "vanished result — advance SE vanished, no location change",
			endingLocation: "MM 1510",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirSE, Result: schema.ResultVanished},
			},
			wantLocations: []schema.Coordinates{"MM 1510"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			endLoc, err := parseCoord(tc.endingLocation)
			if err != nil {
				t.Fatalf("parseCoord(%q): %v", tc.endingLocation, err)
			}

			steps := make([]schema.MoveStep, len(tc.steps))
			copy(steps, tc.steps)

			backwardWalkSteps(steps, endLoc)

			for i, step := range steps {
				if step.EndingLocation != tc.wantLocations[i] {
					t.Errorf("step[%d].EndingLocation = %q, want %q", i, step.EndingLocation, tc.wantLocations[i])
				}
			}
		})
	}
}

func TestDeriveUnitMoveStepLocations(t *testing.T) {
	doc := &schema.Document{
		Clans: []schema.Clan{
			{
				ID: "0249",
				Units: []schema.Unit{
					{
						ID:             "4249",
						EndingLocation: "GE 1513",
						Moves: []schema.Moves{
							{
								ID: "4249",
								Steps: []schema.MoveStep{
									{Intent: schema.IntentAdvance, Advance: schema.DirSW, Result: schema.ResultSucceeded},
									{Intent: schema.IntentAdvance, Advance: schema.DirS, Result: schema.ResultSucceeded},
									{Intent: schema.IntentAdvance, Advance: schema.DirSE, Result: schema.ResultSucceeded},
									{Intent: schema.IntentAdvance, Advance: schema.DirNE, Result: schema.ResultSucceeded},
								},
							},
						},
					},
				},
			},
		},
	}

	DeriveLocations(doc)

	steps := doc.Clans[0].Units[0].Moves[0].Steps
	wantLocations := []schema.Coordinates{"GE 1312", "GE 1313", "GE 1413", "GE 1513"}

	for i, step := range steps {
		if step.EndingLocation != wantLocations[i] {
			t.Errorf("step[%d].EndingLocation = %q, want %q", i, step.EndingLocation, wantLocations[i])
		}
	}
}

func TestForwardWalkSteps(t *testing.T) {
	tests := []struct {
		name             string
		startingLocation schema.Coordinates
		steps            []schema.MoveStep
		wantLocations    []schema.Coordinates
	}{
		{
			name:             "simple advance chain — 3 steps all advancing S, all succeeded",
			startingLocation: "MM 1510",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirS, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirS, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirS, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"MM 1511", "MM 1512", "MM 1513"},
		},
		{
			name:             "failed step mid-chain — advance NE succeeded, advance NE failed, advance N succeeded",
			startingLocation: "MM 1510",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirNE, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirNE, Result: schema.ResultFailed},
				{Intent: schema.IntentAdvance, Advance: schema.DirN, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"MM 1609", "MM 1609", "MM 1608"},
		},
		{
			name:             "single still step",
			startingLocation: "MM 1510",
			steps: []schema.MoveStep{
				{Intent: schema.IntentStill, Still: true, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"MM 1510"},
		},
		{
			name:             "east border crossing — NE then SE across grid boundary",
			startingLocation: "AA 2910",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirNE, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirSE, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"AA 3009", "AB 0110"},
		},
		{
			name:             "west border crossing — NW then SW across grid boundary",
			startingLocation: "AB 0110",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirNW, Result: schema.ResultSucceeded},
				{Intent: schema.IntentAdvance, Advance: schema.DirSW, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"AA 3009", "AA 2910"},
		},
		{
			name:             "north border crossing — N across grid boundary",
			startingLocation: "BA 1501",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirN, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"AA 1521"},
		},
		{
			name:             "south border crossing — S across grid boundary",
			startingLocation: "AA 1521",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirS, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"BA 1501"},
		},
		{
			name:             "corner crossing — SE across both grid boundaries",
			startingLocation: "AA 3021",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirSE, Result: schema.ResultSucceeded},
			},
			wantLocations: []schema.Coordinates{"BB 0101"},
		},
		{
			name:             "vanished result — advance S vanished, no location change",
			startingLocation: "MM 1510",
			steps: []schema.MoveStep{
				{Intent: schema.IntentAdvance, Advance: schema.DirS, Result: schema.ResultVanished},
			},
			wantLocations: []schema.Coordinates{"MM 1510"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			startLoc, err := parseCoord(tc.startingLocation)
			if err != nil {
				t.Fatalf("parseCoord(%q): %v", tc.startingLocation, err)
			}

			steps := make([]schema.MoveStep, len(tc.steps))
			copy(steps, tc.steps)

			forwardWalkSteps(steps, startLoc)

			for i, step := range steps {
				if step.EndingLocation != tc.wantLocations[i] {
					t.Errorf("step[%d].EndingLocation = %q, want %q", i, step.EndingLocation, tc.wantLocations[i])
				}
			}
		})
	}
}

func TestDeriveScoutLocations(t *testing.T) {
	doc := &schema.Document{
		Clans: []schema.Clan{
			{
				ID: "0249",
				Units: []schema.Unit{
					{
						ID:             "4249",
						EndingLocation: "GE 1513",
						Scouts: []schema.ScoutRun{
							{
								ID: "4249s1",
								Steps: []schema.MoveStep{
									{Intent: schema.IntentAdvance, Advance: schema.DirNE, Result: schema.ResultSucceeded},
									{Intent: schema.IntentAdvance, Advance: schema.DirNE, Result: schema.ResultFailed},
								},
							},
							{
								ID: "4249s3",
								Steps: []schema.MoveStep{
									{Intent: schema.IntentAdvance, Advance: schema.DirSW, Result: schema.ResultSucceeded},
									{Intent: schema.IntentAdvance, Advance: schema.DirSW, Result: schema.ResultSucceeded},
									{Intent: schema.IntentAdvance, Advance: schema.DirSW, Result: schema.ResultFailed},
								},
							},
						},
					},
				},
			},
		},
	}

	DeriveLocations(doc)

	unit := doc.Clans[0].Units[0]

	if unit.Scouts[0].StartingLocation != "GE 1513" {
		t.Errorf("scout[0].StartingLocation = %q, want %q", unit.Scouts[0].StartingLocation, "GE 1513")
	}
	wantS1 := []schema.Coordinates{"GE 1612", "GE 1612"}
	for i, step := range unit.Scouts[0].Steps {
		if step.EndingLocation != wantS1[i] {
			t.Errorf("scout[0].step[%d].EndingLocation = %q, want %q", i, step.EndingLocation, wantS1[i])
		}
	}

	if unit.Scouts[1].StartingLocation != "GE 1513" {
		t.Errorf("scout[1].StartingLocation = %q, want %q", unit.Scouts[1].StartingLocation, "GE 1513")
	}
	wantS3 := []schema.Coordinates{"GE 1413", "GE 1314", "GE 1314"}
	for i, step := range unit.Scouts[1].Steps {
		if step.EndingLocation != wantS3[i] {
			t.Errorf("scout[1].step[%d].EndingLocation = %q, want %q", i, step.EndingLocation, wantS3[i])
		}
	}
}

func TestDeriveLocationsSkipsInvalidEndingLocation(t *testing.T) {
	doc := &schema.Document{
		Clans: []schema.Clan{
			{
				ID: "0249",
				Units: []schema.Unit{
					{
						ID:             "0249",
						EndingLocation: "N/A",
						Moves: []schema.Moves{
							{
								ID: "0249",
								Steps: []schema.MoveStep{
									{Intent: schema.IntentStill, Still: true, Result: schema.ResultSucceeded},
								},
							},
						},
					},
				},
			},
		},
	}

	DeriveLocations(doc)

	step := doc.Clans[0].Units[0].Moves[0].Steps[0]
	if step.EndingLocation != "" {
		t.Errorf("step.EndingLocation = %q, want empty (skipped)", step.EndingLocation)
	}
	if len(doc.Notes) == 0 {
		t.Error("expected a warning note for invalid ending location")
	}
}

func TestDeriveObservationLocations(t *testing.T) {
	t.Run("observation present on succeeded step", func(t *testing.T) {
		doc := &schema.Document{
			Clans: []schema.Clan{{
				ID: "0249",
				Units: []schema.Unit{{
					ID:             "4249",
					EndingLocation: "GE 1513",
					Moves: []schema.Moves{{
						ID: "4249",
						Steps: []schema.MoveStep{
							{
								Intent:  schema.IntentAdvance,
								Advance: schema.DirSW,
								Result:  schema.ResultSucceeded,
								Observation: &schema.Observation{
									Terrain: "JG",
								},
							},
							{
								Intent:  schema.IntentAdvance,
								Advance: schema.DirNE,
								Result:  schema.ResultSucceeded,
								Observation: &schema.Observation{
									Terrain: "PR",
								},
							},
						},
					}},
				}},
			}},
		}

		DeriveLocations(doc)

		steps := doc.Clans[0].Units[0].Moves[0].Steps
		if steps[0].Observation.Location != steps[0].EndingLocation {
			t.Errorf("step[0].Observation.Location = %q, want %q", steps[0].Observation.Location, steps[0].EndingLocation)
		}
		if steps[1].Observation.Location != steps[1].EndingLocation {
			t.Errorf("step[1].Observation.Location = %q, want %q", steps[1].Observation.Location, steps[1].EndingLocation)
		}
	})

	t.Run("observation present on failed step", func(t *testing.T) {
		doc := &schema.Document{
			Clans: []schema.Clan{{
				ID: "0249",
				Units: []schema.Unit{{
					ID:             "4249",
					EndingLocation: "GE 1513",
					Moves: []schema.Moves{{
						ID: "4249",
						Steps: []schema.MoveStep{
							{
								Intent:  schema.IntentAdvance,
								Advance: schema.DirSE,
								Result:  schema.ResultFailed,
								Observation: &schema.Observation{
									Terrain: "PR",
								},
							},
						},
					}},
				}},
			}},
		}

		DeriveLocations(doc)

		step := doc.Clans[0].Units[0].Moves[0].Steps[0]
		if step.Observation.Location != step.EndingLocation {
			t.Errorf("observation.Location = %q, want %q", step.Observation.Location, step.EndingLocation)
		}
		if step.EndingLocation != "GE 1513" {
			t.Errorf("step.EndingLocation = %q, want %q (failed step should stay)", step.EndingLocation, "GE 1513")
		}
	})

	t.Run("no observation — nil left alone", func(t *testing.T) {
		doc := &schema.Document{
			Clans: []schema.Clan{{
				ID: "0249",
				Units: []schema.Unit{{
					ID:             "4249",
					EndingLocation: "GE 1513",
					Moves: []schema.Moves{{
						ID: "4249",
						Steps: []schema.MoveStep{
							{
								Intent: schema.IntentStill,
								Still:  true,
								Result: schema.ResultSucceeded,
							},
						},
					}},
				}},
			}},
		}

		DeriveLocations(doc)

		step := doc.Clans[0].Units[0].Moves[0].Steps[0]
		if step.Observation != nil {
			t.Errorf("expected nil observation, got %+v", step.Observation)
		}
	})

	t.Run("multiple observations across steps get correct locations", func(t *testing.T) {
		doc := &schema.Document{
			Clans: []schema.Clan{{
				ID: "0249",
				Units: []schema.Unit{{
					ID:             "4249",
					EndingLocation: "GE 1513",
					Moves: []schema.Moves{{
						ID: "4249",
						Steps: []schema.MoveStep{
							{
								Intent:  schema.IntentAdvance,
								Advance: schema.DirSW,
								Result:  schema.ResultSucceeded,
								Observation: &schema.Observation{
									Terrain: "JG",
								},
							},
							{
								Intent:  schema.IntentAdvance,
								Advance: schema.DirS,
								Result:  schema.ResultSucceeded,
								Observation: &schema.Observation{
									Terrain: "JG",
								},
							},
							{
								Intent:  schema.IntentAdvance,
								Advance: schema.DirSE,
								Result:  schema.ResultSucceeded,
								Observation: &schema.Observation{
									Terrain: "PR",
								},
							},
							{
								Intent:  schema.IntentAdvance,
								Advance: schema.DirNE,
								Result:  schema.ResultSucceeded,
								Observation: &schema.Observation{
									Terrain: "PR",
								},
							},
						},
					}},
				}},
			}},
		}

		DeriveLocations(doc)

		steps := doc.Clans[0].Units[0].Moves[0].Steps
		wantLocations := []schema.Coordinates{"GE 1312", "GE 1313", "GE 1413", "GE 1513"}
		for i, step := range steps {
			if step.Observation.Location != wantLocations[i] {
				t.Errorf("step[%d].Observation.Location = %q, want %q", i, step.Observation.Location, wantLocations[i])
			}
			if step.Observation.Location != step.EndingLocation {
				t.Errorf("step[%d].Observation.Location = %q != step.EndingLocation = %q", i, step.Observation.Location, step.EndingLocation)
			}
		}
	})

	t.Run("scout observations get correct locations", func(t *testing.T) {
		doc := &schema.Document{
			Clans: []schema.Clan{{
				ID: "0249",
				Units: []schema.Unit{{
					ID:             "4249",
					EndingLocation: "GE 1513",
					Scouts: []schema.ScoutRun{{
						ID: "4249s1",
						Steps: []schema.MoveStep{
							{
								Intent:  schema.IntentAdvance,
								Advance: schema.DirNE,
								Result:  schema.ResultSucceeded,
								Observation: &schema.Observation{
									Terrain: "PR",
								},
							},
							{
								Intent:  schema.IntentAdvance,
								Advance: schema.DirNE,
								Result:  schema.ResultFailed,
								Observation: &schema.Observation{
									Terrain: "PR",
								},
							},
						},
					}},
				}},
			}},
		}

		DeriveLocations(doc)

		steps := doc.Clans[0].Units[0].Scouts[0].Steps
		wantLocations := []schema.Coordinates{"GE 1612", "GE 1612"}
		for i, step := range steps {
			if step.Observation.Location != wantLocations[i] {
				t.Errorf("scout step[%d].Observation.Location = %q, want %q", i, step.Observation.Location, wantLocations[i])
			}
		}
	})
}

func TestDeriveCompassPointLocations(t *testing.T) {
	t.Run("all 12 bearings from central hex", func(t *testing.T) {
		obs := &schema.Observation{
			Location: "MM 1510",
			CompassPoints: []schema.CompassPoint{
				{Bearing: "North"},
				{Bearing: "NorthNorthEast"},
				{Bearing: "NorthEast"},
				{Bearing: "East"},
				{Bearing: "SouthEast"},
				{Bearing: "SouthSouthEast"},
				{Bearing: "South"},
				{Bearing: "SouthSouthWest"},
				{Bearing: "SouthWest"},
				{Bearing: "West"},
				{Bearing: "NorthWest"},
				{Bearing: "NorthNorthWest"},
			},
		}
		doc := &schema.Document{
			Clans: []schema.Clan{{
				ID: "0249",
				Units: []schema.Unit{{
					ID:             "0249f1",
					EndingLocation: "MM 1510",
					Moves: []schema.Moves{{
						ID: "0249f1",
						Steps: []schema.MoveStep{{
							Intent:      schema.IntentStill,
							Still:       true,
							Result:      schema.ResultSucceeded,
							Observation: obs,
						}},
					}},
				}},
			}},
		}

		DeriveLocations(doc)

		wantLocations := []schema.Coordinates{
			"MM 1508", // N
			"MM 1608", // NNE
			"MM 1709", // NE
			"MM 1710", // E
			"MM 1711", // SE
			"MM 1611", // SSE
			"MM 1512", // S
			"MM 1411", // SSW
			"MM 1311", // SW
			"MM 1310", // W
			"MM 1309", // NW
			"MM 1408", // NNW
		}
		step := doc.Clans[0].Units[0].Moves[0].Steps[0]
		for i, cp := range step.Observation.CompassPoints {
			if cp.Location != wantLocations[i] {
				t.Errorf("bearing %q: got %q, want %q", cp.Bearing, cp.Location, wantLocations[i])
			}
		}
	})

	t.Run("east bearing from even column", func(t *testing.T) {
		obs := &schema.Observation{
			Location:      "MM 1510",
			CompassPoints: []schema.CompassPoint{{Bearing: "East"}},
		}
		doc := docWithObservation("MM 1510", obs)
		DeriveLocations(doc)
		cp := doc.Clans[0].Units[0].Moves[0].Steps[0].Observation.CompassPoints[0]
		if cp.Location != "MM 1710" {
			t.Errorf("E from even col: got %q, want %q", cp.Location, "MM 1710")
		}
	})

	t.Run("east bearing from odd column", func(t *testing.T) {
		obs := &schema.Observation{
			Location:      "MM 1610",
			CompassPoints: []schema.CompassPoint{{Bearing: "East"}},
		}
		doc := docWithObservation("MM 1610", obs)
		DeriveLocations(doc)
		cp := doc.Clans[0].Units[0].Moves[0].Steps[0].Observation.CompassPoints[0]
		if cp.Location != "MM 1810" {
			t.Errorf("E from odd col: got %q, want %q", cp.Location, "MM 1810")
		}
	})

	t.Run("west bearing from even column", func(t *testing.T) {
		obs := &schema.Observation{
			Location:      "MM 1510",
			CompassPoints: []schema.CompassPoint{{Bearing: "West"}},
		}
		doc := docWithObservation("MM 1510", obs)
		DeriveLocations(doc)
		cp := doc.Clans[0].Units[0].Moves[0].Steps[0].Observation.CompassPoints[0]
		if cp.Location != "MM 1310" {
			t.Errorf("W from even col: got %q, want %q", cp.Location, "MM 1310")
		}
	})

	t.Run("west bearing from odd column", func(t *testing.T) {
		obs := &schema.Observation{
			Location:      "MM 1610",
			CompassPoints: []schema.CompassPoint{{Bearing: "West"}},
		}
		doc := docWithObservation("MM 1610", obs)
		DeriveLocations(doc)
		cp := doc.Clans[0].Units[0].Moves[0].Steps[0].Observation.CompassPoints[0]
		if cp.Location != "MM 1410" {
			t.Errorf("W from odd col: got %q, want %q", cp.Location, "MM 1410")
		}
	})

	t.Run("north bearing at grid top edge", func(t *testing.T) {
		obs := &schema.Observation{
			Location:      "BA 1502",
			CompassPoints: []schema.CompassPoint{{Bearing: "North"}},
		}
		doc := docWithObservation("BA 1502", obs)
		DeriveLocations(doc)
		cp := doc.Clans[0].Units[0].Moves[0].Steps[0].Observation.CompassPoints[0]
		if cp.Location != "AA 1521" {
			t.Errorf("N from grid top edge: got %q, want %q", cp.Location, "AA 1521")
		}
	})

	t.Run("south bearing at grid bottom edge", func(t *testing.T) {
		obs := &schema.Observation{
			Location:      "AA 1520",
			CompassPoints: []schema.CompassPoint{{Bearing: "South"}},
		}
		doc := docWithObservation("AA 1520", obs)
		DeriveLocations(doc)
		cp := doc.Clans[0].Units[0].Moves[0].Steps[0].Observation.CompassPoints[0]
		if cp.Location != "BA 1501" {
			t.Errorf("S from grid bottom edge: got %q, want %q", cp.Location, "BA 1501")
		}
	})

	t.Run("east bearing at grid right edge", func(t *testing.T) {
		obs := &schema.Observation{
			Location:      "AA 2910",
			CompassPoints: []schema.CompassPoint{{Bearing: "East"}},
		}
		doc := docWithObservation("AA 2910", obs)
		DeriveLocations(doc)
		cp := doc.Clans[0].Units[0].Moves[0].Steps[0].Observation.CompassPoints[0]
		if cp.Location != "AB 0110" {
			t.Errorf("E from grid right edge: got %q, want %q", cp.Location, "AB 0110")
		}
	})

	t.Run("corner case — SE crosses both grid boundaries", func(t *testing.T) {
		obs := &schema.Observation{
			Location:      "AA 3020",
			CompassPoints: []schema.CompassPoint{{Bearing: "SouthEast"}},
		}
		doc := docWithObservation("AA 3020", obs)
		DeriveLocations(doc)
		cp := doc.Clans[0].Units[0].Moves[0].Steps[0].Observation.CompassPoints[0]
		if cp.Location != "AB 0221" {
			t.Errorf("SE from corner: got %q, want %q", cp.Location, "AB 0221")
		}
	})

	t.Run("out-of-bounds produces warning and empty location", func(t *testing.T) {
		obs := &schema.Observation{
			Location:      "AA 0101",
			CompassPoints: []schema.CompassPoint{{Bearing: "NorthWest"}},
		}
		doc := docWithObservation("AA 0101", obs)
		DeriveLocations(doc)
		cp := doc.Clans[0].Units[0].Moves[0].Steps[0].Observation.CompassPoints[0]
		if cp.Location != "" {
			t.Errorf("OOB: got %q, want empty", cp.Location)
		}
		if len(doc.Notes) == 0 {
			t.Error("expected a warning note for out-of-bounds compass point")
		}
	})

	t.Run("no compass points — no panic", func(t *testing.T) {
		doc := &schema.Document{
			Clans: []schema.Clan{{
				ID: "0249",
				Units: []schema.Unit{{
					ID:             "0249f1",
					EndingLocation: "MM 1510",
					Moves: []schema.Moves{{
						ID: "0249f1",
						Steps: []schema.MoveStep{{
							Intent: schema.IntentStill,
							Still:  true,
							Result: schema.ResultSucceeded,
							Observation: &schema.Observation{
								Terrain: "SH",
							},
						}},
					}},
				}},
			}},
		}
		DeriveLocations(doc)
	})
}

func docWithObservation(loc schema.Coordinates, obs *schema.Observation) *schema.Document {
	return &schema.Document{
		Clans: []schema.Clan{{
			ID: "0249",
			Units: []schema.Unit{{
				ID:             "0249f1",
				EndingLocation: loc,
				Moves: []schema.Moves{{
					ID: "0249f1",
					Steps: []schema.MoveStep{{
						Intent:      schema.IntentStill,
						Still:       true,
						Result:      schema.ResultSucceeded,
						Observation: obs,
					}},
				}},
			}},
		}},
	}
}
