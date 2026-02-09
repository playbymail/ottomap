// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package wxx builds Worldographer .wxx map file structures from the internal
// tile representation. It maps OttoMap terrain enums to Worldographer tile
// template names via configuration and manages the rendering attributes of each
// hex including elevation, resources, settlements, and edge features. This is
// the final stage before writing the .wxx XML output.
package wxx
