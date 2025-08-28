# AST Parser: Developer Specification

## Scope & Objectives

* Input: a **lossless CST** from `cst.ParseFile(input)`.
* Output: a **normalized AST** that downstream code can use directly.
* Goals:

    * Normalize surface forms (e.g., month names → integers).
    * Enforce semantic constraints (e.g., year/turn are positive integers).
    * Produce actionable **semantic diagnostics** with precise spans (borrowed from CST tokens).
    * Be **panic-free**; recover whenever feasible so the AST is as complete as possible.

The AST is **not** lossless. It intentionally drops trivia and non-essential tokens (kept in CST). Every AST node includes a **Source** field that references the CST origin for diagnostics and tooling.

---

## External Interfaces

### Package and entry point

```go
package ast

import (
    "github.com/yourorg/yourrepo/cst"
)

// Parse produces a normalized AST and semantic diagnostics.
// It calls cst.ParseFile under the hood, merges CST diagnostics (optional),
// and adds AST-level diagnostics. It never panics.
func Parse(input []byte) (*File, []Diagnostic)
```

> If you prefer a two-step pipeline in callers:
>
> ```go
> c, cdiags := cst.ParseFile(input)
> a, adiags := ast.FromCST(c)
> ```
>
> you can provide `FromCST(*cst.File) (*File, []Diagnostic)` as well. The rest
> of this spec describes the transformation and validations.

---

## Diagnostics (AST)

AST diagnostics signal **semantic** issues (unknown month, out-of-range year, etc.). They reference **CST token spans**.

```go
type Severity int
const (
    SeverityError Severity = iota
    SeverityWarning
    SeverityInfo
)

type Code string // short code for programmatic handling, e.g. "E_MONTH", "E_TURN"

type Diagnostic struct {
    Severity Severity
    Code     Code
    Span     Span   // independent copy (not an alias) of CST span
    Message  string
    Notes    []string
}
```

Span mirrors the CST structure (copy, don’t alias):

```go
type Span struct {
    Start int
    End   int
    Line  int
    Col   int
}
```

---

## AST Data Model

All AST nodes implement:

```go
type Node interface {
    NodeKind() Kind
    NodeSpan() Span  // a representative span (e.g., from the primary token)
    Source() Source  // CST origin info (for tooling)
}

type Kind int

const (
    KindFile Kind = iota
    KindHeader
)
```

### Source mapping

```go
type Source struct {
    FileSpan Span       // covering span for the node (for highlighting)
    Origin   any        // implementation detail; typically pointer to CST node
                        // or a compact enum+offset map if you want to avoid pointers
}
```

> **Note**: `Origin` is meant for internal tools and debugging. Keep it opaque to callers.

### File (root)

```go
type File struct {
    Headers []*Header  // multiple headers allowed if they appear; validate below
    src     Source
}

func (f *File) NodeKind() Kind  { return KindFile }
func (f *File) NodeSpan() Span  { return f.src.FileSpan }
func (f *File) Source() Source  { return f.src }
```

### Header (normalized)

We normalize month → `1..12`, year → `int`, turn → `int`. We also carry the **raw** strings for reporting and for partial success.

```go
type Header struct {
    Month       int    // 1..12; 0 if unknown
    Year        int    // e.g., 2025
    Turn        int    // >0 if valid; 0 if unknown/invalid
    RawMonth    string // original text (e.g., "August", "Agust")
    RawYear     string // original "2025"
    RawTurn     string // original "123"
    src         Source

    // Flags let downstream code keep going gracefully:
    UnknownMonth bool
    InvalidYear  bool
    InvalidTurn  bool
}

func (h *Header) NodeKind() Kind  { return KindHeader }
func (h *Header) NodeSpan() Span  { return h.src.FileSpan }
func (h *Header) Source() Source  { return h.src }
```

---

## Mapping CST → AST

### Recognizing a Header

CST shape (from spec):

