# OttoMap --- Unit Walk (Revised Requirements)

## Goal

Rebuild the unit-walk to process **one turn at a time**, generate tile trails per unit, and emit consistent summaries for later consolidation and audit.

------------------------------------------------------------------------

## Inputs & Outputs

### Input

-   `Turn_t` (single turn), containing:
    -   `Units []Unit_t` (includes clan, element, id, type, status line,        steps\[\], starting coords, ending coords, flags like obscured,        scout).
    -   `TurnNo` (e.g., 2305).
-   Definitions available:
    -   `Coord_t` (with `IsNA()`, `IsObscured()`, `ColRow()`).
    -   `Step_t` (direction, action, results payload).

### Output

-   `[]Tile_t` --- all tiles produced while walking this turn (see Tile    additions below).
-   `[]UnitLoc_t` --- per unit:    `{UnitID, Clan, Element, Start Coord_t, End Coord_t}`    (Start/End may be `N/A` if unknown).

------------------------------------------------------------------------

## Data Model Adjustments

### Tile_t (add fields)

-   `Turn int` --- the turn number that produced this tile.
-   `UnitID string`
-   `Clan string`
-   `Element string`
-   `Coord Coord_t`
-   `From *Tile_t` --- previous tile in the unit's path (optional).
-   `To *Tile_t` --- next tile in the unit's path (optional).
-   `StepIndex int` --- index (0...N-1) within the unit's step list;    ending tile gets `StepIndex = len(steps)`; starting tile gets `0`.
-   `Terrain, Borders, Settlement, Etc` --- whatever your status/result    extraction populates.
-   `StatusSource enum{StatusLine, Step}` --- provenance of the tile's    state.

### UnitLoc_t

-   `UnitID, Clan, Element`
-   `Start Coord_t`
-   `End Coord_t`

------------------------------------------------------------------------

## Processing Order

1.  **Accept exactly one turn** per invocation.
2.  **Sort units** by `(Clan asc, Element asc, UnitID asc as tiebreak)`.
3.  For each unit in order, branch:
    -   **Scout**: forward walk (special handling).
    -   **Non-scout**: backward walk from end to start.

------------------------------------------------------------------------

## Logging Policy

Use structured logs with codes and unit context:
- `E-ULK-001` No status line --- **error**, continue.
- `E-ULK-002` Ending coord `N/A` --- **error**, continue.
- `W-ULK-003` Ending coord **obscured** --- **warning**, continue.
- `W-ULK-004` Start coord mismatch: computed start vs parsed start --- **warning**.
- `W-ULK-005` Non-move scout step retained coord (informational warning only if helpful).
- `I-ULK-006` Prior turn end updated due to obscured match --- **info**.

(Keep codes in a central table.)

------------------------------------------------------------------------

## Non-Scout Algorithm (Backward Walk)

Given `u`:
1. **Guardrails**
   - If `u.StatusLine` missing → log `E-ULK-001`, continue.
   - `curLoc := u.EndCoord`.\
     - If `curLoc.IsNA()` → `E-ULK-002`, continue.
     - If `curLoc.IsObscured()` → `W-ULK-003`.
2. **Seed ending tile**
   - `curTile := makeTile(TurnNo, u, curLoc)`
   - Initialize from **status line** + **current location**.
   - Mark as **ending tile**:  `StepIndex = len(u.Steps)`, `StatusSource=StatusLine`.
   - Save as `u.EndTile`.
3. **Walk steps backward**
   - For `i := len(u.Steps)-1; i >= 0; i--`:
     - Apply step **results** to `curTile` (terrain, borders, sightings, etc.).
     - If there **is** a prior step (i-1 \>= 0):
       - Compute **came-from** direction: `(step.Dir + 3) mod 6` (180° turn on hex).
       - `prevCoord := neighbor(curTile.Coord, cameFromDir)`
       - `prevTile := makeTile(TurnNo, u, prevCoord)`
       - Link: `prevTile.To = curTile`; `curTile.From = prevTile`
       - `prevTile.StepIndex = i` (the tile state just **before** applying step i)
       - `curTile = prevTile`
4. **Finish**
   - After loop ends, `curTile` is the **starting tile** for the move this turn. Save as `u.StartTile`.
   - **Compare starts**:
       - If `u.StartCoord` differs from `curTile.Coord` →  `W-ULK-004`.

Collect all tiles created along the way for this unit into the turn's tile list.

------------------------------------------------------------------------

## Scout Algorithm (Forward Walk)

Given `u` with scout flag:
1. **Seed starting tile**
   - `curLoc := u.StartCoord` (may be obscured or exact; if `N/A`, you can still proceed using `N/A` until the first move establishes a coordinate).
   - `curTile := makeTile(TurnNo, u, curLoc)`
   - Initialize from **status line** if present (still log `E-ULK-001` when missing).
   - Mark `StepIndex=0`, `StatusSource=StatusLine`.
2. **Walk steps forward (i := 0...len-1)**
   - If step is **move** with direction `d`:
       - `nextCoord := neighbor(curTile.Coord, d)` (if `curTile.Coord` is `N/A`, treat as error or skip? → **Decision**: log error `E-ULK-002` if `N/A` and cannot move; continue to next unit.)
       - `nextTile := makeTile(TurnNo, u, nextCoord)`
       - Link: `curTile.To = nextTile`; `nextTile.From = curTile`
       - `nextTile.StepIndex = i+1` - `curTile = nextTile`
       - Update **unit's current location** to `nextCoord`.
   - Else (non-move: Still/Patrol/etc.):
     - Keep `curTile.Coord` unchanged (optional `W-ULK-005` once per unit).
   - Apply step **results** to `curTile`.
