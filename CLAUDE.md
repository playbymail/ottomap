# CLAUDE.md

Guide for AI assistants working on the OttoMap codebase.

## Project Overview

OttoMap is a Go CLI tool that converts TribeNet play-by-mail game turn reports into Worldographer `.wxx` map files. It reads `.txt` files extracted from turn report `.docx` files, parses unit movements across a hexagonal grid, and generates visual maps.

**Status:** Maintenance mode (bug fixes only, no new features). The codebase contains abandoned experiments from early development that are being cleaned up.

**License:** AGPLv3

**Version:** Stored in `main.go` as a `semver.Version` struct (currently v0.62.30).

## Quick Reference

```bash
# Build
go build

# Run all tests (with race detector)
make test

# Run specific package tests
make test-lexer
make test-cst
make test-ast

# Golden snapshot tests (compare only)
make golden

# Regenerate golden snapshots after intentional output changes
make golden-update

# Run a specific test
make test GOTESTFLAGS='-run TestHeaderCrosswalk'

# Coverage report
make coverage

# Check version
go run . --version
```

## Architecture

### Processing Pipeline

```
Turn report (.txt) -> Parse -> Walk hex movements -> Build world map -> Render .wxx file
```

The main command is `render`, which orchestrates this full pipeline.

### Two Parser Generations

There are **two parallel parser implementations**. The legacy parser is still used by the main `render` command. The new parser is under active development but not yet integrated into `render`.

1. **Legacy parser** (`internal/parser/`): Pigeon PEG-based. Reports only the first error and quits. Poor error messages. Still powers the main `render` pipeline.

2. **New parser** (`internal/parsers/`): Three-stage Lexer -> CST -> AST architecture. Better error reporting and diagnostics. Not yet wired into the `render` command.
   - `internal/parsers/lexers/` - Tokenization
   - `internal/parsers/cst/` - Concrete Syntax Tree (lossless, preserves all source text)
   - `internal/parsers/ast/` - Abstract Syntax Tree (semantic, simplified for logic)

### New Pipeline (WIP)

`internal/reports/` + `internal/runners/` + `internal/walkers/` form a new orchestration layer:

```
Collect -> Normalize -> Section -> Parse -> Walk
```

This is the intended replacement for the legacy pipeline in `internal/turns/`.

### Key Packages

| Package | Purpose |
|---------|---------|
| `main` (root `.go` files) | CLI commands via Cobra |
| `internal/parser/` | Legacy Pigeon PEG parser (active) |
| `internal/parsers/` | New Lexer/CST/AST parser (WIP) |
| `internal/turns/` | Legacy turn processing pipeline |
| `internal/reports/` | New turn report processing (WIP) |
| `internal/runners/` | New pipeline orchestration (WIP) |
| `internal/walkers/` | New parse tree walking (WIP) |
| `internal/coords/` | Coordinate systems (map, grid, hex, cube) |
| `internal/tiles/` | Hex tile and world map representation |
| `internal/wxx/` | Worldographer .wxx file generation |
| `internal/terrain/` | Terrain type definitions and mappings |
| `internal/direction/` | Compass directions (N, NE, SE, S, SW, NW) |
| `internal/stores/sqlite/` | SQLite persistence (not yet integrated) |
| `internal/config/` | JSON configuration loading |
| `cerrs/` | Constant error type definitions |
| `actions/` | High-level action implementations |

### Abandoned / Incomplete Areas

These areas are vestiges of experiments and should not be extended:

- **Web server code** in legacy files: Moved to separate OttoWeb project.
- **`internal/stores/office/`**: DOCX parsing stub, not implemented.
- **`internal/stores/sqlite/`**: Database infrastructure exists (schema, queries, SQLC config) but is not connected to the render pipeline.
- **`compass/`, `winds/`, `items/`, `resources/`, `norm/`**: Minimal/placeholder packages.
- **Fleet movement**: Not implemented; waiting on real turn report examples.
- **Commented-out code in `render.go`**: Debug/experimental code left from development.

## Conventions

### Naming

