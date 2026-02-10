# AGENT.md — Parser CLI

## Purpose

This CLI parses a single TribeNet turn report and produces a JSON document
conforming to the `internal/tniif` schema. The JSON is consumed by a
separate render application.

## Build & Test

```sh
go build -o dist/local/parser ./cmd/parser
go run ./cmd/parser --debug --game 0300 --clan 0331 \
  --input testdata/0300.0904-01.0331.report.txt \
  --output testdata/0300.0904-01.0331.json
```

Validate output with `jq . <output-file>`.
Cross-check unit count: `grep -c 'Current Hex' <input-file>` should match
`jq '[.clans[].units[]] | length' <output-file>`.

## Key Files

- `cmd/parser/main.go` — CLI entry point and domain→schema mapping
- `cmd/parser/DERIVING_LOCATIONS.md` — plan for populating location fields
- `internal/tniif/schema.go` — JSON schema types (target)
- `internal/domain/types.go` — parsed turn data types (source)
- `internal/parser/parser.go` — turn report parser (`ParseInput`)
- `internal/coords/map.go` — absolute coordinate math (`Map.Add`, `Map.Move`)
- `internal/coords/helpers.go` — coordinate parsing (`HexToMap`)
- `tniif/AGENT.md` — schema design rules and non-goals

## Mapping Status

The mapping converts `domain.Turn_t` → `schema.Document`.

### Completed

- [x] Document header (schema, game, turn, clan, source, created)
- [x] Unit identification and ending location
- [x] Move steps (advance, follows, goesTo, still)
- [x] Move results (succeeded, failed, vanished, unknown)
- [x] Observations: terrain, edges, encounters, resources
- [x] Scouts (`domain.Scout_t` → `schema.ScoutRun`)
- [x] Fleet movement / far horizons (`domain.FarHorizon_t`)
- [x] Settlements (`domain.Settlement_t` → `schema.Settlement`)
- [x] Special hex names (`Turn_t.SpecialNames` → `Document.SpecialHexes`)
- [x] Status lines (garrison/still units — mapped as last move step)
- [x] Scrying results (excluded from schema per design)
- [x] Hidden unit toggle (schema field exists; not populated by parser)
- [x] Notes (deferred to renderer)

### Current Sprint — Derive Locations

See `cmd/parser/DERIVING_LOCATIONS.md` for the full plan.

- [x] Task 1: Unit `MoveStep.EndingLocation` (backward walk from `Unit.EndingLocation`)
- [x] Task 2: Scout `ScoutRun.StartingLocation` and `MoveStep.EndingLocation` (forward walk)
- [x] Task 3: `Observation.Location` (copy from owning `MoveStep.EndingLocation`)
- [x] Task 4: `CompassPoint.Location` (2-hex bearing derivation from `Observation.Location`)

## Known Issues

- `coords.Map.Add()` is unbounded and `Map.ToHex()` silently formats
  out-of-range results as garbage (e.g., row −1 → grid row 00). Any
  multi-step coordinate math near map edges (especially Task 4 compass
  points) must validate bounds before calling `formatCoord()`. See the
  "Map Bounds Validation" section in `DERIVING_LOCATIONS.md`.

## Workflow

1. Implement the next task from the current sprint list above.
2. Add unit tests per the test cases in `DERIVING_LOCATIONS.md`.
3. Run the parser against the test input and validate with `jq`.
4. Update the sprint checklist.
