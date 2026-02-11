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
		Minor: 84,
		Patch: 0,
		Build: semver.Commit(),
	}
	globalConfig *config.Config
)

func main() {
	logWithDefaultFlags, logWithShortFileName, logWithTimestamp := false, true, false
	// if version is on the command line, show it and exit
	for _, arg := range os.Args {
		if arg == "-version" || arg == "--version" {
			fmt.Printf("%s\n", version.Short())
			return
		} else if arg == "-build-info" || arg == "--build-info" {
			fmt.Printf("%s\n", version.String())
			return
		} else if arg == "--log-with-default-flags" {
			logWithDefaultFlags = true
		} else if arg == "--log-with-short-file" {
			logWithShortFileName = true
		} else if arg == "--log-with-short-file=false" {
			logWithShortFileName = false
		} else if arg == "--log-with-timestamp" {
			logWithTimestamp = true
		} else if arg == "--log-with-timestamp=false" {
			logWithTimestamp = false
		}
	}

	logFlags := 0
	if logWithShortFileName {
		logFlags |= log.Lshortfile
	}
	if logWithTimestamp {
		logFlags |= log.Ltime
	}
	if logWithDefaultFlags || logFlags == 0 {
		logFlags = log.LstdFlags
	}
	log.SetFlags(logFlags)

	const quiet, verbose, debug = true, false, false

	const configFileName = "data/input/ottomap.json"
	cfg, err := config.Load(configFileName, quiet, verbose, debug)
	if err != nil {
		log.Fatalf("[config] %q: %v\n", configFileName, err)
	}

	if err := Execute(cfg); err != nil {
		log.Fatal(err)
	}
}

func Execute(cfg *config.Config) error {
	cmdRoot.PersistentFlags().BoolVar(&argsRoot.showVersion, "show-version", false, "show version")
	cmdRoot.PersistentFlags().StringVar(&argsRoot.logFile.name, "log-file", "", "set log file")
	cmdRoot.PersistentFlags().Bool("log-with-default-flags", false, "log with default flags")
	cmdRoot.PersistentFlags().Bool("log-with-shortfile", true, "log with short file name")
	cmdRoot.PersistentFlags().Bool("log-with-timestamp", false, "log with timestamp")

	cmdRoot.PersistentFlags().Bool("debug", false, "log debugging information")
	cmdRoot.PersistentFlags().Bool("quiet", false, "log less information")
	cmdRoot.PersistentFlags().Bool("verbose", false, "log more information")

	cmdRoot.AddCommand(cmdDump)

	cmdRoot.AddCommand(cmdList)
	cmdList.AddCommand(cmdListClans)
	cmdList.AddCommand(cmdListTurns)

	cmdRoot.AddCommand(cmdRender)
	cmdRender.Flags().BoolVar(&argsRender.autoEOL, "auto-eol", true, "automatically convert line endings")
	cmdRender.Flags().BoolVar(&argsRender.warnOnInvalidGrid, "warn-on-invalid-grid", true, "warn on invalid grid id")
	cmdRender.Flags().BoolVar(&argsRender.warnOnNewSettlement, "warn-on-new-settlement", true, "warn on new settlement")
	cmdRender.Flags().BoolVar(&argsRender.warnOnTerrainChange, "warn-on-terrain-change", true, "warn when terrain changes")
	cmdRender.Flags().BoolVar(&argsRender.render.Show.Grid.Coords, "show-grid-coords", false, "show grid coordinates (XX CCRR)")
	cmdRender.Flags().BoolVar(&argsRender.render.Show.Grid.Numbers, "show-grid-numbers", false, "show grid numbers (CCRR)")
	cmdRender.MarkFlagsMutuallyExclusive("show-grid-coords", "show-grid-numbers")
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
	cmdRender.Flags().StringVar(&argsRender.experimental.topLeft, "top-left", "", "experimental: top left corner of rendered map")
	cmdRender.Flags().StringVar(&argsRender.experimental.bottomRight, "bottom-right", "", "experimental: bottom right corner of rendered map")
	cmdRender.MarkFlagsRequiredTogether("top-left", "bottom-right")

	cmdRoot.AddCommand(cmdVersion)

	if cfg == nil || !cfg.AllowConfig {
		globalConfig = config.Default()
	} else {
		globalConfig = cfg
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
			log.Printf("ottomap: version %s\n", version)
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
