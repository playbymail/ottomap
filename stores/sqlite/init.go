// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlite

// initialization functions

import (
	_ "embed"
	"errors"
	"github.com/google/uuid"
	"github.com/playbymail/ottomap/domains"
	"log"
)

var (
	//go:embed sqlc/schema.sql
	schemaDDL string
)

func (db *DB) CreateSchema(adminPassword, assets, templates, salt string) error {
	// we have to assume that the database already exists

	// confirm that the database has foreign keys enabled
	checkPragma := "PRAGMA" + " foreign_keys = ON"
	if rslt, err := db.db.Exec(checkPragma); err != nil {
		log.Printf("[sqldb] error: foreign keys are disabled\n")
		return domains.ErrForeignKeysDisabled
	} else if rslt == nil {
		log.Printf("[sqldb] error: foreign keys pragma failed\n")
		return domains.ErrPragmaReturnedNil
	}

	// create the schema
	if _, err := db.db.Exec(schemaDDL); err != nil {
		log.Printf("[sqldb] failed to initialize schema\n")
		log.Printf("[sqldb] %v\n", err)
		return errors.Join(domains.ErrCreateSchema, err)
	}

	if adminPassword != "" {
		err := db.UpdateAdministrator(adminPassword, true)
		if err != nil {
			log.Printf("[sqldb] failed to update admin user\n")
			log.Printf("[sqldb] %v\n", err)
			return err
		}
		log.Printf("[sqldb] admin user updated\n")
	}

	if err := db.SetServerAssetsPaths(assets); err != nil {
		log.Printf("[sqldb] failed to set assets path\n")
		log.Printf("[sqldb] %v\n", err)
		return err
	}
	log.Printf("[sqldb] added assets %q\n", assets)

	if err := db.SetServerTemplatesPaths(templates); err != nil {
		log.Printf("[sqldb] failed to set templates path\n")
		log.Printf("[sqldb] %v\n", err)
		return err
	}
	log.Printf("[sqldb] added templates %q\n", templates)

	if salt == "" {
		salt = uuid.NewString()
	}
	if err := db.SetServerSalt(salt); err != nil {
		log.Printf("[sqldb] failed to set salt\n")
		log.Printf("[sqldb] %v\n", err)
		return err
	}
	log.Printf("[sqldb] added salt\n")

	return nil
}
