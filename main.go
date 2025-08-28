// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package main implements the ottomap application
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/maloquacious/semver"
	"github.com/playbymail/ottomap/cerrs"
	"github.com/playbymail/ottomap/internal/config"
	"github.com/spf13/cobra"
)

var (
	version = semver.Version{
		Major: 0,
		Minor: 62,
		Patch: 16,
		Build: semver.Commit(),
	}
	globalConfig *config.Config
)

func main() {
	// if version is on the command line, show it and exit
	for _, arg := range os.Args {
		if arg == "-version" || arg == "--version" {
			fmt.Printf("%s\n", version.Short())
			return
		} else if arg == "-build-info" || arg == "--build-info" {
			fmt.Printf("%s\n", version.String())
			return
		}
	}
	log.SetFlags(log.Lshortfile | log.Ltime)

	const configFileName = "data/input/ottomap.json"
	// set the debug flag only if there is a configuration file to debug
	debugConfigFile := false
	if sb, err := os.Stat(configFileName); err == nil && sb.Mode().IsRegular() {
		debugConfigFile = true
	}
	cfg, err := config.Load(configFileName, debugConfigFile)
	if err != nil && debugConfigFile {
		log.Printf("[config] %q: %v\n", configFileName, err)
	}

	if err := Execute(cfg); err != nil {
		log.Fatal(err)
	}
}

