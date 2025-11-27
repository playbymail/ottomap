// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/playbymail/ottomap/actions"
	"github.com/playbymail/ottomap/internal/edges"
	"github.com/playbymail/ottomap/internal/parser"
	"github.com/playbymail/ottomap/internal/results"
	"github.com/playbymail/ottomap/internal/runners"
	"github.com/playbymail/ottomap/internal/terrain"
	"github.com/playbymail/ottomap/internal/tiles"
	"github.com/playbymail/ottomap/internal/turns"
	"github.com/playbymail/ottomap/internal/wxx"
	"github.com/spf13/cobra"
)

var argsRender struct {
	paths struct {
		data   string // path to data folder
		input  string // path to input folder
		output string // path to output folder
	}
	parser              parser.ParseConfig
	mapper              actions.MapConfig
	render              wxx.RenderConfig
	clanId              string
	soloElement         string // when set, only this element is rendered
	originGrid          string
	autoEOL             bool
	quitOnInvalidGrid   bool
	warnOnInvalidGrid   bool
	warnOnNewSettlement bool
	warnOnTerrainChange bool
	maxTurn             struct { // maximum turn id to use
		id    string
		year  int
		month int
	}
	debug struct {
		merge bool
	}
	experimental struct {
		blankMapSmall bool
		blankMapFull  bool
		jsonMapSmall  bool
	}
	saveWithTurnId bool
	show           struct {
		origin   bool
		shiftMap bool
	}
}

