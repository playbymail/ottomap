# AST Parser — Short Reference

## API Summary

### Package

```go
package ast
```

### Entry Points

```go
// Parse parses raw input into an AST, calling the CST parser internally.
// Returns AST and AST-level diagnostics.
func Parse(input []byte) (*File, []Diagnostic)

// FromCST transforms a CST file into an AST file (no lexing/parsing).
func FromCST(cf *cst.File, input []byte) (*File, []Diagnostic)
```

### Diagnostics

```go
type Severity int
const (
    SeverityError Severity = iota
    SeverityWarning
    SeverityInfo
)

type Code string

const (
    CodeMonthUnknown  Code = "E_MONTH"
    CodeYearInvalid   Code = "E_YEAR"
    CodeYearOutOfRange     = "E_YEAR_RANGE"
    CodeTurnInvalid   Code = "E_TURN"
    CodeMissingHash   Code = "W_MISSING_HASH"    // optional
    CodeDupHeader     Code = "W_DUPLICATE_HEADER"// optional
)

type Diagnostic struct {
    Severity Severity
    Code     Code
    Span     Span    // copied from CST token span
    Message  string
    Notes    []string
}
```

### Spans

```go
type Span struct {
    Start int
    End   int
    Line  int
    Col   int
}
```

### Node Interface

```go
type Node interface {
    NodeKind() Kind
    NodeSpan() Span
    Source() Source
}

type Kind int

const (
    KindFile Kind = iota
    KindHeader
)

// Source ties an AST node back to its CST origin.
type Source struct {
    FileSpan Span
    Origin   any // usually pointer to CST node
}
```

---

## Node/Field Table

### Legend

* **Field Type**: `int`, `string`, `bool`, or nested node
* **Req/Opt**: required (always set), optional (may be empty or defaulted)
* **Notes**: describes normalization, flags, or error conditions

---

### `File` (root)

| Field        | Type        | Req/Opt | Notes                                        |
| ------------ | ----------- | ------- | -------------------------------------------- |
| `Headers`    | \[]\*Header | Opt     | Zero or more parsed header nodes.            |
| `src`        | Source      | Req     | Covers full file span, points to CST origin. |
| `NodeKind()` | Kind        | —       | `KindFile`.                                  |
| `NodeSpan()` | Span        | —       | Covers all child headers.                    |
| `Source()`   | Source      | —       | Returns `src`.                               |

---

### `Header`

Represents a normalized header like:
`Current Turn August 2025 (#123)`

| Field          | Type   | Req/Opt | Notes                                           |
| -------------- | ------ | ------- | ----------------------------------------------- |
| `Month`        | int    | Req     | Normalized month, `1..12`, or `0` if unknown.   |
| `Year`         | int    | Req     | Parsed year. Defaults to `0` if invalid.        |
| `Turn`         | int    | Req     | Parsed turn number. Defaults to `0` if invalid. |
| `RawMonth`     | string | Req     | Original text (e.g. `"August"`, `"Agust"`).     |
| `RawYear`      | string | Req     | Original year string.                           |
| `RawTurn`      | string | Req     | Original turn string.                           |
| `UnknownMonth` | bool   | Req     | True if month not recognized.                   |
| `InvalidYear`  | bool   | Req     | True if year not numeric or out of range.       |
| `InvalidTurn`  | bool   | Req     | True if turn not numeric or `< 1`.              |
| `src`          | Source | Req     | Span covers the CST `Header` node.              |
| `NodeKind()`   | Kind   | —       | `KindHeader`.                                   |
| `NodeSpan()`   | Span   | —       | Covering span of CST header.                    |
| `Source()`     | Source | —       | Returns `src`.                                  |

---

## Normalization Rules

* **Months:** Accepts full English names + 3-letter abbreviations (case-insensitive). Anything else → `Month=0`, `UnknownMonth=true`, diagnostic `E_MONTH`.
* **Year:** Must parse as integer, default policy `1900 ≤ Year ≤ 2200`. Out of range → `InvalidYear=true`, diagnostic `E_YEAR_RANGE`. Non-numeric → `InvalidYear=true`, diagnostic `E_YEAR`.
* **Turn:** Must parse as positive integer (`≥1`). Invalid → `Turn=0`, `InvalidTurn=true`, diagnostic `E_TURN`.

---

## Example (Happy Path)

Input:

```
Current Turn August 2025 (#123)
```

AST (abbrev):

```go
&File{
  Headers: []*Header{
    {
      Month: 8,
      Year: 2025,
      Turn: 123,
      RawMonth: "August",
      RawYear:  "2025",
      RawTurn:  "123",
      UnknownMonth: false,
      InvalidYear:  false,
      InvalidTurn:  false,
    },
  },
}
```

Diagnostics: none.

---

## Example (Errors)

Input:

```
Current Turn Agust 202X (#12a3)
```

AST:

```go
Header{
  Month: 0, Year: 0, Turn: 0,
  RawMonth:"Agust", RawYear:"202X", RawTurn:"12a3",
  UnknownMonth:true, InvalidYear:true, InvalidTurn:true,
}
```

Diagnostics:

* `E_MONTH`: unknown month `"Agust"`
* `E_YEAR`: invalid year `"202X"`
* `E_TURN`: invalid turn number `"12a3"`

---

## Recovery & Policy

* AST keeps **nodes with flags** instead of dropping them.
* Multiple headers: policy TBD (warn or pick first/last).
* CST synthesis (missing `#`, `)`) can be surfaced as warnings (optional).
