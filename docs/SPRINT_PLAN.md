# OttoMap Sprint Plan

This document tracks all completed and planned sprints. Sprints 63-74 were
cleanup sprints (dead code removal, dependency cleanup, documentation).
Sprints 75-83 are refactoring sprints to separate the parser and render
pipelines into independent packages with shared domain types.

Current version: **0.76.0**

Each sprint bumps the minor version. Sprint numbering starts at 63.

---

# Phase 1: Cleanup (Completed)

## Sprint 63 — Remove `internal/stores/office/` [COMPLETED]

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

## Sprint 64 — Remove `internal/hexes/` [COMPLETED]

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

## Sprint 65 — Remove the `db` command and SQLite infrastructure [COMPLETED]

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

## Sprint 66 — Remove `load.go` stub and `parse` command skeleton [COMPLETED]

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

## Sprint 67 — Clean commented-out code from `render.go` and `main.go` [COMPLETED]

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

## Sprint 68 — Remove `internal/stores/` parent directory [COMPLETED]

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

## Sprint 69 — Audit and clean `internal/winds/` and `internal/items/` [COMPLETED]

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

## Sprint 70 — Remove `github.com/playbymail/tndocx` dependency [COMPLETED]

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

## Sprint 71 — Remove `github.com/mdhender/semver` indirect dependency [COMPLETED]

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

## Sprint 72 — Update CLAUDE.md to reflect post-cleanup state [COMPLETED]

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

## Sprint 73 — Add package-level doc comments to active internal packages [COMPLETED]

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

## Sprint 74 — Update user-facing documentation in `docs/` [COMPLETED]

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

# Phase 2: Parser/Render Refactoring

**Motivation:** The parser and render pipelines are tightly coupled
through shared types. Types defined in `internal/parser/` (such as
`Encounter_t`, `Settlement_t`, `Special_t`, `UnitId_t`, `Report_t`,
`Move_t`, `Turn_t`) are directly embedded in `internal/tiles/`,
`internal/wxx/`, `internal/turns/`, and `actions/`. This means every
downstream package imports `internal/parser/`, even when it only needs
the data structures and not the parsing logic.

**Goal:** Extract the parser and render into two independent packages.
Parser-specific types stay in `internal/parser/`. Render-specific types
stay in their respective packages (`internal/wxx/`, etc.). Types shared
between both are hoisted into a new `internal/domain/` package.

**Current dependency graph (problematic):**

```
render.go
  ├── internal/parser/    ← parsing logic + shared types (tightly coupled)
  ├── internal/turns/     ← imports parser for types
  ├── internal/tiles/     ← imports parser for types
  ├── actions/            ← imports parser for types
  └── internal/wxx/       ← imports parser for types
```

**Target dependency graph:**

```
render.go
  ├── internal/parser/    ← parsing logic only, imports domain
  ├── internal/turns/     ← imports domain (not parser)
  ├── internal/tiles/     ← imports domain (not parser)
  ├── actions/            ← imports domain (not parser)
  ├── internal/wxx/       ← imports domain (not parser)
  └── internal/domain/    ← shared types used by all
```

---

## Sprint 75 — Dependency Analysis and Migration Plan

**Goal:** Produce a complete, field-level analysis of every type that
crosses the parser/render boundary, and validate the migration strategy
before any code is moved.

**Context:** Before moving any types, we need to know exactly which
fields and methods of each parser type are used by each downstream
consumer. A type that looks shared may have methods only used by the
parser, or fields only used by the renderer. Getting this wrong leads to
circular imports or broken builds in later sprints.

**Scope:**

1. **Type inventory.** For each type currently in `internal/parser/types.go`,
   classify it as:
   - **Parser-only:** used exclusively within `internal/parser/` (e.g.,
     `DirectionTerrain_t`, `Exhausted_t`, `ProhibitedFrom_t`,
     `NearHorizon_t`, `Neighbor_t`, `MissingEdge_t`, `Longhouse_t`,
     `Patrolled_t`, `FoundUnit_t`, `FoundItem_t`)
   - **Shared:** used by both parser and at least one downstream package
     (e.g., `Turn_t`, `Moves_t`, `Move_t`, `Report_t`, `Scout_t`,
     `Border_t`, `Encounter_t`, `Settlement_t`, `Special_t`, `UnitId_t`,
     `FarHorizon_t`, `Scry_t`)
   - **Render-only:** used only by render-side packages (unlikely, but
     check)

