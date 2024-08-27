// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package rest

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/playbymail/ottomap/stores/sqlite"
	"log"
)

type App struct {
	db struct {
		path string
		ctx  context.Context
		db   *sqlite.DB
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
	} else if db, err := sql.Open("sqlite", a.db.path); err != nil {
		return nil, err
	} else {
		a.db.db = sqlite.NewStore(db, a.db.ctx)
	}

	// always create (or update) an operator with a random secret
	operatorSecret := uuid.NewString()
	if err := a.db.db.CreateOperator(operatorSecret); err != nil {
		return nil, err
	}
	log.Printf("store: operator: %q", "operator")
	log.Printf("store: secret::: %q", operatorSecret)

	return a, nil
}
