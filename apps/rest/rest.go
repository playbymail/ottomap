// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package rest

import (
	"context"
	"fmt"
	"github.com/playbymail/ottomap/stores/sqlite"
)

type App struct {
	db struct {
		path  string
		ctx   context.Context
		store *sqlite.DB
	}
}

func New(options ...Option) (*App, error) {
	a := &App{}

	for _, option := range options {
		if err := option(a); err != nil {
			return nil, err
		}
	}

	if a.db.path == "" {
		return nil, fmt.Errorf("missing database path")
	} else if a.db.ctx == nil {
		return nil, fmt.Errorf("missing database context")
	} else if store, err := sqlite.OpenStore(a.db.path, a.db.ctx); err != nil {
		return nil, err
	} else {
		a.db.store = store
	}
	defer func() {
		_ = a.db.store.Close()
	}()

	return a, nil
}