2. **Field-level usage matrix.** For each shared type, document which
   fields are accessed by which packages:
   - `internal/parser/` (producer)
   - `internal/turns/` (consumer)
   - `internal/tiles/` (consumer)
   - `internal/wxx/` (consumer, via `wxx.Features`)
   - `actions/` (consumer)
   - `render.go` (orchestrator)

3. **Method audit.** For each shared type, document which methods exist
   and which packages call them. Methods that are only called by the
   parser should stay in parser (as functions or via embedding). Methods
   used by downstream packages must move to `internal/domain/`.

4. **Constant audit.** Identify constants defined in `internal/parser/`
   that are used downstream (e.g., `LastTurnCurrentLocationObscured`).
   These must also move to `internal/domain/`.

5. **Enum package review.** Confirm that the existing shared enum packages
   (`internal/direction/`, `internal/terrain/`, `internal/edges/`,
   `internal/results/`, `internal/resources/`, `internal/coords/`,
   `internal/compass/`) do not need changes and can remain as-is.

6. **Document the migration order.** Based on the analysis, produce a
   concrete list of:
   - Types moving to `internal/domain/`
   - Methods moving with them
   - Methods staying in `internal/parser/` (parser-only helpers)
   - Expected import changes per package

7. **Risk assessment.** Identify any potential issues:
   - Circular import risks
   - Interface satisfaction that might break
   - Generated code (`grammar.go`) that references moved types
   - The new parser pipeline (`internal/parsers/`) — does it need updates?
   - External packages that import `internal/parser/` types

**Deliverable:** A `docs/REFACTOR_ANALYSIS.md` document with the complete
analysis. No code changes in this sprint.

- Run `go build` and `make test` (verify baseline still passes)
- Version: **0.75.0**

**Outcome:** Complete analysis delivered in `docs/REFACTOR_ANALYSIS.md`.
Key findings: 13 shared types must move to `internal/domain/`, 24+
parser-only types stay in `internal/parser/`. One constant
(`LastTurnCurrentLocationObscured`) must move with `Turn_t`. All six
`Report_t` merge methods must move with the type (they are called only
by the parser, but Go binds methods to their receiver). `FoundItem_t`
was reclassified from parser-only to shared because it is a field of
`Report_t` and a parameter of `tiles.MergeItem()`. `Scry_t` introduces
a dependency on `internal/unit_movement/` in the domain package. No
circular import risks were identified. The generated `grammar.go` will
need updates during alias removal (Sprint 82). The new parser pipeline
(`internal/parsers/`) is completely independent and unaffected.

---

## Sprint 76 — Create `internal/domain/` with Leaf Types

**Goal:** Create the `internal/domain/` package and move the simplest
shared types into it — types with no dependencies on other parser types.

**Context:** Start with "leaf" types that have no internal references to
other parser-defined types. These are the safest to move because they
won't create cascading changes. Each type is a small struct used as a
data carrier between the parser and render pipelines.

**Types to move (leaf types only):**
- `UnitId_t` — string type alias with helper methods (`InClan`,
  `IsFleet`, `Parent`, `String`). Used by every package in the pipeline.
- `Encounter_t` — struct with `TurnId string`, `UnitId UnitId_t`,
  `Friendly bool`. Embedded in `tiles.Tile_t` and `wxx.Features`.
- `Settlement_t` — struct with `TurnId string`, `Name string`. Embedded
  in `tiles.Tile_t` and `wxx.Features`.
- `Special_t` — struct with `TurnId string`, `Id string`, `Name string`.
  Embedded in `tiles.Tile_t` and `wxx.Features`.

**Scope:**
- Create `internal/domain/` directory with `doc.go` and `types.go`
- Move the four types above into `internal/domain/`
- Add type aliases in `internal/parser/types.go` pointing to
  `internal/domain/` to maintain backward compatibility during migration:
  ```go
  type UnitId_t = domain.UnitId_t
  type Encounter_t = domain.Encounter_t
  type Settlement_t = domain.Settlement_t
  type Special_t = domain.Special_t
  ```
