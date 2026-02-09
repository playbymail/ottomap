// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package domain defines the shared data types used by both the parser
// and render pipelines. Types in this package are produced by
// internal/parser and consumed by internal/tiles, internal/turns,
// internal/wxx, and actions. Extracting them here decouples the parser
// from the render side so neither imports the other directly.
package domain
