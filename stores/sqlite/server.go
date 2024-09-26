// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlite

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/playbymail/ottomap/cerrs"
	"os"
	"path/filepath"
)

// SetServerAssetsPaths sets the path for the assets folder for the server.
func (db *DB) SetServerAssetsPaths(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	} else if sb, err := os.Stat(absPath); err != nil {
		return err
	} else if !sb.IsDir() {
		return cerrs.ErrNotDirectory
	}
	return db.q.SetServerAssetsPath(db.ctx, absPath)
}

// SetServerTemplatesPaths sets the path for the templates folder for the server.
func (db *DB) SetServerTemplatesPaths(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	} else if sb, err := os.Stat(absPath); err != nil {
		return err
	} else if !sb.IsDir() {
		return cerrs.ErrNotDirectory
	}
	return db.q.SetServerAssetsPath(db.ctx, absPath)
}

// SetServerSalt sets the salt for the server.
func (db *DB) SetServerSalt(salt string) error {
	// create a hash of the salt
	hash := sha256.New()
	hash.Write([]byte(salt))
	salt = hex.EncodeToString(hash.Sum(nil))

	// and store it in the database
	return db.q.SetServerSalt(db.ctx, salt)
}
