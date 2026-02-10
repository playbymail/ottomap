# TNIIF Project Outline

TNIIF is a shareable, mergeable **map-intelligence interchange format** designed for TribeNet/OttoMap-style workflows.

The goals:

- Preserve **exactly what the report says**, not what we *wish* it said.
- Make player edits safe and obvious.
- Support merging multiple independent contributions without requiring “Scry” or other hacks.
- Remain renderer-friendly and deterministic.

## Scope

TNIIF covers:

- Unit movement chains per turn (including scouts)
- Tile observations gathered at movement endpoints
- Tile borders (directional edges, neighbor terrain when visible)
- Encounters (friendly/hostile)
- Settlements and resources
- Special hex markers (identified by stable `id`)

TNIIF intentionally does **not** cover:

- Scry (removed; schema makes data sharing obsolete)
- Raw report text dumps
- Per-step internal parser debug state beyond minimal `lineNo`/`stepNo` traceability

## Canonical coordinates

Canonical tile coordinates are the **"AA 0101"** format.

- JSON stores coordinates in `TileRef.Key` only.
- Any internal engine conversion (WorldCoord, cube coords, etc.) stays internal to parser/renderer.

## Pipeline placement

Typical flow:

1. **Parse** one or more turn reports (Word/text) into internal types.
2. **Normalize** to TNIIF `Document`:
   - Convert all tile coordinates to `"AA 0101"`
   - Convert movement/observations into `Turn -> Moves -> Steps -> Report`
3. **Merge** multiple documents:
   - Primary key: `(turn.id, unit, tile.key)`
   - Resolve conflicts deterministically (policy documented below).
4. **Render**:
   - Generate Worldographer/WXX output
   - Emit warnings/notes at the relevant tile/unit/turn level

## Merge philosophy

A TNIIF merge should be:

- deterministic (same inputs, same output),
- conservative (never silently discard conflicts),
- additive where possible (keep both, but annotate).

Suggested merge rules (v0 policy sketch):

- **Turns**: union by `turn.id`
- **Moves**:
  - union by `(turn, unit)`
  - steps concatenate only if they are non-conflicting; otherwise keep both chains and attach notes
- **Tile observations**:
  - union by `(turn, tile.key)`
  - if the same field conflicts (e.g., Terrain differs), choose a stable winner (e.g., "newer source", "preferred clan") and attach a `Note` describing the conflict and the loser value

## Renderer notes

Renderer may attach notes such as:

- Unknown edge name (e.g., player typo “Stine Road” vs “Stone Road”)
- Hidden unit suppression toggles
- Out-of-bounds tile keys (e.g., TribeNet map size constraints)

Notes belong as close to the source as possible:

- tile-level (`TileObservation.Notes`)
- turn-level (`Turn.Notes`)
- document-level (`Document.Notes`)

## Roadmap

Planned iterations:

1. **Enum tightening**:
   - Terrain/Edge/Resource values become *report-native codes only*
   - Display names become UI concern (lookup tables)
2. **Coordinate validation**:
   - Strict parsing of `"AA 0101"`
   - Optional checksum/normalization rules (uppercase, spacing)
3. **Conflict policy**:
   - Formalize merge precedence rules
   - Add structured conflict records if needed (beyond free-form notes)
4. **Schema versioning**:
   - `schema` field remains required; bump on breaking changes
