// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlite

import (
	"github.com/google/uuid"
	"github.com/playbymail/ottomap/domains"
	"github.com/playbymail/ottomap/stores/sqlite/sqlc"
	"log"
	"time"
)

func (db *DB) CreateSession(userId domains.ID, ttl time.Duration) (string, error) {
	err := db.q.DeleteUserSessions(db.ctx, int64(userId))
	if err != nil {
		return "", err
	}

	sessionId := uuid.NewString()
	err = db.q.CreateUserSession(db.ctx, sqlc.CreateUserSessionParams{
		SessID:    sessionId,
		UserID:    int64(userId),
		ExpiresAt: time.Now().Add(ttl).UTC(),
	})
	if err != nil {
		return "", err
	}

	return sessionId, nil
}

func (db *DB) DeleteUserSessions(userId domains.ID) error {
	return db.q.DeleteUserSessions(db.ctx, int64(userId))
}

func (db *DB) GetSession(id string) (*domains.User_t, error) {
	row, err := db.q.GetSession(db.ctx, id)
	if err != nil {
		return nil, err
	}
	log.Printf("sessions: clan %s: expires at %v\n", row.Clan, row.ExpiresAt)
	if !time.Now().Before(row.ExpiresAt) {
		return nil, domains.ErrSessionExpired
	}

	// convert row.Timezone to a time.Location
	loc, err := time.LoadLocation(row.Timezone)
	if err != nil {
		return nil, err
	}

	return &domains.User_t{
		ID:       domains.ID(row.UserID),
		Email:    row.Email,
		Timezone: loc,
		Clan:     row.Clan,
		Roles: struct {
			IsActive        bool
			IsAdministrator bool
			IsAuthenticated bool
			IsOperator      bool
			IsUser          bool
		}{
			IsActive:        row.IsActive == 1,
			IsAdministrator: row.IsAdministrator == 1,
			IsOperator:      row.IsOperator == 1,
			IsUser:          row.IsUser == 1,
		},
		Created:   row.CreatedAt,
		Updated:   row.UpdatedAt,
		LastLogin: row.LastLogin,
	}, nil
}