```
Header := Current Turn Month Year "(" "#" Number ")"
```

* We **do not** require that CST tokens are real (they may be synthesized zero-width on error).
* AST must:

    1. Extract raw strings (month, year, turn) from the corresponding token nodes if present.
    2. Normalize strings to numeric forms.
    3. Validate and report issues with spans anchored to the specific CST tokens.
    4. Fill flags and retain raw text for partial success.

### Field-level mapping

| CST Field                             | How to read       | AST Field(s)                | Notes                                     |
| ------------------------------------- | ----------------- | --------------------------- | ----------------------------------------- |
| `Month` (`MonthName` or `Identifier`) | `tok.Text(input)` | `Month` (1..12), `RawMonth` | Unknown → `UnknownMonth=true`, `Month=0`  |
| `Year` (`Number`)                     | numeric parse     | `Year`, `RawYear`           | Non-numeric/overflow → `InvalidYear=true` |
| `TurnNo` (`Number`)                   | numeric parse     | `Turn`, `RawTurn`           | Non-numeric/overflow → `InvalidTurn=true` |

> **Span anchoring**: diagnostics for each field should use that field’s token span. For summary/header-level notes, use the header node’s covering span.

---

## Normalization Rules

### Month

* Accept full English month names (case-insensitive): `January..December`.
* Accept common abbreviations (3-letter): `Jan..Dec`.
* If string is not a recognized month, set `UnknownMonth = true`, `Month = 0`, and emit:

    * Severity: `SeverityError`
    * Code: `Code("E_MONTH")`
    * Message: `unknown month 'Agust'`
    * Notes: `did you mean 'August'?` (optional, if you add a fuzzy matcher)

### Year

* Parse base-10 integer.
* Recommended validation: `Year in [1900, 2200]` (tweak for your domain).
* If non-numeric or out of range, set `InvalidYear = true`, keep `RawYear`, and emit:

    * `E_YEAR` (non-numeric/overflow) or `E_YEAR_RANGE` (out of range).

### Turn

* Parse base-10 integer, require `Turn >= 1`.
* If non-numeric or `< 1`, set `InvalidTurn = true`, and emit `E_TURN`.

> **Policy**: AST **keeps** a `Header` node even if month/year/turn is invalid, to aid recovery and allow partial downstream behavior. Consumers should check flags.

---

## Happy-Path Walkthrough

Input:

```
Current Turn August 2025 (#123)
```

**CST** (abbrev):

```
Header(
  Month: Token(MonthName "August"),
  Year:  Token(Number "2025"),
  Turn:  Token(Number "123"),
  ...
)
```

**AST**:

```go
&ast.File{
  Headers: []*ast.Header{
    {
      Month: 8, Year: 2025, Turn: 123,
      RawMonth: "August", RawYear: "2025", RawTurn: "123",
      UnknownMonth: false, InvalidYear: false, InvalidTurn: false,
      src: Source{FileSpan: <cover of header>, Origin: <*cst.Header>},
    },
  },
}
```

No diagnostics.

---

## Error Handling & Recovery

### General strategy

* **Localize** diagnostics to the field’s token span when possible.
* **Continue** building the AST node if enough context exists (prefer partial nodes over dropping them).
* Never panic; return diagnostics + best-effort AST.

### Examples

#### 1) Unknown month + bad turn

Input:

```
Current Turn Agust 2025 (#12a3)
```

CST (likely):

* `Month` token: `Identifier("Agust")`
* `TurnNo` token: maybe `Identifier("12a3")` (CST may have synthesized a Number and left this as unexpected; choose the real token span for AST diag if available)

AST behavior:

* Month: lookup fails → `Month=0`, `UnknownMonth=true`, diag:

    * `E_MONTH` at span of `Month` token, `"unknown month 'Agust'"`
* Year: OK.
* Turn: parse fails → `Turn=0`, `InvalidTurn=true`, diag:

    * `E_TURN` at span of `TurnNo`, `"invalid turn number '12a3' (digits only)"`

