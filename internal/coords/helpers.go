// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package coords

import (
	"fmt"
	"github.com/playbymail/ottomap/cerrs"
	"github.com/playbymail/ottomap/internal/direction"
	"log"
	"strconv"
	"strings"
)

func Move(hex string, directions ...direction.Direction_e) string {
	if hex == "N/A" || strings.HasPrefix(hex, "##") {
		log.Printf("error: hex %q: %v\n", hex, fmt.Errorf("bad hex"))
		return hex
	}
	to, err := HexToMap(hex)
	if err != nil {
		log.Printf("error: hex %q: %v\n", hex, err)
		panic(err)
	}
	for _, d := range directions {
		to = to.Add(d)
	}
	return to.ToHex()
}

func ColumnRowToMap(column int, row int) Map {
	return Map{Column: column, Row: row}
}

func HexToMap(hex string) (Map, error) {
	if hex == "N/" || strings.HasPrefix(hex, "##") {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	} else if !(len(hex) == 7 && hex[2] == ' ') {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	}
	grid, digits, ok := strings.Cut(hex, " ")
	if !ok {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	} else if len(grid) != 2 {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	} else if strings.TrimRight(grid, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	} else if len(digits) != 4 {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	} else if strings.TrimRight(digits, "0123456789") != "" {
		return Map{}, cerrs.ErrInvalidGridCoordinates
	}
	bigMapRow, bigMapColumn := int(grid[0]-'A'), int(grid[1]-'A')
	littleMapColumn, err := strconv.Atoi(digits[:2])
	if err != nil {
		panic(err)
	}
	littleMapRow, err := strconv.Atoi(digits[2:])
	if err != nil {
		panic(err)
	}
	// log.Printf("hex %q brow %2d bcol %2d mcol %2d mrow %2d\n", hex, bigMapRow, bigMapColumn, littleMapColumn, littleMapRow)
	return Map{
		Column: bigMapColumn*30 + littleMapColumn - 1,
		Row:    bigMapRow*21 + littleMapRow - 1,
	}, nil

	return Map{}, nil
}
