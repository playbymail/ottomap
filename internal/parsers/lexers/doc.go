// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package lexers tokenizes turn report input into a stream of typed tokens
// while tracking source position (line, column, byte offset). It preserves
// whitespace and comments as leading/trailing trivia on tokens, enabling
// accurate error diagnostics. This is the first stage of the new three-stage
// parser pipeline (Lexer -> CST -> AST).
package lexers
