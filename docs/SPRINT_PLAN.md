# OttoMap Sprint Plan

This document tracks all completed and planned sprints. Sprints 63-74 were
cleanup sprints (dead code removal, dependency cleanup, documentation).
Sprints 75-83 are refactoring sprints to separate the parser and render
pipelines into independent packages with shared domain types.

Current version: **0.84.12**

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

## Sprint 75 — Dependency Analysis and Migration Plan [COMPLETED]

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

## Sprint 76 — Create `internal/domain/` with Leaf Types [COMPLETED]

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

## Sprint 77 — Move Report and Border Types to `internal/domain/` [COMPLETED]

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

**Outcome:** Moved four types (`Border_t` with `String` method;
`FarHorizon_t`; `FoundItem_t` with `String` method; `Report_t` with
`MergeBorders`, `MergeEncounters`, `MergeFarHorizons`, `MergeItems`,
`MergeResources`, `MergeSettlements` methods) into `internal/domain/`.
The previously unexported methods `mergeFarHorizons` and `mergeItems`
were exported (`MergeFarHorizons`, `MergeItems`) since they now reside
in a separate package from their sole caller in `internal/parser/`.
Added type aliases in `internal/parser/types.go` (`type Border_t =
domain.Border_t`, etc.) so all downstream packages compile with no
import changes. Updated the call site in `parser.go` to use the new
exported name. Build and all tests pass (pre-existing golden test
failures in the new parser pipeline are unrelated).

---

## Sprint 78 — Move Movement Types to `internal/domain/` [COMPLETED]

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

**Outcome:** Moved five types (`Move_t`, `Scout_t`, `Scry_t`, `Moves_t`,
`Turn_t` with `TopoSortMoves`, `SortMovesByElement`, `FromMayBeObscured`,
`ToMayBeObscured` methods) and the `LastTurnCurrentLocationObscured`
constant into `internal/domain/`. `Scry_t` was previously defined in
`parser.go` rather than `types.go` and was relocated to `domain/types.go`
alongside the other movement types. `ParseConfig` was left in
`internal/parser/` as it is parser configuration, not a shared domain
type. Added type aliases in `internal/parser/types.go` (`type Turn_t =
domain.Turn_t`, `type Moves_t = domain.Moves_t`, `type Move_t =
domain.Move_t`, `type Scout_t = domain.Scout_t`, `type Scry_t =
domain.Scry_t`) and a constant alias (`const
LastTurnCurrentLocationObscured = domain.LastTurnCurrentLocationObscured`)
so all downstream packages compile with no import changes. The
`internal/domain/` package gained two new imports: `results` (for
`Move_t.Result`) and `unit_movement` (for `Scry_t.Type`). Build and all
tests pass (pre-existing golden test failures in the new parser pipeline
are unrelated).

---

## Sprint 79 — Migrate `internal/tiles/` to Import `internal/domain/` [COMPLETED]

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

**Outcome:** Replaced `internal/parser` import with `internal/domain` in
all files under `internal/tiles/`. Updated all type references:
`parser.Encounter_t` → `domain.Encounter_t`, `parser.Settlement_t` →
`domain.Settlement_t`, `parser.Special_t` → `domain.Special_t`,
`parser.UnitId_t` → `domain.UnitId_t`, `parser.Report_t` →
`domain.Report_t`, `parser.Border_t` → `domain.Border_t`,
`parser.FarHorizon_t` → `domain.FarHorizon_t`, `parser.FoundItem_t` →
`domain.FoundItem_t`. Verified no remaining references to
`internal/parser` in the package. Build passes.

---

## Sprint 80 — Migrate `internal/turns/` to Import `internal/domain/` [COMPLETED]

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

**Outcome:** Replaced `internal/parser` import with `internal/domain` in
two files (`walk.go` and `steps.go`). Updated all type references:
`parser.Turn_t` → `domain.Turn_t`, `parser.Special_t` →
`domain.Special_t`, `parser.UnitId_t` → `domain.UnitId_t`,
`parser.Moves_t` → `domain.Moves_t`, `parser.Scout_t` →
`domain.Scout_t`, `parser.Move_t` → `domain.Move_t`. Verified no
remaining references to `internal/parser` in the package. Build passes.

