// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

// worldographer hexes are 300x300, flat top.
// (225, 600) (300, 750) (450, 750) (525, 600) (450, 450) (300, 450)

// FlattenedHexagonVertices returns the vertices of a flattened hexagon centered at the given point,
// adjusted for top and left margins. Assumes that the hexagon is based on a 300x300 square and that
// the canvas has (0,0) at the top left, x increases to the right, and y increases down.
func FlattenedHexagonVertices(center Point, topMargin, leftMargin float64) [6]Point {
	var vertices [6]Point

	// Adjust center by the margins
	adjustedCenter := Point{
		X: center.X + leftMargin,
		Y: center.Y + topMargin,
	}

	// Define the offsets based on the flattened hexagon dimensions
	offsets := [6][2]float64{
		{-150, 0},   // left vertex
		{-75, -150}, // top-left vertex
		{75, -150},  // top-right vertex
		{150, 0},    // right vertex
		{75, 150},   // bottom-right vertex
		{-75, 150},  // bottom-left vertex
	}

	// Calculate vertices based on adjusted center and offsets
	for i, offset := range offsets {
		vertices[i] = Point{
			X: adjustedCenter.X + offset[0],
			Y: adjustedCenter.Y + offset[1],
		}
	}

	return vertices
}
