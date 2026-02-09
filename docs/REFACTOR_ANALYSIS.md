# Parser/Render Refactoring Analysis

Sprint 75 deliverable: field-level dependency analysis for the
`internal/parser/` → downstream package migration.

---

## 1. Type Inventory

### 1.1 Parser-Only Types

These types are used **exclusively** within `internal/parser/` (by
`parser.go`, `grammar.go`, and `types.go`). They must stay in
`internal/parser/`.

| Type | File | Purpose |
|------|------|---------|
| `DirectionTerrain_t` | `types.go` | Successful step result (direction + terrain) |
| `Exhausted_t` | `types.go` | Failed step due to exhaustion |
| `ProhibitedFrom_t` | `types.go` | Failed step due to terrain prohibition |
| `NearHorizon_t` | `types.go` | Deck observation (inner ring neighbor) |
| `Neighbor_t` | `types.go` | Adjacent hex terrain observation |
| `MissingEdge_t` | `types.go` | "No River Adjacent to Hex" result |
| `Longhouse_t` | `types.go` | Longhouse observation (ignored) |
| `Patrolled_t` | `types.go` | Patrol results with found units |
| `FoundUnit_t` | `types.go` | Unit discovered during patrol/move |
| `BlockedByEdge_t` | `parser.go` | Failed step due to edge feature |
| `CarryCapacity_t` | `parser.go` | Carry capacity info (ignored) |
| `DidNotReturn_t` | `parser.go` | Scout did not return |
| `Edge_t` | `parser.go` | Edge feature observation |
| `FoundNothing_t` | `parser.go` | "Nothing of interest found" result |
| `InsufficientCapacity_t` | `parser.go` | Insufficient capacity (ignored) |
| `Location_t` | `parser.go` | Parsed unit location line |
| `Movement_t` | `parser.go` | Internal movement parsing container |
| `NoDirection_t` | `parser.go` | Grammar result (no direction) |
| `NoGroupsFound_t` | `parser.go` | Grammar result (no groups found) |
| `NoGroupsRaided_t` | `parser.go` | Grammar result (no groups raided) |
| `Step_t` (parser.go) | `parser.go` | Internal step parsing container |
| `TurnInfo_t` | grammar | Parsed "Current Turn" line data |
| `Date_t` | grammar | Date component of TurnInfo_t |

**Note:** There is a `Step_t` in `parser.go` (parser-only, internal
parsing container) and a separate `Step_t` in `internal/turns/walk.go`
(turns-only, defined locally). These are different types with the same
name in different packages — no conflict.

### 1.2 Shared Types

These types are produced by `internal/parser/` and consumed by one or
more downstream packages. They must move to `internal/domain/`.

| Type | Downstream Consumers |
|------|---------------------|
| `UnitId_t` | `tiles`, `turns`, `actions`, `render.go` |
| `Encounter_t` | `tiles`, `wxx`, `actions` |
| `Settlement_t` | `tiles`, `wxx`, `actions` |
| `Special_t` | `tiles`, `wxx`, `actions`, `render.go` |
| `Border_t` | `tiles`, `render.go` |
| `FarHorizon_t` | `tiles`, `render.go` |
| `Report_t` | `tiles`, `render.go` |
| `FoundItem_t` | `tiles` (via `Report_t.Items` field and `MergeItem` param) |
| `Move_t` | `turns`, `render.go` |
| `Scout_t` | `turns` |
| `Scry_t` | `turns` |
| `Moves_t` | `turns`, `render.go` |
| `Turn_t` | `turns`, `render.go` |

### 1.3 Render-Only Types

No types were found that are used only by render-side packages and not
by the parser. All downstream types originate in the parser.

### 1.4 Parser Configuration

`ParseConfig` is used in `render.go` (as `argsRender.parser`) and
passed to `parser.ParseInput()`. It configures parsing behavior and
should **stay in `internal/parser/`** — it is not a shared domain type.

---

## 2. Field-Level Usage Matrix

### 2.1 `UnitId_t`

**Definition:** `type UnitId_t string`

| Field/Method | `parser` | `tiles` | `turns` | `wxx` | `actions` | `render.go` |
|-------------|----------|---------|---------|-------|-----------|-------------|
| (value) | create | read | read/write | — | read | read |
| `InClan()` | — | — | — | — | **read** | — |
| `IsFleet()` | — | — | — | — | — | — |
| `Parent()` | — | — | **read** | — | — | — |
| `String()` | read | read | read | — | — | read |