---

## Sprint 81 — Migrate `internal/wxx/` and `actions/` to Import `internal/domain/` [COMPLETED]

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

**Outcome:** Replaced `internal/parser` import with `internal/domain`
in all files under `internal/wxx/` and `actions/`. Updated all type
references: `parser.Encounter_t` → `domain.Encounter_t`,
`parser.Settlement_t` → `domain.Settlement_t`, `parser.Special_t` →
`domain.Special_t`, `parser.UnitId_t` → `domain.UnitId_t`. Verified no
remaining references to `internal/parser` in either package. Build
passes.

---

## Sprint 82 — Remove Type Aliases and Clean Up `internal/parser/` [COMPLETED]

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

**Outcome:** Removed all type aliases from `internal/parser/types.go`
and deleted the file entirely. Updated `render.go` to import
`internal/domain/` for shared types (`Turn_t`, `Moves_t`, `Move_t`,
`Special_t`, `UnitId_t`) and `internal/parser/` only for `ParseInput()`
and `ParseConfig`. Moved parser-only types (`DirectionTerrain_t`,
`Exhausted_t`, `ProhibitedFrom_t`, `NearHorizon_t`, `Neighbor_t`,
`MissingEdge_t`, `Longhouse_t`, `Patrolled_t`, `FoundUnit_t`) into
`internal/parser/parse_types.go`. Verified `internal/parser/` no longer
exports any domain types. Build and all tests pass.

---

## Sprint 83 — Verify Independence and Update Documentation [COMPLETED]

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

**Outcome:** Dependency verification confirmed full independence:
`internal/parser/` does not import any render-side packages
(`internal/turns/`, `internal/tiles/`, `internal/wxx/`, `actions/`), and
none of those packages import `internal/parser/`. Both sides import
`internal/domain/` for shared types. Integration tests pass: `go build`
succeeds, `make test` (with race detector) passes, `make golden` confirms
snapshots are current (golden snapshots were regenerated to reflect
lexer token kind changes from Sprint 69). Updated CLAUDE.md with
`internal/domain/` in Key Packages table, decoupled architecture
description, dependency rules, and updated parser generation notes.
Updated this sprint plan with outcomes for Sprints 79, 81, 82, and 83.

---

# Phase 3: Standalone Render Pipeline

## Sprint 84 — Standalone JSON Render Pipeline (Version 0.84.0)

**Goal:** Implement `cmd/render/main.go` as a standalone CLI that loads
parser-generated JSON Documents (`internal/tniif/schema.go`), merges
them with deterministic conflict resolution, and produces a
Worldographer WXX map file using the existing `internal/wxx` package.

**Context:** Phases 1-2 cleaned up the codebase and decoupled the parser
from the renderer via `internal/domain/` types. The parser CLI
(`cmd/parser`) now produces JSON Documents conforming to the TNIIF
schema. This sprint builds the other half: a render CLI that consumes
those Documents and generates maps — without any dependency on the
legacy `turns.Walk()` pipeline.

**Architecture:** All new code lives in `cmd/render/` (package `main`).
No existing OttoMap code is modified. The pipeline stages are:

```
Load JSON files → Validate → Flatten to observation events →
Sort (Turn, Clan-owning-last, Unit) → Merge (last-writer-wins) →
Convert to wxx.Hex → wxx.Create() → .wxx output
```

**Verification command:**
```sh
go run ./cmd/render testdata/0300.*.json --clan 0249 \
  --output testdata/0300.0904-05.0249.wxx
```

---

### Task 84.1 — Cobra CLI skeleton [COMPLETED]

**Goal:** Create the command-line interface matching `cmd/parser/main.go`
style.

