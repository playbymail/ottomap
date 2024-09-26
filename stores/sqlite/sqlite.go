// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package sqlite implements the Sqlite database store.
package sqlite

//go:generate sqlc generate

import (
	"context"
	"database/sql"
	"errors"
	"github.com/playbymail/ottomap/domains"
	"github.com/playbymail/ottomap/stores/sqlite/sqlc"
	"log"
	"os"
	"path/filepath"
)

type DB struct {
	db  *sql.DB
	ctx context.Context
	q   *sqlc.Queries
}

// CreateStore creates a new store.
// Returns an error if the path is not a directory, or if the database already exists.
func CreateStore(path string, force bool, ctx context.Context) (*DB, error) {
	log.Printf("store: %q\n", path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("store: %q: %s\n", absPath, err)
		return nil, err
	} else if sb, err := os.Stat(absPath); err != nil {
		log.Printf("store: %q: %s\n", absPath, err)
		return nil, err
	} else if !sb.IsDir() {
		log.Printf("store: %q: %s\n", absPath, err)
		return nil, domains.ErrInvalidPath
	}

	path = filepath.Join(absPath, "htmxdata.db")

	// it is an error if the database already exists unless force is true.
	// in that case, we remove the database and create it again.
	if _, err := os.Stat(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			// it's an error if the stat fails for any reason other than the file not existing.
			return nil, err
		}
	} else {
		// database file exists
		if !force {
			// we're not forcing the creation of a new database so this is an error
			return nil, domains.ErrDatabaseExists
		}
		log.Printf("store: removing %s\n", path)
		if err := os.Remove(path); err != nil {
			return nil, err
		}
	}

	// create the database
	log.Printf("store: creating %s\n", path)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// return the store.
	return &DB{db: db, ctx: ctx, q: sqlc.New(db)}, nil
}

// OpenStore opens an existing store.
// Returns an error if the path is not a directory, or if the database does not exist.
func OpenStore(path string, ctx context.Context) (*DB, error) {
	log.Printf("store: %q\n", path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Printf("store: %q: %s\n", absPath, err)
		return nil, err
	} else if sb, err := os.Stat(absPath); err != nil {
		log.Printf("store: %q: %s\n", absPath, err)
		return nil, err
	} else if !sb.IsDir() {
		log.Printf("store: %q: %s\n", absPath, err)
		return nil, domains.ErrInvalidPath
	}

	path = filepath.Join(absPath, "htmxdata.db")

	// it is an error if the database does not already exist
	if _, err := os.Stat(path); err != nil {
		log.Printf("store: %q: %s\n", path, err)
		return nil, err
	}

	log.Printf("store: opening %s\n", path)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// return the store.
	return &DB{db: db, ctx: ctx, q: sqlc.New(db)}, nil
}

func (db *DB) Close() error {
	var err error
	if db != nil && db.db != nil {
		err = db.db.Close()
		db.db = nil
	}
	return err
}
