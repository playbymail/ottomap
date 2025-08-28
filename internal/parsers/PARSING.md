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
```

```go
type CST_Header_t struct {
    CSTBase_t
    KwTribe Token_t
    ID      Token_t
    Dot     Token_t
}
```

### Abstract Syntax Tree (AST)

- Simplified structure representing program meaning.
- Drops punctuation and trivia.
- Uses domain-specific types (`int`, `string`, etc.).
- Diagnostics are carried forward from the CST.

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

## Worked Examples

### Example 1: Correct input

Input:
```
Tribe 42.
```

CST:
- `KwTribe`: Token("Tribe")
- `ID`: Token("42")
- `Dot`: Token(".")
- Diagnostics: none

AST:
```go
HeaderNode_t{ Tribe: "Tribe", ID: 42 }
```

Diagnostics: none

---

### Example 2: Error input

Input:
```
Tribe X.
```

CST:
- `KwTribe`: Token("Tribe")
- `ID`: Missing Token (KindMissing) at position after "Tribe"
- Diagnostic: "expected integer ID" at that span
- `Dot`: Token(".")

AST:
```go
HeaderNode_t{ Tribe: "Tribe", ID: 0 }
```

Diagnostics:
- "expected integer ID"

---

## Benefits

- **Robust error recovery:** parser continues after an error, inserting placeholders.
- **High-quality diagnostics:** precise spans and multiple errors per parse.
- **Round-tripping:** CST preserves original formatting and comments.
- **Simplified AST:** second pass produces clean semantic nodes.

---

## References

- Aho, A.V., Sethi, R., Ullman, J.D. *Compilers: Principles, Techniques, and Tools* (Dragon Book), Ch. 4.3 Error Recovery.
- Parr, T. *The Definitive ANTLR 4 Reference* (error strategy objects).
- Grosch, J. (1992). *Efficient and Comfortable Error Recovery in Recursive Descent Parsers*.
- Scott, E., Johnstone, A. (2000s). *Error Recovery for LR Parsers*.
- Rust compiler internals: error recovery with spans and diagnostics.
- TypeScript compiler architecture: “Missing” nodes in CST.
- Go compiler source: `BadExpr`, `BadStmt`, `BadDecl` nodes in `go/ast`.
