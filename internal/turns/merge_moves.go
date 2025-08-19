// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

//// MergeMoves visits all the moves and returns a slice that consolidates the movement reports.
//// This slice contains one report per hex with the reports merged, most recent turn having priority.
//// The slice is sorted by location (column then row).
//// Assumes that the input is sorted by turn then unit.
//func MergeMoves(turns []*parser.Turn_t, debug bool) ([]*parser.Report_t, error) {
//	var sortedReports []*parser.Report_t
//
//	// collect all the reports into a single map indexed by location.
//	// each entry in the map holds a slice containing all the reports for that location.
//	// each entry's slice will be sorted by turn because the input is sorted by turn.
//	allReports := map[string][]*parser.Report_t{}
//	for _, turn := range turns {
//		for _, unit := range turn.SortedMoves {
//			var moves []*parser.Move_t
//			for _, move := range unit.Moves {
//				moves = append(moves, move)
//			}
//			for _, scout := range unit.Scouts {
//				for _, move := range scout.Moves {
//					moves = append(moves, move)
//				}
//			}
//			for _, move := range moves {
//				if move.Report.TurnId == "" {
//					log.Printf("bug: move.Report.TurnId == %q\n", move.Report.TurnId)
//					move.Report.TurnId = turn.Id
//				}
//				allReports[move.CurrentHex] = append(allReports[move.CurrentHex], move.Report)
//
//				// we have hexes that we've never visited. create reports for them so we can map them.
//				for _, border := range move.Report.Borders {
//					if border.Terrain == terrain.Blank {
//						continue
//					}
//					borderHex := coords.Move(move.CurrentHex, border.Direction)
//					allReports[borderHex] = append(allReports[borderHex], &parser.Report_t{
//						TurnId:  turn.Id,
//						Terrain: border.Terrain,
//					})
//				}
//				for _, fh := range move.Report.FarHorizons {
//					fhHex := move.CurrentHex
//					switch fh.Point {
//					case compass.North:
//						fhHex = coords.Move(fhHex, direction.North, direction.North)
//					case compass.NorthNorthEast:
//						fhHex = coords.Move(fhHex, direction.North, direction.NorthEast)
//					case compass.NorthEast:
//						fhHex = coords.Move(fhHex, direction.NorthEast, direction.NorthEast)
//					case compass.East:
//						fhHex = coords.Move(fhHex, direction.NorthEast, direction.SouthEast)
//					case compass.SouthEast:
//						fhHex = coords.Move(fhHex, direction.SouthEast, direction.SouthEast)
//					case compass.SouthSouthEast:
//						fhHex = coords.Move(fhHex, direction.South, direction.SouthEast)
//					case compass.South:
//						fhHex = coords.Move(fhHex, direction.South, direction.South)
//					case compass.SouthSouthWest:
//						fhHex = coords.Move(fhHex, direction.South, direction.SouthWest)
//					case compass.SouthWest:
//						fhHex = coords.Move(fhHex, direction.SouthWest, direction.SouthWest)
//					case compass.West:
//						fhHex = coords.Move(fhHex, direction.SouthWest, direction.NorthWest)
//					case compass.NorthWest:
//						fhHex = coords.Move(fhHex, direction.NorthWest, direction.NorthWest)
//					case compass.NorthNorthWest:
//						fhHex = coords.Move(fhHex, direction.North, direction.NorthWest)
//					default:
//						panic(fmt.Sprintf("assert(point != %d)", fh.Point))
//					}
//					allReports[fhHex] = append(allReports[fhHex], &parser.Report_t{
//						TurnId:  turn.Id,
//						Terrain: fh.Terrain,
//					})
//				}
//			}
//		}
//	}
//	log.Printf("merge: moves: found %8d hexes in  %8d reports\n", len(allReports), len(sortedReports))
//
//	// for each location, merge the reports into a single entry
//	for hex, reports := range allReports {
//		loc, err := coords.HexToMap(hex)
//		if err != nil {
//			panic(err)
//		}
//		rpt := &parser.Report_t{
//			Location: loc,
//		}
//		sortedReports = append(sortedReports, rpt)
//		for _, report := range reports {
//			rpt.TurnId = report.TurnId
//			if rpt.ScoutedTurnId == "" {
//				rpt.ScoutedTurnId = report.ScoutedTurnId
//			}
//
//			// merge terrain if it is different
//			if report.Terrain != rpt.Terrain {
//				if rpt.Terrain == terrain.Blank {
//					rpt.Terrain = report.Terrain
//				} else if report.Terrain == terrain.Blank {
//					// ignore the new terrain if it is blank
//				} else if report.Terrain == terrain.UnknownLand || report.Terrain == terrain.UnknownWater {
//					// don't overwrite existing terrain with a fleet observation
//				} else {
//					rpt.Terrain = report.Terrain
//				}
//			}
//
//			// merge borders
//			for _, b := range report.Borders {
//				rpt.MergeBorders(b)
//			}
//
//			// merge edges
//			for _, e := range report.Encounters {
//				rpt.MergeEncounters(e)
//			}
//
//			// merge resources
//			// todo: no way to delete a resource?
//			for _, r := range report.Resources {
//				rpt.MergeResources(r)
//			}
//
//			// merge settlement
//			// todo: no way to delete a settlement?
//			for _, s := range report.Settlements {
//				rpt.MergeSettlements(s)
//			}
//		}
//	}
//	log.Printf("sorted %8d reports\n", len(sortedReports))
//
//	// we must sort the returned slice by location
//	sort.Slice(sortedReports, func(i, j int) bool {
//		a, b := sortedReports[i].Location, sortedReports[j].Location
//		if a.Column < b.Column {
//			return true
//		} else if a.Column == b.Column {
//			return a.Row < b.Row
//		}
//		return false
//	})
//
//	if debug {
//		for _, report := range sortedReports {
//			log.Printf("(%5d %5d) %s %s\n", report.Location.Column, report.Location.Row, report.Location.GridString(), report.Terrain)
//			for _, dir := range []direction.Direction_e{direction.North, direction.NorthEast, direction.SouthEast, direction.South, direction.SouthWest, direction.NorthWest} {
//				for _, border := range report.Borders {
//					if border.Direction == dir && border.Terrain != terrain.Blank {
//						log.Printf("(%5d %5d) %s %-2s %s\n", report.Location.Column, report.Location.Row, report.Location.GridString(), border.Direction, border.Terrain)
//					}
//				}
//				for _, border := range report.Borders {
//					if border.Direction == dir && border.Edge != edges.None {
//						log.Printf("(%5d %5d) %s %-2s %s\n", report.Location.Column, report.Location.Row, report.Location.GridString(), border.Direction, border.Edge)
//					}
//				}
//			}
//			for _, encounter := range report.Encounters {
//				log.Printf("(%5d %5d) %s %v\n", report.Location.Column, report.Location.Row, report.Location.GridString(), encounter)
//			}
//		}
//	}
//
//	return sortedReports, nil
//}