**Notes:**
- `InClan()` is called only in `actions/map_world.go:118`.
- `Parent()` is called only in `turns/walk.go:73,80`.
- `IsFleet()` is not called outside the parser but is a public method
  that should move with the type.
- All four methods must move to `internal/domain/`.

### 2.2 `Encounter_t`

| Field | `parser` | `tiles` | `wxx` | `actions` | `render.go` |
|-------|----------|---------|-------|-----------|-------------|
| `TurnId` | write | read | — | — | — |
| `UnitId` | write | read | — | read | — |
| `Friendly` | — | — | — | write | — |

### 2.3 `Settlement_t`

| Field | `parser` | `tiles` | `wxx` | `actions` | `render.go` |
|-------|----------|---------|-------|-----------|-------------|
| `TurnId` | write | — | — | — | — |
| `Name` | write | read | — | read | — |

| Method | `parser` | `tiles` | `wxx` | `actions` | `render.go` |
|--------|----------|---------|-------|-----------|-------------|
| `String()` | — | — | — | — | — |

### 2.4 `Special_t`

| Field | `parser` | `tiles` | `wxx` | `actions` | `render.go` |
|-------|----------|---------|-------|-----------|-------------|
| `TurnId` | write | — | — | — | — |
| `Id` | write | read | — | read | — |
| `Name` | write | — | — | read | — |

### 2.5 `Border_t`

| Field | `parser` | `tiles` | `turns` | `render.go` |
|-------|----------|---------|---------|-------------|
| `Direction` | write | read | — | read |
| `Edge` | write | read | — | read |
| `Terrain` | write | read | — | read |

| Method | `parser` | `tiles` | `turns` | `render.go` |
|--------|----------|---------|---------|-------------|
| `String()` | — | — | — | — |

### 2.6 `FarHorizon_t`

| Field | `parser` | `tiles` | `render.go` |
|-------|----------|---------|-------------|
| `Point` | write | read | read |
| `Terrain` | write | read | read |

### 2.7 `Report_t`

| Field | `parser` | `tiles` | `render.go` |
|-------|----------|---------|-------------|
| `UnitId` | write | read | — |
| `Location` | write | — | — |
| `TurnId` | write | — | — |
| `ScoutedTurnId` | write | — | — |
| `Terrain` | write | read | read |
| `Borders` | write | read | read |
| `Encounters` | write | read | — |
| `Items` | write | read | — |
| `Resources` | write | read | — |
| `Settlements` | write | read | read |
| `FarHorizons` | write | read | read |
| `WasVisited` | write | — | — |
| `WasScouted` | write | — | — |

| Method | `parser` | `tiles` | `render.go` |
|--------|----------|---------|-------------|
| `MergeBorders()` | **call** | — | — |
| `MergeEncounters()` | **call** | — | — |
| `mergeFarHorizons()` | **call** | — | — |
| `mergeItems()` | **call** | — | — |
| `MergeResources()` | **call** | — | — |
| `MergeSettlements()` | **call** | — | — |

**Important:** All six merge methods are called **only** within
`internal/parser/parser.go` (during `parseMove` and
`parseMovementLine`). They operate on `Report_t` instances that the
parser is building. No downstream package calls these methods directly.

However, since `Report_t` moves to `internal/domain/`, its methods
must move with it (methods are tied to their receiver type in Go).
This is fine — the parser will import `domain` and call these methods
on `domain.Report_t` values.

### 2.8 `FoundItem_t`

| Field | `parser` | `tiles` |
|-------|----------|---------|
| `Quantity` | write | — |
| `Item` | write | — |

**Note:** `tiles.MergeItem()` accepts `*parser.FoundItem_t` as a
parameter but is a no-op. The type must still move because it is a
field of `Report_t.Items`. It depends on `items.Item_e` from the
`internal/items/` package.

### 2.9 `Move_t`

