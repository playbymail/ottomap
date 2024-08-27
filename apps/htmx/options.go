// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	store "github.com/playbymail/ottomap/stores/sqlite"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

type Options []Option
type Option func(*App) error

func WithAssets(path string) Option {
	return func(a *App) error {
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else if absPath, err := filepath.Abs(path); err != nil {
			return err
		} else if sb, err := os.Stat(absPath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else {
			a.paths.assets = absPath
		}
		return nil
	}
}

func WithData(path string) Option {
	return func(a *App) error {
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else if absPath, err := filepath.Abs(path); err != nil {
			return err
		} else if sb, err := os.Stat(absPath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else {
			a.paths.data = absPath
		}
		return nil
	}
}

// WithDatabase creates a database connection.
// Has the side effect of creating/updating an operator with a new random secret.
func WithDatabase(path string, ctx context.Context) Option {
	return func(a *App) error {
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if sb.IsDir() || !sb.Mode().IsRegular() {
			return fmt.Errorf("%s: not a file", path)
		} else if db, err := sql.Open("sqlite", path); err != nil {
			return err
		} else {
			a.db.db = store.NewStore(db, ctx)
		}
		operatorSecret := uuid.NewString()
		if err := a.db.db.CreateOperator(operatorSecret); err != nil {
			return err
		}
		log.Printf("store: operator: %q", "operator")
		log.Printf("store: secret::: %q", operatorSecret)
		return nil
	}
}

func WithTemplates(path string) Option {
	return func(a *App) error {
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else if absPath, err := filepath.Abs(path); err != nil {
			return err
		} else if sb, err := os.Stat(absPath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else {
			a.paths.templates = absPath
		}
		return nil
	}
}
