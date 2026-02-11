// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package main implements the render CLI. This program reads one or more
// parser-generated JSON documents and produces a WXX file for Worldographer.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/maloquacious/semver"
	"github.com/playbymail/ottomap/internal/config"
	"github.com/playbymail/ottomap/internal/coords"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/domain"
	"github.com/playbymail/ottomap/internal/edges"
	"github.com/playbymail/ottomap/internal/resources"
	"github.com/playbymail/ottomap/internal/terrain"
	schema "github.com/playbymail/ottomap/internal/tniif"
	"github.com/playbymail/ottomap/internal/wxx"
	"github.com/spf13/cobra"
)

var (
	version = semver.Version{
		Major: 0,
		Minor: 84,
		Patch: 12,
		Build: semver.Commit(),
	}

	// create the default logger for the application
	logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
)

type loadedDoc_t struct {
	File string
	Doc  schema.Document
}

func main() {
	var clanNo, outputPath string
	var dumpMerged bool

	cmdRoot := &cobra.Command{
		Use:           "render [json-files...]",
		Short:         "render JSON documents to WXX",
		Long:          `Render parser-generated JSON documents into a Worldographer WXX map file.`,
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Root().PersistentFlags()
			logLevel, err := flags.GetString("log-level")
			if err != nil {
				return err
			}
			logSource, err := flags.GetBool("log-source")
			if err != nil {
				return err
			}
			debug, err := flags.GetBool("debug")
			if err != nil {
				return err
			}
			quiet, err := flags.GetBool("quiet")
			if err != nil {
				return err
			}
			if debug && quiet {
				return fmt.Errorf("--debug and --quiet are mutually exclusive")
			}
			var lvl slog.Level
			switch {
			case debug:
				lvl = slog.LevelDebug
			case quiet:
				lvl = slog.LevelError
			default:
				switch strings.ToLower(logLevel) {
				case "debug":
					lvl = slog.LevelDebug
				case "info":
					lvl = slog.LevelInfo
				case "warn", "warning":
					lvl = slog.LevelWarn
				case "error":
					lvl = slog.LevelError
				default:
					return fmt.Errorf("log-level: unknown value %q", logLevel)
				}
			}
			handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
				Level:     lvl,
				AddSource: logSource || lvl == slog.LevelDebug,
			})
			logger = slog.New(handler)
			slog.SetDefault(logger)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			started := time.Now()

			if len(clanNo) != 4 || clanNo[0] != '0' {
				return fmt.Errorf("invalid clan %q: must be 4 digits starting with 0", clanNo)
			}
			for _, ch := range clanNo {
				if ch < '0' || ch > '9' {
					return fmt.Errorf("invalid clan %q: must be 4 digits starting with 0", clanNo)
				}
			}

			if outputPath != "" {
				if !strings.HasSuffix(outputPath, ".wxx") {
					return fmt.Errorf("output %q: must end in .wxx", outputPath)
				}
				parentDir := filepath.Dir(outputPath)
				if sb, err := os.Stat(parentDir); err != nil {
					return fmt.Errorf("output %q: parent directory: %w", outputPath, err)
				} else if !sb.IsDir() {
					return fmt.Errorf("output %q: parent %q is not a directory", outputPath, parentDir)
				}
			}

			loaded, err := loadDocuments(args)
			if err != nil {
				return err
			}
			logger.Debug("render", "loaded", len(loaded), "clan", clanNo, "output", outputPath)

			if errs := validateDocuments(loaded); len(errs) > 0 {
				for _, e := range errs {
					logger.Error("validate", "error", e)
				}
				return fmt.Errorf("validation failed (%d errors)", len(errs))
			}

			events, flattenErrs := flattenEvents(loaded)
			if len(flattenErrs) > 0 {
				for _, e := range flattenErrs {
					logger.Error("flatten", "error", e)
				}
				return fmt.Errorf("flatten failed (%d errors)", len(flattenErrs))
			}

			owningClan := schema.ClanID(clanNo)
			sortEvents(events, owningClan)

			tiles := mergeTiles(events)
			if len(tiles) == 0 {
				return fmt.Errorf("no tiles to render")
			}

			if dumpMerged {
				dumpMergedTiles(tiles)
			}

			upperLeft, lowerRight, offset := computeBoundsAndOffset(tiles)
			logger.Info("render",
				"tiles", len(tiles),
				"upperLeft", upperLeft.GridString(),
				"lowerRight", lowerRight.GridString(),
				"offset", offset.GridString(),
			)

			var hexes []*wxx.Hex
			var convertErrs []error
			locs := make([]coords.Map, 0, len(tiles))
			for loc := range tiles {
				locs = append(locs, loc)
			}
			sort.Slice(locs, func(i, j int) bool {
				if locs[i].Column != locs[j].Column {
					return locs[i].Column < locs[j].Column
				}
				return locs[i].Row < locs[j].Row
			})
			for _, loc := range locs {
				hex, errs := convertTileToHex(tiles[loc], offset, owningClan)
				if len(errs) > 0 {
					convertErrs = append(convertErrs, errs...)
				}
				if hex != nil {
					hexes = append(hexes, hex)
				}
			}
			if len(convertErrs) > 0 {
				for _, e := range convertErrs {
					logger.Error("convert", "error", e)
				}
				return fmt.Errorf("convert failed (%d errors)", len(convertErrs))
			}

			specials := collectSpecialHexes(loaded)
			applySpecialHexes(hexes, specials)

			fmt.Printf("bounds: %s to %s\n", upperLeft.GridString(), lowerRight.GridString())

			if outputPath == "" {
				logger.Info("render", "elapsed", time.Since(started).String())
				return nil
			}

			gcfg := config.Default()
			w, err := wxx.NewWXX(gcfg)
			if err != nil {
				return fmt.Errorf("wxx: %w", err)
			}

			for _, hex := range hexes {
				if err := w.MergeHex(hex); err != nil {
					return fmt.Errorf("wxx: merge %s: %w", hex.Location.GridString(), err)
				}
			}

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
				return fmt.Errorf("wxx: create: %w", err)
			}

			fmt.Printf("wrote %s\n", outputPath)
			logger.Info("render", "elapsed", time.Since(started).String())
			return nil
		},
	}

	cmdRoot.PersistentFlags().Bool("debug", false, "enable debug logging (same as --log-level=debug)")
	cmdRoot.PersistentFlags().Bool("quiet", false, "only log errors (same as --log-level=error)")
	cmdRoot.PersistentFlags().String("log-level", "error", "logging level (debug|info|warn|error)")
	cmdRoot.PersistentFlags().Bool("log-source", false, "add file and line numbers to log messages")

	cmdRoot.Flags().StringVar(&clanNo, "clan", "", "clan that owns the data (e.g. 0987)")
	if err := cmdRoot.MarkFlagRequired("clan"); err != nil {
		log.Fatalf("error: clan: %v\n", err)
	}
	cmdRoot.Flags().StringVar(&outputPath, "output", "", "path for WXX output file (must end in .wxx)")
	cmdRoot.Flags().BoolVar(&dumpMerged, "dump-merged", false, "write merged tile state as JSON to stdout")

	cmdRoot.AddCommand(cmdVersion())

	if err := cmdRoot.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func cmdVersion() *cobra.Command {
	showBuildInfo := false
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "display the application's version number",
		RunE: func(cmd *cobra.Command, args []string) error {
			if showBuildInfo {
				fmt.Println(version.String())
				return nil
			}
			fmt.Println(version.Core())
			return nil
		},
	}
	cmd.Flags().BoolVar(&showBuildInfo, "build-info", showBuildInfo, "show build information")
	return cmd
}

