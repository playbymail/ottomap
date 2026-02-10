// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package parser_test

import (
	"github.com/go-test/deep"
	"github.com/playbymail/ottomap/internal/compass"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/domain"
	"github.com/playbymail/ottomap/internal/edges"
	"github.com/playbymail/ottomap/internal/parser"
	"github.com/playbymail/ottomap/internal/resources"
	"github.com/playbymail/ottomap/internal/results"
	"github.com/playbymail/ottomap/internal/terrain"
	"testing"
)

func TestCompassPoint(t *testing.T) {
	var pt compass.Point_e
	for _, tc := range []struct {
		id    string
		line  string
		point compass.Point_e
	}{
		{id: "N/N", line: "N/N", point: compass.North},
		{id: "N/NE", line: "N/NE", point: compass.NorthNorthEast},
		{id: "NE/NE", line: "NE/NE", point: compass.NorthEast},
		{id: "NE/SE", line: "NE/SE", point: compass.East},
		{id: "SE/SE", line: "SE/SE", point: compass.SouthEast},
		{id: "S/SE", line: "S/SE", point: compass.SouthSouthEast},
		{id: "S/S", line: "S/S", point: compass.South},
		{id: "S/SW", line: "S/SW", point: compass.SouthSouthWest},
		{id: "SW/SW", line: "SW/SW", point: compass.SouthWest},
		{id: "SW/NW", line: "SW/NW", point: compass.West},
		{id: "NW/NW", line: "NW/NW", point: compass.NorthWest},
		{id: "N/NW", line: "N/NW", point: compass.NorthNorthWest},
	} {
		va, err := parser.Parse(tc.id, []byte(tc.line), parser.Entrypoint("COMPASSPOINT"))
		if err != nil {
			t.Errorf("id %q: parse failed %v\n", tc.id, err)
			continue
		}
		point, ok := va.(compass.Point_e)
		if !ok {
			t.Errorf("id %q: type: want %T, got %T\n", tc.id, pt, va)
			continue
		}
		if tc.point != point {
			t.Errorf("id %q: point: want %q, got %q\n", tc.id, tc.point, point)
		}
	}
}

func TestCrowsNestObservation(t *testing.T) {
	var fh domain.FarHorizon_t
	for _, tc := range []struct {
		id      string
		line    string
		point   compass.Point_e
		terrain terrain.Terrain_e
	}{
		{id: "land", line: "Sight Land - N/N", point: compass.North, terrain: terrain.UnknownLand},
		{id: "water", line: "Sight Water - N/NE", point: compass.NorthNorthEast, terrain: terrain.UnknownWater},
	} {
		va, err := parser.Parse(tc.id, []byte(tc.line), parser.Entrypoint("CrowsNestObservation"))
		if err != nil {
			t.Errorf("id %q: parse failed %v\n", tc.id, err)
			continue
		}
		cno, ok := va.(domain.FarHorizon_t)
		if !ok {
			t.Errorf("id %q: type: want %T, got %T\n", tc.id, fh, va)
			continue
		}
		if tc.point != cno.Point {
			t.Errorf("id %q: point: want %q, got %q\n", tc.id, tc.point, cno.Point)
		}
		if tc.terrain != cno.Terrain {
			t.Errorf("id %q: terrain: want %q, got %q\n", tc.id, tc.terrain, cno.Terrain)
		}
	}
}

// clearNewMoveFields zeroes out fields added after the original tests were
// written so the deep comparison focuses on the original parser behavior.
func clearNewMoveFields(moves []*domain.Move_t) {
	for _, m := range moves {
		m.UnitId = ""
		m.Debug.PriorMove = nil
		m.Debug.NextMove = nil
		if m.Report != nil {
			m.Report.UnitId = ""
		}
	}
}

