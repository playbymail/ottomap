# CST Parser — Short Reference

## API Summary

### Package

```go
package cst
```

### Entry Point

```go
// Parse a single input buffer into a lossless CST.
// Never panics. Returns a CST (possibly with Bad* nodes) and diagnostics.
func ParseFile(input []byte) (*File, []Diagnostic)
```

### Diagnostics

```go
type Severity int
const (
    SeverityError Severity = iota
    SeverityWarning
    SeverityInfo
)

type Diagnostic struct {
    Severity Severity
    Span     Span     // byte offsets + line/col
    Message  string
    Notes    []string // optional hints
}
```

### Spans

```go
type Span struct {
    Start int // byte offset (inclusive)
    End   int // byte offset (exclusive)
    Line  int // 1-based
    Col   int // 1-based (UTF-8 code points)
}
```

### Node Interface

```go
type Node interface {
    Span() Span
    Kind() Kind
}
```

### Token Node (wraps a lexer token)

```go
type TokenNode struct {
    Tok *lexers.Token // carries Kind, Span, LeadingTrivia, TrailingTrivia
}
```

### Parser Helpers (behavioral contract)

```go
// Single-token lookahead; synthesized zero-width tokens on expectation failures.
want(kind lexers.TokenKind) *TokenNode
wantOneOf(kinds ...lexers.TokenKind) *TokenNode
recoverTo(sync ...lexers.TokenKind)           // skip tokens until any sync kind
```

> **Lexers package**: `lexers.New(line, col int, input []byte) *Lexer` and `(*Lexer).Next() *Token` (returns `nil` on EOF).

---

## Node/Field Table

### Legend

* **T** = `*TokenNode`
* **N** = `Node`
* **\[]N** = slice of nodes
* **Req** = required (synthesized if missing)
* **Opt** = optional
* **Span** = node’s full source span (derived)

---

### `File` (root)

| Field    | Type | Req/Opt | Description                                               |
| -------- | ---- | ------: | --------------------------------------------------------- |
| `Decls`  | \[]N |     Opt | Top-level declarations (e.g., `Header`, future sections). |
| `EOF`    | T    |     Req | Synthetic EOF sentinel at end of file.                    |
| `Span()` | Span |       — | Covers all `Decls` and `EOF`.                             |
| `Kind()` | Kind |       — | `KindFile`.                                               |

---

### `Header`

Represents lines like: `Current Turn August 2025 (#123)`

| Field       | Type | Req/Opt | Expected Token Kinds                               | Description                           |
| ----------- | ---- | ------: | -------------------------------------------------- | ------------------------------------- |
| `KwCurrent` | T    |     Req | `KeywordCurrent`                                   | “Current”.                            |
| `KwTurn`    | T    |     Req | `KeywordTurn`                                      | “Turn”.                               |
| `Month`     | T    |     Req | `MonthName` \| `Identifier`                        | Month literal; not normalized in CST. |
| `Year`      | T    |     Req | `Number`                                           | 4-digit year as written.              |
| `LParen`    | T    |     Req | `LParen`                                           | “(”.                                  |
| `Hash`      | T    |     Req | `Hash`                                             | “#”.                                  |
| `TurnNo`    | T    |     Req | `Number`                                           | Turn number digits only.              |
| `RParen`    | T    |     Req | `RParen`                                           | “)”.                                  |
| `Span()`    | Span |       — | From first to last field (with synthesis allowed). |                                       |
| `Kind()`    | Kind |       — | `KindHeader`.                                      |                                       |

> **Trivia:** All whitespace/comments are preserved on tokens as `LeadingTrivia`/`TrailingTrivia` via `lexers.Token`. The CST does not duplicate trivia into separate nodes.

---

### `BadTopLevel`

Used when an unrecognized run of tokens occurs at the top level and we can’t confidently form a known node.

| Field    | Type           | Req/Opt | Description                                      |
| -------- | -------------- | ------: | ------------------------------------------------ |
| `Tokens` | \[]\*TokenNode |     Req | Raw token sequence preserved until a sync point. |
| `Span()` | Span           |       — | Covers all captured tokens.                      |
| `Kind()` | Kind           |       — | `KindBadTopLevel`.                               |

---

## Token Expectations (Header Happy Path)

Order and kinds expected by `parseHeader()`:

1. `KeywordCurrent`
2. `KeywordTurn`
3. `MonthName` **or** `Identifier`
4. `Number` (year)
5. `LParen`
6. `Hash`
7. `Number` (turn)
8. `RParen`

On mismatch at any step:

* Emit a diagnostic: *“expected X, found Y”* (with spans).
* **Synthesize** a zero-width token of the expected kind at the current insertion point.
* Continue parsing (no panic).

---

## Synchronization (Recovery) Points

When local synthesis isn’t enough (e.g., unexpected tokens accumulate), parsers may call `recoverTo` with context-specific sync kinds:

* Within `Header`: `RParen`, `KeywordCurrent`, or `EOF`.
* At top level: `KeywordCurrent` or `EOF`.

Recovery **discards tokens** until one of the sync kinds is seen (or `EOF`), then resumes normal parsing.

---

## Minimal Construction Rules

* **CST is lossless**: never drop tokens; represent unknown/stray ones inside `BadTopLevel` or keep them as unexpected until a sync point.
* **No semantic normalization** in CST (e.g., don’t map “August” → 8).
* **Synthesis is zero-width** and marks precise insertion sites for tools and diagnostics.

---

## Example (Happy Path)

Input:

```
Current Turn August 2025 (#123)
```

CST shape (abbrev):

```
File
 ├─ Header
 │   ├─ KwCurrent: Token(KeywordCurrent)
 │   ├─ KwTurn:    Token(KeywordTurn)
 │   ├─ Month:     Token(MonthName "August")
 │   ├─ Year:      Token(Number "2025")
 │   ├─ LParen:    Token("(")
 │   ├─ Hash:      Token("#")
 │   ├─ TurnNo:    Token(Number "123")
 │   └─ RParen:    Token(")")
 └─ EOF
```

All tokens retain their original **spans** and **trivia** (via `lexers.Token`).
