# AGENT.md — Working Rules for TNIIF Schema Work

This file is for coding agents and contributors working on the Go types + JSON schema.

## Non-goals (do not add)

- Do **not** add Scry back into the JSON schema.
- Do **not** add alternate coordinate systems to the JSON schema in v0.
- Do **not** embed "pretty names" for enums into shared JSON.

## Canonical coordinates

- Canonical tile coordinate representation is **exactly** `"AA 0101"`. Coordinates represents grid plus column and row in the grid. Grid is AA...ZZ, column is 01...30, and row is 01...21.
- JSON stores this as `TileRef.Key` (required).
- Any internal coordinate representations (cube coords, q/r, x/y) must remain internal.

Validation rules (recommended):

- `Key` should match `^[A-Z]{2} [0-9]{4}$` (two letters, space, four digits)
- Preserve player edits but normalize when safe (uppercase letters, single space)

## Enums

v0 intentionally uses string-backed types:

- `Terrain`, `Edge`, `Resource`, `Direction`

Future tightening:

- Restrict JSON values to **report-native codes only** (e.g. `BH`), not expanded names.
- Any mapping like `BH -> Brush Hill(s)` belongs to the UI/display layer or renderer lookup tables.

## Hidden units

Rules from prior design discussions:

- A unit may include a `"hidden": true` toggle.
- When hidden is true, renderer suppresses:
  - the unit icon itself
  - that unit’s observations
- If another (non-hidden) unit encounters it, that encounter may still appear on the map.
- If *both* units are hidden, the encounter is suppressed.

In the schema:

- `UnitTurnMoves.Hidden` is the per-turn toggle (preferred).
- `UnitMeta.Hidden` is a global default (optional).

## Notes

Use `Note` for anything human-readable the renderer/user should see:

- Unknown edges (typos)
- Merge conflicts
- OOB coordinates

Attach notes at the narrowest sensible scope:
- tile observation, then turn, then document.

Treat `Notes` as a slice of `Note`.

## Versioning

- `Document.Schema` is required.
- Breaking changes require a new schema string (e.g. `tn-map.v1`).

## Style

- Keep types small and JSON-first.
- Prefer stable names and forward-compatible optional fields (`omitempty`).
- Avoid pointers except where "tri-state" matters (e.g., Hidden toggle).
