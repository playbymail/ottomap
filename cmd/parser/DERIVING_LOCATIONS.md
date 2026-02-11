# Deriving Locations

This document describes how to populate the missing `Location` and
`EndingLocation` fields in the schema `Document` after the parser has
finished building the initial structure.

## Why we need this

The parser populates `Unit.EndingLocation` from the turn report's
"Current Hex" line, but it does not fill in `MoveStep.EndingLocation`,
`Observation.Location`, or `CompassPoint.Location`. It also does not set
`ScoutRun.StartingLocation`. This plan describes how to derive all of
these from the one value we trust: `Unit.EndingLocation`.

> **Important:** We cannot trust the starting location from the parser
> because the TribeNet turn report sometimes provides wrong values. We
> can always trust the ending location — that is why unit walks start
> from the end and work backwards.

---

## Shared Primitives

All location math uses `internal/coords`. Convert between schema and
internal representations with:

```go
func parseCoord(c schema.Coordinates) (coords.Map, error) {
    return coords.HexToMap(string(c))
}

func formatCoord(m coords.Map) schema.Coordinates {
    return schema.Coordinates(m.ToHex())
}
```

### Schema ↔ Internal Direction Mapping

Convert `schema.Direction` to `direction.Direction_e`:

| Schema | Internal             |
|--------|----------------------|
| `"N"`  | `direction.North`    |
| `"NE"` | `direction.NorthEast`|
| `"SE"` | `direction.SouthEast`|
| `"S"`  | `direction.South`    |
| `"SW"` | `direction.SouthWest`|
| `"NW"` | `direction.NorthWest`|

### Direction Reversal (for backward walks)

| Forward | Reverse |
|---------|---------|
| N       | S       |
| NE      | SW      |
| SE      | NW      |
| S       | N       |
| SW      | NE      |
| NW      | SE      |

```go
func opposite(d direction.Direction_e) direction.Direction_e {
    switch d {
    case direction.North:     return direction.South
    case direction.NorthEast: return direction.SouthWest
    case direction.SouthEast: return direction.NorthWest
    case direction.South:     return direction.North
    case direction.SouthWest: return direction.NorthEast
    case direction.NorthWest: return direction.SouthEast
    default: panic(fmt.Sprintf("invalid direction %d", d))
    }
}
```

### Column Parity Convention

The diagram in `docs/big_map_crossing.png` uses **1-based grid columns**
(1–30), so grid column 30 is **even** in the diagram. The `internal/coords`
package converts grid columns to **0-based absolute columns**
(`bigMapColumn*30 + gridColumn - 1`), so grid column 30 becomes absolute
column 29, which is **odd**. This means the odd-column vectors in
`vectors.go` apply to grid column 30, not the even-column vectors.

When hand-checking border crossings against the diagram, remember:

| Grid Column (1-based) | Absolute Column (0-based) | Diagram Parity | Code Parity |
|------------------------|---------------------------|----------------|-------------|
| 1                      | 0                         | odd            | even        |
| 2                      | 1                         | even           | odd         |
| 29                     | 28                        | odd            | even        |
| 30                     | 29                        | even           | odd         |

The code is correct and internally consistent — just shifted by one
relative to the diagram's labels.

### Map Bounds Validation

The valid coordinate space is columns 0–779 and rows 0–545 (26 grids
of 30 columns and 26 grids of 21 rows). If `Map.Add()` produces
out-of-range values, add a `Note{Kind:"warn"}` and leave the
coordinate empty rather than emitting garbage.

> **Known hazard (confirmed during Task 2):** `coords.Map.Add()` does
> no bounds checking — it happily produces negative or oversized values.
> Worse, `Map.ToHex()` silently formats them as garbage because Go's
> integer division truncates toward zero. Example: absolute row −1
> gives `BigMapRow = 0`, `GridRow = 0`, producing `"AA xx00"` — a
> syntactically plausible but invalid coordinate with no error.
>
> This is harmless for Tasks 1–3 (game data stays well inside the map),
> but **Task 4 must validate** after each two-step `Move()` because
> compass points from edge hexes will reach outside the valid space.
> A helper like `func validMap(m coords.Map) bool` that checks
> `0 <= m.Column && m.Column < 780 && 0 <= m.Row && m.Row < 546`
> should gate every `formatCoord()` call in Task 4.

---

## Task 1: Unit MoveStep.EndingLocation (backward walk)

### Algorithm

For each `Unit` in each `Clan`:

1. Parse `unit.EndingLocation` to `coords.Map`. If invalid, skip this
   unit (add a warning note).