- Verify all packages still compile with no import changes needed
  (the aliases make this transparent)
- Run `go build` and `make test`
- Version: **0.76.0**

**Outcome:** Created `internal/domain/` package with `doc.go` and
`types.go`. Moved four leaf types (`UnitId_t` with `InClan`, `IsFleet`,
`Parent`, `String` methods; `Encounter_t`; `Settlement_t` with `String`
method; `Special_t`) into `internal/domain/`. Added type aliases in
`internal/parser/types.go` (`type UnitId_t = domain.UnitId_t`, etc.) so
all downstream packages compile with no import changes. Build and all
tests pass.

---

## Sprint 77 — Move Report and Border Types to `internal/domain/`

**Goal:** Move the observation types that bridge parser output and tile
merging into `internal/domain/`.

**Context:** `Report_t` is the central data structure that carries parsed
observations into the tile-building phase. It is populated by the parser
and consumed by `tiles.Tile_t.MergeReports()`. `Border_t` and
`FarHorizon_t` are sub-types of `Report_t` also used during tile merging.

**Types to move:**
- `Border_t` — struct with `Direction`, `Edge`, `Terrain` fields.
  Used by `tiles.Tile_t.MergeBorder()` and `render.go` debug output.
- `FarHorizon_t` — struct with `Point`, `Terrain` fields. Used by
  `tiles.Tile_t.MergeFarHorizon()`.
- `Report_t` — struct with terrain, borders, encounters, settlements,
  resources, far horizons. Used by `tiles.Tile_t.MergeReports()`.
  Methods: `MergeBorders`, `MergeEncounters`, `mergeFarHorizons`,
  `mergeItems`, `MergeResources`, `MergeSettlements`.
- `FoundItem_t` — struct used by `Report_t.Items` field. Even though
  items are discarded during rendering, the type is needed for
  `Report_t` to compile. Move it alongside `Report_t`.

**Scope:**
- Move the four types and their methods to `internal/domain/`
- Add type aliases in `internal/parser/types.go` for backward
  compatibility
- Run `go build` and `make test`
- Version: **0.77.0**

---

## Sprint 78 — Move Movement Types to `internal/domain/`

**Goal:** Move `Move_t`, `Scout_t`, `Scry_t`, `Moves_t`, and `Turn_t`
into `internal/domain/`, completing the shared type migration.

**Context:** These are the core movement and turn types that flow from
the parser through the turn walker and into rendering. They are the most
complex types with the most cross-package references. `Turn_t` in
particular has methods (`TopoSortMoves`, `SortMovesByElement`) and
linked-list pointers (`Prev`, `Next`) that are used by `render.go`.

**Types to move:**
- `Move_t` — single move step. Used by `turns.Step()`, `render.go`
  debug output.
- `Scout_t` — scout movements. Used by `turns.Walk()`.
- `Scry_t` — scrying results. Used by `turns.Walk()`.
- `Moves_t` — unit's moves for a turn. Used by `render.go`,
  `turns.Walk()`.
- `Turn_t` — parsed turn. Used by `render.go`, `turns.Walk()`.
  Methods: `TopoSortMoves`, `SortMovesByElement`, `FromMayBeObscured`,
  `ToMayBeObscured`.

**Also move:**
- `LastTurnCurrentLocationObscured` constant (used by `Turn_t` methods)
- `ParseConfig` struct (if referenced outside parser; otherwise leave)

**Scope:**
- Move the five types, their methods, and the constant to
  `internal/domain/`
- Add type aliases in `internal/parser/types.go` for backward
  compatibility
- Run `go build` and `make test`
- Version: **0.78.0**

**Risk:** Medium. `Turn_t` is referenced heavily in `render.go` (~50
references). The type alias approach should make this transparent, but
careful testing is needed.

---

## Sprint 79 — Migrate `internal/tiles/` to Import `internal/domain/`

**Goal:** Update `internal/tiles/` to import types from
`internal/domain/` instead of `internal/parser/`.