func loadDocuments(paths []string) ([]loadedDoc_t, error) {
	loaded := make([]loadedDoc_t, 0, len(paths))
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		sb, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", absPath, err)
		}
		if !sb.Mode().IsRegular() {
			return nil, fmt.Errorf("%s: not a regular file", absPath)
		}
		data, err := os.ReadFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", absPath, err)
		}
		var doc schema.Document
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("%s: %w", absPath, err)
		}
		logger.Debug("loadDocuments", "file", absPath, "schema", doc.Schema, "game", doc.Game, "turn", doc.Turn)
		loaded = append(loaded, loadedDoc_t{File: filepath.Base(absPath), Doc: doc})
	}
	logger.Debug("loadDocuments", "loaded", len(loaded))
	return loaded, nil
}

func validateDocuments(loaded []loadedDoc_t) []error {
	var errs []error

	for i, ld := range loaded {
		doc := ld.Doc
		prefix := ld.File
		if prefix == "" {
			prefix = fmt.Sprintf("doc[%d]", i)
		}

		if doc.Schema != schema.Version {
			errs = append(errs, fmt.Errorf("%s: schema: got %q, want %q", prefix, doc.Schema, schema.Version))
		}
		if doc.Game == "" {
			errs = append(errs, fmt.Errorf("%s: game is required", prefix))
		}
		if doc.Clan == "" {
			errs = append(errs, fmt.Errorf("%s: clan is required", prefix))
		}

		turn := string(doc.Turn)
		if len(turn) != 7 || turn[4] != '-' {
			errs = append(errs, fmt.Errorf("%s: turn %q: must be YYYY-MM format", prefix, turn))
		} else {
			validDigits := true
			for _, ch := range turn[:4] {
				if ch < '0' || ch > '9' {
					validDigits = false
					break
				}
			}
			for _, ch := range turn[5:] {
				if ch < '0' || ch > '9' {
					validDigits = false
					break
				}
			}
			if !validDigits {
				errs = append(errs, fmt.Errorf("%s: turn %q: must be YYYY-MM format", prefix, turn))
			}
		}

		for ci, clan := range doc.Clans {
			clanPrefix := fmt.Sprintf("%s: clans[%d]", prefix, ci)
			if clan.ID == "" {
				errs = append(errs, fmt.Errorf("%s: id is required", clanPrefix))
			}

			for ui, unit := range clan.Units {
				unitPrefix := fmt.Sprintf("%s: unit %s", prefix, unit.ID)
				if unit.ID == "" {
					unitPrefix = fmt.Sprintf("%s: clans[%d].units[%d]", prefix, ci, ui)
					errs = append(errs, fmt.Errorf("%s: id is required", unitPrefix))
				}

				if unit.EndingLocation != "" {
					if _, err := coords.HexToMap(string(unit.EndingLocation)); err != nil {
						errs = append(errs, fmt.Errorf("%s: endingLocation %q: %w", unitPrefix, unit.EndingLocation, err))
					}
				}

				for mi, moves := range unit.Moves {
					for si, step := range moves.Steps {
						stepPrefix := fmt.Sprintf("%s: step %d", unitPrefix, si+1)
						if len(unit.Moves) > 1 {
							stepPrefix = fmt.Sprintf("%s: moves[%d].step %d", unitPrefix, mi, si+1)
						}

						if step.EndingLocation != "" {
							if _, err := coords.HexToMap(string(step.EndingLocation)); err != nil {
								errs = append(errs, fmt.Errorf("%s: endingLocation %q: %w", stepPrefix, step.EndingLocation, err))
							}
						}

						if step.Observation != nil {
							obsPrefix := stepPrefix + ": observation"
							if step.Observation.Location != "" {
								if _, err := coords.HexToMap(string(step.Observation.Location)); err != nil {
									errs = append(errs, fmt.Errorf("%s: location %q: %w", obsPrefix, step.Observation.Location, err))
								}
							}
							for ei, edge := range step.Observation.Edges {
								if err := edge.Dir.Validate(); err != nil {
									errs = append(errs, fmt.Errorf("%s: edges[%d]: %w", obsPrefix, ei, err))
								}
							}
							for cpi, cp := range step.Observation.CompassPoints {
								if err := cp.Bearing.Validate(); err != nil {
									errs = append(errs, fmt.Errorf("%s: compassPoints[%d]: %w", obsPrefix, cpi, err))
								}
							}
						}
					}
				}

				for sci, scout := range unit.Scouts {
					for si, step := range scout.Steps {
						stepPrefix := fmt.Sprintf("%s: scout %s: step %d", unitPrefix, scout.ID, si+1)
						if scout.ID == "" {
							stepPrefix = fmt.Sprintf("%s: scouts[%d].step %d", unitPrefix, sci, si+1)
						}

						if step.EndingLocation != "" {
							if _, err := coords.HexToMap(string(step.EndingLocation)); err != nil {
								errs = append(errs, fmt.Errorf("%s: endingLocation %q: %w", stepPrefix, step.EndingLocation, err))
							}
						}

						if step.Observation != nil {
							obsPrefix := stepPrefix + ": observation"
							if step.Observation.Location != "" {
								if _, err := coords.HexToMap(string(step.Observation.Location)); err != nil {
									errs = append(errs, fmt.Errorf("%s: location %q: %w", obsPrefix, step.Observation.Location, err))
								}
							}
							for ei, edge := range step.Observation.Edges {
								if err := edge.Dir.Validate(); err != nil {
									errs = append(errs, fmt.Errorf("%s: edges[%d]: %w", obsPrefix, ei, err))
								}
							}
							for cpi, cp := range step.Observation.CompassPoints {
								if err := cp.Bearing.Validate(); err != nil {
									errs = append(errs, fmt.Errorf("%s: compassPoints[%d]: %w", obsPrefix, cpi, err))
								}
							}
						}
					}
				}
			}
		}
	}

	if len(loaded) > 1 {
		game := loaded[0].Doc.Game
		for i := 1; i < len(loaded); i++ {
			if loaded[i].Doc.Game != game {
				errs = append(errs, fmt.Errorf("%s: game %q does not match %s game %q",
					loaded[i].File, loaded[i].Doc.Game, loaded[0].File, game))
			}
		}
	}

	return errs
}

