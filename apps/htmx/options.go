// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

import (
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

type Options []Option
type Option func(*App) error

func WithAssets(path string) Option {
	return func(a *App) error {
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else if absPath, err := filepath.Abs(path); err != nil {
			return err
		} else if sb, err := os.Stat(absPath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else {
			a.paths.assets = absPath
		}
		return nil
	}
}

func WithData(path string) Option {
	return func(a *App) error {
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else if absPath, err := filepath.Abs(path); err != nil {
			return err
		} else if sb, err := os.Stat(absPath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else {
			a.paths.data = absPath
		}
		return nil
	}
}

func WithTemplates(path string) Option {
	return func(a *App) error {
		if sb, err := os.Stat(path); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else if absPath, err := filepath.Abs(path); err != nil {
			return err
		} else if sb, err := os.Stat(absPath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", path)
		} else {
			a.paths.templates = absPath
		}
		return nil
	}
}
