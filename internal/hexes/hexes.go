// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package hexes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/playbymail/ottomap/internal/direction"
	"strconv"
	"strings"
)

// Hex_t is a hex coordinate on the big map.
// It includes the grid coordinates along with the row and column on that grid.
type Hex_t struct {
	GridColumn int
	GridRow    int
	Column     int
	Row        int
}

// IsZero returns true if the hex is zero
func (h Hex_t) IsZero() bool {
	return h == Hex_t{}
}

// String implements the Stringer interface.
func (h Hex_t) String() string {
	if h.IsZero() {
		return "N/A"
	}
	return fmt.Sprintf("%c%c %02d%02d", 'A'+h.GridRow-1, 'A'+h.GridColumn-1, h.Column, h.Row)
}

// MarshalJSON implements the json.Marshaler interface.
func (h Hex_t) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (h *Hex_t) UnmarshalJSON(data []byte) (err error) {
	// expecting a string with quotes
	if bytes.Equal(data, []byte{'"', 'N', '/', 'A', '"'}) {
		*h = Hex_t{}
		return nil
	} else if len(data) != 9 || data[0] != '"' || data[8] != '"' {
		return fmt.Errorf("invalid grid cordinates")
	}
	*h, err = GridStringToHex(string(data[1:8]))
	return err
}

// GridStringToHex converts a string of the form "AZ 0102" to a Hex_t.
//
// Grid id can range from AA to ZZ.
// Row can range from 1 to 21.
// Column can range from 1 to 30.
//
// It explicitly rejects "##", but blindly accepts "N/A".
func GridStringToHex(s string) (h Hex_t, err error) {
	if s == "N/A" {
		return Hex_t{}, nil
	} else if strings.HasPrefix(s, "##") {
		return Hex_t{}, fmt.Errorf("invalid grid coordinates")
	} else if !(len(s) == 7 && s[2] == ' ') {
		return Hex_t{}, fmt.Errorf("invalid grid coordinates")
	} else if h.GridRow = int(s[0] - 'A' + 1); !(0 < h.GridRow && h.GridRow <= 26) {
		return Hex_t{}, fmt.Errorf("invalid grid coordinates")
	} else if h.GridColumn = int(s[1] - 'A' + 1); !(0 < h.GridColumn && h.GridColumn <= 26) {
		return Hex_t{}, fmt.Errorf("invalid grid coordinates")
	} else if h.Column, err = strconv.Atoi(s[3:5]); err != nil {
		return Hex_t{}, fmt.Errorf("invalid grid coordinates")
	} else if !(0 < h.Column && h.Column <= 30) {
		return Hex_t{}, fmt.Errorf("invalid grid coordinates")
	} else if h.Row, err = strconv.Atoi(s[5:]); err != nil {
		return Hex_t{}, fmt.Errorf("invalid grid coordinates")
	} else if !(0 < h.Row && h.Row <= 21) {
		return Hex_t{}, fmt.Errorf("invalid grid coordinates")
	}
	return h, nil
}

// column direction vectors defines the vectors used to determine the coordinates
// of the neighboring column based on the direction and the odd/even column
// property of the starting hex.
//
// NB: grids and hexes start at 1, 1 so "odd" and "even" are based on the hex coordinates.

var OddColumnVectors = map[direction.Direction_e][2]int{
	direction.North:     [2]int{+0, -1}, // ## 1306 -> ## 1305
	direction.NorthEast: [2]int{+1, -1}, // ## 1306 -> ## 1405
	direction.SouthEast: [2]int{+1, +0}, // ## 1306 -> ## 1406
	direction.South:     [2]int{+0, +1}, // ## 1306 -> ## 1307
	direction.SouthWest: [2]int{-1, +0}, // ## 1306 -> ## 1206
	direction.NorthWest: [2]int{-1, -1}, // ## 1306 -> ## 1205
}

var EvenColumnVectors = map[direction.Direction_e][2]int{
	direction.North:     [2]int{+0, -1}, // ## 1206 -> ## 1205
	direction.NorthEast: [2]int{+1, +0}, // ## 1206 -> ## 1306
	direction.SouthEast: [2]int{+1, +1}, // ## 1206 -> ## 1307
	direction.South:     [2]int{+0, +1}, // ## 1206 -> ## 1207
	direction.SouthWest: [2]int{-1, +1}, // ## 1206 -> ## 1107
	direction.NorthWest: [2]int{-1, +0}, // ## 1206 -> ## 1106
}

// Add moves the hex in the given direction and returns the new hex.
// It always moves a single hex and allows for moving between grids
// and wrapping around the big map.
func (h Hex_t) Add(d direction.Direction_e) Hex_t {
	var vec [2]int
	if h.Column%2 == 0 { // even column
		vec = EvenColumnVectors[d]
	} else { // odd column
		vec = OddColumnVectors[d]
	}
	//log.Printf("%q: %q: %2d %2d\n", h, d, vec[0], vec[1])
	n := Hex_t{
		GridColumn: h.GridColumn,
		GridRow:    h.GridRow,
		Column:     h.Column + vec[0],
		Row:        h.Row + vec[1],
	}
	//log.Printf("%q: %q: %2d %2d %2d %2d\n", h, d, n.GridRow, n.GridColumn, n.Column, n.Row)

	// allow for moving between grids and wrapping around the edges of the big map
	if n.Column < 1 {
		n.Column = 30
		if n.GridColumn == 1 {
			n.GridColumn = 26
		} else {
			n.GridColumn = n.GridColumn - 1
		}
	} else if n.Column > 30 {
		n.Column = 1
		if n.GridColumn == 26 {
			n.GridColumn = 1
		} else {
			n.GridColumn = n.GridColumn + 1
		}
	}
	if n.Row < 1 {
		n.Row = 21
		if n.GridRow == 1 {
			n.GridRow = 26
		} else {
			n.GridRow = n.GridRow - 1
		}
	} else if n.Row > 21 {
		n.Row = 1
		if n.GridRow == 26 {
			n.GridRow = 1
		} else {
			n.GridRow = n.GridRow + 1
		}
	}

	return n
}