var cmdRender = &cobra.Command{
	Use:   "render",
	Short: "Create a map from a report",
	Long:  `Load and parse turn report and create a map.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		gcfg := globalConfig
		logFlags := 0
		if gcfg.DebugFlags.LogFile {
			logFlags |= log.Lshortfile
		}
		if gcfg.DebugFlags.LogTime {
			logFlags |= log.Ltime
		}
		log.SetFlags(logFlags)

		if len(argsRender.clanId) != 4 || argsRender.clanId[0] != '0' {
			return fmt.Errorf("clan-id must be a 4 digit number starting with 0")
		} else if n, err := strconv.Atoi(argsRender.clanId[1:]); err != nil || n < 0 || n > 9999 {
			return fmt.Errorf("clan-id must be a 4 digit number starting with 0")
		}

		if argsRender.paths.data == "" {
			return fmt.Errorf("path to data folder is required")
		}

		// do the abs path check for data
		if strings.TrimSpace(argsRender.paths.data) != argsRender.paths.data {
			log.Fatalf("error: data: leading or trailing spaces are not allowed\n")
		} else if path, err := abspath(argsRender.paths.data); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if sb, err := os.Stat(path); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if !sb.IsDir() {
			log.Fatalf("error: data: %v is not a directory\n", path)
		} else {
			argsRender.paths.data = path
		}

		argsRender.paths.input = filepath.Join(argsRender.paths.data, "input")
		if path, err := abspath(argsRender.paths.input); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if sb, err := os.Stat(path); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if !sb.IsDir() {
			log.Fatalf("error: data: %v is not a directory\n", path)
		} else {
			argsRender.paths.input = path
		}

		argsRender.paths.output = filepath.Join(argsRender.paths.data, "output")
		if path, err := abspath(argsRender.paths.output); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if sb, err := os.Stat(path); err != nil {
			log.Fatalf("error: data: %v\n", err)
		} else if !sb.IsDir() {
			log.Fatalf("error: data: %v is not a directory\n", path)
		} else {
			argsRender.paths.output = path
		}

		if len(argsRender.originGrid) == 0 {
			// terminate on ## in location
			argsRender.quitOnInvalidGrid = true
		} else if len(argsRender.originGrid) != 2 {
			log.Fatalf("error: originGrid %q: must be two upper-case letters\n", argsRender.originGrid)
		} else if strings.Trim(argsRender.originGrid, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
			log.Fatalf("error: originGrid %q: must be two upper-case letters\n", argsRender.originGrid)
		} else {
			// don't quit when we replace ## with the location
			argsRender.quitOnInvalidGrid = false
		}

		if argsRender.maxTurn.id == "" {
			argsRender.maxTurn.year, argsRender.maxTurn.month = 9999, 12
		} else if year, month, err := strToTurnId(argsRender.maxTurn.id); err != nil {
			log.Fatalf("error: turn-cutoff %q: must be yyyy-mm format", argsRender.maxTurn.id)
		} else {
			argsRender.maxTurn.year, argsRender.maxTurn.month = year, month
		}
		argsRender.maxTurn.id = fmt.Sprintf("%04d-%02d", argsRender.maxTurn.year, argsRender.maxTurn.month)

		if argsRender.maxTurn.year < 0 {
			argsRender.maxTurn.year = 0
		} else if argsRender.maxTurn.year > 9999 {
			argsRender.maxTurn.year = 9999
		}
		if argsRender.maxTurn.month < 0 {
			argsRender.maxTurn.month = 1
		} else if argsRender.maxTurn.month > 12 {
			argsRender.maxTurn.month = 12
		}
		argsRender.maxTurn.id = fmt.Sprintf("%04d-%02d", argsRender.maxTurn.year, argsRender.maxTurn.month)

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		gcfg := globalConfig

		if argsRoot.showVersion {
			log.Printf("ottomap version %s\n", version)
		}
		if argsRoot.soloClan {
			log.Printf("clan %q: running solo\n", argsRender.clanId)
		}
		argsRender.originGrid = "RR"
		argsRender.quitOnInvalidGrid = false

		started := time.Now()
		log.Printf("data:   %s\n", argsRender.paths.data)
		log.Printf("input:  %s\n", argsRender.paths.input)
		log.Printf("output: %s\n", argsRender.paths.output)

		if gcfg.Experimental.ReverseWalker {
			log.Printf("walk: reverse walk enabled\n")
			err := runners.Run(argsRender.paths.input)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		inputs, err := turns.CollectInputs(argsRender.paths.input, argsRender.maxTurn.year, argsRender.maxTurn.month, argsRoot.soloClan, argsRender.clanId)
		if err != nil {
			log.Fatalf("error: inputs: %v\n", err)
		}
		log.Printf("inputs: found %d turn reports\n", len(inputs))

		// allTurns holds the turn and move data and allows multiple clans to be loaded.
		allTurns := map[string][]*parser.Turn_t{}
		totalUnitMoves := 0
		var turnId, maxTurnId string // will be set to the last/maximum turnId we process
		for _, i := range inputs {
			gcfg := globalConfig

			started := time.Now()
			data, err := os.ReadFile(i.Path)
			if err != nil {
				log.Fatalf("error: read: %v\n", err)
			} else if len(data) == 0 {
				log.Printf("warn: %q: empty file\n", i.Path)
				continue
			}
			if argsRender.autoEOL {
				data = bytes.ReplaceAll(data, []byte{'\r', '\n'}, []byte{'\n'})
				data = bytes.ReplaceAll(data, []byte{'\r'}, []byte{'\n'})
			} else if gcfg.Experimental.StripCR {
				data = bytes.ReplaceAll(data, []byte{'\r', '\n'}, []byte{'\n'})
			}
			if i.Turn.Year < 899 || i.Turn.Year > 9999 || i.Turn.Month < 1 || i.Turn.Month > 12 {
				log.Printf("warn: %q: invalid turn year '%d'\n", i.Id, i.Turn.Year)
				continue
			} else if i.Turn.Month < 1 || i.Turn.Month > 12 {
				log.Printf("warn: %q: invalid turn month '%d'\n", i.Id, i.Turn.Month)
				continue
			}
			pastCutoff := false
			if i.Turn.Year > argsRender.maxTurn.year {
				pastCutoff = true
			} else if i.Turn.Year == argsRender.maxTurn.year {
				if i.Turn.Month > argsRender.maxTurn.month {
					pastCutoff = true
				}
			}
			if pastCutoff {
				log.Printf("warn: %q: past cutoff %04d-%02d\n", i.Id, argsRender.maxTurn.year, argsRender.maxTurn.month)
			}
			turnId = fmt.Sprintf("%04d-%02d", i.Turn.Year, i.Turn.Month)
			if turnId > maxTurnId {
				maxTurnId = turnId
			}
			turn, err := parser.ParseInput(i.Id, turnId, data, gcfg.Parser.AcceptLoneDash, gcfg.DebugFlags.Parser, gcfg.DebugFlags.Sections, gcfg.DebugFlags.Steps, gcfg.DebugFlags.Nodes, gcfg.DebugFlags.FleetMovement, gcfg.Experimental.SplitTrailingUnits, gcfg.Experimental.CleanupScoutStill, argsRender.parser)
			if err != nil {
				log.Fatal(err)
			} else if turnId != fmt.Sprintf("%04d-%02d", turn.Year, turn.Month) {
				if turn.Year == 0 && turn.Month == 0 {
					log.Printf("error: unable to locate turn information in file\n")
					log.Printf("error: this is usually caused by unexpected line endings in the file\n")
					log.Printf("error: try running with --auto-eol\n")
				}
				log.Fatalf("error: expected turn %q: got turn %q\n", turnId, fmt.Sprintf("%04d-%02d", turn.Year, turn.Month))
			}
			//log.Printf("len(turn.SpecialNames) = %d\n", len(turn.SpecialNames))

			allTurns[turnId] = append(allTurns[turnId], turn)
			totalUnitMoves += len(turn.UnitMoves)
			log.Printf("%q: parsed %6d units in %v\n", i.Id, len(turn.UnitMoves), time.Since(started))
		}
		log.Printf("parsed %d inputs in to %d turns and %d units in %v\n", len(inputs), len(allTurns), totalUnitMoves, time.Since(started))

		// consolidate the turns, then sort by year and month
		var consolidatedTurns []*parser.Turn_t
		consolidatedSpecialNames := map[string]*parser.Special_t{}
		foundDuplicates := false
		for _, unitTurns := range allTurns {
			if len(unitTurns) == 0 {
				// we shouldn't have any empty turns, but be safe
				continue
			}
			// create a new turn to hold the consolidated unit moves for the turn
			turn := &parser.Turn_t{
				Id:        fmt.Sprintf("%04d-%02d", unitTurns[0].Year, unitTurns[0].Month),
				Year:      unitTurns[0].Year,
				Month:     unitTurns[0].Month,
				UnitMoves: map[parser.UnitId_t]*parser.Moves_t{},
			}
			consolidatedTurns = append(consolidatedTurns, turn)

			// copy all the unit moves into this new turn, calling out duplicates
			for _, unitTurn := range unitTurns {
				for id, unitMoves := range unitTurn.UnitMoves {
					if turn.UnitMoves[id] != nil {
						foundDuplicates = true
						log.Printf("error: %s: %-6s: duplicate unit\n", turn.Id, id)
					}
					turn.UnitMoves[id] = unitMoves
					turn.SortedMoves = append(turn.SortedMoves, unitMoves)
					if gcfg.Experimental.ReverseWalker {
						turn.MovesSortedByElement = append(turn.MovesSortedByElement, unitMoves)
					}
				}
				if unitTurn.SpecialNames != nil {
					// consolidate any special hexes
					for id, special := range unitTurn.SpecialNames {
						consolidatedSpecialNames[id] = special
					}
				}
			}

			if gcfg.Experimental.ReverseWalker {
				turn.SortMovesByElement()
			}
		}
		if foundDuplicates {
			log.Fatalf("error: please fix the duplicate units and restart\n")
		}
		if len(consolidatedSpecialNames) > 0 {
			log.Printf("consolidated %d special hex names\n", len(consolidatedSpecialNames))
		}
		sort.Slice(consolidatedTurns, func(i, j int) bool {
			a, b := consolidatedTurns[i], consolidatedTurns[j]
			if a.Year < b.Year {
				return true
			} else if a.Year == b.Year {
				return a.Month < b.Month
			}
			return false
		})
		for _, turn := range consolidatedTurns {
			log.Printf("%s: %8d units\n", turn.Id, len(turn.UnitMoves))
			sort.Slice(turn.SortedMoves, func(i, j int) bool {
				return turn.SortedMoves[i].UnitId < turn.SortedMoves[j].UnitId
			})
		}

		// link prev and next turns
		for n, turn := range consolidatedTurns {
			if n > 0 {
				turn.Prev = consolidatedTurns[n-1]
			}
			if n+1 < len(consolidatedTurns) {
				turn.Next = consolidatedTurns[n+1]
			}
		}

		// check for N/A and obscured locations and quit if we find any
		var invalidLocations []string
		naLocationCount, obscuredLocationCount := 0, 0
		for _, turn := range consolidatedTurns {
			for _, unitMoves := range turn.UnitMoves {
				if unitMoves.FromHex == "N/A" {
					naLocationCount++
					invalidLocations = append(invalidLocations, fmt.Sprintf("turn %s: unit %-6s: Previous Hex is 'N/A'", unitMoves.TurnId, unitMoves.UnitId))
				}
				if gcfg.Parser.CheckObscuredGrids {
					if strings.HasPrefix(unitMoves.FromHex, "##") {
						obscuredLocationCount++
						invalidLocations = append(invalidLocations, fmt.Sprintf("turn %s: unit %-6s: Previous Hex starts with '##'", unitMoves.TurnId, unitMoves.UnitId))
					}
					if strings.HasPrefix(unitMoves.ToHex, "##") {
						obscuredLocationCount++
						invalidLocations = append(invalidLocations, fmt.Sprintf("turn %s: unit %-6s: Current  Hex starts with '##'", unitMoves.TurnId, unitMoves.UnitId))
					}
				}
			}
		}
		if len(invalidLocations) > 0 {
			if naLocationCount > 0 && obscuredLocationCount == 0 {
				log.Printf("error: ottomap found units that have 'N/A' in the Previous Hex field\n")
			} else if naLocationCount == 0 && obscuredLocationCount > 0 {
				log.Printf("error: ottomap found units that have '##' in the Current or Previous Hex field\n")
			} else {
				log.Printf("error: ottomap found units that have 'N/A' or '##' in the Current or Previous Hex field\n")
			}
			for _, msg := range invalidLocations {
				log.Println(msg)
			}
			if naLocationCount > 0 || gcfg.Parser.QuitOnObscuredGrids {
				log.Fatalf("please update the location field for these units and restart\n")
			} else {
				// quitting on obscured grids will break many, many maps, so this is just a warning for now
				log.Printf("please update the location field for these units and restart\n")
			}
		}

		// sanity check on the current and prior locations.
		changedLinks, staticLinks := 0, 0
		for _, turn := range consolidatedTurns {
			if turn.Next == nil { // nothing to update
				continue
			}
			for _, unitMoves := range turn.UnitMoves {
				nextUnitMoves := turn.Next.UnitMoves[unitMoves.UnitId]
				if nextUnitMoves == nil {
					continue
				}
				if unitMoves.ToHex[2:] != nextUnitMoves.FromHex[2:] {
					changedLinks++
					log.Printf("warning: %s: %-6s: from %q\n", turn.Id, unitMoves.UnitId, unitMoves.ToHex)
					log.Printf("       : %s: %-6s: to   %q\n", turn.Next.Id, nextUnitMoves.UnitId, nextUnitMoves.FromHex)
				} else {
					staticLinks++
				}
				nextUnitMoves.FromHex = unitMoves.ToHex
			}
		}
		log.Printf("links: %d same, %d changed\n", staticLinks, changedLinks)
		if changedLinks != 0 {
			// this can happen when an element is destroyed and another created with the same name
			// during a single turn.
			log.Printf("warning: the previous and current hexes don't align in some reports\n")
			log.Printf("warning: if you didn't destroy a unit and create another with the\n")
			log.Printf("warning: same name in a single turn, then there may be a bug here.\n")
		}

		// proactively patch some of the obscured locations.
		// turn reports initially gave obscured locations for from and to hexes.
		// around 0902-02, the current location stopped being obscured,
		// but the previous location is still obscured.
		// NB: links between the locations must be validated before patching them!
		updatedCurrentLinks, updatedPreviousLinks := 0, 0
		for _, turn := range consolidatedTurns {
			for _, unitMoves := range turn.UnitMoves {
				var prevTurnMoves *parser.Moves_t
				if turn.Prev != nil {
					prevTurnMoves = turn.Prev.UnitMoves[unitMoves.UnitId]
				}
				var nextTurnMoves *parser.Moves_t
				if turn.Next != nil {
					nextTurnMoves = turn.Next.UnitMoves[unitMoves.UnitId]
				}
				//if unitMoves.UnitId == "0988" {
				//	log.Printf("this: %s: %-6s: this prior %q current %q\n", unitMoves.TurnId, unitMoves.UnitId, unitMoves.FromHex, unitMoves.ToHex)
				//	if prevTurnMoves != nil {
				//		log.Printf("      %s: %-6s: prev prior %q current %q\n", prevTurnMoves.TurnId, prevTurnMoves.UnitId, prevTurnMoves.FromHex, prevTurnMoves.ToHex)
				//	}
				//	if nextTurnMoves != nil {
				//		log.Printf("      %s: %-6s: next prior %q current %q\n", nextTurnMoves.TurnId, nextTurnMoves.UnitId, nextTurnMoves.FromHex, nextTurnMoves.ToHex)
				//	}
				//}

				// link prior.ToHex and this.FromHex if this.FromHex is not obscured
				if !strings.HasPrefix(unitMoves.FromHex, "##") && prevTurnMoves != nil {
					if prevTurnMoves.ToHex != unitMoves.FromHex {
						updatedPreviousLinks++
						prevTurnMoves.ToHex = unitMoves.FromHex
					}
				}

				// link this.ToHex and next.FromHex if this.ToHex is not obscured
				if !strings.HasPrefix(unitMoves.ToHex, "##") && nextTurnMoves != nil {
					if unitMoves.ToHex != nextTurnMoves.FromHex {
						updatedCurrentLinks++
						nextTurnMoves.FromHex = unitMoves.ToHex
					}
				}
			}
		}
		log.Printf("updated %8d obscured 'Previous Hex' locations\n", updatedPreviousLinks)
		log.Printf("updated %8d obscured 'Current Hex'  locations\n", updatedCurrentLinks)

		// dangerous but try to find the origin hex if asked
		if argsRender.show.origin {
			for _, turn := range consolidatedTurns {
				for _, unit := range turn.SortedMoves {
					argsRender.mapper.Origin = unit.Location
					break
				}
			}
			log.Printf("info: origin hex set to %q\n", argsRender.mapper.Origin)
		}

		// dangerous, shift the map
		argsRender.mapper.Render.ShiftMap = argsRender.show.shiftMap
		if argsRender.mapper.Render.ShiftMap {
			log.Printf("warn: will shift map up and left\n")
		}

		// walk the data
		worldMap, err := turns.Walk(consolidatedTurns, consolidatedSpecialNames, argsRender.originGrid, argsRender.quitOnInvalidGrid, argsRender.warnOnInvalidGrid, argsRender.warnOnNewSettlement, argsRender.warnOnTerrainChange, gcfg.DebugFlags.Maps, gcfg.Experimental.ReverseWalker)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}

		if gcfg.Experimental.ReverseWalker {
			// walk through the data
			// todo: we should reverse the sort of the consolidated turns
			for _, turn := range consolidatedTurns {
				_, err = turns.WalkTurn(turn, consolidatedSpecialNames, true)
				if err != nil {
					log.Printf("error: %v\n", err)
				}
			}
		}

		// generate all the solo maps
		type soloMap_t struct {
			unit     string
			worldMap *tiles.Map_t
		}
		var soloMaps []soloMap_t

		if argsRender.soloElement != "" && len(gcfg.Worldographer.Solo) != 0 {
			log.Fatalf("error: solo-element and config.Worldographer.Solo are both set")
		} else if argsRender.soloElement != "" {
			log.Printf("info: rendering only %q\n", argsRender.soloElement)
			solo := worldMap.Solo(argsRender.soloElement)
			log.Printf("info: %s: world %d tiles: solo %d\n", argsRender.soloElement, len(worldMap.Tiles), len(solo.Tiles))
			worldMap = solo
		} else if len(gcfg.Worldographer.Solo) != 0 {
			upperLeft, lowerRight := worldMap.Bounds()
			log.Printf("info: world map:  pre-solo: upper left %4d: lower right %4d\n", upperLeft, lowerRight)

			// create solo maps
			for _, soloUnit := range gcfg.Worldographer.Solo {
				log.Printf("info: %s:%s: solo start %q: stop %q\n", turnId, soloUnit.Unit, soloUnit.StartTurn, soloUnit.StopTurn)
				if !(soloUnit.StartTurn <= turnId && turnId < soloUnit.StopTurn) {
					log.Printf("info: %s:%s: not rendering solo map (inactive)\n", turnId, soloUnit.Unit)
					continue
				}
				soloMap := worldMap.Solo(soloUnit.Unit)
				if len(soloMap.Tiles) == 0 {
					log.Printf("info: %s:%s: not rendering solo map (no tiles)\n", turnId, soloUnit.Unit)
					continue
				}
				log.Printf("info: %s:%s: rendering solo map\n", turnId, soloUnit.Unit)
				log.Printf("info: %s:%s: world %d tiles: solo %d\n", turnId, soloUnit.Unit, len(worldMap.Tiles), len(soloMap.Tiles))
				soloMaps = append(soloMaps, soloMap_t{
					unit:     soloUnit.Unit,
					worldMap: soloMap,
				})
			}

			// remove the solo map tiles from the world map
			for _, soloMap := range soloMaps {
				for _, tile := range soloMap.worldMap.Tiles {
					delete(worldMap.Tiles, tile.Location)
				}
			}
			upperLeft, lowerRight = worldMap.Bounds()
			log.Printf("info: world map: post-solo: upper left %4d: lower right %4d\n", upperLeft, lowerRight)
		}

		if gcfg.DebugFlags.DumpAllTurns {
			log.Printf("hey, dumping it all\n")
			for _, turn := range consolidatedTurns {
				log.Printf("%s: sortedMoves %d\n", turn.Id, len(turn.SortedMoves))
				for _, unit := range turn.SortedMoves {
					for _, move := range unit.Moves {
						if move.Report == nil {
							log.Fatalf("%s: %-6s: %6d: %2d: %s: %s\n", move.TurnId, unit.UnitId, move.LineNo, move.StepNo, move.CurrentHex, "missing report!")
						} else if move.Report.Terrain == terrain.Blank {
							if move.Result == results.Failed {
								log.Printf("%s: %-6s: %s: failed\n", move.TurnId, unit.UnitId, move.CurrentHex)
							} else if move.Still {
								log.Printf("%s: %-6s: %s: stayed in place\n", move.TurnId, unit.UnitId, move.CurrentHex)
							} else if move.Follows != "" {
								log.Printf("%s: %-6s: %s: follows %s\n", move.TurnId, unit.UnitId, move.CurrentHex, move.Follows)
							} else if move.GoesTo != "" {
								log.Printf("%s: %-6s: %s: goes to %s\n", move.TurnId, unit.UnitId, move.CurrentHex, move.GoesTo)
							} else {
								log.Fatalf("%s: %-6s: %6d: %2d: %s: %s\n", move.TurnId, unit.UnitId, move.LineNo, move.StepNo, move.CurrentHex, "missing terrain")
							}
						} else {
							log.Printf("%s: %-6s: %s: terrain %s\n", move.TurnId, unit.UnitId, move.CurrentHex, move.Report.Terrain)
						}
						for _, border := range move.Report.Borders {
							if border.Edge != edges.None {
								log.Printf("%s: %-6s: %s: border  %-14s %q\n", move.TurnId, unit.UnitId, move.CurrentHex, border.Direction, border.Edge)
							}
							if border.Terrain != terrain.Blank {
								log.Printf("%s: %-6s: %s: border  %-14s %q\n", move.TurnId, unit.UnitId, move.CurrentHex, border.Direction, border.Terrain)
							}
						}
						for _, point := range move.Report.FarHorizons {
							log.Printf("%s: %-6s: %s: compass %-14s sighted %q\n", move.TurnId, unit.UnitId, move.CurrentHex, point.Point, point.Terrain)
						}
						for _, settlement := range move.Report.Settlements {
							log.Printf("%s: %-6s: %s: village %q\n", move.TurnId, unit.UnitId, move.CurrentHex, settlement.Name)
						}
					}
				}
			}
		}

		if gcfg.DebugFlags.DumpAllTiles {
			worldMap.Dump()
		}

		// map the data
		upperLeft, lowerRight := worldMap.Bounds()
		log.Printf("map: upper left %4d: lower right %4d\n", upperLeft, lowerRight)
		wxxMap, err := actions.MapWorld(worldMap, consolidatedSpecialNames, parser.UnitId_t(argsRender.clanId), argsRender.mapper, globalConfig)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		log.Printf("map: %8d nodes: elapsed %v\n", worldMap.Length(), time.Since(started))

		// now we can create the Worldographer map!
		var mapName string
		if argsRender.experimental.blankMapSmall {
			mapName = filepath.Join(argsRender.paths.output, "blank-map.wxx")
		} else if argsRender.experimental.blankMapFull {
			mapName = filepath.Join(argsRender.paths.output, "blank-map-full.wxx")
		} else if argsRender.experimental.jsonMapSmall {
			mapName = filepath.Join(argsRender.paths.output, "json-map-small.json")
		} else if argsRender.saveWithTurnId {
			mapName = filepath.Join(argsRender.paths.output, fmt.Sprintf("%s.%s.wxx", maxTurnId, argsRender.clanId))
		} else {
			mapName = filepath.Join(argsRender.paths.output, fmt.Sprintf("%s.wxx", argsRender.clanId))
		}
		if argsRender.experimental.blankMapSmall {
			log.Printf("creating blank map %s\n", mapName)
			if err := wxxMap.CreateBlankMap(mapName, false); err != nil {
				log.Printf("creating %s\n", mapName)
				log.Fatalf("error: %v\n", err)
			}
		} else if argsRender.experimental.blankMapFull {
			log.Printf("creating blank map %s\n", mapName)
			if err := wxxMap.CreateBlankMap(mapName, true); err != nil {
				log.Printf("creating %s\n", mapName)
				log.Fatalf("error: %v\n", err)
			}
		} else if argsRender.experimental.jsonMapSmall {
			log.Printf("creating json map %s\n", mapName)
			if err := wxxMap.CreateJsonMap(mapName, true); err != nil {
				log.Printf("creating %s\n", mapName)
				log.Fatalf("error: %v\n", err)
			}
		} else if err := wxxMap.Create(mapName, turnId, upperLeft, lowerRight, argsRender.render, globalConfig); err != nil {
			log.Printf("creating %s\n", mapName)
			log.Fatalf("error: %v\n", err)
		}
		log.Printf("created  %s\n", mapName)

		// now we can create the solo maps
		for _, su := range soloMaps {
			unit, soloMap := su.unit, su.worldMap
			mapName := filepath.Join(argsRender.paths.output, fmt.Sprintf("%s.%s.wxx", turnId, unit))
			log.Printf("solo: %s:%s: creating %s\n", turnId, unit, mapName)
			upperLeft, lowerRight := soloMap.Bounds()
			log.Printf("solo: %s:%s: upper left %4d: lower right %4d\n", turnId, unit, upperLeft, lowerRight)
			wxxMap, err := actions.MapWorld(soloMap, consolidatedSpecialNames, parser.UnitId_t(argsRender.clanId), argsRender.mapper, globalConfig)
			if err != nil {
				log.Fatalf("error: %v\n", err)
			}
			log.Printf("solo: %s:%s: %8d nodes: elapsed %v\n", turnId, unit, soloMap.Length(), time.Since(started))
			if err := wxxMap.Create(mapName, turnId, upperLeft, lowerRight, argsRender.render, globalConfig); err != nil {
				log.Printf("creating %s\n", mapName)
				log.Fatalf("error: %v\n", err)
			}
			log.Printf("solo: %s:%s: created %s\n", turnId, unit, mapName)
		}

		log.Printf("elapsed: %v\n", time.Since(started))
	},
}

func strToTurnId(t string) (year, month int, err error) {
	fields := strings.Split(t, "-")
	if len(fields) != 2 {
		return 0, 0, fmt.Errorf("invalid date")
	}
	yyyy, mm, ok := strings.Cut(t, "-")
	if !ok {
		return 0, 0, fmt.Errorf("invalid date")
	} else if year, err = strconv.Atoi(yyyy); err != nil {
		return 0, 0, fmt.Errorf("invalid date")
	} else if month, err = strconv.Atoi(mm); err != nil {
		return 0, 0, fmt.Errorf("invalid date")
	} else if year < 899 || year > 9999 {
		return 0, 0, fmt.Errorf("invalid date")
	} else if month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("invalid date")
	}
	return year, month, nil
}
