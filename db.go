// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/playbymail/ottomap/internal/stdlib"
	"github.com/playbymail/ottomap/internal/stores/sqlite"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var (
	argsDb struct {
		paths struct {
			store string // path to the database store
		}
		create struct {
			force bool // if true, overwrite existing database
		}
		load struct {
			clan  string   // clan that owns the reports
			path  string   // path to the directory containing the reports
			files []string // files to load
		}
	}

	cmdDb = &cobra.Command{
		Use:   "db",
		Short: "Database management commands",
	}

	cmdDbCreate = &cobra.Command{
		Use:   "create",
		Short: "Create new database or database objects",
	}

	cmdDbCreateDatabase = &cobra.Command{
		Use:   "database",
		Short: "create and initialize a new database",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if argsDb.paths.store == "" {
				return fmt.Errorf("database: path to store is required\n")
			} else if path, err := filepath.Abs(argsDb.paths.store); err != nil {
				return fmt.Errorf("database: %v\n", err)
			} else {
				argsDb.paths.store = path
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("db: create: %q\n", argsDb.paths.store)

			// it is an error if the database already exists unless force is true.
			// in that case, we remove the database so that we can create it again.
			if !argsDb.create.force {
				if ok, err := stdlib.IsFileExists(argsDb.paths.store); err != nil {
					log.Fatalf("db: %v\n", err)
				} else if ok {
					log.Printf("db: create: %s: removing\n", argsDb.paths.store)
					if err := os.Remove(argsDb.paths.store); err != nil {
						log.Fatalf("db: %v\n", err)
					}
				}
			}

			// create the database
			log.Printf("db: create: %s: creating database\n", argsDb.paths.store)
			err := sqlite.Create(argsDb.paths.store, context.Background())
			if err != nil {
				log.Fatalf("db: create: %v\n", err)
			}

			log.Printf("db: create: %s: created database\n", argsDb.paths.store)
		},
	}

	cmdDbLoad = &cobra.Command{
		Use:   "load",
		Short: "load report files",
	}

	cmdDbLoadFiles = &cobra.Command{
		Use:   "files",
		Short: "load specified files even if they haven't changed",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if argsDb.paths.store == "" {
				return fmt.Errorf("database: path to store is required\n")
			} else if ok, err := stdlib.IsFileExists(argsDb.paths.store); err != nil {
				return fmt.Errorf("database: %v\n", err)
			} else if !ok {
				return fmt.Errorf("database: %s: does not exist\n", argsDb.paths.store)
			}
			if argsDb.load.clan == "" {
				return fmt.Errorf("clan: is required")
			} else if n, err := strconv.Atoi(argsDb.load.clan); err != nil {
				return fmt.Errorf("%q: invalid clan", argsDb.load.clan)
			} else if !(0 < n && n < 1000) {
				return fmt.Errorf("%q: invalid clan", argsDb.load.clan)
			} else {
				argsDb.load.clan = fmt.Sprintf("%04d", n)
			}
			if argsDb.load.path == "" {
				argsDb.load.path = "."
			}
			if path, err := filepath.Abs(argsDb.load.path); err != nil {
				return fmt.Errorf("path: %v\n", err)
			} else {
				argsDb.load.path = path
			}
			if ok, err := stdlib.IsDirExists(argsDb.load.path); err != nil {
				return fmt.Errorf("path: %v\n", err)
			} else if !ok {
				return fmt.Errorf("path: %s: does not exist\n", argsDb.load.path)
			}
			// remaining args are files to load
			for _, arg := range args {
				if bp := filepath.Dir(arg); bp != "." {
					return fmt.Errorf("file: %s: must not contain path\n", arg)
				} else if ok, err := stdlib.IsFileExists(filepath.Join(argsDb.load.path, arg)); err != nil {
					return fmt.Errorf("file: %v\n", err)
				} else if !ok {
					return fmt.Errorf("file: %s: not found in report path\n", arg)
				} else {
					argsDb.load.files = append(argsDb.load.files, arg)
				}
			}
			if len(argsDb.load.files) == 0 {
				return fmt.Errorf("file: no files specified\n")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			clan, err := strconv.Atoi(argsDb.load.clan)
			if err != nil {
				log.Fatalf("db: %v\n", err)
			} else if !(0 < clan && clan < 1000) {
				log.Fatalf("db: %q: invalid clan\n", argsDb.load.clan)
			}
			store, err := sqlite.Open(argsDb.paths.store, context.Background())
			if err != nil {
				log.Fatalf("db: %v\n", err)
			}
			defer store.Close()
			log.Printf("db: %s: opened %p\n", argsDb.paths.store, store)
			log.Printf("db: load: clan: %q\n", argsDb.load.clan)
			log.Printf("db: load: report-path: %q\n", argsDb.load.path)
			for _, file := range argsDb.load.files {
				log.Printf("db: load: file: %q\n", file)
			}
			// get a list of all report files in the path
			reports, err := stdlib.FindInputs(argsDb.load.path, argsDb.load.files...)
			if err != nil {
				log.Fatalf("%s: %v\n", argsDb.load.path, err)
			} else if len(reports) == 0 {
				log.Fatalf("%s: no files found\n", argsDb.load.path)
			}
			// this command forces a load; we'll implement that by deleting all reports
			// before we try loading them.
			for _, report := range reports {
				if err := removeInputFile(store, clan, report); err != nil {
					log.Fatalf("removing %q: %v\n", report.Name, err)
				}
			}
			// now we should be able to load the files.
			for _, report := range reports {
				id, err := loadInputFile(store, clan, report)
				if err != nil {
					log.Fatalf("loading %q: %v\n", report.Name, err)
				}
				log.Printf("db: load: %s: created %8d\n", report.Name, id)
			}
		},
	}

	cmdDbLoadPath = &cobra.Command{
		Use:   "path",
		Short: "load new files in path",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if argsDb.paths.store == "" {
				return fmt.Errorf("database: path to store is required\n")
			} else if ok, err := stdlib.IsFileExists(argsDb.paths.store); err != nil {
				return fmt.Errorf("database: %v\n", err)
			} else if !ok {
				return fmt.Errorf("database: %s: does not exist\n", argsDb.paths.store)
			}
			if argsDb.load.clan == "" {
				return fmt.Errorf("clan: is required")
			} else if n, err := strconv.Atoi(argsDb.load.clan); err != nil {
				return fmt.Errorf("%q: invalid clan", argsDb.load.clan)
			} else if !(0 < n && n < 1000) {
				return fmt.Errorf("%q: invalid clan", argsDb.load.clan)
			} else {
				argsDb.load.clan = fmt.Sprintf("%04d", n)
			}
			if argsDb.load.path == "" {
				argsDb.load.path = "."
			}
			if path, err := filepath.Abs(argsDb.load.path); err != nil {
				return fmt.Errorf("path: %v\n", err)
			} else {
				argsDb.load.path = path
			}
			if ok, err := stdlib.IsDirExists(argsDb.load.path); err != nil {
				return fmt.Errorf("path: %v\n", err)
			} else if !ok {
				return fmt.Errorf("path: %s: does not exist\n", argsDb.load.path)
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			clan, err := strconv.Atoi(argsDb.load.clan)
			if err != nil {
				log.Fatalf("db: %v\n", err)
			} else if !(0 < clan && clan < 1000) {
				log.Fatalf("db: %q: invalid clan\n", argsDb.load.clan)
			}
			store, err := sqlite.Open(argsDb.paths.store, context.Background())
			if err != nil {
				log.Fatalf("db: %v\n", err)
			}
			defer store.Close()
			log.Printf("db: %s: opened %p\n", argsDb.paths.store, store)
			log.Printf("db: load: clan: %q\n", argsDb.load.clan)
			log.Printf("db: load: report-path: %q\n", argsDb.load.path)
			// get a list of all report files in the path
			reports, err := stdlib.FindAllInputs(argsDb.load.path)
			if err != nil {
				log.Fatalf("%s: %v\n", argsDb.load.path, err)
			} else if len(reports) == 0 {
				log.Fatalf("%s: no files found\n", argsDb.load.path)
			}
			// now we should be able to load the files.
			for _, report := range reports {
				// try to load the file. if the error is a duplicate hash or report name, we can ignore it.
				// otherwise, we should report it and continue to the next file.
				id, err := loadInputFile(store, clan, report)
				if errors.Is(err, sqlite.ErrDuplicateHash) {
					log.Printf("%04d: %s: skipped: duplicate hash\n", clan, report.Name)
					continue
				} else if errors.Is(err, sqlite.ErrDuplicateReportName) {
					log.Printf("%04d: %s: skipped: duplicate name\n", clan, report.Name)
					continue
				} else if err != nil {
					log.Printf("%04d: %s: error: %v\n", clan, report.Name, err)
					continue
				}
				log.Printf("%04d: %s: created %8d\n", clan, report.Name, id)
			}
		},
	}
)