- **`_t` suffix** for structs: `Map_t`, `Report_t`, `Turn_t`
- **`_e` suffix** for enums: `Direction_e`, `Terrain_e`, `UnitType_e`
- **`_i` suffix** for interfaces: `Node_i`
- Enums should implement the `Stringer` interface.

### Errors

Errors are defined as string constants using the `cerrs.Error` type:

```go
const ErrSomething = cerrs.Error("description")
```

Per-package errors go in an `errors.go` file. Use `errors.Is()` for comparison and `errors.Join()` for wrapping. Use `log.Fatalf()` only for unrecoverable CLI-level validation failures.

### Copyright Headers

All source files must have:

```go
// Copyright (c) 2024 Michael D Henderson. All rights reserved.
```

### Imports

Standard Go grouping: standard library first, then external packages. Both groups alphabetically sorted.

### Logging

Uses Go's standard `log` package with a `package: context:` prefix pattern:

```go
log.Printf("walk: input: %8d turns\n", len(input))
log.Printf("db: create: path %s\n", path)
```

## Testing

### Test Data

- Test data lives in `testdata/` directories (never committed to git).
- The default test clan ID is `0987`.
- Some tests must run from `testdata/0987` as the working directory.
- Subdirectories: `input/`, `output/`, `logs/`.

### Golden Tests

The new parser packages use golden snapshot testing. Golden test files live alongside test code.

- `make golden` - Run golden tests (read-only comparison)
- `make golden-update` - Regenerate snapshots after intentional changes

### Running Tests

Always run with the race detector:

```bash
go test ./... -race
```

Or use the Makefile, which sets `-race` by default.

## Versioning and Commits

- Semantic versioning via `github.com/maloquacious/semver`.
- Version is stored in the `version` variable in `main.go`.
- The `Build` field contains git commit info via `semver.Commit()`.
- When bumping major: reset minor and patch to 0. When bumping minor: reset patch to 0.
- **Update the version before committing changes** when asked to commit.

## Code Generation

Two code generation tools are used:

1. **Pigeon** (PEG parser generator): `//go:generate pigeon -o grammar.go grammar.peg` in `internal/parser/`.
2. **SQLC** (SQL to Go): `//go:generate sqlc generate` in `internal/stores/sqlite/`. Config in `sqlc.yaml`.

Do not manually edit generated files (`grammar.go`, `models.go`, `queries.sql.go`).

## Configuration

The app reads `data/input/ottomap.json` on startup. Key sections:

- `DebugFlags` - Logging and dump options
- `Experimental` - Feature flags (e.g., `ReverseWalker`, `SplitTrailingUnits`)
- `Parser` - Parser behavior toggles
- `Worldographer` - Map rendering settings (zoom, layers, terrain colors)

## Input Files

Turn report text files follow the naming convention:

```
YYYY-MM.CLAN.report.txt
```

Example: `901-04.0138.report.txt` (year 901, month 04, clan 0138).

Files are plain text extracted from TribeNet `.docx` turn reports. Whitespace (spaces, line breaks, page breaks) is significant to the parser.

## Hex Grid System

The world map uses flat-top hexagons with even columns shifted down:

- 676 grids (26x26, labeled AA through ZZ)
- Each grid: 30 columns x 21 rows
- Hex IDs: 4-digit CCRR (column then row), e.g., `1304`
- Full location: `GridID HexID`, e.g., `KK 1304`
- Coordinates convert between grid, map (absolute 0-based), and cube representations

Direction vectors differ for odd vs. even columns. See `internal/coords/` for the full coordinate system.

## Deployment

Two-script system in `bin/`:

- `bin/deploy.sh` - Local: builds Linux + Windows binaries, creates tarball, transfers via rsync
- `bin/install.sh` - Remote: installs on production server with backups

No CI/CD pipeline. No Docker. Manual deployment process.

## Common Pitfalls

- Do not modify original `.docx` turn report files; only edit the `.txt` copies.
- The `testdata/` directory must never be committed to git.
- The legacy parser (`internal/parser/`) and new parser (`internal/parsers/`) are separate systems; changes to one do not affect the other.
- Many packages under `internal/` are placeholders or experiments; check if a package is actually imported before assuming it is active.
- The `render` command in `render.go` is the primary user-facing functionality; be careful with changes there.
