# TNIIF  
**TribeNet Inferred Input Format**

TNIIF defines the JSON input format consumed by the **OttoMap Renderer**.

It represents a curated, inferred subset of data extracted from parsed TribeNet turn reports.  
TNIIF is the **boundary artifact** between parsing and rendering.

The Parser produces TNIIF.  
The Renderer consumes TNIIF.

Users are encouraged to inspect and edit TNIIF files directly.

---

## Design Goals

TNIIF is designed to be:

- **Explicit** — no hidden inference at render time
- **Strict** — invalid values are rejected, not guessed
- **Deterministic** — same inputs always produce the same map
- **Mergeable** — many files can be rendered into one map
- **Debuggable** — nothing is silently discarded

---

## Scope

TNIIF contains:

- Exactly **one turn** of inferred data
- Metadata needed for attribution, debugging, and output naming
- Units, their movement steps, and the observations they report

TNIIF does **not** contain:

- Hexes or map geometry
- Unit names
- Derived locations
- Rendering instructions
- Worldographer-specific primitives

Hexes are constructed by the renderer.

---

## Terminology

### Clan ID vs Clan Number

- **Clan ID**  
  A **4-digit string**, e.g. `"0987"`  
  This is the canonical external identifier and is used:
  - at the file level
  - in filenames
  - in UI / notes

- **Clan Number**  
  An **integer** in the range `1..999`  
  Used internally by tools. Never stored as a string.

---

## File-Level Metadata

Each file contains exactly one turn and includes the following metadata:

- `game`  
  Four-digit string identifying the TribeNet game (e.g. `"0300"`)

- `turn`  
  String in the form `YYYY-MM`  
  Identifies the turn represented by this file

- `clan` (optional)  
  Clan ID string (`"0987"`) indicating the *intended owner/perspective* of the file  
  Files may contain units from other clans

- `parser_version`  
  Semantic Version string identifying the parser that produced this file

- `source`  
  Filesystem path (Windows, macOS, or Linux) to the original input  
  Stored as raw text; HTML-escaped by the renderer when written to output notes

All metadata is rendered into a note on the output map.

---

## Multi-File Rendering

Users are encouraged to render **multiple TNIIF files** into a single map rather than combining them into one file.

The renderer:

1. Accepts the **target player Clan ID** as an input parameter
2. Loads all input files
3. Sorts data by:
   - turn
   - trust (non-target clan first, target clan last)
   - unit id
   - step sequence
4. Walks the resulting event stream to construct hexes

The output filename uses:
- the target player Clan ID
- the **maximum turn** across all input files

---

## Units

Units are identified solely by **unit id**.

### Unit ID format

DCCC[type][sequence]

Where:
- `D` = single digit `0..9`
- `CCC` = clan number, zero-padded to 3 digits
- optional type code and sequence (e.g. `c4`)

Examples:
- `"0987"` — clan/tribe unit for clan 987
- `"1987"` — tribe unit for clan 987
- `"0987c4"` — 4th Courier unit for clan 987

Clan ownership is derived from the embedded clan number.

### Hidden units

Units may include an optional boolean field:

- `hidden: true`

When a unit is hidden:
- the renderer suppresses **all output derived from that unit**
  - unit icon
  - observations
  - notes
- this suppression persists until a later file sets `hidden: false` for the same unit

If another (visible) unit encounters a hidden unit, that encounter **is still rendered**, because it comes from the visible unit’s data.

If both units are hidden, nothing is rendered 🤪

---

## Moves (Steps)

Internally, the parser produces **steps**; users think of them as **moves**.

Each unit has a list of moves/steps.

A move includes:
- `sequence` — integer `1..n`
- optional `direction` — one of `N, NE, SE, S, SW, NW`
- `observations[]`

Moves may exist without movement:
- A “failed to move” report produces a step with **no direction** but with observations

Moves and observations **do not carry location data**.

The renderer assigns observations to hexes as it generates the map.

---

## Observations

Observations describe what a unit learned during a step.

They may include:

- `terrain` — TribeNet terrain name (string)
- `cities[]` — list of city names
- `resources[]` — list of resource names
- `units_encountered[]` — list of unit ids
- `edge_features` — map of direction → list of features

All values are strings suitable for rendering.

### Edge Features

Edge features are grouped by direction and sorted:

- **Direction order:** clockwise from North  
  `N → NE → SE → S → SW → NW`
- **Then by feature name (lexicographic)**

Each feature is one of:
- road (e.g. `"Stone Road"`)
- border
- terrain

Internally, the parser represents these as `map[direction][]feature`.

---

## Validation, Mapping, and Conflicts

### TribeNet vocabularies

- The parser is the **source of truth** for valid TribeNet names
  - terrain
  - edge features
  - resources
- TNIIF stores TribeNet-native terms only

### Renderer mapping

- Users provide mappings from:
  - TribeNet terrain → Worldographer terrain
  - TribeNet features → Worldographer styles/icons

### Invalid values

If a value is unknown or invalid:
- it is **rejected**
- the map state is **not modified**
- a note is added to the hex, e.g.:

>Unknown terrain: “light forest”
>Unknown edge: “Stine Road”

### Conflict resolution

Anything in an observation may conflict with earlier data.

Rules:
1. Comparisons are **literal**
2. Valid later values replace earlier values (**last write wins**)
3. Replaced values are preserved in a **hex note** for debugging

Example:
- earlier resources: `Iron`, `Copper`
- later resources: `Jade`

Result:
- map shows: `Jade`
- note records: `Iron`, `Copper`

Nothing is silently lost.

---

## Map Construction

- Renderer creates an empty map of the maximum TribeNet size
- Hexes are created as observations are applied
- TribeNet imposes a fixed grid limit (e.g. `"ZZ 3120"`)

TNIIF never contains hex definitions.

---

## Status

This format is **actively evolving**.

The README is authoritative.  
A `schema.json` will be introduced once the field set stabilizes.
