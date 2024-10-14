// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package main implements the ottomap application
package main

import (
	"errors"
	"github.com/mdhender/semver"
	"github.com/playbymail/ottomap/cerrs"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
)

var (
	version = semver.Version{Major: 0, Minor: 20, Patch: 0}
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	// todo: detect when a unit is created as before and after-move action

	// todo: can we use this office package if we fix the tab issue?
	//	if s, err := office.ToStr("input/899-12.0138.Turn-Report.docx"); err != nil {
	//		log.Fatal(err)
	//	} else {
	//		log.Println(s)
	//	}

	if err := Execute(); err != nil {
		log.Fatal(err)
	}
}

func Execute() error {
	cmdRoot.PersistentFlags().BoolVar(&argsRoot.showVersion, "show-version", false, "show version")
	cmdRoot.PersistentFlags().StringVar(&argsRoot.logFile.name, "log-file", "", "set log file")

	cmdRoot.AddCommand(cmdDump)
	cmdDump.Flags().BoolVar(&argsDump.defaultTileMap, "default-tile-map", false, "dump the default tile map")

	cmdRoot.AddCommand(cmdRender)
	cmdRender.Flags().BoolVar(&argsRender.autoEOL, "auto-eol", true, "automatically convert line endings")
	cmdRender.Flags().BoolVar(&argsRender.debug.dumpAllTiles, "debug-dump-all-tiles", false, "dump all tiles")
	cmdRender.Flags().BoolVar(&argsRender.debug.dumpAllTurns, "debug-dump-all-turns", false, "dump all turns")
	cmdRender.Flags().BoolVar(&argsRender.debug.logFile, "debug-log-file", false, "enable file name in log output")
	cmdRender.Flags().BoolVar(&argsRender.debug.logTime, "debug-log-time", false, "enable time in log output")
	cmdRender.Flags().BoolVar(&argsRender.debug.maps, "debug-maps", false, "enable maps debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.nodes, "debug-nodes", false, "enable node debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.parser, "debug-parser", false, "enable parser debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.sections, "debug-sections", false, "enable sections debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.steps, "debug-steps", false, "enable step debugging")
	cmdRender.Flags().BoolVar(&argsRender.experimental.splitTrailingUnits, "x-split-units", false, "experimental: split trailing units")
	cmdRender.Flags().BoolVar(&argsRender.mapper.Dump.BorderCounts, "dump-border-counts", false, "dump border counts")
	cmdRender.Flags().BoolVar(&argsRender.render.FordsAsPills, "fords-as-pills", true, "render fords as pills")
	cmdRender.Flags().BoolVar(&argsRender.parser.Ignore.Scouts, "ignore-scouts", false, "ignore scout reports")
	cmdRender.Flags().BoolVar(&argsRender.noWarnOnInvalidGrid, "no-warn-on-invalid-grid", false, "disable grid id warnings")
	cmdRender.Flags().BoolVar(&argsRender.render.Show.Grid.Coords, "show-grid-coords", false, "show grid coordinates (XX CCRR)")
	cmdRender.Flags().BoolVar(&argsRender.render.Show.Grid.Numbers, "show-grid-numbers", false, "show grid numbers (CCRR)")
	cmdRender.Flags().BoolVar(&argsRender.saveWithTurnId, "save-with-turn-id", false, "add turn id to file name")
	cmdRender.Flags().BoolVar(&argsRender.show.origin, "show-origin", false, "show origin hex")
	cmdRender.Flags().BoolVar(&argsRender.show.shiftMap, "shift-map", true, "shift map up and left")
	cmdRender.Flags().BoolVar(&argsRender.experimental.stripCR, "strip-cr", false, "experimental: enable conversion of DOS EOL")
	cmdRender.Flags().BoolVar(&argsRender.experimental.newWaterTiles, "x-new-water-tiles", false, "experimental: use higher contrast water tiles")
	cmdRender.Flags().StringVar(&argsRender.clanId, "clan-id", "", "clan for output file names")
	if err := cmdRender.MarkFlagRequired("clan-id"); err != nil {
		log.Fatalf("error: clan-id: %v\n", err)
	}
	cmdRender.Flags().StringVar(&argsRender.paths.data, "data", "data", "path to root of data files")
	cmdRender.Flags().StringVar(&argsRender.originGrid, "origin-grid", "", "grid id to substitute for ##")
	cmdRender.Flags().StringVar(&argsRender.maxTurn.id, "max-turn", "", "last turn to map (yyyy-mm format)")

	cmdRoot.AddCommand(cmdVersion)

	return cmdRoot.Execute()
}

var argsRoot struct {
	logFile struct {
		name string
		fd   *os.File
	}
	showVersion bool
}

var cmdRoot = &cobra.Command{
	Use:   "ottomap",
	Short: "Root command for our application",
	Long:  `Create maps from TribeNet turn reports.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if argsRoot.logFile.name != "" {
			if fd, err := os.OpenFile(argsRoot.logFile.name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644); err != nil {
				return err
			} else {
				argsRoot.logFile.fd = fd
			}
			log.SetOutput(argsRoot.logFile.fd)
			argsRoot.showVersion = true
		}
		if argsRoot.showVersion {
			log.Printf("version: %s\n", version)
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if argsRoot.logFile.fd != nil {
			if err := log.Output(2, "log file closed"); err != nil {
				return err
			} else if err = argsRoot.logFile.fd.Close(); err != nil {
				return err
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("Hello from root command\n")
	},
}

func abspath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	} else if sb, err := os.Stat(absPath); err != nil {
		return "", err
	} else if !sb.IsDir() {
		return "", cerrs.ErrNotDirectory
	}
	return absPath, nil
}

func isdir(path string) (bool, error) {
	sb, err := os.Stat(path)
	if err != nil {
		return false, err
	} else if !sb.IsDir() {
		return false, nil
	}
	return true, nil
}

func isfile(path string) (bool, error) {
	sb, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	} else if sb.IsDir() || !sb.Mode().IsRegular() {
		return false, nil
	}
	return true, nil
}
