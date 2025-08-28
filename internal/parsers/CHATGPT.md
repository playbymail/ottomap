# Parsing with diagnostics

Great idea—and you’re pulling on a few proven threads (Ball’s “collect errors, keep parsing”; typed entry points; explicit positions). Here’s feedback to help you keep the good parts and dodge some pitfalls.

## What’s strong

* **Keep-going parsing.** Returning a (partially) built node even when something’s wrong is exactly how you surface *many* errors in one run.
* **Explicit positions.** Passing `(line, col)` into `NewParser` is handy for fragments. Do also track **byte offset** for stable slicing and range reporting.
* **Public entry points.** Clear, testable non-terminals (e.g., `Header()`) are a win.

## What to adjust

### 1) Error nodes *vs.* error fields

Right now you have both `ErrorNode_t` **and** `Error() error` on every node. That’s redundant and forces lots of type-asserts.

Prefer one of these patterns:

**A. NodeBase with diagnostics (recommended)**

* No separate `ErrorNode_t`. Every node can hold 0..n diagnostics.
* Nodes are always the *thing they claim to be* (a `HeaderNode_t` is a header), even when partially filled.

```go
type Position_t struct {
    Line, Col int
    Offset    int   // byte offset
}
type Span_t struct { Start, End Position_t }

type Diagnostic_t struct {
    Span    Span_t
    Message string
    // Optional: Severity, Code, Hints []string
}

type Node_i interface {
    Span() Span_t
    Diagnostics() []Diagnostic_t
    String() string
}

type NodeBase_t struct {
    span        Span_t
    diagnostics []Diagnostic_t
}
func (b *NodeBase_t) Span() Span_t                   { return b.span }
func (b *NodeBase_t) Diagnostics() []Diagnostic_t    { return b.diagnostics }
func (b *NodeBase_t) addDiag(d Diagnostic_t)         { b.diagnostics = append(b.diagnostics, d) }
```

Now your AST stays strongly typed, and you don’t need an `ErrorNode_t` placeholder.

**B. Keep an `ErrorNode_t`, drop `Error()` from the interface**

* Use an actual node only when you *must* stand in for something missing.
* But you’ll still pay the type-assertion tax while walking.

Between the two, A is simpler to consume.

### 2) Reduce type assertions with a visitor

If you stick with `Node_i`, add a visitor so consumers don’t do `switch n := node.(type)`. Example:

```go
type Visitor_i interface {
    VisitHeader(*HeaderNode_t)
    VisitID(*IDNode_t)
    // ...
}

type Visitable_i interface {
    Accept(Visitor_i)
}
```

Have each concrete node implement `Accept`, call `v.VisitX(n)`.

### 3) Strongly-typed fields in nodes

In your example:

```go
type HeaderNode_t struct {
    NodeBase_t
    Tribe Node_i // better: *IdentifierNode_t (or *TokenNode_t{Kind: Tribe})
    ID    Node_i // better: *IntLiteralNode_t
}
```

Choose specific types wherever you can. **Generic `Node_i` fields push complexity to consumers.** If a field is sometimes missing, make it a pointer and add a diagnostic when it’s nil/invalid.

### 4) AST vs CST

Consider a light **CST (concrete syntax tree)** if you need exact tokens/whitespace for messages or reprinting. Then transform to a clean AST. Error collection can live in the CST pass (best messages) and flow through.

If you don’t need exact tokens, a single AST pass with `NodeBase_t` is fine.

### 5) Parser error strategy (how you “keep going”)

Good messages come from good recovery:

* **Expect set reporting.** When you miss, report: *expected {ID, INTEGER} but found '1a'*. Keep the FIRST/FOLLOW sets handy for each production so the parser knows what to sync on.
* **Synchronization points.** Newlines, commas, semicolons, closing braces—whatever makes sense in your grammar. On failure, skip until a sync token. Attach a diagnostic to the node you were building.
* **Single-token insertion/deletion heuristics.** If a missing comma or closing token is *very* likely, virtually insert it (with a diagnostic) and carry on.
* **Cap error cascades.** After N errors in the same production, resync more aggressively.

