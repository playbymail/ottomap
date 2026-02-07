# OttoMap Cleanup Sprint Plan

This document outlines a series of small, focused sprints to clean up the
OttoMap codebase. Each sprint has a single well-defined goal that can be
completed in one coding session. Sprints are ordered to minimize dependency
conflicts: we remove dead code first (easiest, no risk), then orphaned
infrastructure, then clean up active code, and finally add documentation.

Current version: **0.69.0**

Each sprint bumps the minor version. Sprint numbering starts at 63.

---

## Sprint 63 — Remove `internal/stores/office/`

**Goal:** Delete the abandoned DOCX parsing stub package.

**Context:** `internal/stores/office/` contains a single file (`docx.go`,
~180 LOC) that was never implemented. It is not imported by any other
package. This is the safest possible removal — zero impact on any code path.

**Scope:**
- Delete `internal/stores/office/` directory
- Verify no imports reference it (`grep` for `stores/office`)
- Run `go build` and `make test` to confirm nothing breaks
- Version: **0.63.0**

---

## Sprint 64 — Remove `internal/hexes/`

**Goal:** Delete the orphaned hex utilities package.

**Context:** `internal/hexes/` contains `hexes.go` (~160 LOC) and
`hexes_test.go` (~322 LOC). It defines a `Hex_t` coordinate type that
duplicates functionality in `internal/coords/`. It is not imported by any
package outside its own tests. Removing it has zero impact on the build.

**Scope:**
- Delete `internal/hexes/` directory
- Verify no imports reference it
- Run `go build` and `make test`
- Version: **0.64.0**

---

## Sprint 65 — Remove the `db` command and SQLite infrastructure

**Goal:** Remove all database-related code and the `modernc.org/sqlite`
dependency.

**Context:** The SQLite store (`internal/stores/sqlite/`, ~546 LOC Go +
~487 LOC SQL) and the `db` CLI command (`db.go`, ~308 LOC) were an
experiment that was never integrated into the render pipeline. The `db`
subcommand tree (`db create database`, `db load files`, `db load path`) is
the only consumer. Removing this also eliminates the heaviest transitive
dependency tree in `go.mod` (`modernc.org/sqlite` and its 8+ indirect
dependencies: `dustin/go-humanize`, `hashicorp/golang-lru`,
`mattn/go-isatty`, `ncruces/go-strftime`, `remyoudompheng/bigfft`,
`modernc.org/gc`, `modernc.org/libc`, `modernc.org/mathutil`,
`modernc.org/memory`, `modernc.org/strutil`, `modernc.org/token`).

**Scope:**
- Delete `internal/stores/sqlite/` directory (including `sqlc.yaml`,
  `schema.sql`, `queries.sql`)
- Delete `db.go` (root level)
- Remove `cmdDb` registration and all `cmdDb*` flag setup from `main.go`
  (lines ~88-120)
- Remove `loadInputFile` and `removeInputFile` helper functions (only used
  by db commands; also defined in `db.go`)
- Run `go mod tidy` to remove `modernc.org/sqlite` and its indirect deps
- Run `go build` and `make test`
- Version: **0.65.0**

---

## Sprint 66 — Remove `load.go` stub and `parse` command skeleton

**Goal:** Remove the empty `load.go` file and the non-functional `parse`
command.

**Context:** `load.go` is an empty stub (just a package declaration, 3
lines). The `parse` CLI command (`parse.go`, ~149 LOC) reads files but does
nothing with them — `parseScrubbedFiles()` is a no-op returning nil. The
`parse reports` subcommand just logs file names and sizes. Neither is used
by any other code. The types `parsedFileName_t` and `turnId_t` defined in
`parse.go` need to be checked for use elsewhere before removal.

**Scope:**
- Delete `load.go`
- Audit `parsedFileName_t`, `turnId_t`, `validateClanId`,
  `validateTurnId`, and `parseScrubbedFiles` for usage outside `parse.go`
  — move any that are shared; delete the rest
- Delete `parse.go` or reduce to just shared types if needed
- Remove `cmdParse` registration and flag setup from `main.go`
  (lines ~128-137)
- Run `go build` and `make test`
- Version: **0.66.0**

---

## Sprint 67 — Clean commented-out code from `render.go` and `main.go`

**Goal:** Remove all commented-out debug/experimental code blocks.

**Context:** `render.go` and `main.go` contain several blocks of
commented-out debug logging that are no longer useful. These were left from
early development and make the code harder to read.

