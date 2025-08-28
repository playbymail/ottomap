# CST Parser: Developer Specification

## Scope and Objectives

* Build a **lossless CST** (concrete syntax tree) from tokens produced by `lexers.Lexer`.
* **Preserve trivia** (leading/trailing) and **spans** on every token so we can:

    * round-trip or pretty-print with fidelity,
    * produce precise diagnostics,
    * enable robust parser recovery.
* Remain **panic-free**: all errors are surfaced via diagnostics; we attempt recovery and continue.

---

## External Interfaces

### Package and entry points

```go
package cst

import "github.com/yourorg/yourrepo/lexers"

// ParseFile is the main entry point for one logical input (e.g., a report file).
func ParseFile(input []byte) (*File, []Diagnostic)
```

* `ParseFile` must **not** panic. On any error, it returns a CST (possibly with *Bad* nodes) and one or more `Diagnostic`s.
* The parser internally constructs `*Parser` and drives `lexers.New(1, 1, input)`.

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
    Span     Span        // byte offsets + line/col (copied from/derived via tokens)
    Message  string
    Notes    []string    // optional hints
}
```

Diagnostics always point to *something concrete* (token span, synthesized insertion point, or EOF).

---

## Token Consumption Model

* The lexer API:

  ```go
  // New starts at (line, col) with the provided input.
  func New(line, col int, input []byte) *Lexer

  // Next returns the next token, or nil on EOF.
  func (lx *Lexer) Next() *Token
  ```

* The parser wraps `Next()` with a **single-token lookahead** buffer:

  ```go
  type Parser struct {
      lx      *lexers.Lexer
      la      *lexers.Token   // lookahead (nil only at hard EOF)
      input   []byte
      diags   []Diagnostic
  }
  ```

* Invariants:

    * After construction we immediately prime `p.la = lx.Next()`.
    * `p.at(kind)` tests `p.la != nil && p.la.Kind == kind`.
    * `p.bump()` returns the current `la`, then refills `la` with `Next()` (or nil at EOF).
    * `p.expect(kind)` attempts to consume `kind`; on mismatch, it **synthesizes** a token node (zero-length span at the insertion point), emits a diagnostic, and does not advance input (see “Synthesis” below).

---

## CST Data Model

All CST nodes implement:

```go
type Node interface {
    Span() Span    // full span covering this node (computed from children/tokens)
    Kind() Kind    // node kind (Header, File, Token, Bad*, etc.)
}
```

### Spans

```go
type Span struct {
    Start int // byte offset (inclusive)
    End   int // byte offset (exclusive)
    Line  int // 1-based
    Col   int // 1-based (UTF-8 code-points)
}
```

> Implementation note: node spans are computed as `min(child.Start)`..`max(child.End)`; for synthetic zero-width insertions, `Start == End` at the insertion site.

### Token nodes

We **wrap** `*lexers.Token` to keep trivia and raw spans:

```go
type TokenNode struct {
    Tok *lexers.Token // carries Kind, Span, LeadingTrivia, TrailingTrivia
}

func (t *TokenNode) Span() Span { return convertSpan(t.Tok.Span) }
func (t *TokenNode) Kind() Kind { return KindToken } // or reflect subkind if needed
```

> Rationale: we do not copy text here; `Tok.Text(input)` is used lazily for messages/tools.

### File and top-level

```go
type File struct {
    Decls  []Node       // headers, sections, etc. (for this doc, we only show Header)
    EOF    *TokenNode   // EOF sentinel (synthetic node that wraps nil span at end)
    span   Span
}

func (f *File) Span() Span { return f.span }
func (f *File) Kind() Kind { return KindFile }
```

### Example: `Header` node (our “Current Turn …” line)

Grammar (CST-level; lexical kinds are inputs):

```
Header
  := kwCurrent kwTurn Month Year lparen hash Number rparen
```

* “Month” is a **token node**; we do not normalize in CST. The lexer will produce `MonthName` for valid names; otherwise it might be `Identifier`. CST accepts either; semantic normalization is AST’s job.

Go shape:

```go
type Header struct {
    KwCurrent *TokenNode // KeywordCurrent
    KwTurn    *TokenNode // KeywordTurn
    Month     *TokenNode // MonthName | Identifier
    Year      *TokenNode // Number
    LParen    *TokenNode // "("
    Hash      *TokenNode // "#"
    TurnNo    *TokenNode // Number
    RParen    *TokenNode // ")"
    span      Span
}