**Steps:**
1. Replace the stub `cmd/render/main.go` with a Cobra root command:
   - `Use: "render [json-files...]"` (positional file args)
   - `Args: cobra.MinimumNArgs(1)`
2. Add required flags:
   - `--clan` (string, required; validate as 4-digit `"0xxx"` format)
3. Add optional flags:
   - `--output` (string, optional; path for the WXX output file)
4. Add logging flags consistent with parser:
   - `--debug`, `--quiet`, `--log-level`, `--log-source`
5. Add `version` subcommand.
6. Set the render CLI version to `0.84.0` using
   `github.com/maloquacious/semver` with `semver.Commit()`.

Note: If the **--output** flag is not provided, the command will load and validate the input and stop. It will not attempt to create a map file.

**Verification:**
- `go run ./cmd/render --help` prints usage with flags
- `go run ./cmd/render version` prints `0.84.0`
- `go build -o dist/local/render ./cmd/render` succeeds

---

### Task 84.2 — Load JSON Documents from positional args [COMPLETED]

**Goal:** Read and unmarshal all input files into `[]schema.Document`.

**Steps:**
1. Resolve each positional arg to an absolute path.
2. For each path:
   - `os.Stat` — must be a regular file
   - `os.ReadFile` — read contents
   - `json.Unmarshal` into `schema.Document`
3. Return `[]schema.Document` preserving argument order.
4. Log: `"loaded %d documents"` at debug level.

**Implementation:** Create `loadDocuments(paths []string) ([]schema.Document, error)`.

**Verification:**
- Unit test: create temp JSON file, load, assert `Schema`/`Game`/`Turn` fields parsed correctly.
- Manual: `go run ./cmd/render testdata/0300.*.json --clan 0249 --debug` logs document count.

---

### Task 84.3 — Validate all Documents before rendering [COMPLETED]

**Goal:** Collect all validation errors, log them, and stop before any
rendering begins.

**Validation rules (per document):**
- `doc.Schema == schema.Version` (`"tn-map.v0"`)
- Required fields: `doc.Game`, `doc.Turn`, `doc.Clan` must be non-empty
- `doc.Turn` matches `YYYY-MM` format (length 7, `-` at index 4,
  numeric year and month)
- For each `doc.Clans[i]`: `clan.ID` must be non-empty
- For each `unit`: `unit.ID` must be non-empty
- Coordinate validity: `coords.HexToMap(string(loc))` must succeed for:
  - `unit.EndingLocation`
  - `step.EndingLocation` (for each move step)
  - `obs.Location` (for each observation)
- Edge direction validity: `schema.Direction.Validate()` for each edge
- Compass bearing validity: `schema.Bearing.Validate()` for each
  compass point

**Cross-document validation:**
- All documents must have the same `Game` value

**Implementation:** Create `validateDocuments(docs []schema.Document)
[]error`. In `RunE`, if any errors, log each with file context and
return `fmt.Errorf("validation failed (%d errors)", len(errs))`.

**Verification:**
- Unit test: invalid schema version → error
- Unit test: invalid coordinate → error
- Unit test: mismatched game IDs → error
- Manual: corrupt one JSON field; confirm renderer logs error and exits
  non-zero.

---

### Task 84.4 — Define observation event type and flatten Documents [COMPLETED]

**Goal:** Convert all Documents into a flat list of observation events,
avoiding any `turns.Walk` logic.

**Define:**
```go
type obsEvent_t struct {
    Turn       schema.TurnID
    Clan       schema.ClanID
    Unit       string          // UnitID or ScoutID as string
    Loc        coords.Map
    Obs        *schema.Observation
    WasVisited bool
    WasScouted bool
}
```

**Flatten rules:**
- Walk `doc.Clans[] → clan.Units[] → unit.Moves[] → moves.Steps[]`
  - For each step with a non-nil `Observation`:
    - `Turn` = `doc.Turn`
    - `Clan` = containing `clan.ID` (not `doc.Clan`)
    - `Unit` = `string(unit.ID)`
    - `Loc` = `coords.HexToMap(string(obs.Location))`
    - `WasVisited` = `obs.WasVisited`
    - `WasScouted` = `obs.WasScouted`
