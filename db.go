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
)

var argsDb struct {
	force bool // if true, overwrite existing database
	paths struct {
		database  string // path to the database file
		assets    string
		data      string
		templates string
	}
	randomSecret bool   // if true, generate a random secret for signing tokens
	secret       string // secret for signing tokens
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
		if err := store.CreateSchema(argsDb.paths.assets, argsDb.paths.templates); err != nil {
			log.Fatalf("db: init: %v\n", err)
		}

		log.Printf("db: created %q\n", argsDb.paths.database)
	},
}

var cmdDbUpdate = &cobra.Command{
	Use:   "update",
	Short: "Update database configuration",
	Run: func(cmd *cobra.Command, args []string) {
		// Implement your database update logic here
	},
}
