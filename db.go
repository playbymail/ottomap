// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"database/sql"
	"github.com/playbymail/ottomap/stores/sqlite"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var argsDb struct {
	force bool // if true, overwrite existing database
	paths struct {
		database  string // path to the database file
		assets    string
		data      string
		templates string
	}
	secrets struct {
		useRandomSecret bool   // if true, generate a random secret for signing tokens
		admin           string // plain text password for admin user
		salt            string // salt for nothing (unused)
		signing         string // secret for signing tokens
	}
	data struct {
		user struct {
			clan     string // clan number
			email    string // email address for user
			secret   string // secret to use for user
			timezone string // timezone for user
		}
	}
}

var cmdDb = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
}

var cmdDbInit = &cobra.Command{
	Use:   "init",
	Short: "Initialize the database",
	PreRun: func(cmd *cobra.Command, args []string) {
		if argsDb.paths.database == "" {
			argsDb.paths.database = "."
		}
		if path, err := filepath.Abs(argsDb.paths.database); err != nil {
			log.Fatalf("database: %v\n", err)
		} else if ok, err := isdir(path); err != nil {
			log.Fatalf("database: %v\n", err)
		} else if !ok {
			log.Fatalf("database: %s: not a directory\n", path)
		} else {
			argsDb.paths.database = filepath.Join(path, "htmxdata.db")
		}

		if argsDb.paths.assets == "" {
			argsDb.paths.assets = "assets"
		}
		if path, err := abspath(argsDb.paths.assets); err != nil {
			log.Fatalf("assets: %v\n", err)
		} else if ok, err := isdir(path); err != nil {
			log.Fatalf("assets: %v\n", err)
		} else if !ok {
			log.Fatalf("assets: %s: not a directory\n", path)
		} else {
			argsDb.paths.assets = path
		}

		if argsDb.paths.data == "" {
			argsDb.paths.data = "data"
		}
		if path, err := abspath(argsDb.paths.data); err != nil {
			log.Fatalf("data: %v\n", err)
		} else if ok, err := isdir(path); err != nil {
			log.Fatalf("data: %v\n", err)
		} else if !ok {
			log.Fatalf("data: %s: not a directory\n", path)
		} else {
			argsDb.paths.data = path
		}

		if argsDb.paths.templates == "" {
			argsDb.paths.templates = "templates"
		}
		if path, err := abspath(argsDb.paths.templates); err != nil {
			log.Fatalf("templates: %v\n", err)
		} else if ok, err := isdir(path); err != nil {
			log.Fatalf("templates: %v\n", err)
		} else if !ok {
			log.Fatalf("templates: %s: not a directory\n", path)
		} else {
			argsDb.paths.templates = path
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("db: init: database  %s\n", argsDb.paths.database)
		log.Printf("db: init: assets    %s\n", argsDb.paths.assets)
		log.Printf("db: init: data      %s\n", argsDb.paths.data)
		log.Printf("db: init: templates %s\n", argsDb.paths.templates)
		if argsDb.secrets.admin != "" {
			log.Printf("db: init: admin password %q\n", argsDb.secrets.admin)
		}

		// if the database exists, we need to check if we are allowed to overwrite it
		if ok, err := isfile(argsDb.paths.database); err != nil {
			log.Fatalf("db: init: %v\n", err)
		} else if ok {
			if !argsDb.force {
				log.Fatalf("db: init: database %s: already exists\n", argsDb.paths.database)
			} else if err := os.Remove(argsDb.paths.database); err != nil {
				log.Fatalf("db: init: %v\n", err)
			}
			log.Printf("db: init: database %s: removed\n", argsDb.paths.database)
		}

		// create the database.
		log.Printf("db: init: creating database...\n")
		db, err := sql.Open("sqlite", argsDb.paths.database)
		if err != nil {
			log.Fatalf("db: init: %v\n", err)
		}
		defer func() {
			if db != nil {
				_ = db.Close()
			}
		}()

		store := sqlite.NewStore(db, context.Background())

		log.Printf("db: init: creating schema...\n")
		if err := store.CreateSchema(argsDb.secrets.admin, argsDb.paths.assets, argsDb.paths.templates, argsDb.secrets.salt); err != nil {
			log.Fatalf("db: init: %v\n", err)
		}

		log.Printf("db: created %q\n", argsDb.paths.database)
	},
}

var cmdDbCreate = &cobra.Command{
	Use:   "create",
	Short: "Create data-base objects",
}

var cmdDbCreateUser = &cobra.Command{
	Use:   "user",
	Short: "Create a new user",
	PreRun: func(cmd *cobra.Command, args []string) {
		if argsDb.paths.database == "" {
			argsDb.paths.database = "."
		}
		if path, err := filepath.Abs(argsDb.paths.database); err != nil {
			log.Fatalf("database: %v\n", err)
		} else if ok, err := isdir(path); err != nil {
			log.Fatalf("database: %v\n", err)
		} else if !ok {
			log.Fatalf("database: %s: not a directory\n", path)
		} else {
			argsDb.paths.database = filepath.Join(path, "htmxdata.db")
		}

		if len(argsDb.data.user.clan) != 4 {
			log.Fatalf("db: create user: clan must be 4 digits between 1 and 999\n")
		} else if n, err := strconv.Atoi(argsDb.data.user.clan); err != nil {
			log.Fatalf("db: create user: clan must be 4 digits between 1 and 999\n")
		} else if n < 1 || n > 999 {
			log.Fatalf("db: create user: clan must be 4 digits between 1 and 999\n")
		}
		if argsDb.data.user.email != strings.TrimSpace(argsDb.data.user.email) {
			log.Fatalf("db: create user: email must not contain leading or trailing spaces\n")
		}
		if len(argsDb.data.user.secret) < 4 {
			log.Fatalf("db: create user: secret must be at least 4 characters\n")
		}
		if argsDb.data.user.timezone == "" {
			argsDb.data.user.timezone = "UTC"
		} else if argsDb.data.user.timezone != strings.TrimSpace(argsDb.data.user.timezone) {
			log.Fatalf("db: create user: timezone must not contain leading or trailing spaces\n")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("db: create user: db     %s\n", argsDb.paths.database)
		// if the database exists, we need to check if we are allowed to overwrite it
		if ok, err := isfile(argsDb.paths.database); err != nil {
			log.Fatalf("db: create user: %v\n", err)
		} else if !ok {
			log.Fatalf("db: create user: database not found\n")
		}

		// open the database.
		log.Printf("db: create user: opening database...\n")
		db, err := sql.Open("sqlite", argsDb.paths.database)
		if err != nil {
			log.Fatalf("db: create user: %v\n", err)
		}
		defer func() {
			if db != nil {
				_ = db.Close()
			}
		}()

		store := sqlite.NewStore(db, context.Background())

		log.Printf("db: create user: clan   %q\n", argsDb.data.user.clan)
		log.Printf("db: create user: email  %q\n", argsDb.data.user.email)
		log.Printf("db: create user: secret %q\n", argsDb.data.user.secret)

		// validate the timezone
		log.Printf("db: create user: tz     %q\n", argsDb.data.user.timezone)
		loc, err := time.LoadLocation(argsDb.data.user.timezone)
		if err != nil {
			log.Fatalf("db: create user: timezone: %v\n", err)
		}

		user, err := store.CreateUser(argsDb.data.user.email, argsDb.data.user.secret, argsDb.data.user.clan, loc)
		if err != nil {
			log.Fatalf("db: create user: %v\n", err)
		}

		log.Printf("db: create user: user %d created\n", int(user.ID))
	},
}

var cmdDbDelete = &cobra.Command{
	Use:   "delete",
	Short: "Delete data-base objects",
}

var cmdDbDeleteUser = &cobra.Command{
	Use:   "user",
	Short: "Delete a user",
	PreRun: func(cmd *cobra.Command, args []string) {
		if argsDb.paths.database == "" {
			argsDb.paths.database = "."
		}
		if path, err := filepath.Abs(argsDb.paths.database); err != nil {
			log.Fatalf("database: %v\n", err)
		} else if ok, err := isdir(path); err != nil {
			log.Fatalf("database: %v\n", err)
		} else if !ok {
			log.Fatalf("database: %s: not a directory\n", path)
		} else {
			argsDb.paths.database = filepath.Join(path, "htmxdata.db")
		}

		if len(argsDb.data.user.clan) != 4 {
			log.Fatalf("clan: must be 4 digits between 1 and 999\n")
		} else if n, err := strconv.Atoi(argsDb.data.user.clan); err != nil {
			log.Fatalf("clan: must be 4 digits between 1 and 999\n")
		} else if n < 1 || n > 999 {
			log.Fatalf("clan: must be 4 digits between 1 and 999\n")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Printf("db: delete user: db     %s\n", argsDb.paths.database)
		// if the database exists, we need to check if we are allowed to overwrite it
		if ok, err := isfile(argsDb.paths.database); err != nil {
			log.Fatalf("db: delete user: %v\n", err)
		} else if !ok {
			log.Fatalf("db: delete user: database not found\n")
		}

		// open the database.
		log.Printf("db: delete user: opening database...\n")
		db, err := sql.Open("sqlite", argsDb.paths.database)
		if err != nil {
			log.Fatalf("db: delete user: %v\n", err)
		}
		defer func() {
			if db != nil {
				_ = db.Close()
			}
		}()

		store := sqlite.NewStore(db, context.Background())

		log.Printf("db: delete user: clan   %q\n", argsDb.data.user.clan)

		if err := store.DeleteUserByClan(argsDb.data.user.clan); err != nil {
			log.Fatalf("db: delete user: %v\n", err)
		}

		log.Printf("db: delete user: user deleted\n")
	},
}

var cmdDbUpdate = &cobra.Command{
	Use:   "update",
	Short: "Update database configuration",
	Run: func(cmd *cobra.Command, args []string) {
		// Implement your database update logic here
	},
}