- Walk `unit.Scouts[] → scout.Steps[]`
  - Same as above but `Unit` = `string(scout.ID)`
  - `WasScouted` = true (scouts always scout)

**Implementation:** Create `flattenEvents(docs []schema.Document)
[]obsEvent_t`.

**Verification:**
- Unit test: synthetic document with 2 clans, 3 units, and scout runs;
  assert event count and correct `Clan`/`Unit` keys.

---

### Task 84.5 — Sort events with owning-clan-last semantics [COMPLETED]

**Goal:** Deterministic sort order that ensures the owning clan's
observations are applied last.

**Sort comparator (stable sort):**
1. `Turn` ascending (lexicographic — `YYYY-MM` sorts correctly)
2. `Clan` ascending, except the owning clan (from `--clan`) sorts last:
   - `clanRank(clan) = 1 if clan == owningClan, else 0`
   - Compare rank first, then clan ID
3. `Unit` ascending (lexicographic)

**Implementation:** Create `sortEvents(events []obsEvent_t, owningClan
schema.ClanID)` — in-place stable sort.

**Verification:**
- Unit test: events from turns 0901-01 and 0902-01, clans 0249 and
  0331, units 1249e1 and 2331e2. With owning clan 0249:
  - 0901-01/0331 events come before 0901-01/0249 events
  - 0902-01/0331 events come before 0902-01/0249 events
  - Within same turn+clan, units sorted ascending

---

### Task 84.6 — Merge sorted events into per-tile state (last-writer-wins) [COMPLETED]

**Goal:** Fold the sorted event stream into a final map state where
conflicts are resolved by using the most recent value.

**Define per-tile accumulator:**
```go
type tileState_t struct {
    Loc        coords.Map
    Terrain    schema.Terrain
    Edges      map[schema.Direction]schema.Edge
    Resources  []schema.Resource
    Settlements []schema.Settlement
    Encounters []schema.Encounter
    CompassPoints []schema.CompassPoint
    WasVisited bool
    WasScouted bool
    Notes      []schema.Note
}
```

**Merge semantics (applied in sorted order = last-writer-wins):**
- `Terrain`: overwrite if `obs.Terrain != ""`
- `WasVisited` / `WasScouted`: logical OR across all events
- `Edges`: per-direction overwrite when new edge has non-empty fields
- `Resources`: overwrite entire slice if `obs.Resources != nil`
- `Settlements`: overwrite entire slice if `obs.Settlements != nil`
- `Encounters`: overwrite entire slice if `obs.Encounters != nil`
- `CompassPoints`: overwrite entire slice if `obs.CompassPoints != nil`
- `Notes`: append (never overwrite)

This ensures a newer observation can clear a field (empty slice) or
leave a prior value intact (nil/omitted field).

**Implementation:** Create `mergeTiles(events []obsEvent_t)
map[coords.Map]*tileState_t`.

**Verification:**
- Unit test: two events same tile, conflicting terrain → later wins
- Unit test: newer obs with `nil` settlements → older settlements
  preserved; newer obs with empty `[]` settlements → settlements cleared
- Unit test: edge merge overwrites only affected direction

---

### Task 84.7 — Convert merged tile state to `wxx.Hex` structs [COMPLETED]

**Goal:** Map from schema string types to the internal enum types used
by `internal/wxx`.

**String → enum mappings needed:**
- `schema.Terrain` → `terrain.Terrain_e` (invert `terrain.EnumToString`)
- `schema.Direction` → `direction.Direction_e` (invert
  `direction.EnumToString`)
- `schema.Feature` → `edges.Edge_e` (invert `edges.EnumToString`)
- `schema.Resource` → `resources.Resource_e` (invert
  `resources.EnumToString`)

**Build inverted maps at init time.** Validation (Task 3) ensures all
strings are known, so conversion should not fail at this stage.