// loadInputFile loads a report file into the database.
// It reports any errors that occur during the load.
// We assume that the caller has already handled duplicates before calling this function.
func loadInputFile(store *sqlite.Store, clan int, report *stdlib.File_t) (int, error) {
	if !(0 < clan && clan < 1000) {
		return 0, fmt.Errorf("%d: invalid clan", clan)
	}

	// fetch the file's contents
	data, err := os.ReadFile(filepath.Join(report.Path, report.Name))
	if err != nil {
		return 0, errors.Join(fmt.Errorf("reading %q", report.Name), err)
	}
	//log.Printf("%04d: %q: %d bytes (%q)\n", clan, report.Name, len(data), report.Kind)

	// scrub the file
	data, err = scrubData(report.Name, report.Kind, data)
	if err != nil {
		return 0, errors.Join(fmt.Errorf("scrubbing %q", report.Name), err)
	}
	//log.Printf("%04d: %q: %d bytes (%q)\n", clan, report.Name, len(data), report.Kind)

	// insert the file into the database
	id, err := store.CreateNewReport(clan, report.Year, report.Month, report.Unit, report.Hash, data)
	if err != nil {
		return 0, errors.Join(fmt.Errorf("inserting %q", report.Name), err)
	}
	//log.Printf("%04d: %q: %q: loaded %d\n", clan, report.Path, report.Name, id)

	return id, nil
}

// removeInputFile removes a report from the database. It uses both the
// hash and the name to identify the report. If the file is not in the
// database, it does nothing.
func removeInputFile(store *sqlite.Store, clan int, report *stdlib.File_t) error {
	if !(0 < clan && clan < 1000) {
		return fmt.Errorf("%d: invalid clan", clan)
	}
	log.Printf("%04d: remove: %04d %02d %-6s %q\n", clan, report.Year, report.Month, report.Unit, "..."+report.Hash[len(report.Hash)-8:])
	// if the hash is already in the database, remove the associated report.
	if err := store.DeleteReportByHash(clan, report.Hash); err != nil {
		return errors.Join(fmt.Errorf("deleteReportByHash"), err)
	}
	// if the name is already in the database, remove the associated report.
	if err := store.DeleteReportByName(clan, report.Year, report.Month, report.Unit); err != nil {
		return errors.Join(fmt.Errorf("deleteReportByName"), err)
	}
	return nil
}