func TestFleetMovementParse(t *testing.T) {
	for _, tc := range []struct {
		id     string
		line   string
		unitId domain.UnitId_t
		moves  []*domain.Move_t
		debug  bool
	}{
		{id: "900-05.0138f2",
			line:   `STRONG S Fleet Movement: Move NW-GH,`,
			unitId: "0138f2",
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("NW-GH"),
					Result: results.Succeeded, Advance: direction.NorthWest, Report: &domain.Report_t{
						Terrain: terrain.HillsGrassy,
					},
				},
			},
		},
		{id: "900-06.0138f4",
			line:   `MILD NW Fleet Movement: Move NE-LCM,  Lcm NE, SE, S,\NE-LCM,  Lcm NE, SE, SW, S,\NE-LCM,  Lcm NE, SE, SW, S,\`,
			unitId: "0138f4",
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("NE-LCM,  Lcm NE, SE, S"),
					Result: results.Succeeded, Advance: direction.NorthEast, Report: &domain.Report_t{
						Terrain: terrain.LowMountainsConifer,
						Borders: []*domain.Border_t{
							{Direction: direction.NorthEast, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.SouthEast, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.South, Terrain: terrain.LowMountainsConifer},
						},
					},
				},
				{LineNo: 1, StepNo: 2, Line: []byte("NE-LCM,  Lcm NE, SE, SW, S"),
					Result: results.Succeeded, Advance: direction.NorthEast, Report: &domain.Report_t{
						Terrain: terrain.LowMountainsConifer,
						Borders: []*domain.Border_t{
							{Direction: direction.NorthEast, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.SouthEast, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.South, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.SouthWest, Terrain: terrain.LowMountainsConifer},
						},
					},
				},
				{LineNo: 1, StepNo: 3, Line: []byte("NE-LCM,  Lcm NE, SE, SW, S"),
					Result: results.Succeeded, Advance: direction.NorthEast, Report: &domain.Report_t{
						Terrain: terrain.LowMountainsConifer,
						Borders: []*domain.Border_t{
							{Direction: direction.NorthEast, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.SouthEast, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.South, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.SouthWest, Terrain: terrain.LowMountainsConifer},
						},
					},
				},
			},
		},
		{id: "900-06.0138f1",
			line:   `MILD NW Fleet Movement: Move SE-O,-(NE O,  SE LCM,  N O,  S LCM,  SW O,  NW O,  )(Sight Water - N/N, Sight Land - N/NE)`,
			unitId: "0138f1",
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("SE-O,-(NE O,  SE LCM,  N O,  S LCM,  SW O,  NW O,  )(Sight Water - N/N, Sight Land - N/NE)"),
					Result: results.Succeeded, Advance: direction.SouthEast, Report: &domain.Report_t{
						Terrain: terrain.WaterOcean,
						Borders: []*domain.Border_t{
							{Direction: direction.North, Terrain: terrain.WaterOcean},
							{Direction: direction.NorthEast, Terrain: terrain.WaterOcean},
							{Direction: direction.SouthEast, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.South, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.SouthWest, Terrain: terrain.WaterOcean},
							{Direction: direction.NorthWest, Terrain: terrain.WaterOcean},
						},
						FarHorizons: []*domain.FarHorizon_t{
							{Point: compass.North, Terrain: terrain.UnknownWater},
							{Point: compass.NorthNorthEast, Terrain: terrain.UnknownLand},
						},
					},
				},
			},
		},
		{id: "900-06.0138f2",
			line:   `MILD NW Fleet Movement: Move SE-O,-(NE O)(Sight Water - N/N, Sight Land - N/NE)\No River Adjacent to Hex to SW of HEX`,
			unitId: "0138f1",
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("SE-O,-(NE O)(Sight Water - N/N, Sight Land - N/NE)"),
					Result: results.Succeeded, Advance: direction.SouthEast, Report: &domain.Report_t{
						Terrain: terrain.WaterOcean,
						Borders: []*domain.Border_t{
							{Direction: direction.NorthEast, Terrain: terrain.WaterOcean},
						},
						FarHorizons: []*domain.FarHorizon_t{
							{Point: compass.North, Terrain: terrain.UnknownWater},
							{Point: compass.NorthNorthEast, Terrain: terrain.UnknownLand},
						},
					},
				},
				{LineNo: 1, StepNo: 2, Line: []byte("No River Adjacent to Hex to SW of HEX"),
					Result: results.Failed, Still: true, Advance: direction.SouthWest, Report: &domain.Report_t{},
				},
			},
		},
		{id: "900-06.1138f7",
			line:   `MILD N Fleet Movement: Move SW-PR The Dirty Squirrel-(NE GH,  SE O, N GH, S O, SW O, NW O, )(Sight Land - N/N,Sight Land - N/NE,Sight Land - N/NW,Sight Water - NE/NE,Sight Water - NE/SE,Sight Water - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )\NW-O, -(NE GH, SE PR, N SW, S O, SW O, NW O, )(Sight Water - N/N,Sight Land - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Water - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )\NW-O, -(NE SW, SE O, N O, S O, SW O, NW O, )(Sight Water - N/N,Sight Water - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Land - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )\N-O, -(NE O, SE SW, N O, S O, SW O, NW O, )(Sight Land - N/N,Sight Land - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Land - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )\N-O,  Lcm NE, N,-(NE LCM, SE O, N LCM, S O, SW O, NW O, )(Sight Land - N/N,Sight Land - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Land - SE/SE,Sight Land - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )\N-LCM,  Lcm NE, SE,  Ensalada sin Tomate\`,
			unitId: "0138f7",
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("SW-PR The Dirty Squirrel-(NE GH,  SE O, N GH, S O, SW O, NW O, )(Sight Land - N/N,Sight Land - N/NE,Sight Land - N/NW,Sight Water - NE/NE,Sight Water - NE/SE,Sight Water - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )"),
					Result: results.Succeeded, Advance: direction.SouthWest, Report: &domain.Report_t{
						Terrain: terrain.FlatPrairie,
						Borders: []*domain.Border_t{
							{Direction: direction.North, Terrain: terrain.HillsGrassy},
							{Direction: direction.NorthEast, Terrain: terrain.HillsGrassy},
							{Direction: direction.SouthEast, Terrain: terrain.WaterOcean},
							{Direction: direction.South, Terrain: terrain.WaterOcean},
							{Direction: direction.SouthWest, Terrain: terrain.WaterOcean},
							{Direction: direction.NorthWest, Terrain: terrain.WaterOcean},
						},
						FarHorizons: []*domain.FarHorizon_t{
							{Point: compass.North, Terrain: terrain.UnknownLand},
							{Point: compass.NorthNorthEast, Terrain: terrain.UnknownLand},
							{Point: compass.NorthEast, Terrain: terrain.UnknownWater},
							{Point: compass.East, Terrain: terrain.UnknownWater},
							{Point: compass.SouthEast, Terrain: terrain.UnknownWater},
							{Point: compass.SouthSouthEast, Terrain: terrain.UnknownWater},
							{Point: compass.South, Terrain: terrain.UnknownWater},
							{Point: compass.SouthSouthWest, Terrain: terrain.UnknownWater},
							{Point: compass.SouthWest, Terrain: terrain.UnknownWater},
							{Point: compass.West, Terrain: terrain.UnknownWater},
							{Point: compass.NorthWest, Terrain: terrain.UnknownWater},
							{Point: compass.NorthNorthWest, Terrain: terrain.UnknownLand},
						},
						Settlements: []*domain.Settlement_t{{Name: "The Dirty Squirrel"}},
					},
				},
				{LineNo: 1, StepNo: 2, Line: []byte("NW-O, -(NE GH, SE PR, N SW, S O, SW O, NW O, )(Sight Water - N/N,Sight Land - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Water - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )"),
					Result: results.Succeeded, Advance: direction.NorthWest, Report: &domain.Report_t{
						Terrain: terrain.WaterOcean,
						Borders: []*domain.Border_t{
							{Direction: direction.North, Terrain: terrain.FlatSwamp},
							{Direction: direction.NorthEast, Terrain: terrain.HillsGrassy},
							{Direction: direction.SouthEast, Terrain: terrain.FlatPrairie},
							{Direction: direction.South, Terrain: terrain.WaterOcean},
							{Direction: direction.SouthWest, Terrain: terrain.WaterOcean},
							{Direction: direction.NorthWest, Terrain: terrain.WaterOcean},
						},
						FarHorizons: []*domain.FarHorizon_t{
							{Point: compass.North, Terrain: terrain.UnknownWater},
							{Point: compass.NorthNorthEast, Terrain: terrain.UnknownLand},
							{Point: compass.NorthEast, Terrain: terrain.UnknownLand},
							{Point: compass.East, Terrain: terrain.UnknownLand},
							{Point: compass.SouthEast, Terrain: terrain.UnknownWater},
							{Point: compass.SouthSouthEast, Terrain: terrain.UnknownWater},
							{Point: compass.South, Terrain: terrain.UnknownWater},
							{Point: compass.SouthSouthWest, Terrain: terrain.UnknownWater},
							{Point: compass.SouthWest, Terrain: terrain.UnknownWater},
							{Point: compass.West, Terrain: terrain.UnknownWater},
							{Point: compass.NorthWest, Terrain: terrain.UnknownWater},
							{Point: compass.NorthNorthWest, Terrain: terrain.UnknownWater},
						},
					},
				},
				{LineNo: 1, StepNo: 3, Line: []byte("NW-O, -(NE SW, SE O, N O, S O, SW O, NW O, )(Sight Water - N/N,Sight Water - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Land - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )"),
					Result: results.Succeeded, Advance: direction.NorthWest, Report: &domain.Report_t{
						Terrain: terrain.WaterOcean,
						Borders: []*domain.Border_t{
							{Direction: direction.North, Terrain: terrain.WaterOcean},
							{Direction: direction.NorthEast, Terrain: terrain.FlatSwamp},
							{Direction: direction.SouthEast, Terrain: terrain.WaterOcean},
							{Direction: direction.South, Terrain: terrain.WaterOcean},
							{Direction: direction.SouthWest, Terrain: terrain.WaterOcean},
							{Direction: direction.NorthWest, Terrain: terrain.WaterOcean},
						},
						FarHorizons: []*domain.FarHorizon_t{
							{Point: compass.North, Terrain: terrain.UnknownWater},
							{Point: compass.NorthNorthEast, Terrain: terrain.UnknownWater},
							{Point: compass.NorthEast, Terrain: terrain.UnknownLand},
							{Point: compass.East, Terrain: terrain.UnknownLand},
							{Point: compass.SouthEast, Terrain: terrain.UnknownLand},
							{Point: compass.SouthSouthEast, Terrain: terrain.UnknownWater},
							{Point: compass.South, Terrain: terrain.UnknownWater},
							{Point: compass.SouthSouthWest, Terrain: terrain.UnknownWater},
							{Point: compass.SouthWest, Terrain: terrain.UnknownWater},
							{Point: compass.West, Terrain: terrain.UnknownWater},
							{Point: compass.NorthWest, Terrain: terrain.UnknownWater},
							{Point: compass.NorthNorthWest, Terrain: terrain.UnknownWater},
						},
					},
				},
				{LineNo: 1, StepNo: 4, Line: []byte("N-O, -(NE O, SE SW, N O, S O, SW O, NW O, )(Sight Land - N/N,Sight Land - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Land - SE/SE,Sight Water - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )"),
					Result: results.Succeeded, Advance: direction.North, Report: &domain.Report_t{
						Terrain: terrain.WaterOcean,
						Borders: []*domain.Border_t{
							{Direction: direction.North, Terrain: terrain.WaterOcean},
							{Direction: direction.NorthEast, Terrain: terrain.WaterOcean},
							{Direction: direction.SouthEast, Terrain: terrain.FlatSwamp},
							{Direction: direction.South, Terrain: terrain.WaterOcean},
							{Direction: direction.SouthWest, Terrain: terrain.WaterOcean},
							{Direction: direction.NorthWest, Terrain: terrain.WaterOcean},
						},
						FarHorizons: []*domain.FarHorizon_t{
							{Point: compass.North, Terrain: terrain.UnknownLand},
							{Point: compass.NorthNorthEast, Terrain: terrain.UnknownLand},
							{Point: compass.NorthEast, Terrain: terrain.UnknownLand},
							{Point: compass.East, Terrain: terrain.UnknownLand},
							{Point: compass.SouthEast, Terrain: terrain.UnknownLand},
							{Point: compass.SouthSouthEast, Terrain: terrain.UnknownWater},
							{Point: compass.South, Terrain: terrain.UnknownWater},
							{Point: compass.SouthSouthWest, Terrain: terrain.UnknownWater},
							{Point: compass.SouthWest, Terrain: terrain.UnknownWater},
							{Point: compass.West, Terrain: terrain.UnknownWater},
							{Point: compass.NorthWest, Terrain: terrain.UnknownWater},
							{Point: compass.NorthNorthWest, Terrain: terrain.UnknownWater},
						},
					},
				},
				{LineNo: 1, StepNo: 5, Line: []byte("N-O,  Lcm NE, N,-(NE LCM, SE O, N LCM, S O, SW O, NW O, )(Sight Land - N/N,Sight Land - N/NE,Sight Water - N/NW,Sight Land - NE/NE,Sight Land - NE/SE,Sight Land - SE/SE,Sight Land - S/SE,Sight Water - S/S,Sight Water - S/SW,Sight Water - SW/SW,Sight Water - SW/NW,Sight Water - NW/NW, )"),
					Result: results.Succeeded, Advance: direction.North, Report: &domain.Report_t{
						Terrain: terrain.WaterOcean,
						Borders: []*domain.Border_t{
							{Direction: direction.North, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.NorthEast, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.SouthEast, Terrain: terrain.WaterOcean},
							{Direction: direction.South, Terrain: terrain.WaterOcean},
							{Direction: direction.SouthWest, Terrain: terrain.WaterOcean},
							{Direction: direction.NorthWest, Terrain: terrain.WaterOcean},
						},
						FarHorizons: []*domain.FarHorizon_t{
							{Point: compass.North, Terrain: terrain.UnknownLand},
							{Point: compass.NorthNorthEast, Terrain: terrain.UnknownLand},
							{Point: compass.NorthEast, Terrain: terrain.UnknownLand},
							{Point: compass.East, Terrain: terrain.UnknownLand},
							{Point: compass.SouthEast, Terrain: terrain.UnknownLand},
							{Point: compass.SouthSouthEast, Terrain: terrain.UnknownLand},
							{Point: compass.South, Terrain: terrain.UnknownWater},
							{Point: compass.SouthSouthWest, Terrain: terrain.UnknownWater},
							{Point: compass.SouthWest, Terrain: terrain.UnknownWater},
							{Point: compass.West, Terrain: terrain.UnknownWater},
							{Point: compass.NorthWest, Terrain: terrain.UnknownWater},
							{Point: compass.NorthNorthWest, Terrain: terrain.UnknownWater},
						},
					},
				},
				{LineNo: 1, StepNo: 6, Line: []byte("N-LCM,  Lcm NE, SE,  Ensalada sin Tomate"),
					Result: results.Succeeded, Advance: direction.North, Report: &domain.Report_t{
						Terrain: terrain.LowMountainsConifer,
						Borders: []*domain.Border_t{
							{Direction: direction.NorthEast, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.SouthEast, Terrain: terrain.LowMountainsConifer},
						},
						Settlements: []*domain.Settlement_t{{Name: "Ensalada sin Tomate"}},
					},
				},
			},
		},
	} {
		fm, err := parser.ParseFleetMovementLine(tc.id, "", tc.unitId, 1, []byte(tc.line), false, tc.debug, tc.debug, tc.debug, false)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		clearNewMoveFields(fm)
		i1, i2 := 0, 0
		for i1 < len(tc.moves) && i2 < len(fm) {
			m1, m2 := tc.moves[i1], fm[i2]
			//t.Errorf("id: %q: step %3d: %q %q\n", tc.id, m1.StepNo, m1.Line, m2.Line)
			if diff := deep.Equal(m1, m2); diff != nil {
				for _, d := range diff {
					t.Errorf("id: %q: step %3d: %s\n", tc.id, m1.StepNo, d)
				}
			}
			i1, i2 = i1+1, i2+1
		}
		for i1 < len(tc.moves) {
			m1 := tc.moves[i1]
			t.Errorf("id: %q: step %3d: missing step %q\n", tc.id, m1.StepNo, m1.Line)
			i1 = i1 + 1
		}
		for i2 < len(fm) {
			m2 := fm[i2]
			t.Errorf("id: %q: step %3d: extra   step %q\n", tc.id, m2.StepNo, m2.Line)
			i2 = i2 + 1
		}
	}
}

func TestLocationParse(t *testing.T) {
	var lt parser.Location_t
	for _, tc := range []struct {
		id      string
		line    string
		unitId  domain.UnitId_t
		msg     string
		currHex string
		prevHex string
	}{
		{id: "0138", line: "Tribe 0138, , Current Hex = ## 1108, (Previous Hex = OO 1615)", unitId: "0138", msg: "", currHex: "## 1108", prevHex: "OO 1615"},
		{id: "1138", line: "Tribe 1138, , Current Hex = ## 0709, (Previous Hex = ## 0709)", unitId: "1138", msg: "", currHex: "## 0709", prevHex: "## 0709"},
		{id: "2138", line: "Tribe 2138, , Current Hex = ## 0709, (Previous Hex = ## 0709)", unitId: "2138", msg: "", currHex: "## 0709", prevHex: "## 0709"},
		{id: "3138", line: "Tribe 3138, , Current Hex = ## 0708, (Previous Hex = ## 0708)", unitId: "3138", msg: "", currHex: "## 0708", prevHex: "## 0708"},
		{id: "4138", line: "Tribe 4138, , Current Hex = ## 0709, (Previous Hex = OO 0709)", unitId: "4138", msg: "", currHex: "## 0709", prevHex: "OO 0709"},
		{id: "0138c1", line: "Courier 0138c1, , Current Hex = ## 0709, (Previous Hex = ## 1010)", unitId: "0138c1", msg: "", currHex: "## 0709", prevHex: "## 1010"},
		{id: "0138c2", line: "Courier 0138c2, , Current Hex = ## 1103, (Previous Hex = ## 0709)", unitId: "0138c2", msg: "", currHex: "## 1103", prevHex: "## 0709"},
		{id: "0138c3", line: "Courier 0138c3, , Current Hex = ## 1103, (Previous Hex = ## 0709)", unitId: "0138c3", msg: "", currHex: "## 1103", prevHex: "## 0709"},
		{id: "0138e1", line: "Element 0138e1, , Current Hex = ## 1106, (Previous Hex = ## 2002)", unitId: "0138e1", msg: "", currHex: "## 1106", prevHex: "## 2002"},
		{id: "0138e9", line: "Element 0138e9, , Current Hex = OO 0602, (Previous Hex = OO 0302)", unitId: "0138e9", msg: "", currHex: "OO 0602", prevHex: "OO 0302"},
		{id: "1138e1", line: "Element 1138e1, , Current Hex = ## 0709, (Previous Hex = ## 1010)", unitId: "1138e1", msg: "", currHex: "## 0709", prevHex: "## 1010"},
		{id: "2138e1", line: "Element 2138e1, , Current Hex = ## 0904, (Previous Hex = ## 1507)", unitId: "2138e1", msg: "", currHex: "## 0904", prevHex: "## 1507"},
		{id: "0138f1", line: "Fleet 0138f1, , Current Hex = OO 1508, (Previous Hex = OO 1508)", unitId: "0138f1", msg: "", currHex: "OO 1508", prevHex: "OO 1508"},
		{id: "0138f3", line: "Fleet 0138f3, , Current Hex = OQ 1210, (Previous Hex = OQ 0713)", unitId: "0138f3", msg: "", currHex: "OQ 1210", prevHex: "OQ 0713"},
		{id: "0138f8", line: "Fleet 0138f8, , Current Hex = QP 1210, (Previous Hex = QP 0713)", unitId: "0138f8", msg: "", currHex: "QP 1210", prevHex: "QP 0713"},
		{id: "1138f2", line: "Fleet 1138f2, , Current Hex = RO 2415, (Previous Hex = RO 2415)", unitId: "1138f2", msg: "", currHex: "RO 2415", prevHex: "RO 2415"},
		{id: "3138g1", line: "Garrison 3138g1, , Current Hex = ## 0708, (Previous Hex = OO 0708)", unitId: "3138g1", msg: "", currHex: "## 0708", prevHex: "OO 0708"},
	} {
		va, err := parser.Parse(tc.id, []byte(tc.line), parser.Entrypoint("Location"))
		if err != nil {
			t.Errorf("id %q: parse failed %v\n", tc.id, err)
			continue
		}
		location, ok := va.(parser.Location_t)
		if !ok {
			t.Errorf("id %q: type: want %T, got %T\n", tc.id, lt, va)
			continue
		}
		if tc.unitId != location.UnitId {
			t.Errorf("id %q: follows: want %q, got %q\n", tc.id, tc.unitId, location.UnitId)
		}
		if tc.msg != location.Message {
			t.Errorf("id %q: message: want %q, got %q\n", tc.id, tc.msg, location.Message)
		}
		if tc.currHex != location.CurrentHex {
			t.Errorf("id %q: currentHex: want %q, got %q\n", tc.id, tc.currHex, location.CurrentHex)
		}
		if tc.prevHex != location.PreviousHex {
			t.Errorf("id %q: previousHex: want %q, got %q\n", tc.id, tc.prevHex, location.PreviousHex)
		}
	}
}

func TestScoutMovementParse(t *testing.T) {
	for _, tc := range []struct {
		id      string
		line    string
		unitId  domain.UnitId_t
		scoutNo int
		moves   []*domain.Move_t
		debug   bool
	}{
		{id: "900-05.0138e1s1",
			line: `Scout 1:Scout N-PR,  \N-GH,  \N-RH,  O NW,  N, Find Iron Ore, 1590,  0138c2,  0138c3\ Can't Move on Ocean to N of HEX,  Patrolled and found 1590,  0138c2,  0138c3`, unitId: "0138e1s1", scoutNo: 1,
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("N-PR"),
					Result: results.Succeeded, Advance: direction.North, Report: &domain.Report_t{
						Terrain: terrain.FlatPrairie,
					},
				},
				{LineNo: 1, StepNo: 2, Line: []byte("N-GH"),
					Result: results.Succeeded, Advance: direction.North, Report: &domain.Report_t{
						Terrain: terrain.HillsGrassy,
					},
				},
				{LineNo: 1, StepNo: 3, Line: []byte("N-RH,  O NW,  N, Find Iron Ore, 1590,  0138c2,  0138c3"),
					Result: results.Succeeded, Advance: direction.North, Report: &domain.Report_t{
						Terrain: terrain.HillsRocky,
						Borders: []*domain.Border_t{
							{Direction: direction.North, Terrain: terrain.WaterOcean},
							{Direction: direction.NorthWest, Terrain: terrain.WaterOcean},
						},
						Resources:  []resources.Resource_e{resources.IronOre},
						Encounters: []*domain.Encounter_t{{UnitId: "1590"}, {UnitId: "0138c2"}, {UnitId: "0138c3"}},
					},
				},
				{LineNo: 1, StepNo: 4, Line: []byte("Can't Move on Ocean to N of HEX,  Patrolled and found 1590,  0138c2,  0138c3"),
					Result: results.Failed, Advance: direction.North, Report: &domain.Report_t{
						Borders: []*domain.Border_t{
							{Direction: direction.North, Terrain: terrain.WaterOcean},
						},
						Encounters: []*domain.Encounter_t{{UnitId: "1590"}, {UnitId: "0138c2"}, {UnitId: "0138c3"}},
					},
				},
			},
		},
		{id: "900-05.0138e1s3",
			line: `Scout 3:Scout SE-PR,  River S, 0590\ Not enough M.P's to move to SE into ROCKY HILLS,  Patrolled and found 0590`, unitId: "0138e1s3", scoutNo: 3,
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("SE-PR,  River S, 0590"),
					Result: results.Succeeded, Advance: direction.SouthEast, Report: &domain.Report_t{
						Terrain: terrain.FlatPrairie,
						Borders: []*domain.Border_t{
							{Direction: direction.South, Edge: edges.River},
						},
						Encounters: []*domain.Encounter_t{{UnitId: "0590"}},
					},
				},
				{LineNo: 1, StepNo: 2, Line: []byte("Not enough M.P's to move to SE into ROCKY HILLS,  Patrolled and found 0590"),
					Result: results.Failed, Advance: direction.SouthEast, Report: &domain.Report_t{
						Borders: []*domain.Border_t{
							{Direction: direction.SouthEast, Terrain: terrain.HillsRocky},
						},
						Encounters: []*domain.Encounter_t{{UnitId: "0590"}},
					},
				},
			},
		},
		{id: "900-05.0138e1s7",
			line: `Scout 7:Scout N-PR,  O NW,  N,  River S, 3138\ Can't Move on Ocean to N of HEX,  Patrolled and found 3138`, unitId: "0138e1s7", scoutNo: 7,
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("N-PR,  O NW,  N,  River S, 3138"),
					Result: results.Succeeded, Advance: direction.North, Report: &domain.Report_t{
						Terrain: terrain.FlatPrairie,
						Borders: []*domain.Border_t{
							{Direction: direction.North, Terrain: terrain.WaterOcean},
							{Direction: direction.South, Edge: edges.River},
							{Direction: direction.NorthWest, Terrain: terrain.WaterOcean},
						},
						Encounters: []*domain.Encounter_t{{UnitId: "3138"}},
					},
				},
				{LineNo: 1, StepNo: 2, Line: []byte("Can't Move on Ocean to N of HEX,  Patrolled and found 3138"),
					Result: results.Failed, Advance: direction.North, Report: &domain.Report_t{
						Borders: []*domain.Border_t{
							{Direction: direction.North, Terrain: terrain.WaterOcean},
						},
						Encounters: []*domain.Encounter_t{{UnitId: "3138"}},
					},
				},
			},
		},
		{id: "900-05.0138e1s8",
			line: `Scout 8:Scout SW-GH,  \NW-PR,  \NW-PR,  \NW-PR,  \ Not enough M.P's to move to NW into PRAIRIE,  Nothing of interest found`, unitId: "0138e1s8", scoutNo: 8,
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("SW-GH"),
					Result: results.Succeeded, Advance: direction.SouthWest, Report: &domain.Report_t{
						Terrain: terrain.HillsGrassy,
					},
				},
				{LineNo: 1, StepNo: 2, Line: []byte("NW-PR"),
					Result: results.Succeeded, Advance: direction.NorthWest, Report: &domain.Report_t{
						Terrain: terrain.FlatPrairie,
					},
				},
				{LineNo: 1, StepNo: 3, Line: []byte("NW-PR"),
					Result: results.Succeeded, Advance: direction.NorthWest, Report: &domain.Report_t{
						Terrain: terrain.FlatPrairie,
					},
				},
				{LineNo: 1, StepNo: 4, Line: []byte("NW-PR"),
					Result: results.Succeeded, Advance: direction.NorthWest, Report: &domain.Report_t{
						Terrain: terrain.FlatPrairie,
					},
				},
				{LineNo: 1, StepNo: 5, Line: []byte("Not enough M.P's to move to NW into PRAIRIE,  Nothing of interest found"),
					Result: results.Failed, Advance: direction.NorthWest, Report: &domain.Report_t{
						Borders: []*domain.Border_t{
							{Direction: direction.NorthWest, Terrain: terrain.FlatPrairie},
						},
					},
				},
			},
		},
	} {
		sm, err := parser.ParseScoutMovementLine(tc.id, "", tc.unitId, 1, []byte(tc.line), false, tc.debug, tc.debug, false, false)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		clearNewMoveFields(sm.Moves)
		if tc.scoutNo != sm.No {
			t.Errorf("id %q: scoutNo: want %d, got %d\n", tc.id, tc.scoutNo, sm.No)
		}
		i1, i2 := 0, 0
		for i1 < len(tc.moves) && i2 < len(sm.Moves) {
			m1, m2 := tc.moves[i1], sm.Moves[i2]
			//t.Errorf("id: %q: step %3d: %q %q\n", tc.id, m1.StepNo, m1.Line, m2.Line)
			if diff := deep.Equal(m1, m2); diff != nil {
				for _, d := range diff {
					t.Errorf("id: %q: step %3d: %s\n", tc.id, m1.StepNo, d)
				}
			}
			i1, i2 = i1+1, i2+1
		}
		for i1 < len(tc.moves) {
			m1 := tc.moves[i1]
			t.Errorf("id: %q: step %3d: missing step %q\n", tc.id, m1.StepNo, m1.Line)
			i1 = i1 + 1
		}
		for i2 < len(sm.Moves) {
			m2 := sm.Moves[i2]
			t.Errorf("id: %q: step %3d: extra   step %q\n", tc.id, m2.StepNo, m2.Line)
			i2 = i2 + 1
		}
	}
}

