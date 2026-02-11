// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package main implements the parser CLI. This program parses a single
// turn report and logs some information about it. In a future version,
// it will create a JSON file.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/maloquacious/semver"
	"github.com/playbymail/ottomap/internal/compass"
	"github.com/playbymail/ottomap/internal/direction"
	"github.com/playbymail/ottomap/internal/domain"
	"github.com/playbymail/ottomap/internal/edges"
	"github.com/playbymail/ottomap/internal/parser"
	"github.com/playbymail/ottomap/internal/resources"
	"github.com/playbymail/ottomap/internal/results"
	"github.com/playbymail/ottomap/internal/terrain"
	"github.com/playbymail/ottomap/internal/tndocx"
	schema "github.com/playbymail/ottomap/internal/tniif"
	"github.com/spf13/cobra"
)

var (
	version = semver.Version{
		Major: 0,
		Minor: 3,
		Patch: 3,
		Build: semver.Commit(),
	}
	logger *slog.Logger
)

func main() {
	var clanNo, game, path, outputPath string
	var acceptLoneDash, debugParser, debugSections, debugSteps, debugNodes, debugFleetMovement bool
	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	addFlags := func(cmd *cobra.Command) error {
		cmd.PersistentFlags().Bool("debug", false, "enable debug logging (same as --log-level=debug)")
		cmd.PersistentFlags().Bool("quiet", false, "only log errors (same as --log-level=error)")
		cmd.PersistentFlags().String("log-level", "error", "logging level (debug|info|warn|error))")
		cmd.PersistentFlags().Bool("log-source", false, "add file and line numbers to log messages")
		cmd.Flags().StringVar(&path, "input", path, "turn report to parse")
		if err := cmd.MarkFlagRequired("input"); err != nil {
			return err
		}
		cmd.Flags().StringVar(&game, "game", game, "game code")
		if err := cmd.MarkFlagRequired("game"); err != nil {
			log.Fatalf("error: game: %v\n", err)
		}
		cmd.Flags().StringVar(&clanNo, "clan", clanNo, "clan that owns the input data")
		if err := cmd.MarkFlagRequired("clan"); err != nil {
			log.Fatalf("error: clan: %v\n", err)
		}
		cmd.Flags().StringVar(&outputPath, "output", "", "write results to file instead of stdout")
		cmd.Flags().BoolVar(&acceptLoneDash, "accept-lone-dash", false, "accept lone dash in movement lines")
		cmd.Flags().BoolVar(&debugParser, "debug-parser", false, "enable parser debug logging")
		cmd.Flags().BoolVar(&debugSections, "debug-sections", false, "enable section debug logging")
		cmd.Flags().BoolVar(&debugSteps, "debug-steps", false, "enable step debug logging")
		cmd.Flags().BoolVar(&debugNodes, "debug-nodes", false, "enable node debug logging")
		cmd.Flags().BoolVar(&debugFleetMovement, "debug-fleet-movement", false, "enable fleet movement debug logging")
		return nil
	}

	cmdRoot := &cobra.Command{
		Use:           "parser",
		Short:         "tribenet report parser",
		Long:          `Parse TribeNet turn reports.`,
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
			slog.SetDefault(logger) // optional, but convenient
			if len(game) != 4 {
				return fmt.Errorf("invalid game %q", game)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := strconv.Atoi(clanNo)
			if err != nil || n < 1 || n > 999 {
				return fmt.Errorf("clan-id must be a number in the range 1..999")
			}
			clanID := fmt.Sprintf("%04d", n)
			logger.Info("parser", "clan-id", clanID)

			path, err = filepath.Abs(path)
			if err != nil {
				logger.Error("parser: invalid path", "error", err)
				return err
			}

			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".txt" && ext != ".docx" {
				return fmt.Errorf("input must be a .txt or .docx file")
			}

			if sb, err := os.Stat(path); err != nil {
				logger.Error("parser: invalid path", "error", err)
				return err
			} else if sb.IsDir() {
				return fmt.Errorf("path is a folder, not a file")
			} else if !sb.Mode().IsRegular() {
				return fmt.Errorf("path must be a regular file")
			}

			started := time.Now()
			logger.Info("parser", "input", path)

			var data []byte
			switch ext {
			case ".docx":
				data, err = tndocx.ParsePath(path, false, true)
			case ".txt":
				data, err = os.ReadFile(path)
			}
			if err != nil {
				logger.Error("parser", "error", err)
				return err
			}
			if len(data) == 0 {
				logger.Error("parser", "error", "empty input file")
				return fmt.Errorf("empty input file")
			}

			fid := filepath.Base(path)
			turn, err := parser.ParseInput(
				fid,
				"",
				data,
				acceptLoneDash,
				debugParser,
				debugSections,
				debugSteps,
				debugNodes,
				debugFleetMovement,
				false, // experimentalUnitSplit
				false, // experimentalScoutStill
				parser.ParseConfig{},
			)
			if err != nil {
				logger.Error("parser", "file", fid, "error", err)
				return err
			}

			doc := schema.Document{
				Schema:  schema.Version,
				Source:  "parser@" + version.Core(),
				Created: schema.Timestamp(time.Now().UTC().Format(schema.TimestampFormat)),
				Game:    schema.GameID(game),
				Turn:    schema.TurnID(fmt.Sprintf("%04d-%02d", turn.Year, turn.Month)),
				Clan:    schema.ClanID(clanID),
			}
			if len(turn.SpecialNames) > 0 {
				specialIDs := make([]string, 0, len(turn.SpecialNames))
				for id := range turn.SpecialNames {
					specialIDs = append(specialIDs, id)
				}
				sort.Strings(specialIDs)
				for _, id := range specialIDs {
					sp := turn.SpecialNames[id]
					name := sp.Name
					if name == "" {
						name = sp.Id
					}
					doc.SpecialHexes = append(doc.SpecialHexes, schema.SpecialHex{
						Name: name,
					})
				}
			}

			clans := make(map[schema.ClanID]*schema.Clan)
			for _, moves := range turn.UnitMoves {
				unit := schema.Unit{
					ID:             schema.UnitID(moves.UnitId),
					EndingLocation: schema.Coordinates(moves.ToHex),
				}

				sm := schema.Moves{
					ID:      unit.ID,
					Follows: schema.UnitID(moves.Follows),
					GoesTo:  schema.Coordinates(moves.GoesTo),
				}
				for _, m := range moves.Moves {
					sm.Steps = append(sm.Steps, convertMoveStep(m))
				}
				if len(sm.Steps) > 0 || sm.Follows != "" || sm.GoesTo != "" {
					unit.Moves = append(unit.Moves, sm)
				}

				for _, scout := range moves.Scouts {
					run := convertScoutRun(moves.UnitId, scout)
					if len(run.Steps) > 0 {
						unit.Scouts = append(unit.Scouts, run)
					}
				}

				cid := schema.ClanID("0" + string(moves.UnitId)[1:4])
				clan, ok := clans[cid]
				if !ok {
					clan = &schema.Clan{ID: cid}
					clans[cid] = clan
				}
				clan.Units = append(clan.Units, unit)
			}

			clanIDs := make([]schema.ClanID, 0, len(clans))
			for id := range clans {
				clanIDs = append(clanIDs, id)
			}
			sort.Slice(clanIDs, func(i, j int) bool { return clanIDs[i] < clanIDs[j] })
			for _, id := range clanIDs {
				clan := clans[id]
				sort.Slice(clan.Units, func(i, j int) bool { return clan.Units[i].ID < clan.Units[j].ID })
				doc.Clans = append(doc.Clans, *clan)
			}

			DeriveLocations(&doc)

			outputData, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				logger.Error("parser", "output", err)
				return err
			}

			if outputPath == "" {
				fmt.Println(string(outputData))
			} else {
				err = os.WriteFile(outputPath, outputData, 0o644)
				if err != nil {
					logger.Error("parser", "error", err)
					return err
				}
				logger.Info("parser", "created", outputPath)
			}

			logger.Info("parser", "elapsed time", time.Since(started).String())
			return nil
		},
	}
	if err := addFlags(cmdRoot); err != nil {
		logger.Error("parser", "error", err)
		os.Exit(1)
	}
	cmdRoot.AddCommand(cmdVersion())

	if err := cmdRoot.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func cmdVersion() *cobra.Command {
	showBuildInfo := false
	addFlags := func(cmd *cobra.Command) error {
		cmd.Flags().BoolVar(&showBuildInfo, "build-info", showBuildInfo, "show build information")
		return nil
	}
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
	if err := addFlags(cmd); err != nil {
		logger.Error("version", "error", err)
		os.Exit(1)
	}
	return cmd
}