**Convert each `tileState_t` → `wxx.Hex`:**
- `Location` and `RenderAt` (after offset — Task 8)
- `Terrain`: mapped enum
- `WasVisited` / `WasScouted`: direct copy
- `Features.Edges.*`: group by edge feature type (Canal, Ford, Pass,
  River, StoneRoad)
- `Features.Resources`: `[]resources.Resource_e`
- `Features.Settlements`: `[]*domain.Settlement_t`
- `Features.Encounters`: `[]*domain.Encounter_t` (mark friendly using
  `domain.UnitId_t.InClan(clan)`)

**Implementation:** Create `convertTileToHex(state *tileState_t,
renderOffset coords.Map, owningClan schema.ClanID) *wxx.Hex`.

**Verification:**
- Unit test: tile with "River" edge on "NE" → `hex.Features.Edges.River`
  contains `direction.NE`
- Unit test: resource string maps to expected enum

---

### Task 84.8 — Compute map bounds and render offset [COMPLETED]

**Goal:** Determine the bounding box and shift offset so the rendered
map is reasonably sized.

**Steps:**
1. Iterate all merged tiles to find:
   - `upperLeft = (min(Column), min(Row))`
   - `lowerRight = (max(Column), max(Row))`
2. Compute `renderOffset` (matching `actions/map_world.go` logic):
   - `borderWidth = 4`, `borderHeight = 4`
   - If `upperLeft.Column > borderWidth`:
     offset = `upperLeft.Column - borderWidth` (make even if odd)
   - If `upperLeft.Row > borderHeight`:
     offset = `upperLeft.Row - borderHeight`
3. Set `hex.RenderAt = coords.Map{Column: loc.Column - offset.Column,
   Row: loc.Row - offset.Row}` for each hex.

The TribeNet coordinates of the top-left corner of the bounding box must have an odd column and odd row (e.g., "AA 0101") or the map will not render correctly in Worldographer.

Note: The rule above is for the TribeNet coordinates. The `coords.Map` location uses 0-based columns and rows, so 
- If the loc.Row is **odd**, move the corner North one step.
- If the loc.Column is **odd**, move the corner Northwest one step.
If you apply the moves in this order, you do not need to worry about moving to an invalid location.

Examples (using TribeNet coordinates):
- "AA 0101" would not move since both row and column are odd.
- "AA 0102" would move to "AA 0101". It would move N because the row is even.
- "AA 0201" would move to "AA 0101". It would move NW because the column is even.
- "AA 0202" would move to "AA 0101". It would first move N because the row is even, then move NW because the column is even.

**Implementation:** Create `computeBoundsAndOffset(tiles
map[coords.Map]*tileState_t) (upperLeft, lowerRight, offset
coords.Map)`.

**Verification:**
- Unit test: known tile positions produce expected bounds and even
  column offset.
- Manual: debug log prints bounds and offsets.

---

### Task 84.9 — Merge SpecialHexes into tile features [COMPLETED]

**Goal:** Replicate the "special settlement labeling" mechanism.

**Steps:**
1. Collect `SpecialHexes` across all documents into
   `map[string]*domain.Special_t` keyed by `strings.ToLower(name)`.
2. For each final hex:
   - For each settlement name, if `strings.ToLower(name)` exists in the
     special map, move it from `Features.Settlements` to
     `Features.Special` (avoid duplicates).

**Implementation:** Create `applySpecialHexes(hexes []*wxx.Hex,
specials map[string]*domain.Special_t)`.

**Verification:**
- Unit test: settlement "Foo" + special hex "Foo" → one `Special` entry,
  zero `Settlements` entries for that name.

---

### Task 84.10 — Build WXX and write output file [COMPLETED]

**Goal:** Wire everything together and produce the `.wxx` file.

**Steps:**
1. Load default config: `gcfg := config.Default()`
2. Create WXX: `w, err := wxx.NewWXX(gcfg)`
3. For each final `wxx.Hex` (iterate in deterministic order, e.g.,
   sorted by `Location`):
   - Call `w.MergeHex(hex)` exactly once per tile
