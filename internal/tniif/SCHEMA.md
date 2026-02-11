# TNIIF JSON Schema (v0)

This folder contains the initial Go types for the **TNIIF** JSON structures.

## Canonical tile coordinates

The canonical tile coordinate representation is the human-friendly **`"AA 0101"`** format:

- It is what the renderer already supports.
- It is what players can easily read and edit.
- It is the only coordinate representation in this v0 schema (`TileRef.Key`).

## Current status

- **Scry is intentionally not present.** Sharing data is done by exchanging one or more JSON `Document` files.
- Enums (`Terrain`, `Edge`, `Resource`, `Direction`) are **string-backed** for stability and forward compatibility.

## Planned tightening (future iteration)

You mentioned that the pipeline currently supports "player convenience" expansions like `"BH" -> "Brush Hill(s)"`.
In a future iteration, we will:

- Restrict JSON to **report-native codes only** (e.g. `"BH"`).
- Move any "pretty name" / convenience mapping to:
  - the UI layer (display),
  - and/or optional lookup tables shipped with the renderer (not embedded in the shared schema).

## File layout

- `schema/schema.go` — Go types for the v0 schema
- `TNIIF.md` — project outline (scope, pipeline integration, roadmap)
- `AGENT.md` — working rules for coding agents and contributors

## Quick usage

```go
import "github.com/mdhender/tniif/schema"

var doc schema.Document
// json.Unmarshal(..., &doc)
```

## License

**MIT License**

  Copyright (c) 2026 Michael D Henderson
  
  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:
  
  The above copyright notice and this permission notice shall be included in all
  copies or substantial portions of the Software.
  
  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
  SOFTWARE.
