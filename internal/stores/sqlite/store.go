// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"github.com/playbymail/ottomap/internal/stdlib"
	"log"
	_ "modernc.org/sqlite"
)

var (
	//go:embed schema.sql
	schemaDDL string
)

// Create creates a new store.
// Returns an error if the database file already exists.
// The caller must delete the database file if they want to start fresh.
func Create(path string, ctx context.Context) error {
	// if the stat fails because the file doesn't exist, we're okay.
	// if it fails for any other reason, it's an error.
	if ok, err := stdlib.IsFileExists(path); err != nil {
		log.Printf("db: create: %q: %s\n", path, err)
		return err
	} else if ok {
		// we're not forcing the creation of a new database so this is an error
		log.Printf("db: create: %q: %s\n", path, "database already exists")
		return ErrDatabaseExists
	}

	// create the database
	log.Printf("db: create: path %s\n", path)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Printf("db: create: %v\n", err)
		return err
	}
	defer db.Close()

	// confirm that the database has foreign keys enabled
	checkPragma := "PRAGMA" + " foreign_keys = ON"
	if rslt, err := db.Exec(checkPragma); err != nil {
		log.Printf("db: create: foreign keys are disabled\n")
		return ErrForeignKeysDisabled
	} else if rslt == nil {
		log.Printf("db: create: foreign keys pragma failed\n")
		return ErrPragmaReturnedNil
	}

	// create the schema
	if _, err := db.Exec(schemaDDL); err != nil {
		log.Printf("db: create: failed to initialize schema\n")
		log.Printf("db: create: %v\n", err)
		return errors.Join(ErrCreateSchema, err)
	}

	log.Printf("db: create: created %s\n", path)

	// initialize the database with default data
	log.Printf("db: create: inserted default data\n")

	return nil
}

// Open opens an existing store.
// Returns an error if the database file is not a valid file.
// Caller must call Close() when done.
func Open(path string, ctx context.Context) (*Store, error) {
	// it is an error if the database does not already exist,
	// or it is not a file.
	if ok, err := stdlib.IsFileExists(path); err != nil {
		log.Printf("db: open: %q: %v\n", path, err)
		return nil, err
	} else if !ok {
		log.Printf("db: open: %q: %s\n", path, "not a database")
		return nil, ErrInvalidPath
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Printf("db: open: %s: %v\n", path, err)
		return nil, err
	}

	// confirm that the database has foreign keys enabled
	checkPragma := "PRAGMA" + " foreign_keys = ON"
	if rslt, err := db.Exec(checkPragma); err != nil {
		_ = db.Close()
		log.Printf("db: open: foreign keys are disabled\n")
		return nil, ErrForeignKeysDisabled
	} else if rslt == nil {
		_ = db.Close()
		log.Printf("db: open: foreign keys pragma failed\n")
		return nil, ErrPragmaReturnedNil
	}

	// return the store.
	return &Store{path: path, db: db, ctx: ctx, q: New(db)}, nil
}