4. Determine `maxTurn` = lexicographic max of all `doc.Turn` values
5. Set up `wxx.RenderConfig`:
   - `Version` = render CLI version (0.84.0)
   - `Meta.IncludeMeta = true`
   - `Meta.IncludeOrigin = true`
6. Call `w.Create(outputPath, string(maxTurn), upperLeft, lowerRight,
   renderCfg, gcfg)`

**Verification:**
- End-to-end manual test:
  ```sh
  go run ./cmd/render testdata/0300.*.json --clan 0249 \
    --output testdata/0300.0904-05.0249.wxx
  ```
- Output file exists and has non-zero size
- `gunzip -t testdata/0300.0904-05.0249.wxx` succeeds (valid gzip)

---

### Task 84.11 — Unit tests for the render pipeline [COMPLETED]

**Goal:** Focused tests covering the pipeline logic. All tests live in
`cmd/render/` only.

**Test files:**
- `cmd/render/sort_test.go`:
  - `TestSortOrder_OwningClanLast`
  - `TestSortOrder_MultipleTurns`
- `cmd/render/merge_test.go`:
  - `TestMerge_LWW_Terrain`
  - `TestMerge_NilVsEmptySlices`
  - `TestMerge_EdgePerDirection`
- `cmd/render/validate_test.go`:
  - `TestValidate_InvalidCoords`
  - `TestValidate_MismatchedGame`
  - `TestValidate_InvalidSchema`
- `cmd/render/convert_test.go`:
  - `TestConvert_TerrainMapping`
  - `TestConvert_EdgeFeatureMapping`
- `cmd/render/integration_test.go`:
  - `TestIntegration_RenderSmallDoc` — create 1-2 tiny synthetic docs,
    run full pipeline to a temp `.wxx`, assert output is valid gzip

Note: Tests should write to an in-memory filesystem (like Afero) rather than creating files in `tmp` or scratch folders.

**Verification:**
- `go test ./cmd/render/...`

---

### Task 84.12 — Developer ergonomics and error quality [COMPLETED]

**Goal:** Ensure errors are actionable and debugging is easy.

**Steps:**
- Error messages must include: file name, unit ID, and field path
  (e.g., `"0300.0901-01.0249.json: unit 1249e1: step 3:
  invalid coordinate \"## 1316\""`)
- Add `--dump-merged` optional flag that writes merged tile state as
  JSON to stdout (defer if time is short)
- Log elapsed time at end of pipeline

**Verification:**
- Run with deliberately bad input; confirm error messages identify the
  exact file and field.

**Outcome:** Introduced `loadedDoc_t` struct pairing each document with
its source file name. Updated `loadDocuments`, `validateDocuments`,
`flattenEvents`, and `collectSpecialHexes` to thread file names through
the pipeline. Error messages now use the format
`"filename.json: unit 0987: step 3: observation: location \"## 1316\""`.
Added `--dump-merged` flag that serializes merged tile state as sorted
JSON to stdout for debugging. Added elapsed time logging at end of
pipeline (following parser CLI pattern). All 124 tests updated and
passing. Version: 0.84.12.

---

### Scope

- **In scope:** Tasks 84.1-84.12 above (CLI, load, validate, flatten, sort,
  merge, convert, bounds, specials, write, tests, errors)
- **Out of scope for Sprint 84:**
  - Player-specific config loading (`--config` flag)
  - Solo maps
  - Mentee maps
  - Movement cost labels
  - Visited/scouted labels (can be added in Sprint 85)
  - Blank map generation

### Risk Assessment

| Risk                                      | Mitigation                                     |
|-------------------------------------------|------------------------------------------------|
| Enum string mapping gaps (terrain/edges)  | Task 84.3 validation fails fast on unknowns    |
| `wxx.MergeHex` panics on terrain change   | Task 84.6 merges to one final state per tile   |
| Silent data loss from overwrite semantics | Task 84.6 tests cover nil vs empty slice cases |
| Coordinate out-of-bounds                  | Task 84.3 validates all coords before merge    |

