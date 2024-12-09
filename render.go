// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"bytes"
	"fmt"
	"github.com/playbymail/ottomap/actions"
	"github.com/playbymail/ottomap/internal/edges"
	"github.com/playbymail/ottomap/internal/parser"
	"github.com/playbymail/ottomap/internal/results"
	"github.com/playbymail/ottomap/internal/terrain"
	"github.com/playbymail/ottomap/internal/turns"
	"github.com/playbymail/ottomap/internal/wxx"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
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
	acceptLoneDash      bool
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
		dumpAllTiles  bool
		dumpAllTurns  bool
		fleetMovement bool
		logFile       bool
		logTime       bool
		maps          bool
		merge         bool
		nodes         bool
		parser        bool
		sections      bool
		steps         bool
	}
	experimental struct {
		newWaterTiles      bool
		cleanUpScoutStill  bool
		splitTrailingUnits bool
		stripCR            bool
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
		logFlags := 0
		if argsRender.debug.logFile {
			logFlags |= log.Lshortfile
		}
		if argsRender.debug.logTime {
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
		} else if yyyy, mm, ok := strings.Cut(argsRender.maxTurn.id, "-"); !ok {
			log.Fatalf("error: turn-cutoff %q: must be yyyy-mm format", argsRender.maxTurn.id)
		} else if year, err := strconv.Atoi(yyyy); err != nil {
			log.Fatalf("error: turn-cutoff %q: must be yyyy-mm format", argsRender.maxTurn.id)
		} else if month, err := strconv.Atoi(mm); err != nil {
			log.Fatalf("error: turn-cutoff %q: must be yyyy-mm format", argsRender.maxTurn.id)
		} else if year < 899 || year > 9999 {
			log.Fatalf("error: turn-cutoff %q: invalid year %d", argsRender.maxTurn.id, year)
		} else if month < 1 || month > 12 {
			log.Fatalf("error: turn-cutoff %q: invalid month %d", argsRender.maxTurn.id, month)
		} else {
			argsRender.maxTurn.year, argsRender.maxTurn.month = year, month
		}

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
		if argsRoot.showVersion {
			log.Printf("ottomap version %s\n", version)
		}
		if argsRoot.soloClan {
			log.Printf("clan %q: running solo\n", argsRender.clanId)
		}

		if argsRender.experimental.newWaterTiles {
			log.Printf("experimental: newWaterTiles enabled\n")
			terrain.TileTerrainNames[terrain.Lake] = "Water Sea"
			terrain.TileTerrainNames[terrain.Ocean] = "Water Ocean"
		}

		argsRender.originGrid = "RR"
		argsRender.quitOnInvalidGrid = false

		started := time.Now()
		log.Printf("data:   %s\n", argsRender.paths.data)
		log.Printf("input:  %s\n", argsRender.paths.input)
		log.Printf("output: %s\n", argsRender.paths.output)

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
			} else if argsRender.experimental.stripCR {
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
			turn, err := parser.ParseInput(i.Id, turnId, data, argsRender.acceptLoneDash, argsRender.debug.parser, argsRender.debug.sections, argsRender.debug.steps, argsRender.debug.nodes, argsRender.debug.fleetMovement, argsRender.experimental.splitTrailingUnits, argsRender.experimental.cleanUpScoutStill, argsRender.parser)
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
				}
				if unitTurn.SpecialNames != nil {
					// consolidate any the special hexes
					for id, special := range unitTurn.SpecialNames {
						consolidatedSpecialNames[id] = special
					}
				}
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

		// check for N/A values in locations and quit if we find any
		naLocationCount := 0
		for _, turn := range consolidatedTurns {
			for _, unitMoves := range turn.UnitMoves {
				if unitMoves.FromHex == "N/A" {
					naLocationCount++
					log.Printf("%s: %-6s: location %q: invalid location\n", unitMoves.TurnId, unitMoves.UnitId, unitMoves.FromHex)
				}
			}
		}
		if naLocationCount != 0 {
			log.Fatalf("please update the invalid locations and restart\n")
		}

		// sanity check on the current and prior locations.
		badLinks, goodLinks := 0, 0
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
					badLinks++
					log.Printf("error: %s: %-6s: from %q\n", turn.Id, unitMoves.UnitId, unitMoves.ToHex)
					log.Printf("     : %s: %-6s: to   %q\n", turn.Next.Id, nextUnitMoves.UnitId, nextUnitMoves.FromHex)
				} else {
					goodLinks++
				}
				nextUnitMoves.FromHex = unitMoves.ToHex
			}
		}
		log.Printf("links: %d good, %d bad\n", goodLinks, badLinks)
		if badLinks != 0 {
			// this should never happen. if it does then something is wrong with the report generator.
			log.Printf("sorry: the previous and current hexes don't align in some reports\n")
			log.Fatalf("please report this error")
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
				//if unitMoves.Id == "0138" {
				//	log.Printf("this: %s: %-6s: this prior %q current %q\n", unitMoves.TurnId, unitMoves.Id, unitMoves.FromHex, unitMoves.ToHex)
				//	if prevTurnMoves != nil {
				//		log.Printf("      %s: %-6s: prev prior %q current %q\n", prevTurnMoves.TurnId, prevTurnMoves.Id, prevTurnMoves.FromHex, prevTurnMoves.ToHex)
				//	}
				//	if nextTurnMoves != nil {
				//		log.Printf("      %s: %-6s: next prior %q current %q\n", nextTurnMoves.TurnId, nextTurnMoves.Id, nextTurnMoves.FromHex, nextTurnMoves.ToHex)
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
		worldMap, err := turns.Walk(consolidatedTurns, consolidatedSpecialNames, argsRender.originGrid, argsRender.quitOnInvalidGrid, argsRender.warnOnInvalidGrid, argsRender.warnOnNewSettlement, argsRender.warnOnTerrainChange, argsRender.debug.maps)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		if argsRender.soloElement != "" {
			log.Printf("info: rendering only %q\n", argsRender.soloElement)
			solo := worldMap.Solo(argsRender.soloElement)
			log.Printf("info: %s: world %d tiles: solo %d\n", argsRender.soloElement, len(worldMap.Tiles), len(solo.Tiles))
			worldMap = solo
		}

		if argsRender.debug.dumpAllTurns {
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
		upperLeft, lowerRight := worldMap.Bounds()

		if argsRender.debug.dumpAllTiles {
			worldMap.Dump()
		}

		// map the data
		wxxMap, err := actions.MapWorld(worldMap, consolidatedSpecialNames, parser.UnitId_t(argsRender.clanId), argsRender.mapper)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		log.Printf("map: %8d nodes: elapsed %v\n", worldMap.Length(), time.Since(started))

		// now we can create the Worldographer map!
		var mapName string
		if argsRender.saveWithTurnId {
			mapName = filepath.Join(argsRender.paths.output, fmt.Sprintf("%s.%s.wxx", maxTurnId, argsRender.clanId))
		} else {
			mapName = filepath.Join(argsRender.paths.output, fmt.Sprintf("%s.wxx", argsRender.clanId))
		}
		if err := wxxMap.Create(mapName, turnId, upperLeft, lowerRight, argsRender.render); err != nil {
			log.Printf("creating %s\n", mapName)
			log.Fatalf("error: %v\n", err)
		}
		log.Printf("created  %s\n", mapName)

		log.Printf("elapsed: %v\n", time.Since(started))
	},
}
