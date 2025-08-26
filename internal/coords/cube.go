// Copyright (c) 2025 Michael D Henderson. All rights reserved.

package coords

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/playbymail/ottomap/internal/direction"
)

// this should be temporary. github.com/maloquacious/hexg is implementing
// a complete hexagon grid package, and we should be using CubeCoords from
// there. but that got pushed into github.com/malquacious/wxx for some work.
// it's all quite nicely mixed up. so, here we are implementing it again.

// you need to read the Red Blob Games hex grid pages to make sense of some of this.
// see https://www.redblobgames.com/grids/hexagons/
// i'm going to call their style "hexg" for short.

// Error implements constant errors
type Error string

// Error implements the Errors interface
func (e Error) Error() string {
	return string(e)
}

const (
	ErrInvalidGridCoordinates = Error("invalid grid coordinates")
)

const ODD = -1

var cube_directions = []CubeCoord{
	{q: 1, r: 0, s: -1},
	{q: 1, r: -1, s: 0},
	{q: 0, r: -1, s: 1},
	{q: -1, r: 0, s: 1},
	{q: -1, r: 1, s: 0},
	{q: 0, r: 1, s: -1},
}

// CubeCoord implements cube coordinates
type CubeCoord struct {
	q, r, s int
}

func (a CubeCoord) Add(b CubeCoord) CubeCoord {
	return CubeCoord{q: a.q + b.q, r: a.r + b.r, s: a.s + b.s}
}

func (hex CubeCoord) Neighbor(direction int) CubeCoord {
	return hex.Add(cube_directions[(6+(direction%6))%6])
}

func (h CubeCoord) ToOddQ() OddQCoord {
	parity := h.q & 1
	col, row := h.q, h.r+((h.q+ODD*parity)/2)
	return OddQCoord{col: col, row: row}
}

func (h CubeCoord) ToWorldMapCoord() WorldMapCoord {
	oddq := h.ToOddQ()
	gridRow, gridColumn := oddq.row/rowsPerGrid, oddq.col/columnsPerGrid
	subGridColumn, subGridRow := oddq.col-gridColumn*columnsPerGrid+1, oddq.row-gridRow*rowsPerGrid+1
	var gridRowCode, gridColumnCode byte
	switch {
	case gridRow < 0:
		gridRowCode = '<'
	case gridRow > 25:
		gridRowCode = '>'
	default:
		gridRowCode = 'A' + byte(gridRow)
	}
	switch {
	case gridColumn < 0:
		gridColumnCode = '<'
	case gridColumn > 25:
		gridColumnCode = '>'
	default:
		gridColumnCode = 'A' + byte(gridColumn)
	}
	var subGridColumnCode, subGridRowCode string
	switch {
	case subGridColumn < 1:
		subGridColumnCode = "<<"
	case subGridColumn > columnsPerGrid:
		subGridColumnCode = ">>"
	default:
		subGridColumnCode = fmt.Sprintf("%02d", subGridColumn)
	}
	switch {
	case subGridRow < 1:
		subGridRowCode = "<<"
	case subGridRow > rowsPerGrid:
		subGridRowCode = ">>"
	default:
		subGridRowCode = fmt.Sprintf("%02d", subGridRow)
	}
	return WorldMapCoord{
		id:   fmt.Sprintf("%c%c %s%s", gridRowCode, gridColumnCode, subGridColumnCode, subGridRowCode),
		cube: h,
	}
}

// OddQCoord implements "odd-q," an offset coordinate with flat top hexes and odd columns pushed down.
type OddQCoord struct {
	col, row int
}

func (h OddQCoord) ToCube() CubeCoord {
	parity := h.col & 1
	q, r := h.col, h.row-((h.col+ODD*parity)/2)
	return CubeCoord{q: q, r: r, s: -q - r}
}

// TribeNet uses a map composed of grids (see https://tribenet.wiki/mapping/grid)
//
// we know that TribeNet world maps have an origin of (gridRow: A, gridColumn: A, column: 1, row: 1) and
// (gridRow: A, gridRow: A, column: 2, row: 1) is pushed down (it is "south-east" of the origin).
// in hexg terms, that would be offset coordinates with an even-q layout
// (vertical columns, even columns shoved down).
//
// but...
//
// hexg uses (column: 0, row: 0) as the origin, so when we translate (1, 1) => (0, 0),
// we find that we actually have an odd-q layout (vertical columns, odd columns shoved down).
//
// WorldMapCoord implements "odd-q," an offset coordinate with flat top hexes and odd columns pushed down.
// It implements a different Stringer, displaying coordinates as "AB 0102," where:
//   - "A"  is grid row        with a range of A  ... Z
//   - "B"  is grid column     with a range of A  ... Z
//   - "01" is sub-grid column with a range of 1 ... 30
//   - "02" is sub-grid row    with a range of 1 ... 21
//
// Note that the TribeNet world maps:
// * use "AA 0101" for the origin
// * accepts "##" as an anonymous grid
// * considers "N/A" to be the null coordinate

const (
	columnsPerGrid = 30
	rowsPerGrid    = 21
)

// WorldMapCoord stores the location of a tile on the TribeNet "world map."
// It works through the translation between world map coordinates, offset coordinates, and cube coordinates.
type WorldMapCoord struct {
	id   string
	cube CubeCoord
}