func TestStatusLine(t *testing.T) {
	for _, tc := range []struct {
		id     string
		line   string
		unitId domain.UnitId_t
		moves  []*domain.Move_t
		debug  bool
	}{
		{id: "899-12.0138.0138",
			line:   `0138 Status: PRAIRIE, 0138`,
			unitId: "0138",
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("PRAIRIE, 0138"),
					Result: results.StatusLine, Still: true, Report: &domain.Report_t{
						Terrain:    terrain.FlatPrairie,
						Encounters: []*domain.Encounter_t{{UnitId: "0138"}},
					},
				},
			},
		},
		{id: "900-01.0138.0138e1",
			line:   `0138e1 Status: PRAIRIE,River S, 0138e1`,
			unitId: "0138e1",
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("PRAIRIE,River S, 0138e1"),
					Result: results.StatusLine, Still: true, Report: &domain.Report_t{
						Terrain: terrain.FlatPrairie,
						Borders: []*domain.Border_t{
							{Direction: direction.South, Edge: edges.River},
						},
						Encounters: []*domain.Encounter_t{{UnitId: "0138e1"}},
					},
				},
			},
		},
		{id: "900-02.0138.0138",
			line:   `0138 Status: PRAIRIE, O S,Ford SE, 2138, 0138`,
			unitId: "0138",
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("PRAIRIE, O S,Ford SE, 2138, 0138"),
					Result: results.StatusLine, Still: true, Report: &domain.Report_t{
						Terrain: terrain.FlatPrairie,
						Borders: []*domain.Border_t{
							{Direction: direction.SouthEast, Edge: edges.Ford},
							{Direction: direction.South, Terrain: terrain.WaterOcean},
						},
						Encounters: []*domain.Encounter_t{{UnitId: "2138"}, {UnitId: "0138"}},
					},
				},
			},
		},
		{id: "900-02.0138.0138e1",
			line:   `0138e1 Status: PRAIRIE, O NW, 0138e1`,
			unitId: "0138e1",
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("PRAIRIE, O NW, 0138e1"),
					Result: results.StatusLine, Still: true, Report: &domain.Report_t{
						Terrain: terrain.FlatPrairie,
						Borders: []*domain.Border_t{
							{Direction: direction.NorthWest, Terrain: terrain.WaterOcean},
						},
						Encounters: []*domain.Encounter_t{{UnitId: "0138e1"}},
					},
				},
			},
		},
		{id: "900-04.0138.0138",
			line:   `0138 Status: CONIFER HILLS, O SW, NW, S, 2138, 0138c1, 0138, 1138`,
			unitId: "0138",
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("CONIFER HILLS, O SW, NW, S, 2138, 0138c1, 0138, 1138"),
					Result: results.StatusLine, Still: true, Report: &domain.Report_t{
						Terrain: terrain.HillsConifer,
						Borders: []*domain.Border_t{
							{Direction: direction.South, Terrain: terrain.WaterOcean},
							{Direction: direction.SouthWest, Terrain: terrain.WaterOcean},
							{Direction: direction.NorthWest, Terrain: terrain.WaterOcean},
						},
						Encounters: []*domain.Encounter_t{{UnitId: "2138"}, {UnitId: "0138c1"}, {UnitId: "0138"}, {UnitId: "1138"}},
					},
				},
			},
		},
	} {
		sl, err := parser.ParseStatusLine(tc.id, "", tc.unitId, 1, []byte(tc.line), false, tc.debug, tc.debug, false)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		clearNewMoveFields(sl)
		i1, i2 := 0, 0
		for i1 < len(tc.moves) && i2 < len(sl) {
			m1, m2 := tc.moves[i1], sl[i2]
			//t.Errorf("id: %q: step %3d: %d %q %d %q\n", tc.id, m1.StepNo, i1, m1.Line, i2, m2.Line)
			if diff := deep.Equal(m1, m2); diff != nil {
				for _, d := range diff {
					t.Errorf("id: %q: step %3d: %s\n", tc.id, m1.StepNo, d)
				}
			}
			i1, i2 = i1+1, i2+1
		}
		for i1 < len(tc.moves) {
			m1 := tc.moves[i1]
			t.Errorf("id: %q: step %3d: missing step %q\n", tc.id, m1.StepNo, m1.Line)
			i1 = i1 + 1
		}
		for i2 < len(sl) {
			m2 := sl[i2]
			t.Errorf("id: %q: step %3d: extra   step %q\n", tc.id, m2.StepNo, m2.Line)
			i2 = i2 + 1
		}
	}
}

