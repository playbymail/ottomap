// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package coords

import (
	"github.com/playbymail/ottomap/internal/direction"
)

// column direction vectors defines the vectors used to determine the coordinates
// of the neighboring column based on the direction and the odd/even column
// property of the starting hex.
//
// NB: grids start at 0101 and hexes at (0,0), so "odd" and "even" are based
//     on the hex coordinates, not the grid.

var OddColumnVectors = map[direction.Direction_e][2]int{
	direction.North:     [2]int{+0, -1}, // ## 1206 -> (11, 05) -> (11, 04) -> ## 1205
	direction.NorthEast: [2]int{+1, +0}, // ## 1206 -> (11, 05) -> (12, 05) -> ## 1306
	direction.SouthEast: [2]int{+1, +1}, // ## 1206 -> (11, 05) -> (12, 06) -> ## 1307
	direction.South:     [2]int{+0, +1}, // ## 1206 -> (11, 05) -> (11, 06) -> ## 1207
	direction.SouthWest: [2]int{-1, +1}, // ## 1206 -> (11, 05) -> (10, 06) -> ## 1107
	direction.NorthWest: [2]int{-1, +0}, // ## 1206 -> (11, 05) -> (10, 05) -> ## 1106
}

var EvenColumnVectors = map[direction.Direction_e][2]int{
	direction.North:     [2]int{+0, -1}, // ## 1306 -> (12, 05) -> (12, 04) -> ## 1305
	direction.NorthEast: [2]int{+1, -1}, // ## 1306 -> (12, 05) -> (13, 04) -> ## 1405
	direction.SouthEast: [2]int{+1, +0}, // ## 1306 -> (12, 05) -> (13, 05) -> ## 1406
	direction.South:     [2]int{+0, +1}, // ## 1306 -> (12, 05) -> (12, 06) -> ## 1307
	direction.SouthWest: [2]int{-1, +0}, // ## 1306 -> (12, 05) -> (11, 05) -> ## 1206
	direction.NorthWest: [2]int{-1, -1}, // ## 1306 -> (12, 05) -> (11, 04) -> ## 1205
}
