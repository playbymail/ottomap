// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package ffs

import (
	"database/sql"
	"errors"
	"log"
)

func (s *Store) GetClans(id string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

type Clan_t struct {
	Id    string    // id of the player's clan
	Turns []*Turn_t // list of turns that the clan has uploaded reports for
}

type Turn_t struct {
	Id      string
	Reports []*Report_t // list of reports that the clan has uploaded for this turn
}

type Report_t struct {
	Id    int64    // id of the report from the database table
	Clan  string   // id of the clan that owns the report
	Units []Unit_t // list of units included in this report
	Path  string   // path to the report file
	Map   string   // path to the map, only set when there is a map file
}

type Unit_t struct {
	Id          string
	CurrentHex  string
	PreviousHex string
}

func (s *Store) GetClan(uid int64) (Clan_t, error) {
	var c Clan_t

	user, err := s.queries.GetUser(s.ctx, uid)
	if err != nil {
		return c, err
	}
	c.Id = user.Clan

	r, err := s.queries.GetUserReports(s.ctx, uid)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return c, err
	}
	log.Printf("ffs: getClan: reports %d\n", len(r))
	for _, rpt := range r {
		var t *Turn_t
		for _, tt := range c.Turns {
			if tt.Id == rpt.Turn {
				t = tt
				break
			}
		}
		if t == nil {
			t = &Turn_t{Id: rpt.Turn}
			c.Turns = append(c.Turns, t)
		}
		t.Reports = append(t.Reports, &Report_t{
			Id:   rpt.ID,
			Clan: rpt.Clan,
			Path: rpt.Path,
		})
	}

	return c, nil
}

func (s *Store) createSchema() error {
	if _, err := s.mdb.Exec(schema); err != nil {
		return err
	}
	return nil
}
