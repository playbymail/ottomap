// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package rest

import (
	"context"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

type Options []Option
type Option func(*App) error

// WithDatabase sets the path to the database file and the context for the connection.
func WithDatabase(path string, ctx context.Context) Option {
	return func(a *App) (err error) {
		if path == "" {
			return fmt.Errorf("missing database path")
		} else if ctx == nil {
			return fmt.Errorf("missing database context")
		}

		path, err = filepath.Abs(path)
		if err != nil {
			return err
		} else if sb, err := os.Stat(path); err != nil {
			return err
		} else if sb.IsDir() || !sb.Mode().IsRegular() {
			return fmt.Errorf("%s: not a file", path)
		}

		a.db.path = path
		a.db.ctx = ctx

		return nil
	}
}