### 6) Unicode + spans

Columns are tricky with multibyte UTF-8. Consider:

* Track **byte offsets** for slicing, and compute line/column only for diagnostics (cache line starts to make it cheap).
* Or store both: `Span_t{Start: {Line, Col, Offset}, End: ...}`.

### 7) Panics

Keep panics strictly for **internal invariants**. For input problems, use diagnostics. Wrap panics with `recover` at top-level entry points to convert to a fatal diagnostic that points back to a bug; include a stack snippet in debug builds.

### 8) API surface

Entry points can return concrete types for convenience:

```go
func (p *Parser_t) Header() *HeaderNode_t
```

You can still satisfy `Node_i` via embedding `NodeBase_t`, but callers don’t need to downcast.

### 9) Pretty-printing

`String()` is handy for tests, but for users prefer a **formatter** that also prints diagnostics. Keep `String()` minimal; add a `Format(w io.Writer, opts ...)` for structured output.

### 10) Performance niceties (optional)

* Pool token slices and small nodes via `sync.Pool` if parsing huge files.
* Avoid copying `[]byte`—store token spans and slice on demand.

## Tiny end-to-end sketch

```go
type HeaderNode_t struct {
    NodeBase_t
    Tribe *IdentifierNode_t
    ID    *IntLiteralNode_t
}

func (p *Parser_t) Header() *HeaderNode_t {
    n := &HeaderNode_t{}
    start := p.pos()

    tr, ok := p.parseTribe()
    if !ok {
        n.addDiag(p.expectedDiag("Tribe"))
        p.syncToIDOrEol()
    }
    n.Tribe = tr

    id, ok := p.parseID()
    if !ok {
        n.addDiag(p.expectedDiag("integer ID"))
        // optionally synthesize a dummy IntLiteralNode_t so downstream is simpler
        id = &IntLiteralNode_t{ NodeBase_t: NodeBase_t{span: p.zeroWidthSpan()}, Value: 0 }
    }
    n.ID = id

    n.span = Span_t{Start: start, End: p.pos()}
    return n
}
```

This keeps:

* **Typed fields** (`*IdentifierNode_t`, `*IntLiteralNode_t`)
* **Diagnostics on the node** (no special error node)
* **Graceful recovery** and continued parsing

## Verdict

Your core idea—returning nodes even on errors to report many issues—is solid. The main refinement is to **embed diagnostics on every node and keep concrete types** so consumers don’t drown in type assertions. Add a visitor, adopt clear synchronization points, and you’ll get Monkey-style friendly errors without sacrificing structure or performance.

---

Yes — you’re not the first person to bump into the “PEG-style parsers give me garbage error messages, but I want a full AST and decent diagnostics” problem 🙂. There’s a whole little literature on it, scattered between compiler textbooks, papers, and blog posts. Some starting points:

---

## Classic compiler texts

* **Aho, Sethi, Ullman, *Compilers: Principles, Techniques, and Tools* (“Dragon Book”):**

    * Ch. 4.3 “Error Recovery” covers *panic-mode*, *phrase-level*, and *error productions*.
    * While old-school, it lays out the basic tradeoffs you’re circling (skip tokens vs. keep nodes with diagnostics).
* **Cooper & Torczon, *Engineering a Compiler***:

    * Their sections on syntax error recovery talk about attaching diagnostics while still building partial parse trees.
* **Appel, *Modern Compiler Implementation in ML/Java/C***:

    * More modern treatment, shows how to embed error tokens in AST and carry on.

---

## Papers

* **“Practical Error Recovery in LR Parsers” (Börger, 1989)** — one of the early attempts to formalize building usable trees while continuing after an error.
* **Elizabeth Scott & Adrian Johnstone, “Error recovery for LR parsers”** (multiple papers through the 1990s–2000s). They focus on algorithmic error recovery strategies, but the idea of *error nodes* in the syntax tree shows up here.
* **Grosch, “Efficient and Comfortable Error Recovery in Recursive Descent Parsers” (1992, Universität Karlsruhe)** — specifically about recursive descent, with design patterns similar to what you’re sketching.
* **Gingerich & Fisher, “Error Recovery in Parsing Expression Grammars” (2015)** — addresses PEGs (like Pigeon) and how to get less awful diagnostics.

