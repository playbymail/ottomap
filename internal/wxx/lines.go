// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

// GetMidpoint calculates the midpoint of a given line segment.
//
// This function takes two points, `a` and `b`, which represent the endpoints of a line segment.
// It returns a new point that represents the midpoint of the line segment connecting `a` and `b`.
//
// Parameters:
//   - a (Point): The first endpoint of the line segment.
//   - b (Point): The second endpoint of the line segment.
//
// Returns:
//   - Point: A new point that represents the midpoint of the line segment.
//
// Example:
//
//	a := Point{X: 1, Y: 2}
//	b := Point{X: 3, Y: 4}
//	midpoint := GetMidpoint(a, b)
//	fmt.Println(midpoint) // Output: {X: 2, Y: 3}
func GetMidpoint(a, b Point) Point {
	return Point{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
}

// GetCenteredHalfSegment calculates the midline segment of a given line segment.
//
// This function takes two points, `a` and `b`, which represent the endpoints of a line segment.
// It returns two points that define a new segment located in the middle of the original segment.
// The resulting segment is one-half the length of the original segment, offset by one-quarter
// of the length of the original segment.
//
// Parameters:
//   - a (Point): The first endpoint of the original line segment.
//   - b (Point): The second endpoint of the original line segment.
//
// Returns:
//   - start (Point): The starting point of the midline segment.
//   - end (Point): The ending point of the midline segment.
//
// Example:
//
//	a := Point{X: 1, Y: 2}
//	b := Point{X: 5, Y: 6}
//	start, end := GetMidlineSegment(a, b)
//	fmt.Println(start, end) // Output: {X: 2, Y: 3} {X: 4, Y: 5}
func GetCenteredHalfSegment(a, b Point) (start, end Point) {
	centerPoint := GetMidpoint(a, b)
	return GetMidpoint(a, centerPoint), GetMidpoint(centerPoint, b)
}