| Field | `parser` | `turns` | `render.go` |
|-------|----------|---------|-------------|
| `UnitId` | write | read | — |
| `Advance` | write | read | read |
| `Follows` | write | read | read |
| `GoesTo` | write | read | read |
| `Still` | write | read | read |
| `Result` | write | read | read |
| `Report` | write | read | read |
| `LineNo` | write | read | read |
| `StepNo` | write | read | read |
| `Line` | write | — | — |
| `TurnId` | write | — | read |
| `CurrentHex` | write | — | read |
| `FromCoordinates` | write | — | — |
| `ToCoordinates` | write | — | — |
| `Location` | write | write | — |
| `Debug.FleetMoves` | write | read | — |
| `Debug.PriorMove` | write | — | — |
| `Debug.NextMove` | write | — | — |

### 2.10 `Scout_t`

| Field | `parser` | `turns` |
|-------|----------|---------|
| `No` | write | read |
| `TurnId` | write | read |
| `Moves` | write | read |
| `LineNo` | write | — |
| `Line` | write | — |

### 2.11 `Scry_t`

| Field | `parser` | `turns` |
|-------|----------|---------|
| `UnitId` | write | — |
| `Type` | write | — |
| `Origin` | write | — |
| `Coordinates` | write | read |
| `Location` | write | read |
| `Text` | write | — |
| `Moves` | write | read |
| `Scouts` | write | read |

**Note:** `Scry_t` depends on `unit_movement.Type_e` from
`internal/unit_movement/`. This is an existing shared enum package,
no change needed.

### 2.12 `Moves_t`

| Field | `parser` | `turns` | `render.go` |
|-------|----------|---------|-------------|
| `TurnId` | write | read | read |
| `UnitId` | write | read | read |
| `Moves` | write | read | read |
| `Follows` | write | read | — |
| `GoesTo` | write | read | — |
| `Scries` | write | read | — |
| `Scouts` | write | read | — |
| `FromHex` | write | read | read |
| `ToHex` | write | read | read |
| `Coordinates` | write | read/write | — |
| `Location` | write | read/write | read |

### 2.13 `Turn_t`

| Field | `parser` | `turns` | `render.go` |
|-------|----------|---------|-------------|
| `Id` | write | read | read |
| `Year` | write | — | read |
| `Month` | write | — | read |
| `UnitMoves` | write | read | read/write |
| `SortedMoves` | write | read | read/write |
| `MovesSortedByElement` | write | read | read/write |
| `SpecialNames` | write | — | read |
| `Next` | — | — | read/write |
| `Prev` | — | — | read/write |

| Method | `parser` | `turns` | `render.go` |
|--------|----------|---------|-------------|
| `FromMayBeObscured()` | — | — | — |
| `ToMayBeObscured()` | — | — | — |
| `TopoSortMoves()` | — | **call** | — |
| `SortMovesByElement()` | — | — | **call** |

**Notes:**
- `FromMayBeObscured()` is defined but not called anywhere in the
  codebase. It returns a constant `true`. Keep it on the type for
  future use.
- `ToMayBeObscured()` is defined but not called anywhere currently.
  It references `LastTurnCurrentLocationObscured`, so the constant
  must move with the type.
- `TopoSortMoves()` is called in `turns/walk.go:90`.
- `SortMovesByElement()` is called in `render.go:353`.

---

## 3. Method Audit

### 3.1 Methods That Must Move to `internal/domain/`

These methods are on shared types and are called (or could be called)
by downstream packages:

| Type | Method | Called By |
|------|--------|----------|
| `UnitId_t` | `InClan(clan)` | `actions/map_world.go` |
| `UnitId_t` | `IsFleet()` | (public API, no current callers outside parser) |
| `UnitId_t` | `Parent()` | `turns/walk.go` |
| `UnitId_t` | `String()` | multiple packages (implicit via `%s`/`%q`) |
| `Turn_t` | `TopoSortMoves()` | `turns/walk.go` |
| `Turn_t` | `SortMovesByElement()` | `render.go` |
| `Turn_t` | `FromMayBeObscured()` | (no current callers) |
| `Turn_t` | `ToMayBeObscured()` | (no current callers) |
| `Report_t` | `MergeBorders()` | `internal/parser/` only |
| `Report_t` | `MergeEncounters()` | `internal/parser/` only |
| `Report_t` | `mergeFarHorizons()` | `internal/parser/` only |
| `Report_t` | `mergeItems()` | `internal/parser/` only |
| `Report_t` | `MergeResources()` | `internal/parser/` only |
| `Report_t` | `MergeSettlements()` | `internal/parser/` only |
| `Settlement_t` | `String()` | (no current callers outside parser) |
| `Border_t` | `String()` | (no current callers) |
| `FoundItem_t` | `String()` | (no current callers) |