// NewWorldMapCoord converts a world map coordinate to cube coordinates, returning any errors.
// For historical reasons, we treat grid "##" as "QQ" and an id of "N/A" as cube coordinates (0,0,0).
//
// Note that we always convert the grid id to uppercase.
func NewWorldMapCoord(id string) (WorldMapCoord, error) {
	// force to uppercase before converting
	id = strings.ToUpper(id)

	if validGridId := id == "N/A" || (len(id) == 7 && id[2] == ' '); !validGridId {
		return WorldMapCoord{}, ErrInvalidGridCoordinates
	}

	if id == "N/A" {
		return WorldMapCoord{id: id}, nil
	}

	// extract and validate the grid row and column
	gridRow, gridColumn := int(id[0]), int(id[1])
	if gridRow == '#' && gridColumn == '#' {
		// we have to put obscured coordinates somewhere, so we will put them in "QQ"
		gridRow, gridColumn = 'Q', 'Q'
	} else if isValidGridRow := 'A' <= gridRow && gridRow <= 'Z'; !isValidGridRow {
		return WorldMapCoord{}, ErrInvalidGridCoordinates
	} else if isValidGridColumn := 'A' <= gridColumn && gridColumn <= 'Z'; !isValidGridColumn {
		return WorldMapCoord{}, ErrInvalidGridCoordinates
	}
	// convert from "A" ... "Z" to 0 ... 25
	gridRow, gridColumn = gridRow-'A', gridColumn-'A'

	// extract and validate the sub-grid column and row
	subGridColumn, err := strconv.Atoi(id[3:5])
	if err != nil {
		return WorldMapCoord{}, ErrInvalidGridCoordinates
	}
	subGridRow, err := strconv.Atoi(id[5:])
	if err != nil {
		return WorldMapCoord{}, ErrInvalidGridCoordinates
	}
	if isValidSubGridColumn := 1 <= subGridColumn && subGridColumn <= columnsPerGrid; !isValidSubGridColumn {
		return WorldMapCoord{}, ErrInvalidGridCoordinates
	} else if isValidSubGridRow := 1 <= subGridRow && subGridRow <= rowsPerGrid; !isValidSubGridRow {
		return WorldMapCoord{}, ErrInvalidGridCoordinates
	}
	// convert from 1 based to 0 based
	subGridColumn, subGridRow = subGridColumn-1, subGridRow-1

	return WorldMapCoord{
		id: id,
		cube: OddQCoord{
			col: gridColumn*columnsPerGrid + subGridColumn,
			row: gridRow*rowsPerGrid + subGridRow,
		}.ToCube(),
	}, nil
}

// Equals returns true if the coordinates were created from the same world map coordinates.
// This is wonky because of "N/A" and obscured grids, but seems like the best compromise.
func (c WorldMapCoord) Equals(b WorldMapCoord) bool {
	return c.id == b.id
}

// IsNA returns true if the id of the coordinates is "N/A"
func (c WorldMapCoord) IsNA() bool {
	return c.id == "N/A"
}

func (c WorldMapCoord) Move(ds ...direction.Direction_e) WorldMapCoord {
	to := c.cube
	for _, d := range ds {
		switch d {
		case direction.North:
			to = to.Neighbor(2)
		case direction.NorthEast:
			to = to.Neighbor(1)
		case direction.SouthEast:
			to = to.Neighbor(0)
		case direction.South:
			to = to.Neighbor(5)
		case direction.SouthWest:
			to = to.Neighbor(4)
		case direction.NorthWest:
			to = to.Neighbor(3)
		default:
			panic(fmt.Sprintf("assert(direction != %d)", d))
		}

	}
	return to.ToWorldMapCoord()
}

func (c WorldMapCoord) MoveReverse(ds ...direction.Direction_e) WorldMapCoord {
	to := c.cube
	for _, d := range ds {
		switch d {
		case direction.North:
			to = to.Neighbor(5)
		case direction.NorthEast:
			to = to.Neighbor(4)
		case direction.SouthEast:
			to = to.Neighbor(3)
		case direction.South:
			to = to.Neighbor(2)
		case direction.SouthWest:
			to = to.Neighbor(1)
		case direction.NorthWest:
			to = to.Neighbor(0)
		default:
			panic(fmt.Sprintf("assert(direction != %d)", d))
		}
	}
	return to.ToWorldMapCoord()
}

// String implements the strings.Stringer interface.
// It returns the internal coordinates converted to a world map coordinate.
// "N/A" and obscured coordinates may cause some surprise.
// We use "<" and ">" in the result to signal out-of-bound values.
func (c WorldMapCoord) String() string {
	oddq := c.cube.ToOddQ()
	gridRow, gridColumn := oddq.row/rowsPerGrid, oddq.col/columnsPerGrid
	subGridColumn, subGridRow := oddq.col-gridColumn*columnsPerGrid+1, oddq.row-gridRow*rowsPerGrid+1
	var gridRowCode, gridColumnCode byte
	switch {
	case gridRow < 0:
		gridRowCode = '<'
	case gridRow > 25:
		gridRowCode = '>'
	default:
		gridRowCode = 'A' + byte(gridRow)
	}
	switch {
	case gridColumn < 0:
		gridColumnCode = '<'
	case gridColumn > 25:
		gridColumnCode = '>'
	default:
		gridColumnCode = 'A' + byte(gridColumn)
	}
	var subGridColumnCode, subGridRowCode string
	switch {
	case subGridColumn < 1:
		subGridColumnCode = "<<"
	case subGridColumn > columnsPerGrid:
		subGridColumnCode = ">>"
	default:
		subGridColumnCode = fmt.Sprintf("%02d", subGridColumn)
	}
	switch {
	case subGridRow < 1:
		subGridRowCode = "<<"
	case subGridRow > rowsPerGrid:
		subGridRowCode = ">>"
	default:
		subGridRowCode = fmt.Sprintf("%02d", subGridRow)
	}
	return fmt.Sprintf("%c%c %s%s", gridRowCode, gridColumnCode, subGridColumnCode, subGridRowCode)
}
