// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package ast transforms the lossless CST into a simplified Abstract Syntax
// Tree with semantic validation and normalization. It normalizes month names
// to numeric values, validates year ranges, and converts turn numbers to
// integers, emitting diagnostics for validation failures. This is the third
// stage of the new three-stage parser pipeline (Lexer -> CST -> AST).
package ast
