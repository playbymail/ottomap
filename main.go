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
	version = semver.Version{Major: 0, Minor: 16, Patch: 17}
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
	cmdRoot.AddCommand(cmdDb)
	cmdDb.PersistentFlags().StringVar(&argsDb.paths.database, "database", "", "path to the database file")

	cmdDb.AddCommand(cmdDbInit)
	cmdDbInit.Flags().BoolVarP(&argsDb.force, "force", "f", false, "force the creation even if the database exists")
	cmdDbInit.Flags().StringVar(&argsDb.secrets.admin, "admin-password", "", "optional password for the admin user")
	cmdDbInit.Flags().StringVarP(&argsDb.paths.assets, "assets", "a", "", "path to the assets directory")
	cmdDbInit.Flags().StringVarP(&argsDb.paths.data, "data", "d", "", "path to the data files directory")
	cmdDbInit.Flags().StringVarP(&argsDb.paths.templates, "templates", "t", "", "path to the templates directory")
	cmdDbInit.Flags().StringVarP(&argsDb.secrets.signing, "secret", "s", "", "new secret for signing tokens")

	cmdDb.AddCommand(cmdDbCreate)
	cmdDbCreate.AddCommand(cmdDbCreateUser)
	cmdDbCreateUser.Flags().StringVarP(&argsDb.data.user.clan, "clan-id", "c", "", "clan number for user")
	if err := cmdDbCreateUser.MarkFlagRequired("clan-id"); err != nil {
		log.Fatalf("error: clan-id: %v\n", err)
	}
	cmdDbCreateUser.Flags().StringVarP(&argsDb.data.user.email, "email", "e", "", "email for user")
	if err := cmdDbCreateUser.MarkFlagRequired("email"); err != nil {
		log.Fatalf("error: email: %v\n", err)
	}
	cmdDbCreateUser.Flags().StringVarP(&argsDb.data.user.secret, "secret", "s", "", "secret for user")
	if err := cmdDbCreateUser.MarkFlagRequired("secret"); err != nil {
		log.Fatalf("error: secret: %v\n", err)
	}
	cmdDbCreateUser.Flags().StringVarP(&argsDb.data.user.timezone, "timezone", "t", "UTC", "timezone for user")

	cmdDb.AddCommand(cmdDbDelete)
	cmdDbDelete.AddCommand(cmdDbDeleteUser)
	cmdDbDeleteUser.Flags().StringVarP(&argsDb.data.user.clan, "clan-id", "c", "", "clan number for user")
	if err := cmdDbCreateUser.MarkFlagRequired("clan-id"); err != nil {
		log.Fatalf("error: clan-id: %v\n", err)
	}

	cmdDb.AddCommand(cmdDbUpdate)
	cmdDbUpdate.Flags().BoolVar(&argsDb.secrets.useRandomSecret, "use-random-secret", false, "generate a new random secret for signing tokens")
	cmdDbUpdate.Flags().StringVar(&argsDb.secrets.admin, "admin-password", "", "update password for the admin user")
	cmdDbUpdate.Flags().StringVarP(&argsDb.paths.assets, "assets", "a", "", "new path to the assets directory")
	cmdDbUpdate.Flags().StringVarP(&argsDb.paths.data, "data", "d", "", "new path to the data files directory")
	cmdDbUpdate.Flags().StringVarP(&argsDb.paths.templates, "templates", "t", "", "new path to the templates directory")
	cmdDbUpdate.Flags().StringVarP(&argsDb.secrets.signing, "secret", "s", "", "new secret for signing tokens")

	cmdRoot.AddCommand(cmdDump)
	cmdDump.Flags().BoolVar(&argsDump.defaultTileMap, "default-tile-map", false, "dump the default tile map")

	cmdRoot.AddCommand(cmdRender)
	cmdRender.Flags().BoolVar(&argsRender.autoEOL, "auto-eol", false, "automatically convert line endings")
	cmdRender.Flags().BoolVar(&argsRender.debug.dumpAllTiles, "debug-dump-all-tiles", false, "dump all tiles")
	cmdRender.Flags().BoolVar(&argsRender.debug.dumpAllTurns, "debug-dump-all-turns", false, "dump all turns")
	cmdRender.Flags().BoolVar(&argsRender.debug.maps, "debug-maps", false, "enable maps debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.nodes, "debug-nodes", false, "enable node debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.parser, "debug-parser", false, "enable parser debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.sections, "debug-sections", false, "enable sections debugging")
	cmdRender.Flags().BoolVar(&argsRender.debug.steps, "debug-steps", false, "enable step debugging")
	cmdRender.Flags().BoolVar(&argsRender.experimental.stripCR, "debug-strip-cr", false, "experimental: enable conversion of DOS EOL")
	cmdRender.Flags().BoolVar(&argsRender.experimental.splitTrailingUnits, "x-split-units", false, "experimental: split trailing units")
	cmdRender.Flags().BoolVar(&argsRender.mapper.Dump.BorderCounts, "dump-border-counts", false, "dump border counts")
	cmdRender.Flags().BoolVar(&argsRender.parser.Ignore.Scouts, "ignore-scouts", false, "ignore scout reports")
	cmdRender.Flags().BoolVar(&argsRender.noWarnOnInvalidGrid, "no-warn-on-invalid-grid", false, "disable grid id warnings")
	cmdRender.Flags().BoolVar(&argsRender.render.Show.Grid.Coords, "show-grid-coords", false, "show grid coordinates (XX CCRR)")
	cmdRender.Flags().BoolVar(&argsRender.render.Show.Grid.Numbers, "show-grid-numbers", false, "show grid numbers (CCRR)")
	cmdRender.Flags().BoolVar(&argsRender.saveWithTurnId, "save-with-turn-id", false, "add turn id to file name")
	cmdRender.Flags().BoolVar(&argsRender.show.origin, "show-origin", false, "show origin hex")
	cmdRender.Flags().BoolVar(&argsRender.show.shiftMap, "shift-map", false, "shift map up and left")
	cmdRender.Flags().StringVar(&argsRender.clanId, "clan-id", "", "clan for output file names")
	if err := cmdRender.MarkFlagRequired("clan-id"); err != nil {
		log.Fatalf("error: clan-id: %v\n", err)
	}
	cmdRender.Flags().StringVar(&argsRender.paths.data, "data", "data", "path to root of data files")
	cmdRender.Flags().StringVar(&argsRender.originGrid, "origin-grid", "", "grid id to substitute for ##")
	cmdRender.Flags().StringVar(&argsRender.maxTurn.id, "max-turn", "", "last turn to map (yyyy-mm format)")

	cmdRoot.AddCommand(cmdServe)

	cmdServe.AddCommand(cmdServeHtmx)
	cmdServeHtmx.Flags().StringVar(&argsServeHtmx.paths.assets, "assets", "assets", "path to public assets")
	cmdServeHtmx.Flags().StringVar(&argsServeHtmx.paths.data, "data", "userdata", "path to root of user data files")
	cmdServeHtmx.Flags().StringVar(&argsServeHtmx.paths.templates, "templates", "templates", "path to template files")
	cmdServeHtmx.Flags().StringVar(&argsServeHtmx.server.host, "host", "localhost", "host to serve on")
	cmdServeHtmx.Flags().StringVar(&argsServeHtmx.server.port, "port", "29631", "port to bind to")

	cmdServe.AddCommand(cmdServeRest)
	cmdServeRest.Flags().StringVar(&argsServeRest.paths.database, "database", "ottomap.db", "path to database file")
	cmdServeRest.Flags().StringVar(&argsServeRest.server.host, "host", "localhost", "host to serve on")
	cmdServeRest.Flags().StringVar(&argsServeRest.server.port, "port", "29642", "port to bind to")

	cmdRoot.AddCommand(cmdUser)

	cmdUser.AddCommand(cmdUserCreate)
	cmdUserCreate.Flags().StringVarP(&argsUser.clan, "clan-id", "c", "", "clan for the new user (required)")
	if err := cmdUserCreate.MarkFlagRequired("clan-id"); err != nil {
		log.Fatalf("error: clan-id: %v\n", err)
	}
	cmdUserCreate.Flags().StringVarP(&argsUser.email, "email", "e", "", "email address (required)")
	if err := cmdUserCreate.MarkFlagRequired("email"); err != nil {
		log.Fatalf("error: email: %v\n", err)
	}
	cmdUserCreate.Flags().StringVarP(&argsUser.password, "password", "p", "", "password for the new user")
	cmdUserCreate.Flags().StringVarP(&argsUser.role, "role", "r", "user", "user role (default to 'user')")

	cmdUser.AddCommand(cmdUserDelete)
	cmdUserDelete.Flags().StringVarP(&argsUser.clan, "clan-id", "c", "", "clan of the user to delete (required)")
	if err := cmdUserDelete.MarkFlagRequired("clan-id"); err != nil {
		log.Fatalf("error: clan-id: %v\n", err)
	}
	cmdUserDelete.Flags().BoolVarP(&argsUser.force, "force", "f", false, "force deletion without confirmation")

	cmdUser.AddCommand(cmdUserList)
	cmdUserList.Flags().StringVarP(&argsUser.role, "role", "r", "", "filter users by role")
	cmdUserList.Flags().IntVarP(&argsUser.limit, "limit", "l", 0, "limit the number of results")

	cmdUser.AddCommand(cmdUserUpdate)
	cmdUserUpdate.Flags().StringP("clan-id", "c", "", "clan for the user (required)")
	if err := cmdUserUpdate.MarkFlagRequired("clan-id"); err != nil {
		log.Fatalf("error: clan-id: %v\n", err)
	}
	cmdUserUpdate.Flags().StringP("email", "e", "", "new email address")
	cmdUserUpdate.Flags().StringP("password", "p", "", "new password")
	cmdUserUpdate.Flags().StringP("role", "r", "", "new role for the user")

	cmdRoot.AddCommand(cmdVersion)

	return cmdRoot.Execute()
}

var cmdRoot = &cobra.Command{
	Use:   "ottomap",
	Short: "Root command for our application",
	Long:  `Create maps from TribeNet turn reports.`,
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
