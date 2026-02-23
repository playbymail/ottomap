// Copyright (c) 2026 Michael D Henderson. All rights reserved.

// Package main implements the extract CLI. This program extracts the text
// from a turn report.
package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/maloquacious/semver"
	"github.com/playbymail/ottomap/internal/tndocx"
	"github.com/spf13/cobra"
)

var (
	version = semver.Version{
		Major: 0,
		Minor: 1,
		Patch: 0,
		Build: semver.Commit(),
	}
	logger *slog.Logger
)

func main() {
	var path, outputPath string
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
		cmd.Flags().StringVar(&outputPath, "output", "", "write results to file instead of stdout")
		return nil
	}

	cmdRoot := &cobra.Command{
		Use:           "extract",
		Short:         "tribenet report extract",
		Long:          `Extract TribeNet turn reports.`,
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
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			path, err = filepath.Abs(path)
			if err != nil {
				logger.Error("parser: invalid path", "error", err)
				return err
			}

			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".docx" {
				fmt.Printf("%s: is not a Word document: nothing to do\n", path)
				return nil
			}

			data, err := tndocx.ParsePath(path, true, true)
			if err != nil {
				logger.Error("extract", "error", err)
				return err
			}
			if len(data) == 0 {
				fmt.Printf("%s: is empty\n", path)
				return nil
			}

			if outputPath == "" {
				fmt.Println(string(data))
			} else {
				err = os.WriteFile(outputPath, data, 0o644)
				if err != nil {
					logger.Error("extract", "error", err)
					return err
				}
				fmt.Printf("%s: created\n", outputPath)
			}

			return nil
		},
	}
	if err := addFlags(cmdRoot); err != nil {
		logger.Error("extract", "error", err)
		os.Exit(1)
	}
	cmdRoot.AddCommand(cmdVersion())

	if err := cmdRoot.Execute(); err != nil {
		log.Fatal(err)
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