type obsEvent_t struct {
	Turn       schema.TurnID
	Clan       schema.ClanID
	Unit       string
	Loc        coords.Map
	Obs        *schema.Observation
	WasVisited bool
	WasScouted bool
}

func flattenEvents(loaded []loadedDoc_t) ([]obsEvent_t, []error) {
	var events []obsEvent_t
	var errs []error

	for _, ld := range loaded {
		doc := ld.Doc
		for _, clan := range doc.Clans {
			for _, unit := range clan.Units {
				for _, moves := range unit.Moves {
					for si, step := range moves.Steps {
						if step.Observation == nil {
							continue
						}
						loc, err := coords.HexToMap(string(step.Observation.Location))
						if err != nil {
							errs = append(errs, fmt.Errorf("%s: unit %s: step %d: observation location %q: %w",
								ld.File, unit.ID, si+1, step.Observation.Location, err))
							continue
						}
						events = append(events, obsEvent_t{
							Turn:       doc.Turn,
							Clan:       clan.ID,
							Unit:       string(unit.ID),
							Loc:        loc,
							Obs:        step.Observation,
							WasVisited: step.Observation.WasVisited,
							WasScouted: step.Observation.WasScouted,
						})
					}
				}

				for _, scout := range unit.Scouts {
					for si, step := range scout.Steps {
						if step.Observation == nil {
							continue
						}
						loc, err := coords.HexToMap(string(step.Observation.Location))
						if err != nil {
							errs = append(errs, fmt.Errorf("%s: unit %s: scout %s: step %d: observation location %q: %w",
								ld.File, unit.ID, scout.ID, si+1, step.Observation.Location, err))
							continue
						}
						events = append(events, obsEvent_t{
							Turn:       doc.Turn,
							Clan:       clan.ID,
							Unit:       string(scout.ID),
							Loc:        loc,
							Obs:        step.Observation,
							WasVisited: step.Observation.WasVisited,
							WasScouted: true,
						})
					}
				}
			}
		}
	}

	logger.Debug("flattenEvents", "events", len(events), "errors", len(errs))
	return events, errs
}