2. Set `cur` to the parsed map coordinate.
3. Iterate `unit.Moves[i].Steps` from **last to first** (backward):
   a. Set `step.EndingLocation = formatCoord(cur)`.
   b. Determine whether `cur` changes for the *prior* step:
      - If `step.Result != "succeeded"`: `cur` does not change.
      - If `step.Intent == "still"`: `cur` does not change.
      - If `step.Intent == "advance"`: `cur = cur.Add(opposite(step.Advance))`.
      - If `step.Intent == "follows"` or `"goesTo"`: report an error
        and panic — these should never have a prior step.

### Worked Example

Unit ending at `ID 1610`, steps: `[NE✓, NE✓, S✓, S✓, S✓, SW✓, Still]`

Walk backward from `ID 1610`:
```
Step 6 (Still)  → EndingLocation = ID 1610, cur stays ID 1610
Step 5 (SW✓)    → EndingLocation = ID 1610, cur = Add(opposite(SW)=NE) → ID 1709
Step 4 (S✓)     → EndingLocation = ID 1709, cur = Add(opposite(S)=N) → ID 1708
Step 3 (S✓)     → EndingLocation = ID 1708, cur = Add(opposite(S)=N) → ID 1707
Step 2 (S✓)     → EndingLocation = ID 1707, cur = Add(opposite(S)=N) → ID 1706
Step 1 (NE✓)    → EndingLocation = ID 1706, cur = Add(opposite(NE)=SW) → ID 1607
Step 0 (NE✓)    → EndingLocation = ID 1607, cur = Add(opposite(NE)=SW) → ID 1507
```

After the walk, `cur` holds the derived starting location (`ID 1507`).
We do not store this in the schema because we do not trust it, but it
can be used for validation during development.

### Unit Tests

Test the backward walk with table-driven tests. Each test case
provides a `Unit.EndingLocation`, a slice of `MoveStep`s (with Intent,
Advance, and Result populated), and the expected `EndingLocation` for
each step after derivation.

**Required test cases:**

1. **Simple advance chain** — 3 steps all advancing N, all succeeded.
2. **Mixed results** — advance NE (succeeded), advance SE (failed),
   still. Verify the failed step gets the same location as the prior
   step.
3. **Single still step** — unit didn't move; EndingLocation matches
   Unit.EndingLocation.
4. **East border crossing** — unit at `AA 3001`, advance NE succeeded.
   The prior hex should be in grid `AA` column 29 (i.e., `AA 2901` or
   `AA 2902` depending on row parity). Verify the grid letter changes
   correctly.
5. **West border crossing** — unit at `AB 0101`, advance NW succeeded.
   The prior hex should be back in grid `AA`.
6. **North border crossing** — unit at `BA 1501`, advance N succeeded.
   The prior hex should be in grid `AA` row 21.
7. **South border crossing** — unit at `AA 1521`, advance S succeeded.
   The prior hex should be in grid `BA` row 1.
8. **Corner crossing (NE from bottom-right of grid)** — unit at
   `AA 3021`, advance NE succeeded. Verify the prior hex (SW of there)
   is in the correct grid.
9. **Corner crossing (SW from top-left of grid)** — unit at `BB 0101`,
   advance SW succeeded. Verify the prior hex (NE of there) is in the
   correct grid.
10. **Vanished result** — advance SE (vanished). Location should not
    change (treat as non-movement).

---

## Task 2: Scout MoveStep.EndingLocation (forward walk)

### Algorithm

For each `Unit` in each `Clan`, for each `ScoutRun`:

1. Set `scout.StartingLocation = unit.EndingLocation`.
2. Parse `scout.StartingLocation` to `coords.Map`. If invalid, skip
   (add a warning note).
3. Set `cur` to the parsed map coordinate.
4. Iterate `scout.Steps` from **first to last** (forward):
   a. Determine whether `cur` changes:
      - If `step.Result != "succeeded"`: `cur` does not change.
      - If `step.Intent == "still"`: `cur` does not change.
      - If `step.Intent == "advance"`: `cur = cur.Add(step.Advance)`.
      - If `step.Intent == "follows"` or `"goesTo"`: report an error
        and panic — scouts are not allowed these moves.
   b. Set `step.EndingLocation = formatCoord(cur)`.

### Worked Example

Scout starting at `ID 1610`, steps: `[N✓, N✓, NW✓, NW(failed)]`

Walk forward from `ID 1610`:
```
Step 0 (N✓)      → cur = Add(N)  → ID 1609, EndingLocation = ID 1609
Step 1 (N✓)      → cur = Add(N)  → ID 1608, EndingLocation = ID 1608
Step 2 (NW✓)     → cur = Add(NW) → ID 1508, EndingLocation = ID 1508
Step 3 (NW fail) → cur unchanged → ID 1508, EndingLocation = ID 1508
```

### Unit Tests

**Required test cases:**

1. **Simple advance chain** — 3 steps all advancing S, all succeeded.
2. **Failed step mid-chain** — advance NE (succeeded), advance NE
   (failed), advance N (succeeded). Verify that the failed step shares
   its location with the prior step, and the subsequent step moves
   from that same location.
