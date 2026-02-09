// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package turns implements the legacy turn processing pipeline that orchestrates
// parsing turn reports and walking unit movements across the hexagonal map. It
// reads turn report files, processes movement results, updates tile observations
// for terrain, edges, and neighbors, and tracks unit encounters. This is the
// primary engine used by the render command.
package turns