**Key insight:** The `Report_t` merge methods are called only by the
parser, but they must move with `Report_t` because methods are bound
to their receiver type in Go. This is safe — the parser will call
them via `domain.Report_t` after the migration.

### 3.2 Methods That Stay in `internal/parser/`

These methods are on parser-only types:

| Type | Method |
|------|--------|
| `DirectionTerrain_t` | `String()` |
| `Exhausted_t` | `String()` |
| `ProhibitedFrom_t` | `String()` |
| `Neighbor_t` | `String()` |
| `BlockedByEdge_t` | `String()` |
| `DidNotReturn_t` | `String()` |
| `Edge_t` | `String()` |
| `FoundNothing_t` | `String()` |

---

## 4. Constant Audit

| Constant | Defined In | Used By | Action |
|----------|-----------|---------|--------|
| `LastTurnCurrentLocationObscured` | `parser.go:41` | `types.go:46` (`Turn_t.ToMayBeObscured()`), `parser.go:101` | **Move to `internal/domain/`** |

The constant is used by `Turn_t.ToMayBeObscured()`, which is a method
on a shared type. It is also used within `parser.go:101` (during
`ParseInput`), but the parser will import `domain` anyway.

---

## 5. Enum Package Review

These packages are already shared and **require no changes**:

| Package | Used By | Status |
|---------|---------|--------|
| `internal/direction/` | parser, tiles, turns, wxx, actions | Stable |
| `internal/terrain/` | parser, tiles, wxx, render.go | Stable |
| `internal/edges/` | parser, tiles, actions, render.go | Stable |
| `internal/results/` | parser, turns, render.go | Stable |
| `internal/resources/` | parser, tiles, wxx | Stable |
| `internal/coords/` | parser, tiles, turns, wxx, actions, render.go | Stable |
| `internal/compass/` | parser, tiles | Stable |
| `internal/items/` | parser (via `FoundItem_t`) | Stable (external dep) |
| `internal/winds/` | parser (grammar only) | Stable (external dep) |
| `internal/unit_movement/` | parser (via `Scry_t.Type`) | Stable |

**Note:** `FoundItem_t` references `items.Item_e`. When `FoundItem_t`
moves to `internal/domain/`, `domain` will need to import
`internal/items/`. This is safe — `items` has no dependencies on
`parser` or `domain`.

Similarly, `Scry_t` references `unit_movement.Type_e`. When `Scry_t`
moves, `domain` will import `internal/unit_movement/`.

---

## 6. Migration Order

### 6.1 Types Moving to `internal/domain/`

Ordered by dependency (leaf types first):

**Sprint 76 — Leaf types (no parser-type dependencies):**
- `UnitId_t` + methods (`InClan`, `IsFleet`, `Parent`, `String`)
- `Encounter_t`
- `Settlement_t` + `String()` method
- `Special_t`

**Sprint 77 — Report types (depend on leaf types):**
- `Border_t` + `String()` method
- `FarHorizon_t`
- `FoundItem_t` + `String()` method
- `Report_t` + all six merge methods

**Sprint 78 — Movement types (depend on leaf + report types):**
- `Move_t`
- `Scout_t`
- `Scry_t`
- `Moves_t`
- `Turn_t` + methods (`TopoSortMoves`, `SortMovesByElement`,
  `FromMayBeObscured`, `ToMayBeObscured`)
- `LastTurnCurrentLocationObscured` constant

### 6.2 Types Staying in `internal/parser/`

All parser-only types listed in section 1.1, plus:
- `ParseConfig` (parser configuration, not a domain type)

### 6.3 Expected Import Changes Per Package

| Package | Current Import | New Import | Sprint |
|---------|---------------|------------|--------|
| `internal/tiles/` | `internal/parser` | `internal/domain` | 79 |
| `internal/turns/` | `internal/parser` | `internal/domain` | 80 |
| `internal/wxx/` | `internal/parser` | `internal/domain` | 81 |
| `actions/` | `internal/parser` | `internal/domain` | 81 |
| `render.go` | `internal/parser` | `internal/domain` + `internal/parser` | 82 |
| `internal/parser/` | (self) | `internal/domain` | 82 |