AST still emits:

```go
Header{ Month:0, Year:2025, Turn:0, RawMonth:"Agust", RawTurn:"12a3", UnknownMonth:true, InvalidTurn:true }
```

#### 2) Missing tokens synthesized by CST

Input:

```
Current Turn August 2025 (123)   // missing '#'
```

CST: `Hash` synthesized zero-width at insertion site.

AST:

* Month/Year/Turn parse fine (turn text from the real number token).
* Optionally, emit **info** or **warning** if you want to surface the missing hash at AST level:

    * `W_MISSING_HASH` anchored to the synthesized `Hash` span (zero-width).
    * This is **optional**; CST diagnostic already covered structure.

#### 3) Multiple headers

If multiple headers appear in the file, we keep them in order. You may want an **AST-level policy**:

* If multiple headers conflict (e.g., different months/years), decide whether to:

    * keep all and **warn** (`W_DUPLICATE_HEADER`), or
    * keep the **first** and **warn** on later ones, or
    * keep the **last** and **warn** on earlier ones.

This is domain-specific. The AST layer should **not** drop nodes silently; always emit a diagnostic if there’s a policy conflict.

---

## Builder Design

### Entry

```go
func Parse(input []byte) (*File, []Diagnostic) {
    cf, cdiags := cst.ParseFile(input)
    af, adiags := FromCST(cf, input) // input needed for token text
    // Optionally merge CST diagnostics as well:
    // return af, append(cstToASTDiags(cdiags), adiags...)
    return af, adiags
}
```

### Transform

```go
func FromCST(cf *cst.File, input []byte) (*File, []Diagnostic) {
    b := &builder{input: input}
    return b.file(cf)
}

type builder struct {
    input []byte
    diags []Diagnostic
}

func (b *builder) file(cf *cst.File) (*File, []Diagnostic) {
    f := &File{}
    for _, decl := range cf.Decls {
        switch d := decl.(type) {
        case *cst.Header:
            h := b.header(d)
            f.Headers = append(f.Headers, h)
        default:
            // Ignore other decls for now, or add diags if needed.
        }
    }
    // Compute file span from children (optional).
    if len(f.Headers) > 0 {
        f.src.FileSpan = f.Headers[0].NodeSpan()
        for _, h := range f.Headers[1:] {
            f.src.FileSpan = cover(f.src.FileSpan, h.NodeSpan())
        }
    }
    return f, b.diags
}
```

### Header normalization

```go
func (b *builder) header(ch *cst.Header) *Header {
    h := &Header{ src: Source{ FileSpan: toSpan(ch.Span()), Origin: ch } }

    // Month
    rawMonth := tokenText(b.input, ch.Month)
    h.RawMonth = rawMonth
    m, known := normalizeMonth(rawMonth)
    if !known {
        h.UnknownMonth = true
        b.err(ch.Month, "E_MONTH", "unknown month %q", rawMonth)
    }
    h.Month = m

    // Year
    rawYear := tokenText(b.input, ch.Year)
    h.RawYear = rawYear
    y, yok := parseYear(rawYear)
    if !yok {
        h.InvalidYear = true
        b.err(ch.Year, "E_YEAR", "invalid year %q", rawYear)
    } else if !validYearRange(y) {
        h.InvalidYear = true
        b.err(ch.Year, "E_YEAR_RANGE", "year %d out of range", y)
    }
    h.Year = y

    // Turn
    rawTurn := tokenText(b.input, ch.TurnNo)
    h.RawTurn = rawTurn
    t, tok := parsePositiveInt(rawTurn)
    if !tok || t < 1 {
        h.InvalidTurn = true
        b.err(ch.TurnNo, "E_TURN", "invalid turn number %q", rawTurn)
    }
    h.Turn = t

    return h
}
```

Helpers:

