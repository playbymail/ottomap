// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"bytes"
	"github.com/playbymail/ottomap/internal/coords"
)

type WXX struct {
	buffer *bytes.Buffer

	tiles map[coords.Map]*Tile
}

func NewWXX() *WXX {
	return &WXX{
		tiles: map[coords.Map]*Tile{},
	}
}

// GetTile returns the tile at the given coordinates.
func (w *WXX) GetTile(location coords.Map) *Tile {
	t, ok := w.tiles[location]
	if !ok {
		panic("tile not defined")
	}
	return t
}