func sortEvents(events []obsEvent_t, owningClan schema.ClanID) {
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].Turn != events[j].Turn {
			return events[i].Turn < events[j].Turn
		}
		ri, rj := 0, 0
		if events[i].Clan == owningClan {
			ri = 1
		}
		if events[j].Clan == owningClan {
			rj = 1
		}
		if ri != rj {
			return ri < rj
		}
		if events[i].Clan != events[j].Clan {
			return events[i].Clan < events[j].Clan
		}
		return events[i].Unit < events[j].Unit
	})
	logger.Debug("sortEvents", "events", len(events), "owningClan", owningClan)
}

type tileState_t struct {
	Loc           coords.Map
	Terrain       schema.Terrain
	Edges         map[schema.Direction]schema.Edge
	Resources     []schema.Resource
	Settlements   []schema.Settlement
	Encounters    []schema.Encounter
	CompassPoints []schema.CompassPoint
	WasVisited    bool
	WasScouted    bool
	Notes         []schema.Note
}

func dumpMergedTiles(tiles map[coords.Map]*tileState_t) {
	type dumpTile struct {
		Location      string                        `json:"location"`
		Terrain       schema.Terrain                `json:"terrain,omitempty"`
		Edges         map[schema.Direction]schema.Edge `json:"edges,omitempty"`
		Resources     []schema.Resource             `json:"resources,omitempty"`
		Settlements   []schema.Settlement           `json:"settlements,omitempty"`
		Encounters    []schema.Encounter            `json:"encounters,omitempty"`
		CompassPoints []schema.CompassPoint         `json:"compassPoints,omitempty"`
		WasVisited    bool                          `json:"wasVisited,omitempty"`
		WasScouted    bool                          `json:"wasScouted,omitempty"`
		Notes         []schema.Note                 `json:"notes,omitempty"`
	}

	locs := make([]coords.Map, 0, len(tiles))
	for loc := range tiles {
		locs = append(locs, loc)
	}
	sort.Slice(locs, func(i, j int) bool {
		if locs[i].Column != locs[j].Column {
			return locs[i].Column < locs[j].Column
		}
		return locs[i].Row < locs[j].Row
	})

	out := make([]dumpTile, 0, len(locs))
	for _, loc := range locs {
		ts := tiles[loc]
		out = append(out, dumpTile{
			Location:      ts.Loc.GridString(),
			Terrain:       ts.Terrain,
			Edges:         ts.Edges,
			Resources:     ts.Resources,
			Settlements:   ts.Settlements,
			Encounters:    ts.Encounters,
			CompassPoints: ts.CompassPoints,
			WasVisited:    ts.WasVisited,
			WasScouted:    ts.WasScouted,
			Notes:         ts.Notes,
		})
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		logger.Error("dump-merged", "error", err)
		return
	}
	fmt.Println(string(data))
}

