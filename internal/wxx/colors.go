// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package wxx

import (
	"fmt"
	"strconv"
)

// hexToRGBA converts a hex color string without alpha channel (e.g. #ffff4d) to an RGBA tuple.
func hexToRGB(hex string) (float64, float64, float64, error) {
	if len(hex) != 7 {
		return 0, 0, 0, fmt.Errorf("invalid hex color length: %s", hex)
	}

	// Parse the red, green, and blue components
	r, err := strconv.ParseUint(hex[1:3], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}

	g, err := strconv.ParseUint(hex[3:5], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}

	b, err := strconv.ParseUint(hex[5:7], 16, 8)
	if err != nil {
		return 0, 0, 0, err
	}

	// Normalize the RGB values to the range [0, 1]
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	return rf, gf, bf, nil
}

// hexToRGBA converts a hex color string with alpha channel (e.g. #ffffff4d) to an RGBA tuple.
func hexToRGBA(hex string) (float64, float64, float64, float64, error) {
	if len(hex) != 9 {
		return 0, 0, 0, 0, fmt.Errorf("invalid hex color length: %s", hex)
	}

	// Parse the red, green, blue, and alpha components
	r, err := strconv.ParseUint(hex[1:3], 16, 8)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	g, err := strconv.ParseUint(hex[3:5], 16, 8)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	b, err := strconv.ParseUint(hex[5:7], 16, 8)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	a, err := strconv.ParseUint(hex[7:9], 16, 8)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	// Normalize the RGB values to the range [0, 1]
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0
	af := float64(a) / 255.0

	return rf, gf, bf, af, nil
}