func Execute(cfg *config.Config) error {
	cmdRoot.PersistentFlags().BoolVar(&argsRoot.showVersion, "show-version", false, "show version")
	cmdRoot.PersistentFlags().StringVar(&argsRoot.logFile.name, "log-file", "", "set log file")

	cmdRoot.AddCommand(cmdDb)
	cmdDb.PersistentFlags().StringVar(&argsDb.paths.store, "store", argsDb.paths.store, "path to the database file")

	cmdDb.AddCommand(cmdDbCreate)
	cmdDbCreate.AddCommand(cmdDbCreateDatabase)
	cmdDbCreateDatabase.Flags().BoolVar(&argsDb.create.force, "force", false, "force the creation if the database exists")
	cmdDbCreateDatabase.Flags().StringVar(&argsDb.paths.store, "store", argsDb.paths.store, "path to the database file")
	if err := cmdDbCreateDatabase.MarkFlagRequired("store"); err != nil {
		log.Fatalf("store: %v\n", err)
	}

	cmdDb.AddCommand(cmdDbLoad)
	cmdDbLoad.AddCommand(cmdDbLoadFiles)
	cmdDbLoadFiles.Flags().StringVar(&argsDb.load.clan, "clan", argsDb.load.clan, "clan that owns reports")
	if err := cmdDbLoadFiles.MarkFlagRequired("clan"); err != nil {
		log.Fatalf("clan: %v\n", err)
	}
	cmdDbLoadFiles.Flags().StringVar(&argsDb.paths.store, "store", argsDb.paths.store, "path to the database file")
	if err := cmdDbLoadFiles.MarkFlagRequired("store"); err != nil {
		log.Fatalf("store: %v\n", err)
	}
	cmdDbLoadFiles.Flags().StringVar(&argsDb.load.path, "report-path", argsDb.load.path, "path to report files")

	cmdDbLoad.AddCommand(cmdDbLoadPath)
	cmdDbLoadPath.Flags().StringVar(&argsDb.load.clan, "clan", argsDb.load.clan, "clan that owns reports")
	if err := cmdDbLoadPath.MarkFlagRequired("clan"); err != nil {
		log.Fatalf("clan: %v\n", err)
	}
	cmdDbLoadPath.Flags().StringVar(&argsDb.paths.store, "store", argsDb.paths.store, "path to the database file")
	if err := cmdDbLoadPath.MarkFlagRequired("store"); err != nil {
		log.Fatalf("store: %v\n", err)
	}
	cmdDbLoadPath.Flags().StringVar(&argsDb.load.path, "report-path", argsDb.load.path, "path to report files")

	cmdRoot.AddCommand(cmdDump)

	cmdRoot.AddCommand(cmdList)
	cmdList.AddCommand(cmdListClans)
	cmdList.AddCommand(cmdListTurns)

	cmdRoot.AddCommand(cmdParse)
	cmdParse.AddCommand(cmdParseFile)
	//cmdParseFile.Flags().StringVar(&argsParseFiles.clanId, "clan-id", "", "clan id")
	//if err := cmdParseFile.MarkFlagRequired("clan-id"); err != nil {
	//	log.Fatalf("error: clan-id: %v\n", err)
	//}
	//cmdParseFile.Flags().StringVar(&argsParseFiles.turnId, "turn-id", "", "turn id")
	//if err := cmdParseFile.MarkFlagRequired("turn-id"); err != nil {
	//	log.Fatalf("error: turn-id: %v\n", err)
	//}

	cmdRoot.AddCommand(cmdRender)
	cmdRender.Flags().BoolVar(&argsRender.autoEOL, "auto-eol", true, "automatically convert line endings")
	cmdRender.Flags().BoolVar(&argsRender.warnOnInvalidGrid, "warn-on-invalid-grid", true, "warn on invalid grid id")
	cmdRender.Flags().BoolVar(&argsRender.warnOnNewSettlement, "warn-on-new-settlement", true, "warn on new settlement")
	cmdRender.Flags().BoolVar(&argsRender.warnOnTerrainChange, "warn-on-terrain-change", true, "warn when terrain changes")
	cmdRender.Flags().BoolVar(&argsRender.render.Show.Grid.Coords, "show-grid-coords", false, "show grid coordinates (XX CCRR)")
	cmdRender.Flags().BoolVar(&argsRender.render.Show.Grid.Numbers, "show-grid-numbers", false, "show grid numbers (CCRR)")
	cmdRender.Flags().BoolVar(&argsRender.saveWithTurnId, "save-with-turn-id", false, "add turn id to file name")
	cmdRender.Flags().BoolVar(&argsRoot.soloClan, "solo", false, "limit parsing to a single clan")
	cmdRender.Flags().BoolVar(&argsRender.show.origin, "show-origin", false, "show origin hex")
	cmdRender.Flags().BoolVar(&argsRender.show.shiftMap, "shift-map", true, "shift map up and left")
	cmdRender.Flags().StringVar(&argsRender.clanId, "clan-id", "", "clan for output file names")
	if err := cmdRender.MarkFlagRequired("clan-id"); err != nil {
		log.Fatalf("error: clan-id: %v\n", err)
	}
	cmdRender.Flags().StringVar(&argsRender.paths.data, "data", "data", "path to root of data files")
	cmdRender.Flags().StringVar(&argsRender.maxTurn.id, "max-turn", "", "last turn to map (yyyy-mm format)")
	cmdRender.Flags().StringVar(&argsRender.originGrid, "origin-grid", "", "grid id to substitute for ##")
	// todo: remove support for the solo-element flag. can't do it now because it breaks one player's map.
	cmdRender.Flags().StringVar(&argsRender.soloElement, "solo-element", "", "limit parsing to a single element of a clan")
	cmdRender.Flags().BoolVar(&argsRender.experimental.blankMapSmall, "blank-map", false, "experimental: create a blank map, AA..ZP")
	cmdRender.Flags().BoolVar(&argsRender.experimental.blankMapFull, "blank-map-full", false, "experimental: create a blank map, AA..ZZ")

	cmdRoot.AddCommand(cmdScrub)
	cmdScrub.AddCommand(cmdScrubFile)
	cmdScrub.AddCommand(cmdScrubFiles)

	cmdRoot.AddCommand(cmdVersion)

	if cfg == nil || !cfg.AllowConfig {
		globalConfig = config.Default()
	} else {
		globalConfig = cfg
		//if data, err := json.MarshalIndent(globalConfig, "", "  "); err == nil {
		//	log.Printf("[global] %s\n", data)
		//}
	}

	return cmdRoot.Execute()
}

var argsRoot struct {
	logFile struct {
		name string
		fd   *os.File
	}
	showVersion bool
	soloClan    bool // when set, only clans with this id are processed
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