**Known locations:**
- `main.go:174-176` — commented-out JSON config dump
- `render.go:244` — commented-out `maxYear` logging
- `render.go:477-485` — commented-out unit tracking debug block
- Any other commented-out blocks found during the sprint

**Scope:**
- Remove all commented-out code blocks in `render.go` and `main.go`
- Do NOT remove TODO comments or comments that explain intent
- Run `go build` and `make test`
- Version: **0.67.0**

---

## Sprint 68 — Remove `internal/stores/` parent directory

**Goal:** Clean up the now-empty `internal/stores/` directory tree.

**Context:** After Sprint 63 removed `office/` and Sprint 65 removed
`sqlite/`, the `internal/stores/` directory should be empty (or contain
only the parent directory). This sprint verifies that and cleans up any
remaining artifacts.

**Scope:**
- Verify `internal/stores/` is empty after prior sprints
- Delete the `internal/stores/` directory
- Check for any lingering references to `stores` in imports
- If `internal/stores/` still contains files not covered by earlier sprints,
  assess and remove them
- Run `go build` and `make test`
- Version: **0.68.0**

**Note:** This sprint can be merged into Sprint 65 if the stores directory
will clearly be empty after that sprint completes. Listed separately for
safety.

---

## Sprint 69 — Audit and clean `internal/winds/` and `internal/items/`

**Goal:** Determine if `winds` and `items` packages can be removed, and
remove them if safe.

**Context:** `internal/winds/` (~62 LOC) defines wind strength enums and
`internal/items/` (~461 LOC) defines 200+ item type enums. Both are
imported only by the legacy PEG parser's generated `grammar.go` and
`types.go`. The parser references these types in its grammar rules but the
parsed values are never propagated to the render pipeline — they are
effectively dead data. However, removing them requires either modifying the
PEG grammar (which regenerates `grammar.go`) or confirming the parser still
compiles without the removed types.

**Scope:**
- Trace how `winds` and `items` types flow through `grammar.go` →
  `parser.go` → render pipeline to confirm values are discarded
- If values are truly unused past the parser: remove the grammar rules that
  reference them, regenerate `grammar.go` via `go generate`, then delete
  the packages
- If removal would break the parser: document findings and defer to a
  future sprint
- Run `go build` and `make test`
- Version: **0.69.0**

**Outcome:** Audit confirmed that parsed wind/item values are discarded
within the parser (items are explicitly ignored during move processing;
wind data is never read after being stored in `Movement_t`). However,
these types are used by external packages (e.g. `github.com/playbymail/tndocx`)
and must not be removed from the grammar. Both packages were documented
with package-level comments explaining the external dependency constraint.
Additionally, pre-existing build failures in the new parser pipeline
(`internal/parsers/`) were fixed: the lexer was updated to produce
keyword-specific token kinds (`KeywordCurrent`, `KeywordTurn`, `MonthName`,
`Identifier`) per the LEXING.md specification, and the CST parser was
updated to handle EOL tokens and mixed alphanumeric turn numbers.

---

## Sprint 70 — Remove `github.com/playbymail/tndocx` dependency

**Goal:** Evaluate whether the `tndocx` dependency can be removed.

**Context:** `scrub.go` imports `github.com/playbymail/tndocx` for
DOCX-to-text conversion. The `scrub` command is a user-facing utility that
pre-processes turn report files. If DOCX scrubbing has been moved to the
separate OttoWeb project (as CLAUDE.md suggests for web server code), this
dependency may be removable. If the `scrub` command is still needed, this
sprint should document that decision and skip the removal.

**Scope:**
- Determine if the `scrub` command is still actively used by players
- If no longer needed: remove `scrub.go`, its registration in `main.go`,
  and run `go mod tidy`
- If still needed: document the decision and close the sprint with no
  code changes
- Run `go build` and `make test`
- Version: **0.70.0**

---

## Sprint 71 — Remove `github.com/mdhender/semver` indirect dependency

**Goal:** Clean up the stale indirect semver dependency.

**Context:** `go.mod` lists `github.com/mdhender/semver` as an indirect
dependency alongside the actively used `github.com/maloquacious/semver`.
This appears to be a leftover from a package rename or migration. A simple
`go mod tidy` may resolve it, but it should be verified.

**Scope:**
- Run `go mod tidy`
- Verify `github.com/mdhender/semver` is removed from `go.mod` and
  `go.sum`