func convertMoveStep(m *domain.Move_t) schema.MoveStep {
	step := schema.MoveStep{
		Result: convertResult(m.Result),
	}
	switch {
	case m.Advance != direction.Unknown:
		step.Intent = schema.IntentAdvance
		step.Advance = convertDirection(m.Advance)
	case m.Follows != "":
		step.Intent = schema.IntentFollows
		step.Follows = schema.UnitID(m.Follows)
	case m.GoesTo != "":
		step.Intent = schema.IntentGoesTo
		step.GoesTo = m.GoesTo
	case m.Still:
		step.Intent = schema.IntentStill
		step.Still = true
	}
	if m.Report != nil {
		step.Observation = convertObservation(m.Report)
	}
	return step
}

func convertScoutRun(unitId domain.UnitId_t, s *domain.Scout_t) schema.ScoutRun {
	run := schema.ScoutRun{
		ID: schema.ScoutID(fmt.Sprintf("%ss%d", unitId, s.No)),
	}
	for _, m := range s.Moves {
		run.Steps = append(run.Steps, convertMoveStep(m))
	}
	return run
}

func convertObservation(r *domain.Report_t) *schema.Observation {
	obs := &schema.Observation{
		Terrain: schema.Terrain(terrain.EnumToString[r.Terrain]),
	}
	for _, b := range r.Borders {
		obs.Edges = append(obs.Edges, schema.Edge{
			Dir:             convertDirection(b.Direction),
			Feature:         schema.Feature(edges.EnumToString[b.Edge]),
			NeighborTerrain: schema.Terrain(terrain.EnumToString[b.Terrain]),
		})
	}
	for _, e := range r.Encounters {
		obs.Encounters = append(obs.Encounters, schema.Encounter{
			Unit: schema.UnitID(e.UnitId),
		})
	}
	for _, s := range r.Settlements {
		if s != nil && s.Name != "" {
			obs.Settlements = append(obs.Settlements, schema.Settlement{
				Name: s.Name,
			})
		}
	}
	for _, rs := range r.Resources {
		if rs != resources.None {
			obs.Resources = append(obs.Resources, schema.Resource(resources.EnumToString[rs]))
		}
	}
	for _, fh := range r.FarHorizons {
		obs.CompassPoints = append(obs.CompassPoints, schema.CompassPoint{
			Bearing:         convertBearing(fh.Point),
			NeighborTerrain: schema.Terrain(terrain.EnumToString[fh.Terrain]),
		})
	}
	return obs
}

func convertBearing(p compass.Point_e) schema.Bearing {
	return schema.Bearing(compass.EnumToString[p])
}

func convertDirection(d direction.Direction_e) schema.Direction {
	return schema.Direction(direction.EnumToString[d])
}

func convertResult(r results.Result_e) schema.MoveResult {
	switch r {
	case results.Succeeded, results.Teleported:
		return schema.ResultSucceeded
	case results.Failed, results.Blocked, results.Prohibited, results.ExhaustedMovementPoints:
		return schema.ResultFailed
	case results.Vanished:
		return schema.ResultVanished
	default:
		return schema.ResultUnknown
	}
}
