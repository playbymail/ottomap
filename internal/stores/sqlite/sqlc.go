// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlite

import (
	"context"
	"database/sql"
)

//go:generate sqlc generate

type Store struct {
	path string
	db   *sql.DB
	ctx  context.Context
	q    *Queries
}

func (s *Store) Close() error {
	var err error
	if s != nil && s.db != nil {
		err = s.db.Close()
		s.db = nil
	}
	return err
}

func (s *Store) Delete(stmt string) (sql.Result, error) {
	return s.db.ExecContext(s.ctx, stmt)
}

func (s *Store) Update(stmt string) (sql.Result, error) {
	return s.db.ExecContext(s.ctx, stmt)
}
