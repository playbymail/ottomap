// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package coords

import (
	"github.com/playbymail/ottomap/cerrs"
	"strconv"
	"strings"
)

func StringToGridCoords(s string) (Grid, error) {
	var gc Grid

	if !(len(s) == 7 && s[2] == ' ') {
		return Grid{}, cerrs.ErrInvalidGridCoordinates
	} else if strings.HasPrefix(s, "##") {
		// todo: find the best way to deal with the "##" case
		// gc.BigMapRow, gc.BigMapColumn = 0, 0
		gc.BigMapRow, gc.BigMapColumn = int('O'-'A'), int('O'-'A')
	} else if gc.BigMapRow = int(s[0] - 'A'); !(0 <= gc.BigMapRow && gc.BigMapRow < 26) {
		return Grid{}, cerrs.ErrInvalidGridCoordinates
	} else if gc.BigMapColumn = int(s[1] - 'A'); !(0 <= gc.BigMapColumn && gc.BigMapColumn < 26) {
		return Grid{}, cerrs.ErrInvalidGridCoordinates
	}

	var err error

	if gc.GridColumn, err = strconv.Atoi(s[3:5]); err != nil {
		return Grid{}, cerrs.ErrInvalidGridCoordinates
	} else if !(1 <= gc.GridColumn && gc.GridColumn <= 30) {
		return Grid{}, cerrs.ErrInvalidGridCoordinates
	}

	if gc.GridRow, err = strconv.Atoi(s[5:]); err != nil {
		return Grid{}, cerrs.ErrInvalidGridCoordinates
	} else if !(1 <= gc.GridRow && gc.GridRow <= 21) {
		return Grid{}, cerrs.ErrInvalidGridCoordinates
	}

	return gc, nil
}