---

## Blog posts / practical write-ups

* **Thorsten Ball, *Writing An Interpreter In Go*** (the Monkey interpreter you mentioned). His parser is a model for “collect errors, keep parsing, return partial ASTs.”
* **Rust compiler team blogs** (e.g. “Error recovery in the Rust parser”): Rust moved from panicking on first error to embedding diagnostics on AST nodes, with spans. They document the shift from error tokens to richer diagnostics.
* **TypeScript team design notes**: their parser intentionally builds a complete tree with *Missing* nodes. See their [“Compiler Internals” wiki](https://github.com/microsoft/TypeScript/wiki/Architectural-Overview#parser) — very close to your “ErrorNode” vs. “expected-but-not-found” idea.
* **Tree-sitter docs**: they talk about how they guarantee a full parse tree, even if parts are marked “ERROR,” to support editor tooling.
* **ANTLR v4 book (Parr)**: whole chapter on *error strategy objects* (`DefaultErrorStrategy`, `BailErrorStrategy`), which is another concrete embodiment of what you’re proposing.

---

## Keywords to search

If you want to dig deeper, search on:

* *“error recovery recursive descent parser”*
* *“diagnostic nodes abstract syntax tree”*
* *“error productions parser”*
* *“panic-mode recovery parser”*
* *“PEG parser error reporting”*
* *“tree-sitter error node”*

---

👉 If you want **Go-specific precedents**:

* Thorsten Ball’s Monkey interpreter (chapters 2–4).
* The [Go compiler’s own frontend](https://github.com/golang/go/tree/master/src/cmd/compile/internal/syntax) — it produces full ASTs with `BadExpr`, `BadDecl`, `BadStmt` nodes when syntax is invalid, rather than aborting. That’s almost exactly your “ErrorNode\_t.”

---

Would you like me to pull out **concrete design examples** from the Go compiler (`BadExpr`, `BadStmt`, etc.) and TypeScript parser (“Missing” nodes) so you can compare how they solved the `ErrorNode vs Diagnostics` question?

---

Here’s the Go AST version you can inspect and copy from:

```go
// go/ast
type BadExpr struct {
    From, To token.Pos // position range of bad expression
}
```

A `BadExpr` is “a placeholder for an expression containing syntax errors for which a correct expression node cannot be created.” ([Go Packages][1], [Go][2])

Notes:

* There are matching placeholders for other syntactic classes: `BadStmt` and `BadDecl`, each carrying a span. ([Go][2])
* The compiler’s *internal* `cmd/compile/internal/syntax` package also has a `BadExpr` (fields unexported) with the same intent: a node inserted when parsing fails so the tree stays complete. ([Go Packages][3])

This pattern maps directly to your idea: instead of panicking or returning `nil`, the parser emits a typed node with a precise source range you can report on later.

[1]: https://pkg.go.dev/go/ast?utm_source=chatgpt.com "ast package"
[2]: https://go.dev/pkg/go/ast/?m=old&utm_source=chatgpt.com "Package ast"
[3]: https://pkg.go.dev/cmd/compile/internal/syntax?utm_source=chatgpt.com "cmd/compile/internal/syntax"

---

Awesome—CSTs are a great fit for “small inputs, keep everything, give great errors.”

Here’s a compact guide to using a **Concrete Syntax Tree (CST)** in your setup, plus Go-ish sketches you can drop in.

# What a CST is (and why you might want one)

* **CST = every token, in order**: It keeps *all* tokens (including punctuation) and usually the “trivia” around them (whitespace, comments). It mirrors the grammar closely.
* **AST = meaning only**: It drops syntax sugar and punctuation and keeps semantic structure.

### Why parse to CST first?

* **Excellent diagnostics**: Every node is a precise span of the input; you can attach errors exactly where they occur (and still return a full tree).
* **Round-tripping**: You can pretty-print either (a) *exactly* the original text or (b) a formatted version, because you didn’t lose whitespace/comments.
* **Robust recovery**: You can insert explicit **Missing** tokens/nodes when something’s absent; downstream passes don’t crash.
* **Stable tooling**: Editors/linters love CSTs; small edits don’t cascade.

Cost: the CST is a bit heavier in memory and you’ll write one more pass (CST → AST). With your small inputs, that’s a non-issue.

---

# Suggested shape

## Tokens (with trivia)

```go
type Kind_t int

type Trivia_t struct {
    Kind   Kind_t   // Whitespace, LineComment, BlockComment
    Lexeme []byte
    Span   Span_t
}
type Token_t struct {
    Kind     Kind_t
    Lexeme   []byte   // you said copying is fine
    Span     Span_t   // byte offsets + (line,col) from lexer
    Leading  []Trivia_t
    Trailing []Trivia_t
}
```

> Keep line/col from the lexer; also store byte offsets in `Span_t` for slicing.

## CST nodes

CST nodes keep **tokens as children** (and sometimes nested nodes). They do not coerce types or drop punctuation.

```go
type Span_t struct{ Start, End Position_t } // (Line, Col, Offset) each

type Diagnostic_t struct {
    Span    Span_t
    Message string
    // Severity, Code, Hints []string (optional)
}

type CSTNode_i interface {
    Span() Span_t
    Diagnostics() []Diagnostic_t
}

type CSTBase_t struct {
    span        Span_t
    diagnostics []Diagnostic_t
}
func (b *CSTBase_t) Span() Span_t                { return b.span }
func (b *CSTBase_t) Diagnostics() []Diagnostic_t { return b.diagnostics }
func (b *CSTBase_t) addDiag(d Diagnostic_t)      { b.diagnostics = append(b.diagnostics, d) }
```

### Example: `header := "Tribe" ID "."`

```go
type CST_Header_t struct {
    CSTBase_t
    KwTribe Token_t      // required keyword
    ID      Token_t      // integer token
    Dot     Token_t      // "."
    // If something is missing, store a synthesized token (see below)
}
```

> The CST mirrors the grammar production: you literally see the keyword, the ID token, and the dot token.

---

# Error recovery with CST

### Missing & Bad placeholders

Instead of returning `nil`, **synthesize** placeholders:

* **Missing tokens**: zero-length span at the current position, special kind like `KindMissing`.
* **Bad tokens / nodes**: spans that cover the junk you skipped; add a diagnostic.

```go
func (p *Parser_t) synthMissing(kind Kind_t) Token_t {
    pos := p.pos()
    return Token_t{
        Kind: kind, Lexeme: nil,
        Span: Span_t{Start: pos, End: pos}, // zero-width
    }
}
```

### Sync points

For a production, define where you can safely resume (e.g., end of line, comma, right-paren). On failure:

1. Add a diagnostic: `expected integer ID`.
2. Option A: insert `MissingID` token (preferred).
3. Option B: create a `Bad` subnode/token that spans the junk you skipped.
4. Skip to sync and continue.

Because you keep **every token**, the CST remains structurally valid and walkable.

---

# From CST to AST (the “lowering” pass)

The second pass transforms the verbose CST into a cleaner AST:

* Coerce types (`ID` → `int`).
* Drop punctuation and trivia.
* Decide policy for errors: either

    * carry CST diagnostics up, or
    * re-issue AST-level diagnostics (usually you propagate).

### Example AST types

```go
type Node_i interface {
    Span() Span_t
    Diagnostics() []Diagnostic_t
    String() string
}
type HeaderNode_t struct {
    NodeBase_t
    Tribe string // Or a domain type
    ID    int
}
```

### Lowering function

```go
func LowerHeader(cst *CST_Header_t) *HeaderNode_t {
    n := &HeaderNode_t{}
    n.SetSpan(cst.Span())
    n.AddDiagnostics(cst.Diagnostics()...)

    // Extract semantics
    n.Tribe = "Tribe" // keyword is fixed; or derive from token if needed

    // Convert ID token -> int
    if cst.ID.Kind == KindInt {
        if v, err := parseInt(cst.ID.Lexeme); err == nil {
            n.ID = v
        } else {
            n.AddDiag(Diagnostic_t{Span: cst.ID.Span, Message: "invalid integer"})
        }
    } else if cst.ID.Kind == KindMissing {
        n.AddDiag(Diagnostic_t{Span: cst.ID.Span, Message: "missing ID"})
    } else {
        n.AddDiag(Diagnostic_t{Span: cst.ID.Span, Message: "expected integer ID"})
    }

    return n
}
```

> The AST is now pleasant to consume—no punctuation or trivia—while diagnostics remain precise.

---

# API shape

* **Lexer**: returns `Token_t` with (line, col, offset) and trivia.
* **Parser**: returns a **CST root** and a top-level diagnostics slice (also stored on nodes).

  ```go
  func NewParser(line, col int, input []byte) *Parser_t
  func (p *Parser_t) Header() *CST_Header_t
  ```
* **Lowerer**: CST → AST

  ```go
  func LowerHeader(*CST_Header_t) *HeaderNode_t
  ```

---

# Pretty-printing / round-tripping

Because the CST holds all tokens + trivia:

* **Lossless reprint**: Concatenate `Leading trivia + token + trailing trivia` for each token in order.
* **Formatted print**: Ignore original trivia and emit normalized whitespace from the AST (or reflow CST tokens with a formatter that discards old trivia).

You can offer both: `cst.ReprintExact()` and `ast.Format(opts)`.

---

# Testing strategy

1. **Golden tests** on CST tokenization (including trivia capture).
2. **Parse error tests**: small inputs with known diagnostics (assert message, span).
3. **Lowering tests**: snapshot AST + propagated diagnostics.
4. **Round-trip**: `parse → CST → exact print` equals original (modulo newline normalization).

---

# When to choose CST→AST vs AST-only

* **Choose CST→AST** if you want excellent error messages, IDE-friendly trees, or need to preserve comments/formatting.
* **Choose AST-only** if inputs are tiny *and* you never need round-tripping or comments—then embed diagnostics on AST nodes and skip the CST.

Given your goals (“parse whole input, great errors”), **CST→AST** is the sweet spot.

---

If you’d like, I can sketch a minimal working parser function for the `header` rule that:

* inserts `Missing` tokens,
* uses a sync set,
* returns a `CST_Header_t`,
* and a tiny `LowerHeader` that produces your `HeaderNode_t`.

---

Here’s a clean Markdown document you can drop straight into your repo or notes.

---

# Parsing Strategy: CST → AST with Diagnostics

This document describes a parsing strategy where the parser first constructs a **Concrete Syntax Tree (CST)** that mirrors the grammar and retains *all* tokens (including punctuation, comments, and whitespace). A second pass then lowers the CST into a simplified **Abstract Syntax Tree (AST)** suitable for semantic analysis.

This approach is designed to produce **excellent error diagnostics** while preserving the ability to continue parsing after errors.

---

## Assumptions

- Inputs are small enough that copying byte slices for tokens is acceptable.
- The lexer returns tokens with `(line, col, offset)` so the parser does not need to handle UTF-8 character widths directly.
- The parser’s job is to build a complete CST, inserting **Missing** or **Bad** tokens when input is incomplete or incorrect.
- Diagnostics (error messages) are attached to nodes in the CST, and propagated to the AST during lowering.

---

## Key Concepts

### Concrete Syntax Tree (CST)

- Mirrors the grammar rules exactly.
- Contains every token (keywords, punctuation, identifiers).
- Retains **trivia** (comments, whitespace).
- Uses synthesized tokens for **missing** elements when errors occur.
- Attaches diagnostics with precise spans.

Example CST for a grammar rule:

```

header := "Tribe" ID "."

````

```go
type CST_Header_t struct {
    CSTBase_t
    KwTribe Token_t
    ID      Token_t
    Dot     Token_t
}
````

### Abstract Syntax Tree (AST)

* Simplified structure representing program meaning.
* Drops punctuation and trivia.
* Uses domain-specific types (`int`, `string`, etc.).
* Diagnostics are carried forward from the CST.

Example AST node:

```go
type HeaderNode_t struct {
    NodeBase_t
    Tribe string
    ID    int
}
```

---

## Tokens

```go
type Kind_t int

const (
    KindTribe Kind_t = iota
    KindInt
    KindDot
    KindMissing
    // ...
)

type Token_t struct {
    Kind   Kind_t
    Lexeme []byte
    Span   Span_t
}

type Span_t struct {
    Line, Col, Offset int
}
```

---

## Diagnostics

```go
type Diagnostic_t struct {
    Span    Span_t
    Message string
}

type CSTBase_t struct {
    span        Span_t
    diagnostics []Diagnostic_t
}
```

---

## Minimal Parser Function (Header Rule)

This example shows how to parse `header := "Tribe" ID "."`.

```go
func (p *Parser_t) Header() *CST_Header_t {
    start := p.pos()
    n := &CST_Header_t{}

    // Expect "Tribe"
    if tok := p.next(); tok.Kind == KindTribe {
        n.KwTribe = tok
    } else {
        n.KwTribe = p.synthMissing(KindTribe)
        n.addDiag(Diagnostic_t{Span: tok.Span, Message: "expected 'Tribe'"})
        p.syncTo(KindInt, KindDot)
    }

    // Expect ID (integer)
    if tok := p.peek(); tok.Kind == KindInt {
        n.ID = p.next()
    } else {
        n.ID = p.synthMissing(KindInt)
        n.addDiag(Diagnostic_t{Span: p.pos(), Message: "expected integer ID"})
        p.syncTo(KindDot)
    }

    // Expect "."
    if tok := p.peek(); tok.Kind == KindDot {
        n.Dot = p.next()
    } else {
        n.Dot = p.synthMissing(KindDot)
        n.addDiag(Diagnostic_t{Span: p.pos(), Message: "expected '.'"})
    }

    n.span = Span_t{Line: start.Line, Col: start.Col, Offset: start.Offset}
    return n
}
```

---

## Lowering Function

```go
func LowerHeader(cst *CST_Header_t) *HeaderNode_t {
    n := &HeaderNode_t{}
    n.SetSpan(cst.Span())
    n.AddDiagnostics(cst.Diagnostics()...)

    if cst.ID.Kind == KindInt {
        if v, err := strconv.Atoi(string(cst.ID.Lexeme)); err == nil {
            n.ID = v
        } else {
            n.AddDiag(Diagnostic_t{Span: cst.ID.Span, Message: "invalid integer"})
        }
    }
    n.Tribe = "Tribe"
    return n
}
```

---

## Benefits

* **Robust error recovery:** parser continues after an error, inserting placeholders.
* **High-quality diagnostics:** precise spans and multiple errors per parse.
* **Round-tripping:** CST preserves original formatting and comments.
* **Simplified AST:** second pass produces clean semantic nodes.

---

## References

* Aho, A.V., Sethi, R., Ullman, J.D. *Compilers: Principles, Techniques, and Tools* (Dragon Book), Ch. 4.3 Error Recovery.
* Parr, T. *The Definitive ANTLR 4 Reference* (error strategy objects).
* Grosch, J. (1992). *Efficient and Comfortable Error Recovery in Recursive Descent Parsers*.
* Scott, E., Johnstone, A. (2000s). *Error Recovery for LR Parsers*.
* Rust compiler internals: error recovery with spans and diagnostics.
* TypeScript compiler architecture: “Missing” nodes in CST.
* Go compiler source: `BadExpr`, `BadStmt`, `BadDecl` nodes in `go/ast`.

---

Would you like me to extend this doc with a **worked example input** (like `Tribe 42.` and `Tribe X.`) showing the CST structure, the diagnostics, and the final AST? That might make the strategy easier to “see” for non-parser folks.

---