```go
func tokenText(input []byte, tn *cst.TokenNode) string {
    if tn == nil || tn.Tok == nil { return "" }
    // CST wraps lexers.Token; expose a helper/method if available.
    s := tn.Tok.Span
    return string(input[s.Start:s.End])
}

func toSpan(s cst.Span) Span {
    return Span{Start: s.Start, End: s.End, Line: s.Line, Col: s.Col}
}

func (b *builder) err(tn *cst.TokenNode, code Code, format string, args ...any) {
    var sp Span
    if tn != nil && tn.Tok != nil {
        sp = toSpan(tn.Tok.Span)
    }
    b.diags = append(b.diags, Diagnostic{
        Severity: SeverityError,
        Code:     code,
        Span:     sp,
        Message:  fmt.Sprintf(format, args...),
    })
}
```

Normalization utilities:

```go
var months = map[string]int{
    "january":1,"jan":1,
    "february":2,"feb":2,
    "march":3,"mar":3,
    "april":4,"apr":4,
    "may":5,
    "june":6,"jun":6,
    "july":7,"jul":7,
    "august":8,"aug":8,
    "september":9,"sep":9,"sept":9,
    "october":10,"oct":10,
    "november":11,"nov":11,
    "december":12,"dec":12,
}

func normalizeMonth(s string) (int, bool) {
    if s == "" { return 0, false }
    return months[strings.ToLower(strings.TrimSpace(s))], months[strings.ToLower(strings.TrimSpace(s))] != 0
}

func parseYear(s string) (int, bool) {
    v, err := strconv.Atoi(strings.TrimSpace(s))
    return v, err == nil
}

func validYearRange(y int) bool { return y >= 1900 && y <= 2200 }

func parsePositiveInt(s string) (int, bool) {
    v, err := strconv.Atoi(strings.TrimSpace(s))
    if err != nil { return 0, false }
    return v, true
}
```

(Import `fmt`, `strconv`, `strings` where used.)

---

## Policy Notes & Recovery

* **Keep nodes, flag fields**: prefer building a `Header` with flags over dropping it. This supports partial downstream logic and better user feedback.
* **De-duplication**: If only one `Header` should apply, pick a policy:

    * Keep all, `W_DUPLICATE_HEADER` for extras.
    * Keep first/last and warn on the others.
* **Synthesis-awareness**: If a CST token is synthetic (zero-width), you may **downgrade** some errors to warnings, since the structural issue was already flagged at CST. You can detect synthetic tokens by `Start == End`.

---

## Testing Strategy

* **Happy path**: round-trip from sample inputs; check `Header{Month, Year, Turn}` and that diagnostics are empty.
* **Months**: full names and abbreviations; unknown strings produce `E_MONTH`.
* **Year**: non-numeric, large negatives, overflow, and range edges `[1900,2200]`.
* **Turn**: non-numeric, `0`, `-1`.
* **Synthesis**: inputs that trigger CST synthesis (missing `#`, missing `)`) should still produce usable AST nodes; optionally assert warning vs error policies if adopted.
* **Duplicates**: multiple headers; assert policy diagnostics.

---

## Worked Examples

### OK

```
Current Turn August 2025 (#123)
```

AST:

```
Header{ Month:8, Year:2025, Turn:123, flags=false }
```

Diags: none.

### Unknown month, invalid turn

```
Current Turn Agust 2025 (#12a3)
```

AST:

```
Header{
  Month:0, Year:2025, Turn:0,
  RawMonth:"Agust", RawTurn:"12a3",
  UnknownMonth:true, InvalidTurn:true,
}
```

Diags:

* `E_MONTH` at span("Agust")
* `E_TURN` at span("12a3")

### Missing hash (CST synthesized)

```
Current Turn August 2025 (123)
```

AST:

```
Header{ Month:8, Year:2025, Turn:123 }
```

Diags: optional `W_MISSING_HASH` (policy), anchored to synthesized span.