func mergeTiles(events []obsEvent_t) map[coords.Map]*tileState_t {
	tiles := make(map[coords.Map]*tileState_t)

	for _, ev := range events {
		if ev.Obs == nil {
			continue
		}

		ts, ok := tiles[ev.Loc]
		if !ok {
			ts = &tileState_t{
				Loc:   ev.Loc,
				Edges: make(map[schema.Direction]schema.Edge),
			}
			tiles[ev.Loc] = ts
		}

		obs := ev.Obs

		if obs.Terrain != "" {
			ts.Terrain = obs.Terrain
		}

		if obs.WasVisited {
			ts.WasVisited = true
		}
		if obs.WasScouted {
			ts.WasScouted = true
		}

		for _, edge := range obs.Edges {
			if edge.Feature != "" || edge.NeighborTerrain != "" || edge.RawEdge != "" {
				ts.Edges[edge.Dir] = edge
			}
		}

		if obs.Resources != nil {
			ts.Resources = obs.Resources
		}
		if obs.Settlements != nil {
			ts.Settlements = obs.Settlements
		}
		if obs.Encounters != nil {
			ts.Encounters = obs.Encounters
		}
		if obs.CompassPoints != nil {
			ts.CompassPoints = obs.CompassPoints
		}

		ts.Notes = append(ts.Notes, obs.Notes...)
	}

	logger.Debug("mergeTiles", "tiles", len(tiles))
	return tiles
}

func convertTileToHex(ts *tileState_t, renderOffset coords.Map, owningClan schema.ClanID) (*wxx.Hex, []error) {
	var errs []error

	hex := &wxx.Hex{
		Location: ts.Loc,
		RenderAt: coords.Map{
			Column: ts.Loc.Column - renderOffset.Column,
			Row:    ts.Loc.Row - renderOffset.Row,
		},
		WasVisited: ts.WasVisited,
		WasScouted: ts.WasScouted,
	}

	if ts.Terrain != "" {
		t, ok := terrain.StringToEnum[string(ts.Terrain)]
		if !ok {
			errs = append(errs, fmt.Errorf("tile %s: unknown terrain %q", ts.Loc.GridString(), ts.Terrain))
		} else {
			hex.Terrain = t
		}
	}

	for schemaDir, edge := range ts.Edges {
		dir, ok := direction.StringToEnum[string(schemaDir)]
		if !ok {
			errs = append(errs, fmt.Errorf("tile %s: unknown direction %q", ts.Loc.GridString(), schemaDir))
			continue
		}
		if edge.Feature == "" {
			continue
		}
		feat, ok := edges.StringToEnum[string(edge.Feature)]
		if !ok {
			errs = append(errs, fmt.Errorf("tile %s: unknown edge feature %q", ts.Loc.GridString(), edge.Feature))
			continue
		}
		switch feat {
		case edges.Canal:
			hex.Features.Edges.Canal = append(hex.Features.Edges.Canal, dir)
		case edges.Ford:
			hex.Features.Edges.Ford = append(hex.Features.Edges.Ford, dir)
		case edges.Pass:
			hex.Features.Edges.Pass = append(hex.Features.Edges.Pass, dir)
		case edges.River:
			hex.Features.Edges.River = append(hex.Features.Edges.River, dir)
		case edges.StoneRoad:
			hex.Features.Edges.StoneRoad = append(hex.Features.Edges.StoneRoad, dir)
		}
	}

	for _, r := range ts.Resources {
		res, ok := resources.StringToEnum[string(r)]
		if !ok {
			errs = append(errs, fmt.Errorf("tile %s: unknown resource %q", ts.Loc.GridString(), r))
			continue
		}
		if res != resources.None {
			hex.Features.Resources = append(hex.Features.Resources, res)
		}
	}

	for _, s := range ts.Settlements {
		hex.Features.Settlements = append(hex.Features.Settlements, &domain.Settlement_t{
			Name: string(s.Name),
		})
	}

	clanAsUnitId := domain.UnitId_t(owningClan)
	for _, e := range ts.Encounters {
		unitId := domain.UnitId_t(e.Unit)
		hex.Features.Encounters = append(hex.Features.Encounters, &domain.Encounter_t{
			UnitId:   unitId,
			Friendly: unitId.InClan(clanAsUnitId),
		})
	}

	logger.Debug("convertTileToHex", "loc", ts.Loc.GridString(), "terrain", hex.Terrain, "edges", len(ts.Edges), "errors", len(errs))
	return hex, errs
}