func (h *Header) Span() Span { return h.span }
func (h *Header) Kind() Kind { return KindHeader }
```

> Trivia: stays attached to each token via `Tok.LeadingTrivia` / `Tok.TrailingTrivia`. We do **not** duplicate trivia as separate nodes.

---

## Happy-Path Parsing

### High-level driver

```go
func ParseFile(input []byte) (*File, []Diagnostic) {
    p := &Parser{
        lx:    lexers.New(1, 1, input),
        input: input,
    }
    p.la = p.lx.Next()

    f := &File{}
    for p.la != nil {
        // For now we only recognize Header at top-level.
        // Future: switch on first-token kind to select section parsers.
        if p.at(lexers.KeywordCurrent) {
            hdr := p.parseHeader()
            f.Decls = append(f.Decls, hdr)
        } else {
            // Unknown top-level token: produce Bad node with recovery.
            bad := p.parseBadTopLevel()
            f.Decls = append(f.Decls, bad)
        }
    }
    f.EOF = p.synthesizeEOF()
    f.span = cover(f.Decls, f.EOF)
    return f, p.diags
}
```

### Header parser (strict happy path)

```go
func (p *Parser) parseHeader() *Header {
    h := &Header{}
    h.KwCurrent = p.want(lexers.KeywordCurrent) // strict; diag on miss; synthesize
    h.KwTurn    = p.want(lexers.KeywordTurn)
    h.Month     = p.wantOneOf(lexers.MonthName, lexers.Identifier)
    h.Year      = p.want(lexers.Number)
    h.LParen    = p.want(lexers.LParen)
    h.Hash      = p.want(lexers.Hash)
    h.TurnNo    = p.want(lexers.Number)
    h.RParen    = p.want(lexers.RParen)
    h.span      = cover(h.KwCurrent, h.RParen)
    return h
}
```

Helper semantics:

* `want(kind)`:

    * If `at(kind)`: `return wrap(bump())`.
    * Else: emit diagnostic “expected {kind}, found {foundKind}”; **synthesize** a zero-width `TokenNode` of `kind` at the current insertion point; return it (does **not** consume input). (See *Synthesis*.)
* `wantOneOf(k1, k2, …)`:

    * If matches any: consume and return actual token.
    * Else: diag; synthesize `k1` (or a designated representative kind).

### Example: happy path for

```
Current Turn August 2025 (#123)
```

Token flow (from lexer), abbreviated kinds:

```
KeywordCurrent, KeywordTurn, MonthName("August"), Number("2025"),
LParen, Hash, Number("123"), RParen
```

`parseHeader()` returns a `*Header` whose token fields wrap those tokens. `File` contains one `Header` and an `EOF` sentinel.

---

## Error Handling and Recovery

We separate **local synthesis** (to satisfy an expected token) from **synchronization** (skipping unexpected tokens until a safe point).

### 1) Token Synthesis (local, non-consuming)

When an expected token is missing, we create a zero-length token node and emit a diagnostic. This lets CST shape remain intact and predictable.

```go
func (p *Parser) synthToken(kind lexers.TokenKind) *TokenNode {
    // insertion point is at current lookahead; if la==nil, we anchor at EOF position
    span := insertionSpan(p.la)           // zero-width span
    tok  := &lexers.Token{Kind: kind, Span: convertBack(span)}
    return &TokenNode{Tok: tok}
}

func (p *Parser) expect(kind lexers.TokenKind) *TokenNode {
    if p.at(kind) {
        return wrap(p.bump())
    }
    p.errorExpected(kind, p.la)
    return p.synthToken(kind)
}
```

**Why zero-width?** It allows tooling to display “we inserted `)` here” with a caret between characters, and it does not distort neighboring real token spans.

### 2) Synchronization (skipping unexpected input)

When the parser observes an **unexpected** token (e.g., a stray comma after `Current`), it:

* emits a diagnostic for the unexpected token (and possibly surrounding context),
* consumes tokens until hitting a **sync point**.

**Sync points** (for header context):

* `lexers.RParen` (close of `(#N)` group),
* `lexers.KeywordCurrent` (start of another header),
* **newline** if exposed by the lexer as a trivia event or token (we recommend line trivia on tokens; if so, sync to `RParen` or `KeywordCurrent` or EOF),
* `EOF`.

Implementation pattern:

```go
func (p *Parser) recoverTo(sync ...lexers.TokenKind) {
    for p.la != nil {
        if p.atAny(sync...) { return }
        p.bump() // discard
    }
}
```

Use this in a specific context, e.g., inside `parseHeader()` after multiple failures:

```go
// Example: too many errors in header; skip to end of header or next header start
p.recoverTo(lexers.RParen, lexers.KeywordCurrent)
```

### 3) *Bad* nodes

When we can’t confidently construct a valid node shape (e.g., we lose the opening keyword), we build a **Bad** node that captures the raw token run so the tree remains complete and testable.

```go
type BadTopLevel struct {
    Tokens []*TokenNode // raw tokens preserved
    span   Span
}
func (b *BadTopLevel) Span() Span { return b.span }
func (b *BadTopLevel) Kind() Kind { return KindBadTopLevel }
```

`parseBadTopLevel()` typically grabs tokens until a sync point (start of a recognizable construct or EOF).

---

## Concrete Error Examples (Header)

Below, `•` indicates insertion/caret position for synthesized tokens.

### A) Misspelled `Current` and month; missing `)`

Input:

```
Currnet  Turn  Agust 2025 (##12a3
```

Behavior:

* `parseHeader()` is entered only if we see `KeywordCurrent`. Here we **don’t**; top-level driver won’t call `parseHeader()`. We instead build a `BadTopLevel` until we hit a sync point (`KeywordCurrent` or EOF). Since none appears, we collect the whole line and diag:

    * “expected ‘Current’, found Identifier ‘Currnet’ at line 1 col 1”
    * “garbled top-level; looking for header start”

Alternative (if you want to be *aggressive*):

* Allow `Identifier` at top level to *tentatively* try `parseHeader()` and let `want(KeywordCurrent)` synthesize `Current` before the identifier. This often produces more targeted diagnostics:

Emitted diagnostics (aggressive mode):

* Inserted `Current` at `•Currnet` (zero width at line start).
* Expected `Month`, found `Identifier("Agust")` — **do not** error here at CST level; simply accept `Identifier` via `wantOneOf(MonthName, Identifier)` and leave it to AST-phase to validate. If you *do* wish to warn here, use a **warning** severity (“unknown month name”).
* For `##12a3`, you will consume `Hash` then accept `Identifier("12a3")` only if you loosen `TurnNo` to `Identifier|Number`; **recommended** is to keep CST strict (`Number` required), produce:

    * “invalid turn number token ‘12a3’ (digits only)”
    * synthesis of a `Number` token at `•12a3` (zero-width), **without** consuming the bad identifier (then enter recovery that discards up to newline/EOF).
* Missing `)`:

    * After consuming as much as possible, if `RParen` is not present, synthesize `)` at end of line; message: “missing ‘)’ to close turn number”.