**`render.go`** will continue to import `internal/parser/` for
`ParseInput()` and `ParseConfig`. It will additionally import
`internal/domain/` for all shared types.

### 6.4 `internal/domain/` Import List

The new package will import:

```
internal/compass
internal/coords
internal/direction
internal/edges
internal/items
internal/resources
internal/results
internal/terrain
internal/unit_movement
```

Plus standard library: `fmt`, `sort`, `strings`.

---

## 7. Risk Assessment

### 7.1 Circular Import Risks

**Risk: None.**

`internal/domain/` imports only leaf enum/coord packages. It does not
import `internal/parser/`, `internal/tiles/`, `internal/turns/`,
`internal/wxx/`, or `actions/`. The dependency graph remains acyclic:

```
internal/domain/  →  enum/coord packages (direction, terrain, etc.)
internal/parser/  →  internal/domain/ + enum/coord packages
internal/tiles/   →  internal/domain/ + enum/coord packages
internal/turns/   →  internal/domain/ + internal/tiles/
internal/wxx/     →  internal/domain/ + enum/coord packages
actions/          →  internal/domain/ + internal/tiles/ + internal/wxx/
render.go         →  all of the above
```

### 7.2 Interface Satisfaction

**Risk: None.**

No parser types implement interfaces defined in downstream packages.
The only interface reference is `parser.Node_i` in
`internal/walkers/tree.go`, but that refers to
`internal/reports/parser.Node_i` (the **new** parser pipeline), not
`internal/parser/` (the legacy parser). No conflict.

### 7.3 Generated Code (`grammar.go`)

**Risk: Low, mitigated by type aliases.**

`grammar.go` is generated from `grammar.peg` and references types
from `types.go` (e.g., `Encounter_t`, `Settlement_t`, `Border_t`,
`FarHorizon_t`, `FoundItem_t`, `Move_t`, `Report_t`, etc.).

During the alias phase (Sprints 76-78), `grammar.go` will continue to
work because the aliases make `parser.Encounter_t` equivalent to
`domain.Encounter_t`.

In Sprint 82 (alias removal), `grammar.go` will need to reference
`domain` types directly. This requires either:
1. Updating `grammar.peg` to use `domain.` prefixes and regenerating
   `grammar.go`, **or**
2. Adding a `domain` import to `grammar.go` and doing a sed-style
   replacement on the generated file.

Option 1 is preferred. The grammar rules that construct types
(e.g., `&Encounter_t{...}`) will become `&domain.Encounter_t{...}`.

### 7.4 New Parser Pipeline (`internal/parsers/`)

**Risk: None.**

The new parser pipeline (`internal/parsers/lexers/`, `cst/`, `ast/`)
does **not** import `internal/parser/`. It is a completely independent
implementation. No updates are needed.

### 7.5 New Orchestration Layer

**Risk: None.**

`internal/reports/`, `internal/runners/`, and `internal/walkers/`
do not import `internal/parser/` (the legacy parser). The walkers
package imports `internal/reports/parser` (the new parser), which is
a different package entirely.

### 7.6 External Packages

**Risk: None.**

All types being moved are under `internal/`, which Go's visibility
rules prevent external packages from importing. No external consumers
can be affected.

### 7.7 The `FoundItem_t` / `items` Dependency

**Risk: Low.**

`FoundItem_t` references `items.Item_e`. Moving `FoundItem_t` to
`internal/domain/` means `domain` must import `internal/items/`.
This is safe — `items` is a leaf package with no dependencies on
parser or render code. The `items` package is retained for external
package compatibility (per Sprint 69 findings).

### 7.8 The `Scry_t` / `unit_movement` Dependency

**Risk: Low.**

`Scry_t` references `unit_movement.Type_e`. Moving `Scry_t` to
`internal/domain/` means `domain` must import
`internal/unit_movement/`. This is safe — `unit_movement` is a leaf
package.

---

## 8. Summary

The migration is well-bounded. Thirteen types, six merge methods,
four `UnitId_t` methods, four `Turn_t` methods, and one constant
will move from `internal/parser/` to `internal/domain/`.

The type alias strategy (Sprints 76-78) ensures backward compatibility
during the transition. Each downstream package can be migrated
independently (Sprints 79-81), and aliases can be removed only after
all consumers have been updated (Sprint 82).

No circular imports are possible. The generated `grammar.go` is the
only file requiring special attention during alias removal.