3. **Finish**
   - Set `u.EndTile = curTile` - `u.EndCoord = curTile.Coord`  (may remain obscured/unchanged)

Append all scout tiles to the turn's tile list.

------------------------------------------------------------------------

## Turn-Level Return & Outer Loop Checks

After iterating all units in the turn:
- **Return**:
  - `tilesThisTurn []Tile_t`
  - `unitLocsThisTurn []UnitLoc_t` (for every processed unit, even with `N/A` coords)

**Outer loop (caller)** responsibilities (when processing multiple turns sequentially):
1. Compare each unit's **returned** `{Start, End}` with the **parsed** counterparts for this **same** turn; report any diffs.
2. Compare this turn's **Start** with **prior turn's End** for the same unit; report diffs.
3. **Obscured carry-forward repair**:
   - If prior turn's `End` was **obscured** and this turn's `Start` has a concrete coordinate whose **col+row** matches prior's **col+row**, update prior turn's `End` to concrete and log `I-ULK-006`.
4. Continue to next turn using the same procedure.

------------------------------------------------------------------------

## Consolidation (Post-Walk Merge)

After all turns walked:

1.  **Sort all tiles** by:
    -  `Turn asc`, 2) `Coord asc` (col,row or cube), 3) `UnitID asc`.
2.  **Merge into a map keyed by `Coord`**:
    -   For each `Coord`, aggregate all tiles across turns.
    -   Build a consolidated record containing:
        -   Latest known Terrain/Borders per turn.
        -   List of contributing units.
        -   Full audit trail of changes.
3.  **Diff/Conflict reporting** (per coordinate):
    -   For a given **turn**, if multiple tiles disagree on **terrain**
        or **borders**, emit a report:
        -   Fields: `Turn`, `Coord`, `Units[]`,
            `Delta{Field, From, To}`.
4.  **Settlements --- special turn-scoped rule**:
    -   If **any** unit reports a settlement at `Coord` in turn `T`        where none existed before:
        -   Report **once**: "Settlement newly observed at T, Coord by            Units\[...\]."
        -   Do **not** report non-settlement observations at the **same            turn/coord** as conflicts.
    -   If a settlement existed before turn `T` but is **missing** in        turn `T`:
        -   Report **once**: "Settlement missing at T, Coord (previously            present)."
        -   Do not multiply-report by unit.
5.  **Return** the **merged list** (stable-ordered by `Coord`, then by    last-observed `Turn`).

------------------------------------------------------------------------

## Determinism & Idempotency

-   Given the same input turn and unit ordering rule, the walk must be deterministic.
-   Re-walking the same turn must produce identical `Tile_t` sequences and summaries.

------------------------------------------------------------------------

## Acceptance Criteria (High-Value Tests)

1.  **Non-scout backward chain**
    -   Given end coord and 3 steps, path length is 4 tiles (start + 3).
    -   `StepIndex` counts down to 0 at start.
    -   Start coord warns if mismatch with parsed start (`W-ULK-004`).
2.  **Scout forward chain**
    -   Starting at a known coord, 2 moves produce 3 tiles and final coord advances twice.
    -   Non-move step keeps coord.
3.  **Missing status line**
    -   Logs `E-ULK-001`; unit is skipped but other units continue.
4.  **N/A ending coord (non-scout)**
    -   Logs `E-ULK-002`; unit skipped.
5.  **Obscured end (non-scout)**
    -   Logs `W-ULK-003`; proceeds.
6.  **Direction reverse math**
    -   For each hex dir `d`, came-from is `(d+3)%6`; validated by neighbor back-and-forth.
7.  **Outer loop prior-end repair**
    -   Prior turn end obscured at (col,row=10,12); next turn start exact at (10,12) updates prior and logs `I-ULK-006`.
8.  **Consolidation conflicts**
    -   Two units claim different terrain on same coord/turn → single conflict record with both units.
9.  **Settlement rule**
    -   First appearance in turn T: single "new settlement" record; no per-unit "missing" noise in T.
    -   Missing in later turn: single "settlement missing" record.

------------------------------------------------------------------------

## Helper Signatures (Go-ish)

``` go
func WalkTurn(t Turn_t) (tiles []Tile_t, unitLocs []UnitLoc_t)

func makeTile(turn int, u Unit_t, c Coord_t) *Tile_t

func neighbor(c Coord_t, dir int) Coord_t // dir 0..5 (flat per your layout)

func Consolidate(allTiles []Tile_t) (merged []MergedTile_t, reports []DiffReport_t)
```

`MergedTile_t` contains final per-coord view with history;
`DiffReport_t` is your structured difference record including settlement cases.

------------------------------------------------------------------------

## Notes / Decisions

-   **Scouts with `N/A` start**: if a move occurs but we have no concrete starting coord, log `E-ULK-002` and skip the unit (safer), or permit if your engine supports inferred starts; above spec chooses the **safer skip**.
-   **Hex layout**: assumes standard 6-dir hex with `(d+3)%6` being opposite; wire to your axial/offset utilities.
-   **Status application**: Status line initializes ending (non-scout) / starting (scout) tile; step results refine tile state thereafter.
