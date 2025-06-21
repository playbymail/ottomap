// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"fmt"
	"github.com/playbymail/ottomap/internal/direction"
	"math"
)

type Point struct {
	X float64
	Y float64
}

func (p Point) Scale(s float64) Point {
	return Point{
		X: p.X * s,
		Y: p.Y * s,
	}
}

func (p Point) Translate(t Point) Point {
	return Point{
		X: p.X + t.X,
		Y: p.Y + t.Y,
	}
}

var (
	// Define the offsets based on the flattened hexagon dimensions
	flattenedHexOffsets = [6]Point{
		{-150, 0},   // left vertex
		{-75, -150}, // top-left vertex
		{75, -150},  // top-right vertex
		{150, 0},    // right vertex
		{75, 150},   // bottom-right vertex
		{-75, 150},  // bottom-left vertex
	}
)

// coordsToPoints returns the center point and vertices of a hexagon centered at the
// given column and row. It converts the column, row to the pixel at the center of the
// corresponding tile, then calculates the vertices based on that point.
// The center point is the first value in the returned slice.
func coordsToPoints(column, row int) [7]Point {
	const height, width = 300, 300
	const halfHeight, oneQuarterWidth, threeQuarterWidth = height / 2, width / 4, width * 3 / 4
	const leftMargin, topMargin = width / 2, halfHeight

	// points is the center plus the six vertices
	var points [7]Point

	points[0].X = float64(column)*threeQuarterWidth + leftMargin
	if column&1 == 1 { // shove odd rows down half the height of a tile
		points[0].Y = float64(row)*height + halfHeight + topMargin
	} else {
		points[0].Y = float64(row)*height + topMargin
	}

	// Calculate vertices based on offsets from center
	for i, offset := range flattenedHexOffsets {
		points[i+1] = Point{
			X: points[0].X + offset.X,
			Y: points[0].Y + offset.Y,
		}
	}

	return points
}

func bottomLeftCenter(v [7]Point) Point {
	bc := edgeCenter(direction.South, v)
	return Point{X: (v[6].X + bc.X) / 2, Y: bc.Y}
}

func offsetSouthEastCenter(v [7]Point) Point {
	bc := edgeCenter(direction.SouthEast, v)
	return Point{X: (v[6].X + bc.X) / 2, Y: bc.Y}
}

func distance(p1, p2 Point) float64 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func edgeCenter(edge direction.Direction_e, v [7]Point) Point {
	var from, to int
	switch edge {
	case direction.North:
		from, to = 2, 3
	case direction.NorthEast:
		from, to = 3, 4
	case direction.SouthEast:
		from, to = 4, 5
	case direction.South:
		from, to = 5, 6
	case direction.SouthWest:
		from, to = 6, 1
	case direction.NorthWest:
		from, to = 1, 2
	default:
		panic(fmt.Sprintf("assert(direction != %d)", edge))
	}
	return midpoint(v[from], v[to])
}

func midpoint(p1, p2 Point) Point {
	return Point{
		X: (p1.X + p2.X) / 2,
		Y: (p1.Y + p2.Y) / 2,
	}
}

func settlementLabelXY(label string, v [7]Point) Point {
	return edgeCenter(direction.South, v).Translate(Point{X: float64(len(label)), Y: -35})
}