**Context:** `internal/tiles/` is the heaviest consumer of parser types.
`Tile_t` embeds `[]*parser.Encounter_t`, `[]*parser.Settlement_t`,
`[]*parser.Special_t` as fields. Multiple methods accept parser types as
parameters (`MergeReports`, `MergeBorder`, `MergeEncounter`,
`MergeFarHorizon`, `MergeItem`, `MergeSettlement`). `Map_t.FetchTile()`
accepts `parser.UnitId_t`.

**Scope:**
- Replace `import "internal/parser"` with `import "internal/domain"`
  in all files under `internal/tiles/`
- Update all type references: `parser.Encounter_t` → `domain.Encounter_t`,
  `parser.Settlement_t` → `domain.Settlement_t`, etc.
- Verify no remaining references to `internal/parser` in the package
- Run `go build` and `make test`
- Version: **0.79.0**

---

## Sprint 80 — Migrate `internal/turns/` to Import `internal/domain/`

**Goal:** Update `internal/turns/` to import types from
`internal/domain/` instead of `internal/parser/`.

**Context:** `internal/turns/` uses parser types for the `Walk()` and
`Step()` function signatures, plus `Step_t` references `parser.UnitId_t`.
The `Walk()` function takes `[]*parser.Turn_t` as input and processes
`parser.Move_t` structs via `Step()`.

**Scope:**
- Replace `import "internal/parser"` with `import "internal/domain"`
  in all files under `internal/turns/`
- Update all type references in function signatures and local usage
- Verify no remaining references to `internal/parser` in the package
- Run `go build` and `make test`
- Version: **0.80.0**

---

## Sprint 81 — Migrate `internal/wxx/` and `actions/` to Import `internal/domain/`

**Goal:** Update the final two downstream packages to import from
`internal/domain/` instead of `internal/parser/`.

**Context:** `internal/wxx/` embeds `[]*parser.Encounter_t`,
`[]*parser.Settlement_t`, and `[]*parser.Special_t` in the `Features`
struct. `actions/map_world.go` references `parser.Special_t`,
`parser.UnitId_t`, and the `parser.Encounter_t.InClan()` method. Both
packages only use these as data carriers — they never call any parsing
logic.

**Scope:**
- Replace `import "internal/parser"` with `import "internal/domain"`
  in all files under `internal/wxx/` and `actions/`
- Update all type references
- Verify no remaining references to `internal/parser` in either package
- Run `go build` and `make test`
- Version: **0.81.0**

---

## Sprint 82 — Remove Type Aliases and Clean Up `internal/parser/`

**Goal:** Remove the backward-compatibility type aliases from
`internal/parser/types.go` and update `render.go` to import from
`internal/domain/` directly.

**Context:** After Sprints 79-81, all downstream packages import from
`internal/domain/`. The type aliases in `internal/parser/types.go` were
a transitional mechanism. Now `render.go` (the orchestrator) and
`internal/parser/` itself are the only remaining consumers of the
aliases. This sprint completes the decoupling.

**Scope:**
- Update `render.go` to import `internal/domain/` for shared types and
  `internal/parser/` only for `ParseInput()`
- Update `internal/parser/` to import `internal/domain/` for the shared
  types it populates during parsing
- Remove all type aliases from `internal/parser/types.go`
- Move parser-only types (`DirectionTerrain_t`, `Exhausted_t`,
  `ProhibitedFrom_t`, `NearHorizon_t`, `Neighbor_t`, `MissingEdge_t`,
  `Longhouse_t`, `Patrolled_t`, `FoundUnit_t`) into a separate file
  (e.g., `internal/parser/parse_types.go`) for clarity
- Verify `internal/parser/` no longer exports any types that belong in
  `internal/domain/`
- Run `go build` and `make test`
- Version: **0.82.0**

**Risk:** Medium. Removing aliases is a breaking change for any external
consumer of `internal/parser/` types. Since these are under `internal/`,
external access is already prohibited by Go's visibility rules.

---

## Sprint 83 — Verify Independence and Update Documentation

**Goal:** Validate that the parser and render packages are fully
independent, and update project documentation to reflect the new
architecture.

**Context:** After the refactoring, the dependency graph should be clean:
`internal/parser/` and the render packages (`internal/turns/`,
`internal/tiles/`, `internal/wxx/`, `actions/`) should share no direct
imports — they communicate only through `internal/domain/` types.