### B) Extra commas and stray `)`:

Input:

```
Current,,  Turn /*comment*/ August   2025  (#123))
```

Behavior:

* After `KeywordCurrent`, we see `Comma`. That is **unexpected** in this position:

    * Diagnostic: “unexpected ‘,’ after ‘Current’”.
    * Recovery: bump until we see `KeywordTurn` (safe local sync), but also keep a budget (don’t skip too far).
* Everything else parses; after `RParen` we see another `RParen`:

    * Diagnostic: “extra ‘)’ after closing turn number”.
    * Strategy: consume the stray `)` and continue; no synthesis needed.

---

## Helpers (recommended minimal set)

```go
func (p *Parser) at(k lexers.TokenKind) bool
func (p *Parser) atAny(ks ...lexers.TokenKind) bool
func (p *Parser) bump() *lexers.Token
func (p *Parser) want(k lexers.TokenKind) *TokenNode
func (p *Parser) wantOneOf(k ...lexers.TokenKind) *TokenNode
func (p *Parser) synthToken(k lexers.TokenKind) *TokenNode
func (p *Parser) errorExpected(k lexers.TokenKind, found *lexers.Token)
func (p *Parser) recoverTo(sync ...lexers.TokenKind)
func (p *Parser) synthesizeEOF() *TokenNode
```

**Sane defaults:**

