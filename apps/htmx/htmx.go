// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

import (
	"fmt"
	"os"
)

type App struct {
	paths struct {
		assets    string
		data      string
		templates string
	}
}

func New(options ...Option) (*App, error) {
	a := &App{}

	for _, option := range options {
		if err := option(a); err != nil {
			return nil, err
		}
	}

	if a.paths.assets == "" {
		return nil, fmt.Errorf("missing assets path")
	} else if sb, err := os.Stat(a.paths.assets); err != nil {
		return nil, err
	} else if !sb.IsDir() {
		return nil, fmt.Errorf("%s: not a directory", a.paths.assets)
	} else if a.paths.data == "" {
		return nil, fmt.Errorf("missing data path")
	} else if sb, err := os.Stat(a.paths.data); err != nil {
		return nil, err
	} else if !sb.IsDir() {
		return nil, fmt.Errorf("%s: not a directory", a.paths.data)
	} else if a.paths.templates == "" {
		return nil, fmt.Errorf("missing templates path")
	} else if sb, err := os.Stat(a.paths.templates); err != nil {
		return nil, err
	} else if !sb.IsDir() {
		return nil, fmt.Errorf("%s: not a directory", a.paths.templates)
	}

	return a, nil
}
