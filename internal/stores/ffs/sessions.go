// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package ffs

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/google/uuid"
	"github.com/playbymail/ottomap/domains"
	"github.com/playbymail/ottomap/internal/stores/ffs/sqlc"
	"log"
	"net/http"
	"time"
)

func (s *Store) CreateSession(clan, magicKey string) (domains.Session_t, error) {
	var sess domains.Session_t

	// todo: should use bcrypt, not a simple hash
	hash := sha256.Sum256([]byte(magicKey))

	user, err := s.queries.AuthenticateUser(s.ctx, sqlc.AuthenticateUserParams{
		Clan:           clan,
		HashedPassword: hex.EncodeToString(hash[:]),
	})
	if err != nil {
		return sess, err
	}

	sid := uuid.NewString()
	err = s.queries.CreateSession(s.ctx, sqlc.CreateSessionParams{
		ID:          sid,
		Uid:         user.ID,
		ExpiresDttm: time.Now().Add(2 * 7 * 24 * time.Hour), // two weeks
	})
	if err != nil {
		return sess, err
	}

	sess.Id, sess.UserId, sess.Clan = sid, domains.ID(user.ID), user.Clan

	return sess, nil
}

func (s *Store) DeleteSession(sid string) error {
	return s.queries.DeleteSession(s.ctx, sid)
}

func (s *Store) GetSession(r *http.Request) Session_t {
	var sess Session_t

	//log.Printf("session: fromRequest: cookie %q\n", "ottomap")
	cookie, err := r.Cookie("ottomap")
	if err != nil {
		log.Printf("ffs: session: fromRequest: no cookie: %v\n", err)
		return sess
	}
	log.Printf("%s: %s: getSession: cookie %q\n", r.Method, r.URL.Path, cookie.Value)

	ss, err := s.queries.GetSession(s.ctx, cookie.Value)
	if err != nil {
		log.Printf("ffs: session: fromRequest: get session: %v\n", err)
		return sess
	}

	user, err := s.queries.GetUser(s.ctx, ss.Uid)
	if err != nil {
		log.Printf("ffs: session: fromRequest: get user: %v\n", err)
		return sess
	}

	sess.Id, sess.Uid, sess.Clan, sess.ExpiresAt = ss.ID, user.ID, user.Clan, ss.ExpiresDttm

	return sess
}

type Session_t struct {
	Id        string
	Uid       int64
	Clan      string
	ExpiresAt time.Time
}

func (sess Session_t) IsAuthenticated() bool {
	return sess.Uid > 0
}
