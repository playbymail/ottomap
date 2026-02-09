// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package parser implements the legacy Pigeon PEG-based parser that converts
// TribeNet turn report text into structured movement and observation data.
// It processes unit sections (Courier, Element, Fleet, Garrison, Tribe) and
// extracts movement sequences with terrain and edge observations. This parser
// is still used by the main render pipeline.
package parser
