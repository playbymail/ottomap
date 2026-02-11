    # Parser CLI — Implementation Plan

This document describes the plan for completing `cmd/parser/main.go`,
the standalone CLI that parses a single TribeNet turn report file.

## Overview

The parser CLI reads a single turn report file (`.txt` or `.docx`),
parses it using `internal/parser.ParseInput` (the same parser used
by `render.go`), and logs a summary.  The goal is to split the
current parse+render pipeline into two separate tools because the
combined implementation in `render.go` is unmaintainable.

## Current State

The `RunE` function in `cmd/parser/main.go` already handles:
- CLI flags (`--input`, `--clan`, logging flags)
- Clan-id validation (with bugs — see below)
- Path resolution and file existence check (with bugs — see below)
- Reading the file into `[]byte`
- Timing the operation

It has four TODO markers where the real work should happen:
```go
// todo: call parser.ParseInput with the input
// todo: log any errors and return
// todo: sort the parsed input
// todo: log the number of units parsed
```

## Bugs to Fix

### Bug 1 — Inverted file-type check (lines 126–135)

The current code rejects regular files and accepts directories:
```go
} else if !sb.IsDir() {
    // error says "path is a folder" — but this branch runs for non-dirs
```

**Fix:** Invert the condition:
```go
} else if sb.IsDir() {
    return fmt.Errorf("path is a folder, not a file")
} else if !sb.Mode().IsRegular() {
    return fmt.Errorf("path must be a regular file")
}
```

### Bug 2 — Redundant `filepath.IsAbs` check (lines 122–125)

`filepath.Abs` always returns an absolute path (or an error).
Checking `filepath.IsAbs` after a successful `Abs` call is redundant.

**Fix:** Remove lines 122–125.

### Bug 3 — Fragile clan-id validation (lines 108–114)

`clanNo[1:]` will panic if `clanNo` is empty.  The code is
unnecessarily clever about the leading zero — just convert the
whole string to a number and check the range.  `fmt.Sprintf("%04d")`
handles zero-padding for the output.

**Fix:**
```go
n, err := strconv.Atoi(clanNo)
if err != nil || n < 1 || n > 999 {
    return fmt.Errorf("clan-id must be a number in the range 1..999")
}
clanID := fmt.Sprintf("%04d", n)
```

## Missing Feature — File Extension Validation

The CLI must accept only `.txt` and `.docx` files.  Add this check
after resolving the absolute path:
```go
ext := strings.ToLower(filepath.Ext(path))
if ext != ".txt" && ext != ".docx" {
    return fmt.Errorf("input must be a .txt or .docx file")
}
```

## Package Rename — `internal/docx` → `internal/tndocx`

Before implementing, rename `internal/docx` to `internal/tndocx`
to reflect that this is a TribeNet-specific docx adapter, not a
general-purpose docx library.  Update the package declaration and
all import paths accordingly.

## Import Changes

Replace the legacy import:
```go
// remove
"github.com/playbymail/ottomap/internal/turns"

// add
"github.com/playbymail/ottomap/internal/parser"
"github.com/playbymail/ottomap/internal/tndocx"
```

The `turns.TurnReportFile_t` usage (currently only used to hold
`Path`) is no longer needed — `path` is already a local variable.

## Implementing the TODOs

The four TODOs map to the call pattern in `render.go` line 297,
which calls `parser.ParseInput`.

### Reading the input (before the TODOs)

Replace the current `os.ReadFile` block with extension-aware loading.
For `.docx` files, the docx extractor converts the Word XML into
plain text (the same format as a `.txt` turn report).  The extracted
text is then used as the input to `ParseInput`, not the raw Word data:
```go
var data []byte
switch ext {
case ".docx":
    data, err = tndocx.ParsePath(path, false, true)
case ".txt":
    data, err = os.ReadFile(path)
}
if err != nil {
    logger.Error("parser", "error", err)
    return err
}
if len(data) == 0 {
    logger.Error("parser", "error", "empty input file")
    return fmt.Errorf("empty input file")
}
```

### TODO 1 — Call `parser.ParseInput`

The `ParseInput` signature is:
```go
func ParseInput(fid, tid string, input []byte,
    acceptLoneDash, debugParser, debugSections, debugSteps, debugNodes,
    debugFleetMovement, experimentalUnitSplit, experimentalScoutStill bool,
    cfg ParseConfig) (*domain.Turn_t, error)
```

Add Cobra bool flags for the parser parameters in `addFlags`.
All default to `false`.  Do **not** alias `--debug-parser` to
`--debug` — the `--debug` flag controls only the logging level.

```go
cmd.Flags().StringVar(&outputPath, "output", "", "write results to file instead of stdout")
cmd.Flags().BoolVar(&acceptLoneDash, "accept-lone-dash", false, "accept lone dash in movement lines")
cmd.Flags().BoolVar(&debugParser, "debug-parser", false, "enable parser debug logging")
cmd.Flags().BoolVar(&debugSections, "debug-sections", false, "enable section debug logging")
cmd.Flags().BoolVar(&debugSteps, "debug-steps", false, "enable step debug logging")
cmd.Flags().BoolVar(&debugNodes, "debug-nodes", false, "enable node debug logging")
cmd.Flags().BoolVar(&debugFleetMovement, "debug-fleet-movement", false, "enable fleet movement debug logging")
```

The `fid` (file id) is the base filename; `tid` (turn id) starts
empty and is set by `ParseInput` from the "Current Turn" line in
the report:
```go
fid := filepath.Base(path)
turn, err := parser.ParseInput(
    fid,
    "",    // tid — will be set from the report
    data,
    acceptLoneDash,
    debugParser,
    debugSections,
    debugSteps,
    debugNodes,
    debugFleetMovement,
    false, // experimentalUnitSplit
    false, // experimentalScoutStill
    parser.ParseConfig{},
)
```

### TODO 2 — Log errors and return

```go
if err != nil {
    logger.Error("parser", "file", fid, "error", err)
    return err
}
```

### TODO 3 — Log the number of units parsed

Write the summary to `--output` if set, otherwise to stdout:
```go
w := os.Stdout
if outputPath != "" {
    f, err := os.Create(outputPath)
    if err != nil {
        logger.Error("parser", "error", err)
        return err
    }
    defer f.Close()
    w = f
}
fmt.Fprintf(w, "turn %s: %d units parsed\n", turn.Id, len(turn.UnitMoves))
```

## Summary of Changes

| File                                  | Change                                                                                                                                    |
|---------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------|
| `cmd/parser/main.go`                  | Fix bugs #1–#3, add extension validation, replace `internal/turns` import with `internal/parser` + `internal/tndocx`, implement TODOs 1–4 |
| `internal/docx/` → `internal/tndocx/` | Rename package directory and update package declaration                                                                                   |
