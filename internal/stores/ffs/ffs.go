// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package ffs

//go:generate sqlc generate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/playbymail/ottomap/internal/stores/ffs/sqlc"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var (
	//go:embed sqlc/schema.sql
	schema string
)

type Store struct {
	path    string        // path to the store and data files
	file    string        // path to the store file
	mdb     *sql.DB       // the in-memory database
	queries *sqlc.Queries // the sqlc database query functions
	ctx     context.Context
}

func New(options ...Option) (*Store, error) {
	started := time.Now()

	s := &Store{
		ctx: context.Background(),
	}

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.path == "" {
		return nil, fmt.Errorf("path not set")
	} else if sb, err := os.Stat(s.path); err != nil {
		return nil, err
	} else if !sb.IsDir() {
		return nil, fmt.Errorf("%s: not a directory", s.path)
	}
	s.file = filepath.Join(s.path, "store.db")
	log.Printf("ffs: store: %s\n", s.file)
	_ = os.Remove(s.file)

	if mdb, err := sql.Open("sqlite", s.file); err != nil {
		return nil, err
	} else {
		s.mdb = mdb
	}
	// todo: uncomment this when the schema is fixed and we are saving the store to disk
	// defer func() {
	//	if s.mdb != nil {
	//		_ = s.mdb.Close()
	//	}
	// }()

	// create the schema
	if err := s.createSchema(); err != nil {
		return nil, errors.Join(ErrCreateSchema, err)
	}

	// confirm that the database has foreign keys enabled
	if rslt, err := s.mdb.Exec("PRAGMA" + " foreign_keys = ON"); err != nil {
		log.Printf("error: foreign keys are disabled\n")
		return nil, ErrForeignKeysDisabled
	} else if rslt == nil {
		log.Printf("error: foreign keys pragma failed\n")
		return nil, ErrPragmaReturnedNil
	}

	// compile the regular expressions that we'll use when processing the files
	rxMagicKey, err := regexp.Compile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if err != nil {
		return nil, err
	}
	rxTurnReports, err := regexp.Compile(`^([0-9]{4}-[0-9]{2})\.([0-9]{4})\.report\.txt`)
	if err != nil {
		return nil, err
	}
	rxTurnMap, err := regexp.Compile(`^([0-9]{4}-[0-9]{2})\.([0-9]{4})\.wxx`)
	if err != nil {
		return nil, err
	}
	rxCourierSection, err := regexp.Compile(`^Courier (\d{4}c\d), `)
	if err != nil {
		return nil, err
	}
	rxElementSection, err := regexp.Compile(`^Element (\d{4}e\d), `)
	if err != nil {
		return nil, err
	}
	rxFleetSection, err := regexp.Compile(`^Fleet (\d{4}f\d), `)
	if err != nil {
		return nil, err
	}
	rxGarrisonSection, err := regexp.Compile(`^Garrison (\d{4}g\d), `)
	if err != nil {
		return nil, err
	}
	rxTribeSection, err := regexp.Compile(`^Tribe (\d{4}), `)
	if err != nil {
		return nil, err
	}

	// create the sqlc interface to our database
	s.queries = sqlc.New(s.mdb)

	// find all paths in the root directory that contain clan data
	entries, err := os.ReadDir(s.path)
	if err != nil {
		log.Printf("ffs: readRoot: %v\n", err)
		return nil, err
	}
	for _, entry := range entries {
		// is the entry a directory and is it a valid magic key?
		if !entry.IsDir() || !rxMagicKey.MatchString(entry.Name()) {
			continue
		}
		log.Printf("ffs: %q: found key\n", entry.Name())
		// does the entry contain a clan file
		var clan struct {
			Id   string
			Clan string
		}
		keyPath := filepath.Join(s.path, entry.Name())
		if data, err := os.ReadFile(filepath.Join(keyPath, "clan.json")); err != nil {
			log.Printf("warn: %q: %v\n", entry.Name(), err)
			continue
		} else if err = json.Unmarshal(data, &clan); err != nil {
			log.Printf("warn: %q: %v\n", entry.Name(), err)
			continue
		} else if clan.Id != entry.Name() {
			log.Printf("warn: %q: clan.json: id mismatch\n", entry.Name())
			continue
		}

		// create a fake user for the clan with hashed password for authentication and session management
		hash := sha256.Sum256([]byte(entry.Name()))
		hashStr := hex.EncodeToString(hash[:])
		uid, err := s.queries.CreateUser(s.ctx, sqlc.CreateUserParams{
			Clan:           clan.Clan,
			Handle:         clan.Clan,
			HashedPassword: hashStr,
			MagicKey:       clan.Id,
			Path:           keyPath,
		})
		if err != nil {
			log.Printf("ffs: %q: %v\n", clan.Id, err)
			continue
		}
		log.Printf("ffs: user %d: key %q\n", uid, clan.Id)

		// load the clan details for this user
		details, err := os.ReadDir(keyPath)
		if err != nil {
			log.Printf("ffs: %q: %v\n", clan.Id, err)
			continue
		}
		for _, detail := range details {
			if match := rxTurnMap.FindStringSubmatch(detail.Name()); len(match) == 3 {
				log.Printf("ffs: %s: %q: turn map\n", clan.Clan, detail.Name())
				mapFile := filepath.Join(keyPath, detail.Name())
				mid, err := s.queries.CreateTurnMap(s.ctx, sqlc.CreateTurnMapParams{
					Clan: clan.Clan,
					Path: mapFile,
					Turn: match[1],
					Uid:  uid,
				})
				if err != nil {
					log.Printf("ffs: %s: %q: %v\n", clan.Clan, detail.Name(), err)
					continue
				}
				log.Printf("ffs: %s: %q: map -> %d\n", clan.Clan, detail.Name(), mid)
			} else if match := rxTurnReports.FindStringSubmatch(detail.Name()); len(match) == 3 {
				log.Printf("ffs: %s: %q: turn report\n", clan.Clan, detail.Name())
				reportFile := filepath.Join(keyPath, detail.Name())
				turn := match[1]
				rid, err := s.queries.CreateTurnReport(s.ctx, sqlc.CreateTurnReportParams{
					Clan: clan.Clan,
					Path: reportFile,
					Turn: turn,
					Uid:  uid,
				})
				if err != nil {
					log.Printf("ffs: %s: %q: %v\n", clan.Clan, detail.Name(), err)
					continue
				}

				data, err := os.ReadFile(reportFile)
				if err != nil {
					log.Printf("ffs: %s: %q: %v\n", clan.Clan, detail.Name(), err)
					continue
				}

				type unitDetails_t struct {
					Id          string
					CurrentHex  string
					PreviousHex string
					Line        int
				}
				var units []unitDetails_t
				for no, line := range bytes.Split(data, []byte("\n")) {
					if matches := rxCourierSection.FindStringSubmatch(string(line)); len(matches) == 2 {
						units = append(units, unitDetails_t{
							Id:   matches[1],
							Line: no + 1,
						})
					} else if matches = rxElementSection.FindStringSubmatch(string(line)); len(matches) == 2 {
						units = append(units, unitDetails_t{
							Id:   matches[1],
							Line: no + 1,
						})
					} else if matches = rxFleetSection.FindStringSubmatch(string(line)); len(matches) == 2 {
						units = append(units, unitDetails_t{
							Id:   matches[1],
							Line: no + 1,
						})
					} else if matches = rxGarrisonSection.FindStringSubmatch(string(line)); len(matches) == 2 {
						units = append(units, unitDetails_t{
							Id:   matches[1],
							Line: no + 1,
						})
					} else if matches = rxTribeSection.FindStringSubmatch(string(line)); len(matches) == 2 {
						units = append(units, unitDetails_t{
							Id:   matches[1],
							Line: no + 1,
						})
					}
				}
				for _, unit := range units {
					log.Printf("ffs: %s: %q: %4d: %d: %q: %q\n", clan.Clan, detail.Name(), unit.Line, rid, turn, unit.Id)
					err = s.queries.CreateUnit(s.ctx, sqlc.CreateUnitParams{
						Name: unit.Id,
						Rid:  rid,
						Turn: turn,
					})
					if err != nil {
						log.Printf("ffs: %s: %q: %v\n", clan.Clan, detail.Name(), err)
						continue
					}
				}
				log.Printf("ffs: %s: %q: turn report -> %d\n", clan.Clan, detail.Name(), rid)
			}
		}

		//sm.sessions[hashStr] = session_t{
		//	clan:    clan.Clan,
		//	id:      entry.Name(),
		//	key:     hashStr,
		//	expires: time.Now().Add(sm.ttl),
		//}
		//log.Printf("session: load %q -> %q\n", entry.Name(), hashStr)
	}

	// todo: remove this stuff in the testing session
	// 97084c3e-0a4e-462f-bf6b-b2fa15bc10a9|3|2024-08-04 07:19:59.625609 -0600 MDT m=+50446.881029751
	if err := s.queries.CreateSession(s.ctx, sqlc.CreateSessionParams{
		ID:          "97084c3e-0a4e-462f-bf6b-b2fa15bc10a9",
		Uid:         3,
		ExpiresDttm: time.Now().Add(2 * 7 * 24 * time.Hour),
	}); err != nil {
		return nil, err
	}

	log.Printf("ffs: store: created in %v\n", time.Since(started))

	return s, nil
}

func (s *Store) Close() error {
	if s.mdb != nil {
		return s.mdb.Close()
	}
	return nil
}