### Version: **0.84.0**

---

## Phase 2 Sprint Summary

| Sprint | Version | Goal                                           | Risk   | Est. Files Changed  |
|--------|---------|------------------------------------------------|--------|---------------------|
| 75     | 0.75.0  | Dependency analysis and migration plan         | None   | 0 (docs only)       |
| 76     | 0.76.0  | Create `internal/domain/` with leaf types      | Low    | ~6 created/modified |
| 77     | 0.77.0  | Move report and border types to domain         | Low    | ~4 modified         |
| 78     | 0.78.0  | Move movement types to domain                  | Medium | ~6 modified         |
| 79     | 0.79.0  | Migrate `internal/tiles/` imports              | Low    | 2 modified          |
| 80     | 0.80.0  | Migrate `internal/turns/` imports              | Low    | ~5 modified         |
| 81     | 0.81.0  | Migrate `internal/wxx/` and `actions/` imports | Low    | ~4 modified         |
| 82     | 0.82.0  | Remove aliases, clean up parser                | Medium | ~5 modified         |
| 83     | 0.83.0  | Verify independence, update docs               | None   | ~2 modified         |

**Risk levels:**
- **None** — Analysis or documentation only
- **Low** — Moving types with aliases preserving backward compatibility
- **Medium** — Removing aliases or moving heavily-referenced types

---

## Complete Sprint Summary

### Phase 1: Cleanup (Sprints 63-74) [ALL COMPLETED]

| Sprint | Version | Goal                                        | Status    |
|--------|---------|---------------------------------------------|-----------|
| 63     | 0.63.0  | Remove `internal/stores/office/`            | COMPLETED |
| 64     | 0.64.0  | Remove `internal/hexes/`                    | COMPLETED |
| 65     | 0.65.0  | Remove `db` command + SQLite infrastructure | COMPLETED |
| 66     | 0.66.0  | Remove `load.go` + `parse` command          | COMPLETED |
| 67     | 0.67.0  | Clean commented-out code                    | COMPLETED |
| 68     | 0.68.0  | Remove empty `internal/stores/`             | COMPLETED |
| 69     | 0.69.0  | Audit/remove `winds` and `items`            | COMPLETED |
| 70     | 0.70.0  | Evaluate `tndocx` dependency                | COMPLETED |
| 71     | 0.71.0  | Clean stale indirect dependency             | COMPLETED |
| 72     | 0.72.0  | Update CLAUDE.md                            | COMPLETED |
| 73     | 0.73.0  | Add package doc comments                    | COMPLETED |
| 74     | 0.74.0  | Update user documentation                   | COMPLETED |

### Phase 2: Parser/Render Refactoring (Sprints 75-83)

| Sprint | Version | Goal                                           | Status    |
|--------|---------|------------------------------------------------|-----------|
| 75     | 0.75.0  | Dependency analysis and migration plan         | COMPLETED |
| 76     | 0.76.0  | Create `internal/domain/` with leaf types      | COMPLETED |
| 77     | 0.77.0  | Move report and border types to domain         | COMPLETED |
| 78     | 0.78.0  | Move movement types to domain                  | COMPLETED |
| 79     | 0.79.0  | Migrate `internal/tiles/` imports              | COMPLETED |
| 80     | 0.80.0  | Migrate `internal/turns/` imports              | COMPLETED |
| 81     | 0.81.0  | Migrate `internal/wxx/` and `actions/` imports | COMPLETED |
| 82     | 0.82.0  | Remove aliases, clean up parser                | COMPLETED |
| 83     | 0.83.0  | Verify independence, update docs               | COMPLETED |

### Phase 3: Standalone Render Pipeline (Sprints 84+)

| Sprint | Version | Goal                                           | Status    |
|--------|---------|------------------------------------------------|-----------|
| 84     | 0.84.0  | Standalone JSON render pipeline (`cmd/render`) | COMPLETED |