func TestTribeFollowsParse(t *testing.T) {
	for _, tc := range []struct {
		id      string
		line    string
		unitId  domain.UnitId_t
		follows domain.UnitId_t
		debug   bool
	}{
		{id: "1812", line: "Tribe Follows 1812", follows: "1812"},
		{id: "1812f3", line: "Tribe Follows 1812f3", follows: "1812f3"},
	} {
		tf, err := parser.ParseTribeFollowsLine(tc.id, "", tc.unitId, 1, []byte(tc.line), tc.debug)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		if tc.follows != tf.Follows {
			t.Errorf("id %q: follows: want %q, got %q\n", tc.id, tc.follows, tf.Follows)
		}
	}
}

func TestTribeGoesParse(t *testing.T) {
	for _, tc := range []struct {
		id     string
		line   string
		unitId domain.UnitId_t
		goesTo string
		debug  bool
	}{
		{id: "1", line: "Tribe Goes to DT 1812", goesTo: "DT 1812"},
		{id: "2", line: "Tribe Goes to ## 1812", goesTo: "## 1812"},
		{id: "3", line: "Tribe Goes to N/A", goesTo: "N/A"},
	} {
		gt, err := parser.ParseTribeGoesToLine(tc.id, "", tc.unitId, 1, []byte(tc.line), tc.debug)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		if tc.goesTo != gt.GoesTo {
			t.Errorf("id %q: goesTo: want %q, got %q\n", tc.id, tc.goesTo, gt.GoesTo)
		}
	}
}