- If it persists, trace what still references it and resolve
- Run `go build` and `make test`
- Version: **0.71.0**

---

## Sprint 72 — Update CLAUDE.md to reflect post-cleanup state

**Goal:** Revise CLAUDE.md to accurately describe the codebase after
cleanup sprints.

**Context:** CLAUDE.md currently documents packages and features that will
have been removed by earlier sprints (stores/office, stores/sqlite, db
command, parse command, hexes, etc.). The "Abandoned / Incomplete Areas"
section, package table, and other references need updating.

**Scope:**
- Remove references to deleted packages (`stores/office`, `stores/sqlite`,
  `hexes`, `db` command, `parse` command, etc.)
- Update the "Abandoned / Incomplete Areas" section
- Update the "Key Packages" table
- Update the "Code Generation" section (remove SQLC references)
- Update the "Configuration" section if the `db` command docs were there
- Verify all remaining package descriptions are accurate
- Version: **0.72.0**

---

## Sprint 73 — Add package-level doc comments to active internal packages

**Goal:** Add Go doc comments to each actively used internal package.

**Context:** Many internal packages lack `doc.go` files or package-level
documentation. Adding brief doc comments improves discoverability via
`go doc` and helps future contributors understand package responsibilities.

**Packages to document (active only):**
- `internal/parser/`
- `internal/parsers/` (and sub-packages: `lexers/`, `cst/`, `ast/`)
- `internal/coords/`
- `internal/tiles/`
- `internal/wxx/`
- `internal/terrain/`
- `internal/direction/`
- `internal/edges/`
- `internal/config/`
- `internal/stdlib/`
- `internal/turns/`
- `internal/units/`
- `internal/unit_movement/`
- `internal/results/`
- `actions/`
- `cerrs/`

**Scope:**
- Add a `doc.go` with a package comment to each listed package that lacks
  one
- Keep comments to 2-4 sentences: what the package does, who uses it
- Do not modify any existing code
- Run `go build` and `make test`
- Version: **0.73.0**

---

## Sprint 74 — Update user-facing documentation in `docs/`

**Goal:** Refresh the user documentation to match current CLI behavior.

**Context:** The docs directory contains user guides (`OttoMap_CLI.adoc`,
`OttoMap_Quick_Start_Guide.adoc`, `OttoMap_Users_Manual.adoc`) and
reference docs (`ERRORS.md`, `LANGUAGE.md`, `TERRAIN.md`). After removing
the `db` and `parse` commands, the CLI documentation needs updating. The
user guides may also reference removed features.

**Scope:**
- Update `OttoMap_CLI.adoc` to remove references to deleted commands
- Review and update `OttoMap_Quick_Start_Guide.adoc` and
  `OttoMap_Users_Manual.adoc`
- Verify `ERRORS.md`, `LANGUAGE.md`, and `TERRAIN.md` are still accurate
- Run `go build` and `make test` (docs changes shouldn't break anything,
  but verify)
- Version: **0.74.0**

---

## Sprint Summary

| Sprint | Version | Goal | Risk | Est. Files Changed |
|--------|---------|------|------|--------------------|
| 63 | 0.63.0 | Remove `internal/stores/office/` | None | 1 deleted |
| 64 | 0.64.0 | Remove `internal/hexes/` | None | 2 deleted |
| 65 | 0.65.0 | Remove `db` command + SQLite infrastructure | Low | ~12 deleted/modified |
| 66 | 0.66.0 | Remove `load.go` + `parse` command | Low | 3 deleted/modified |
| 67 | 0.67.0 | Clean commented-out code | None | 2 modified |
| 68 | 0.68.0 | Remove empty `internal/stores/` | None | 0-1 deleted |
| 69 | 0.69.0 | Audit/remove `winds` and `items` | Medium | 2-5 modified/deleted |
| 70 | 0.70.0 | Evaluate `tndocx` dependency | Low | 0-3 deleted/modified |
| 71 | 0.71.0 | Clean stale indirect dependency | None | 1 modified |
| 72 | 0.72.0 | Update CLAUDE.md | None | 1 modified |
| 73 | 0.73.0 | Add package doc comments | None | ~16 created |
| 74 | 0.74.0 | Update user documentation | None | 3-6 modified |

**Risk levels:**
- **None** — Deleting provably unused code or editing docs only
- **Low** — Removing code with a single known consumer, easy to verify
- **Medium** — Requires tracing through generated code or making a judgment call
