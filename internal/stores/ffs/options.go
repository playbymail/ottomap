// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package ffs

import (
	"fmt"
	"os"
	"path/filepath"
)

type Option func(*Store) error

func WithPath(path string) Option {
	return func(s *Store) error {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		} else if sb, err := os.Stat(absPath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", absPath)
		}
		s.path = absPath
		return nil
	}
}
