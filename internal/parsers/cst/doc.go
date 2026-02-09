// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package cst builds a lossless Concrete Syntax Tree from tokens, preserving
// every lexical detail including punctuation and whitespace. It reports detailed
// parse diagnostics with precise source locations and recovers from errors by
// capturing malformed constructs in BadTopLevel nodes. This is the second stage
// of the new three-stage parser pipeline (Lexer -> CST -> AST).
package cst