3. **Single still step** — scout didn't move; EndingLocation matches
   StartingLocation.
4. **East border crossing** — scout starts at `AA 3001`, advances NE
   then SE. Verify grid transitions.
5. **West border crossing** — scout starts at `AB 0101`, advances NW
   then SW. Verify grid transitions.
6. **North border crossing** — scout starts at `BA 1501`, advances N.
   Should end in grid `AA`.
7. **South border crossing** — scout starts at `AA 1521`, advances S.
   Should end in grid `BA`.
8. **Corner crossing** — scout starts at `AA 3021`, advances SE.
   Verify the resulting grid is correct.
9. **Vanished result** — advance S (vanished). Location should not
    change.

---

## Task 3: Observation.Location

### Algorithm

After Task 1 and Task 2 have populated all `MoveStep.EndingLocation`
values:

For every `MoveStep` that has a non-nil `Observation`:
- Set `observation.Location = step.EndingLocation`.

The observation is always made at the hex the unit/scout is in after
completing the step.

### Unit Tests

1. **Observation present on succeeded step** — verify
   `observation.Location` matches `step.EndingLocation`.
2. **Observation present on failed step** — verify the location matches
   (the unit stayed where it was).
3. **No observation** — verify nil observations are left alone (no
   panic, no side effects).
4. **Multiple observations across steps** — verify each observation
   gets the correct step's location, not a shared/stale value.

---

## Task 4: CompassPoint.Location

### Algorithm

After Task 3 has populated `Observation.Location`:

For each `Observation` that has `CompassPoints`:
1. Parse `observation.Location` to `coords.Map`.
2. For each `CompassPoint`:
   - Look up the bearing in the bearing-to-directions map (below).
   - Compute `loc = m.Move(d1, d2)`.
   - Set `compassPoint.Location = formatCoord(loc)`.

### Bearing → Direction Pairs

These are two sequential hex moves from the observation hex. This
mapping is confirmed by the existing implementation in
`internal/tiles/tile.go` (`MergeFarHorizon`).

| Bearing | Move 1    | Move 2    |
|---------|-----------|-----------|
| N       | North     | North     |
| NNE     | North     | NorthEast |
| NE      | NorthEast | NorthEast |
| E       | NorthEast | SouthEast |
| SE      | SouthEast | SouthEast |
| SSE     | South     | SouthEast |
| S       | South     | South     |
| SSW     | South     | SouthWest |
| SW      | SouthWest | SouthWest |
| W       | SouthWest | NorthWest |
| NW      | NorthWest | NorthWest |
| NNW     | North     | NorthWest |

### Unit Tests

**Required test cases:**

1. **All 12 bearings from a central hex** — pick a hex well inside a
   grid (e.g., `MM 1510`) and verify all 12 compass point locations
   against hand-calculated expected values.
2. **East bearing from even column** — verify parity handling.
3. **East bearing from odd column** — verify parity handling.
4. **West bearing from even column** — same.
5. **West bearing from odd column** — same.
6. **North bearing at grid top edge** — e.g., observation at
   `BA 1502`, bearing N → N+N should cross into grid `AA`.
7. **South bearing at grid bottom edge** — e.g., observation at
   `AA 1520`, bearing S → S+S should cross into grid `BA`.
8. **East bearing at grid right edge** — e.g., observation at
   `AA 2910`, bearing E → NE+SE should cross into grid `AB`.
9. **Corner case** — observation at `AA 3020`, bearing SE → SE+SE
   should cross both column and row grid boundaries.

---

## Execution Order

The tasks must be executed in this order:

1. **Task 1** — Unit backward walk (populates `MoveStep.EndingLocation`
   for units).
2. **Task 2** — Scout forward walk (sets `ScoutRun.StartingLocation`
   and populates `MoveStep.EndingLocation` for scouts).
3. **Task 3** — Observation location (copies `step.EndingLocation` to
   `observation.Location`).
4. **Task 4** — Compass point location (derives 2-hex-ring locations
   from `observation.Location` and bearing).

All four tasks should be called in a single `DeriveLocations(doc *schema.Document)`
function in `cmd/parser/main.go` (or a dedicated helper file) that is
invoked after the document is fully built but before JSON marshaling.

---

## Implementation Notes

- The `coords.Map.Add()` and `coords.Map.Move()` methods already handle
  even/odd column parity for flat-top hexes. All movement math should
  go through these methods.
- Grid boundary crossings are handled automatically by the Map→Grid
  conversion (`Map.ToGrid()` / `Map.ToHex()`), since `Map` uses
  absolute column/row coordinates.
- The `follows` and `goesTo` intents should panic if encountered in
  contexts where they have prior steps (units) or appear at all
  (scouts). These are terminal movement types that do not have
  preceding or subsequent advance steps.
- `vanished` results should be treated the same as `failed` — no
  location change.
