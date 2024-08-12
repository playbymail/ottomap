// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package coords

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Grid represents coordinates on the big map.
// They start at 1,1 and increase to the right and down.
type Grid struct {
	BigMapRow    int // range A .. Z
	BigMapColumn int // range A .. Z
	GridColumn   int // range 01 .. 30
	GridRow      int // range 01 .. 21
}

func (g Grid) String() string {
	if g.IsZero() {
		return "N/A"
	}
	return fmt.Sprintf("%c%c %02d%02d", 'A'+g.BigMapRow, 'A'+g.BigMapColumn, g.GridColumn, g.GridRow)
}

func (g Grid) IsZero() bool {
	return g.BigMapRow == 0 && g.BigMapColumn == 0 && g.GridColumn == 0 && g.GridRow == 0
}

func (g Grid) ToMapCoords() (Map, error) {
	return Map{
		Column: g.BigMapColumn*30 + g.GridColumn - 1,
		Row:    g.BigMapRow*21 + g.GridRow - 1,
	}, nil
}

// MarshalJSON implements the json.Marshaler interface.
func (g Grid) MarshalJSON() ([]byte, error) {
	return json.Marshal(g.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (g *Grid) UnmarshalJSON(data []byte) error {
	// expecting a string with quotes
	if bytes.Equal(data, []byte{'"', 'N', '/', 'A', '"'}) {
		*g = Grid{}
		return nil
	} else if len(data) != 9 || data[0] != '"' || data[8] != '"' {
		return fmt.Errorf("invalid grid cordinates")
	}
	tmp, err := StringToGridCoords(string(data[1:8]))
	if err != nil {
		return err
	}
	*g = tmp
	return nil
}