func TestTribeMovementParse(t *testing.T) {
	for _, tc := range []struct {
		id     string
		line   string
		unitId domain.UnitId_t
		moves  []*domain.Move_t
		debug  bool
	}{
		{id: "900-01.0138",
			line: `Tribe Movement: Move \`,
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte{},
					Result: results.Succeeded, Still: true, Report: &domain.Report_t{},
				},
			},
		},
		{id: "900-02.0138",
			line: "Tribe Movement: Move NW-GH,",
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("NW-GH"),
					Result: results.Succeeded, Advance: direction.NorthWest, Report: &domain.Report_t{
						Terrain: terrain.HillsGrassy,
					},
				},
			},
		},
		{id: "900-02.0138f1",
			line: `Tribe Movement: Move SW-PR The Dirty Squirrel\N-LCM,  Lcm NE, SE,  Ensalada sin Tomate\`,
			moves: []*domain.Move_t{
				{LineNo: 1, StepNo: 1, Line: []byte("SW-PR The Dirty Squirrel"),
					Result: results.Succeeded, Advance: direction.SouthWest, Report: &domain.Report_t{
						Terrain:     terrain.FlatPrairie,
						Settlements: []*domain.Settlement_t{{Name: "The Dirty Squirrel"}},
					},
				},
				{LineNo: 1, StepNo: 2, Line: []byte("N-LCM,  Lcm NE, SE,  Ensalada sin Tomate"),
					Result: results.Succeeded, Advance: direction.North, Report: &domain.Report_t{
						Terrain: terrain.LowMountainsConifer,
						Borders: []*domain.Border_t{
							{Direction: direction.NorthEast, Terrain: terrain.LowMountainsConifer},
							{Direction: direction.SouthEast, Terrain: terrain.LowMountainsConifer},
						},
						Settlements: []*domain.Settlement_t{{Name: "Ensalada sin Tomate"}},
					},
				},
			},
		},
	} {
		tm, err := parser.ParseTribeMovementLine(tc.id, "", tc.unitId, 1, []byte(tc.line), false, tc.debug, tc.debug, false)
		if err != nil {
			t.Errorf("id %q: parse failed: %v\n", tc.id, err)
			continue
		}
		clearNewMoveFields(tm)
		i1, i2 := 0, 0
		for i1 < len(tc.moves) || i2 < len(tm) {
			m1, m2 := tc.moves[i1], tm[i2]
			//t.Errorf("id: %q: step %3d: %d %q %d %q\n", tc.id, m1.StepNo, i1, m1.Line, i2, m2.Line)
			if diff := deep.Equal(m1, m2); diff != nil {
				for _, d := range diff {
					t.Errorf("id: %q: step %3d: %s\n", tc.id, m1.StepNo, d)
				}
			}
			i1, i2 = i1+1, i2+1
		}
		for i1 < len(tc.moves) {
			m1 := tc.moves[i1]
			t.Errorf("id: %q: step %3d: missing step %q\n", tc.id, m1.StepNo, m1.Line)
			i1 = i1 + 1
		}
		for i2 < len(tm) {
			m2 := tm[i2]
			t.Errorf("id: %q: step %3d: extra   step %q\n", tc.id, m2.StepNo, m2.Line)
			i2 = i2 + 1
		}
	}
}

func TestTurnInfoParse(t *testing.T) {
	var ti parser.TurnInfo_t
	for _, tc := range []struct {
		id        string
		line      string
		thisYear  int
		thisMonth int
		nextYear  int
		nextMonth int
	}{
		{id: "900-01", line: "Current Turn 900-01 (#1), Spring, FINE\tNext Turn 900-02 (#2), 12/11/2023", thisYear: 900, thisMonth: 1, nextYear: 900, nextMonth: 2},
		{id: "900-02", line: "Current Turn 900-02 (#2), Spring, FINE", thisYear: 900, thisMonth: 2},
	} {
		va, err := parser.Parse(tc.id, []byte(tc.line), parser.Entrypoint("TurnInfo"))
		if err != nil {
			t.Errorf("id %q: parse failed %v\n", tc.id, err)
			continue
		}
		turnInfo, ok := va.(parser.TurnInfo_t)
		if !ok {
			t.Errorf("id %q: type: want %T, got %T\n", tc.id, ti, va)
			continue
		}
		if tc.thisYear != turnInfo.CurrentTurn.Year {
			t.Errorf("id %q: thisYear: want %d, got %d\n", tc.id, tc.thisYear, turnInfo.CurrentTurn.Year)
		}
		if tc.thisMonth != turnInfo.CurrentTurn.Month {
			t.Errorf("id %q: thisMonth: want %d, got %d\n", tc.id, tc.thisMonth, turnInfo.CurrentTurn.Month)
		}
		if tc.nextYear == 0 && tc.nextMonth == 0 && !turnInfo.NextTurn.IsZero() {
			t.Errorf("id %q: nextTurn: want %v, got %v\n", tc.id, parser.Date_t{}, turnInfo.NextTurn)
		} else {
			if tc.nextYear != turnInfo.NextTurn.Year {
				t.Errorf("id %q: nextYear: want %d, got %d\n", tc.id, tc.nextYear, turnInfo.NextTurn.Year)
			}
			if tc.nextMonth != turnInfo.NextTurn.Month {
				t.Errorf("id %q: nextMonth: want %d, got %d\n", tc.id, tc.nextMonth, turnInfo.NextTurn.Month)
			}
		}
	}
}

func diffBorderSets(set1, set2 []*domain.Border_t) ([]*domain.Border_t, []*domain.Border_t) {
	map1 := map[string]*domain.Border_t{}
	map2 := map[string]*domain.Border_t{}

	// Fill map1 with elements from set1
	for _, i := range set1 {
		map1[i.String()] = i
	}

	// Fill map2 with elements from set2
	for _, i := range set2 {
		map2[i.String()] = i
	}

	// Find elements in set1 but not in set2
	var onlyInSet1 []*domain.Border_t
	for k, v := range map1 {
		if _, found := map2[k]; !found {
			onlyInSet1 = append(onlyInSet1, v)
		}
	}

	// Find elements in set2 but not in set1
	var onlyInSet2 []*domain.Border_t
	for k, v := range map2 {
		if _, found := map1[k]; !found {
			onlyInSet2 = append(onlyInSet2, v)
		}
	}

	return onlyInSet1, onlyInSet2
}

func diffUnitSets(set1, set2 []domain.UnitId_t) ([]domain.UnitId_t, []domain.UnitId_t) {
	// Create maps to store the presence of UnitIDs in each slice
	map1 := map[domain.UnitId_t]bool{}
	map2 := map[domain.UnitId_t]bool{}

	// Fill map1 with elements from set1
	for _, id := range set1 {
		map1[id] = true
	}

	// Fill map2 with elements from set2
	for _, id := range set2 {
		map2[id] = true
	}

	// Find elements in set1 but not in set2
	var onlyInSet1 []domain.UnitId_t
	for id := range map1 {
		if _, found := map2[id]; !found {
			onlyInSet1 = append(onlyInSet1, id)
		}
	}

	// Find elements in set2 but not in set1
	var onlyInSet2 []domain.UnitId_t
	for id := range map2 {
		if _, found := map1[id]; !found {
			onlyInSet2 = append(onlyInSet2, id)
		}
	}

	return onlyInSet1, onlyInSet2
}