* `want()` always emits a **single** diagnostic on mismatch and returns a synthetic token.
* `wantOneOf()` emits a diagnostic listing the set.
* Recovery functions must be **contextual**: the caller decides the sync set.
* A **per-production error budget** (e.g., 3 errors) prevents cascades; on exceed, escalate to recovery.

---

## “Current Turn …” Walkthroughs

### 1) Perfect input

```
Current Turn August 2025 (#123)
```

* Tokens: `KeywordCurrent, KeywordTurn, MonthName, Number, LParen, Hash, Number, RParen`
* CST: `File{ Decl: [Header{KwCurrent, KwTurn, Month, Year, LParen, Hash, TurnNo, RParen}], EOF }`
* No diagnostics.

### 2) Missing `#`

```
Current Turn August 2025 (123)
```

* At `Hash` position, `want(Hash)` synthesizes a zero-width `Hash` at `•`.
* Diagnostic: “expected ‘#’ before turn number”.
* No recovery needed; continue normally.

### 3) Non-digit turn

```
Current Turn August 2025 (#12a3)
```

* `want(Number)` mismatches on `Identifier("12a3")` → synthesize `Number` (zero-width), diag “invalid turn number”.
* Recovery choice:

    * **Simple**: leave the bad `Identifier` in the stream, then attempt to parse `RParen` (likely mismatches and triggers one more synth + diag).
    * **Better**: consume the bad identifier with a one-token recovery (“skipped invalid token in turn number”).

We prefer **better**: consume the invalid token (with a “skipped …” note), then continue, so you don’t stack follow-on errors.

---

## Trivia Preservation

* Trivia exists only on **tokens**. CST nodes reference tokens; we do **not** lift trivia to separate nodes at CST layer.
* Formatter/pretty-printer works off the token sequence, reading `LeadingTrivia`/`TrailingTrivia`.
* This design allows the CST to remain structurally clean while still being **lossless**.

---

## Testing Strategy

* **Golden tests**: for each input, assert:

    * serialized CST shape (node kinds + token kinds + key spans),
    * diagnostic list (message prefixes + spans),
    * round-trip token stream including trivia equals original input (optional tool).
* **Error injection tests**:

    * single missing token per slot in `Header`,
    * unknown tokens inserted between every pair,
    * malformed numbers/months,
    * large recovery: skip to next `KeywordCurrent`.

---

## Performance Notes

* Single lookahead is sufficient for the current grammar.
* Zero allocations for token text (spans only). `Token.Text(input)` only when formatting diags or tools.
* Avoid quadratic behavior in recovery: implement max skip budget per recovery call (e.g., stop after 256 tokens and emit a throttle diagnostic).

---

## Extensibility Hooks

* New sections (beyond `Header`) should:

    * define a **CST node** with only **token fields and child nodes**,
    * implement a strict **happy-path** using `want()/wantOneOf()`,
    * specify **local sync points** for recovery,
    * never normalize semantics (that belongs in AST).

---

### Appendix: Minimal Parser Skeleton (compilable)

```go
type Kind int

const (
    KindFile Kind = iota
    KindHeader
    KindBadTopLevel
    KindToken
)

// cover computes span of nodes/tokens (nil-safe).
func cover(parts ...interface{}) Span {
    var s Span
    first := true
    for _, p := range parts {
        if p == nil { continue }
        var ps Span
        switch v := p.(type) {
        case *TokenNode:
            ps = convertSpan(v.Tok.Span)
        case Node:
            ps = v.Span()
        }
        if first {
            s = ps
            first = false
        } else {
            if ps.Start < s.Start { s.Start = ps.Start }
            if ps.End   > s.End   { s.End = ps.End }
        }
    }
    return s
}

func (p *Parser) want(k lexers.TokenKind) *TokenNode {
    if p.at(k) { return &TokenNode{Tok: p.bump()} }
    p.errorExpected(k, p.la)
    return p.synthToken(k)
}

func (p *Parser) wantOneOf(ks ...lexers.TokenKind) *TokenNode {
    for _, k := range ks {
        if p.at(k) { return &TokenNode{Tok: p.bump()} }
    }
    p.errorExpectedSet(ks, p.la)
    // choose the 0th as representative for synthesis
    return p.synthToken(ks[0])
}
```

(Left out: trivial helpers `at`, `bump`, `recoverTo`, and diag formatters for brevity—they follow the semantics described above.)