**Scope:**

1. **Dependency verification.** Confirm with `go list` that:
   - `internal/parser/` does not import `internal/turns/`, `internal/tiles/`,
     `internal/wxx/`, or `actions/`
   - `internal/turns/`, `internal/tiles/`, `internal/wxx/`, and `actions/`
     do not import `internal/parser/`
   - Both sides import `internal/domain/` for shared types

2. **Integration test.** Run the full pipeline end-to-end:
   - `go build` to confirm compilation
   - `make test` with race detector
   - `make golden` to confirm output hasn't changed

3. **Update CLAUDE.md:**
   - Add `internal/domain/` to the Key Packages table
   - Update the Architecture section to describe the decoupled pipeline
   - Update the "Two Parser Generations" section
   - Document the new dependency rules

4. **Update this sprint plan** with outcomes for Sprints 75-83.

- Version: **0.83.0**

---

## Phase 2 Sprint Summary

| Sprint | Version | Goal | Risk | Est. Files Changed |
|--------|---------|------|------|--------------------|
| 75 | 0.75.0 | Dependency analysis and migration plan | None | 0 (docs only) |
| 76 | 0.76.0 | Create `internal/domain/` with leaf types | Low | ~6 created/modified |
| 77 | 0.77.0 | Move report and border types to domain | Low | ~4 modified |
| 78 | 0.78.0 | Move movement types to domain | Medium | ~6 modified |
| 79 | 0.79.0 | Migrate `internal/tiles/` imports | Low | 2 modified |
| 80 | 0.80.0 | Migrate `internal/turns/` imports | Low | ~5 modified |
| 81 | 0.81.0 | Migrate `internal/wxx/` and `actions/` imports | Low | ~4 modified |
| 82 | 0.82.0 | Remove aliases, clean up parser | Medium | ~5 modified |
| 83 | 0.83.0 | Verify independence, update docs | None | ~2 modified |

**Risk levels:**
- **None** — Analysis or documentation only
- **Low** — Moving types with aliases preserving backward compatibility
- **Medium** — Removing aliases or moving heavily-referenced types

---

## Complete Sprint Summary

### Phase 1: Cleanup (Sprints 63-74) [ALL COMPLETED]

| Sprint | Version | Goal | Status |
|--------|---------|------|--------|
| 63 | 0.63.0 | Remove `internal/stores/office/` | COMPLETED |
| 64 | 0.64.0 | Remove `internal/hexes/` | COMPLETED |
| 65 | 0.65.0 | Remove `db` command + SQLite infrastructure | COMPLETED |
| 66 | 0.66.0 | Remove `load.go` + `parse` command | COMPLETED |
| 67 | 0.67.0 | Clean commented-out code | COMPLETED |
| 68 | 0.68.0 | Remove empty `internal/stores/` | COMPLETED |
| 69 | 0.69.0 | Audit/remove `winds` and `items` | COMPLETED |
| 70 | 0.70.0 | Evaluate `tndocx` dependency | COMPLETED |
| 71 | 0.71.0 | Clean stale indirect dependency | COMPLETED |
| 72 | 0.72.0 | Update CLAUDE.md | COMPLETED |
| 73 | 0.73.0 | Add package doc comments | COMPLETED |
| 74 | 0.74.0 | Update user documentation | COMPLETED |

### Phase 2: Parser/Render Refactoring (Sprints 75-83)

| Sprint | Version | Goal | Status |
|--------|---------|------|--------|
| 75 | 0.75.0 | Dependency analysis and migration plan | COMPLETED |
| 76 | 0.76.0 | Create `internal/domain/` with leaf types | COMPLETED |
| 77 | 0.77.0 | Move report and border types to domain | PLANNED |
| 78 | 0.78.0 | Move movement types to domain | PLANNED |
| 79 | 0.79.0 | Migrate `internal/tiles/` imports | PLANNED |
| 80 | 0.80.0 | Migrate `internal/turns/` imports | PLANNED |
| 81 | 0.81.0 | Migrate `internal/wxx/` and `actions/` imports | PLANNED |
| 82 | 0.82.0 | Remove aliases, clean up parser | PLANNED |
| 83 | 0.83.0 | Verify independence, update docs | PLANNED |