func computeBoundsAndOffset(tiles map[coords.Map]*tileState_t) (upperLeft, lowerRight, offset coords.Map) {
	const (
		borderWidth  = 4
		borderHeight = 4
	)

	first := true
	for loc := range tiles {
		if first {
			upperLeft, lowerRight = loc, loc
			first = false
			continue
		}
		if loc.Column < upperLeft.Column {
			upperLeft.Column = loc.Column
		}
		if loc.Row < upperLeft.Row {
			upperLeft.Row = loc.Row
		}
		if loc.Column > lowerRight.Column {
			lowerRight.Column = loc.Column
		}
		if loc.Row > lowerRight.Row {
			lowerRight.Row = loc.Row
		}
	}

	if first {
		return
	}

	minColInGrid := (upperLeft.Column / 30) * 30
	minRowInGrid := (upperLeft.Row / 21) * 21

	if upperLeft.Column > borderWidth {
		offset.Column = upperLeft.Column - borderWidth
	}
	if offset.Column < minColInGrid {
		offset.Column = minColInGrid
	}

	if upperLeft.Row > borderHeight {
		offset.Row = upperLeft.Row - borderHeight
	}
	if offset.Row < minRowInGrid {
		offset.Row = minRowInGrid
	}

	if offset.Row%2 != 0 {
		if offset.Row-1 >= minRowInGrid {
			offset.Row--
		}
	}

	if offset.Column%2 != 0 {
		if offset.Column-1 >= minColInGrid {
			offset.Column--
		}
	}

	logger.Debug("computeBoundsAndOffset",
		"upperLeft", upperLeft.GridString(),
		"lowerRight", lowerRight.GridString(),
		"offset", offset.GridString(),
	)
	return upperLeft, lowerRight, offset
}

func collectSpecialHexes(loaded []loadedDoc_t) map[string]*domain.Special_t {
	specials := make(map[string]*domain.Special_t)
	for _, ld := range loaded {
		for _, sh := range ld.Doc.SpecialHexes {
			key := strings.ToLower(sh.Name)
			if _, exists := specials[key]; !exists {
				specials[key] = &domain.Special_t{
					Id:   key,
					Name: sh.Name,
				}
			}
		}
	}
	logger.Debug("collectSpecialHexes", "specials", len(specials))
	return specials
}

func applySpecialHexes(hexes []*wxx.Hex, specials map[string]*domain.Special_t) {
	if len(specials) == 0 {
		return
	}
	for _, hex := range hexes {
		var remaining []*domain.Settlement_t
		for _, s := range hex.Features.Settlements {
			key := strings.ToLower(s.Name)
			if sp, ok := specials[key]; ok {
				duplicate := false
				for _, existing := range hex.Features.Special {
					if strings.ToLower(existing.Name) == key {
						duplicate = true
						break
					}
				}
				if !duplicate {
					hex.Features.Special = append(hex.Features.Special, sp)
				}
			} else {
				remaining = append(remaining, s)
			}
		}
		hex.Features.Settlements = remaining
	}
	logger.Debug("applySpecialHexes", "hexes", len(hexes), "specials", len(specials))
}
