# Lexer Briefing Document

## Goals of the Lexer

The lexer’s job is to transform raw input text into a sequence of **tokens** that the parser can consume. Our lexer must support two primary goals:

1. **Excellent diagnostics**

    * Every token will carry **span information** (line, column, and byte offsets).
    * This allows us to highlight the exact portion of the input when reporting errors.

2. **Full fidelity CST construction**

    * Tokens will include both **significant tokens** (keywords, identifiers, numbers) and **trivia** (whitespace, comments).
    * The parser can then build a CST that preserves *everything the user wrote*, enabling round-tripping, pretty-printing, and context-rich error messages.

---

## Token Structure

Each token produced by the lexer will include:

* **Kind** – the type of token (e.g., `KeywordCurrent`, `Identifier`, `Number`).
* **Span** – the exact location in the input:

  ```go
  type Span struct {
      Start int // byte offset (inclusive)
      End   int // byte offset (exclusive)
      Line  int
      Col   int
  }
  ```
* **Text** – the raw text slice corresponding to the span.
* **Trivia** – any leading and trailing trivia (whitespace, comments) associated with the token.

Example definition:

```go
type Token struct {
    Kind   TokenKind
    Span   Span
    Text   string
    Trivia []Trivia
}
```

Trivia entries record their own spans and text:

```go
type Trivia struct {
    Kind TriviaKind
    Span Span
    Text string
}
```

---

## Example: “Current Turn August 2025 (#123)”

Input text:

```
Current Turn August 2025 (#123)
```

Lexed tokens (simplified view):

```
[0] Token{Kind: KeywordCurrent, Text: "Current", Span: [0..7]}
[1] Trivia{Kind: Whitespace, Text: " ", Span: [7..8]}
[2] Token{Kind: KeywordTurn, Text: "Turn", Span: [8..12]}
[3] Trivia{Kind: Whitespace, Text: " ", Span: [12..13]}
[4] Token{Kind: IdentifierMonth, Text: "August", Span: [13..19]}
[5] Trivia{Kind: Whitespace, Text: " ", Span: [19..20]}
[6] Token{Kind: Number, Text: "2025", Span: [20..24]}
[7] Trivia{Kind: Whitespace, Text: " ", Span: [24..25]}
[8] Token{Kind: LParen, Text: "(", Span: [25..26]}
[9] Token{Kind: Hash, Text: "#", Span: [26..27]}
[10] Token{Kind: Number, Text: "123", Span: [27..30]}
[11] Token{Kind: RParen, Text: ")", Span: [30..31]}
```

Notes:

* Whitespace is captured as **trivia**, not discarded.
* Each token knows exactly where it came from (`Span`), which means diagnostics can highlight text like `(#123)` precisely.
* The parser will be able to group trivia with surrounding tokens for CST preservation.

---

Great—here’s a compact “bad input” section you can drop into the lexer briefing. It shows exactly what the lexer will emit, how spans and trivia are preserved, and why that helps diagnostics and parser recovery.

---

# Bad-Input Examples (Diagnostics + Recovery)

## Token model (small refinement)

To avoid copying substrings, tokens carry **spans** into the original buffer. Text is resolved **lazily** (e.g., for debug printing or messages).

```go
type Span struct {
    Start int // byte offset (inclusive)
    End   int // byte offset (exclusive)
    Line  int // 1-based
    Col   int // 1-based, in UTF-8 code points
}

type Trivia struct {
    Kind TriviaKind // Whitespace, LineComment, BlockComment
    Span Span
}

type Token struct {
    Kind           TokenKind
    Span           Span
    LeadingTrivia  []Trivia
    TrailingTrivia []Trivia
}

// Helper for diagnostics / debugging:
func (t Token) Text(src []byte) string { return string(src[t.Span.Start:t.Span.End]) }
```

Notes:

* We distinguish **leading** vs **trailing** trivia to make CST placement deterministic.
* No eager substring copies; everything is span-based.

---

## Handling Bad Input

### Bad input #1: Misspelled keyword and month, malformed turn number

Input (line 1):

```
Currnet  Turn  Agust 2025 (##12a3
```

Issues:

* `Currnet` (misspelled “Current”)
* `Agust` (misspelled “August”)
* `##12a3` (extra `#`, alphanumeric number)
* Missing closing `)` at end of line

#### Lexer output (simplified)

```
[0] Token{Kind: Identifier, Span: [0..7]  L1:C1 }      // "Currnet"
    TrailingTrivia: [Whitespace [7..9] "  "]

[1] Token{Kind: KeywordTurn, Span: [9..13]  L1:C10}    // "Turn"
    TrailingTrivia: [Whitespace [13..15] "  "]

[2] Token{Kind: Identifier, Span: [15..20] L1:C16}     // "Agust"
    TrailingTrivia: [Whitespace [20..21] " "]

[3] Token{Kind: Number, Span: [21..25] L1:C22}         // "2025"
    TrailingTrivia: [Whitespace [25..26] " "]

[4] Token{Kind: LParen, Span: [26..27] L1:C27}         // "("

[5] Token{Kind: Hash, Span: [27..28] L1:C28}           // "#"
[6] Token{Kind: Hash, Span: [28..29] L1:C29}           // "#"

[7] Token{Kind: Identifier, Span: [29..33] L1:C30}     // "12a3"
    // End of line reached; no RParen produced
```

#### Why this helps diagnostics

Because the lexer **doesn’t collapse or discard** anything:

* We can point to the exact span of the typo (`Currnet`) and suggest a fix.
* We can flag `Agust` as an **unknown month** at parse time, but the span is ready now.
* We keep both `#` tokens, so the parser can diagnose “double # before turn number.”
* Missing `)` is detectable during parsing with a precise “expected ‘)’ here” pointing to end-of-line.

**Example diagnostic (parser using token spans):**

```
error: expected 'Current' keyword
  --> line 1, col 1
  1 | Currnet  Turn  Agust 2025 (##12a3
      ^^^^^^^
  hint: did you mean 'Current'?
```

**Another diagnostic:**

```
error: invalid month name 'Agust'
  --> line 1, col 16
  1 | Currnet  Turn  Agust 2025 (##12a3
                    ^^^^^
  hint: expected one of: January, February, ..., August, ...
```

**And for the turn number:**

```
error: invalid turn number token '12a3'
  --> line 1, col 30
  1 | Currnet  Turn  Agust 2025 (##12a3
                              ^^^^
  note: digits only are allowed after '#'
```

**Missing right paren (with recovery):**

```
error: missing ')' to close turn number
  --> line 1, col 34
  1 | Currnet  Turn  Agust 2025 (##12a3
                                   ^
  note: parser will continue from here
```

The parser’s recovery strategy (sync to next header, end-of-line, or a safe delimiter) is effective because **all** tokens and trivia (including the second `#`) survive lexing.

---

### Bad input #2: Extra punctuation, embedded comment, odd spacing

Input:

```
Current,,  Turn /*comment*/ August   2025  (#123))
```

#### Lexer output (simplified)

```
[0] Token{Kind: KeywordCurrent, Span: [0..7] L1:C1}
[1] Token{Kind: Comma, Span: [7..8] L1:C8}
[2] Token{Kind: Comma, Span: [8..9] L1:C9}
    TrailingTrivia: [Whitespace [9..11] "  "]

[3] Token{Kind: KeywordTurn, Span: [11..15] L1:C12}
    TrailingTrivia: [
        Whitespace   [15..16] " ",
        BlockComment [16..27] "/*comment*/",
        Whitespace   [27..28] " "
    ]

[4] Token{Kind: MonthName, Span: [28..34] L1:C29}      // "August"
    TrailingTrivia: [Whitespace [34..37] "   "]

[5] Token{Kind: Number, Span: [37..41] L1:C38}         // "2025"
    TrailingTrivia: [Whitespace [41..43] "  "]

[6] Token{Kind: LParen, Span: [43..44] L1:C44}         // "("
[7] Token{Kind: Hash, Span: [44..45] L1:C45}
[8] Token{Kind: Number, Span: [45..48] L1:C46}         // "123"
[9] Token{Kind: RParen, Span: [48..49] L1:C49}         // ")"
[10] Token{Kind: RParen, Span: [49..50] L1:C50}        // stray ")"
```

**Diagnostics examples enabled by spans + trivia:**

* “Unexpected `,` after `Current` (two commas in a row).”
* “Comment between `Turn` and month is allowed” (not an error), but if policy disallows it, we can point at `/*comment*/`.
* “Stray `)` after closing turn number.”

---

### Parser recovery (why keeping everything matters)

Because the lexer never throws away oddities:

* The parser can **skip** unexpected tokens (`Comma`, extra `RParen`) until a **sync point** (e.g., end of header line) is found.
* The CST can **retain** commas, comments, and spacing exactly as typed—useful for tools (pretty printers, formatters) and for testers who compare CST to input.

---

### Summary for implementers

* **Do not normalize** or drop unusual characters in the lexer; emit them as tokens (or a dedicated `Unknown` token) with spans.
* **Attach trivia** (leading/trailing) to the **nearest significant token**; this enables faithful CST and precise pointer ranges in errors.
* **Keep months/keywords as recognized tokens** when they match; produce `Identifier` (or `Unknown`) otherwise—never “fix” in the lexer.
* With spans on every token and trivia, downstream diagnostics can highlight **exact bytes** and offer contextual hints without guesswork.

---

## Approach and Benefits

* **Span-based tokens**

    * Instead of copying substrings early, we store spans into the original input.
    * This makes lexing faster and avoids duplication.
    * Substrings are resolved only when needed (e.g., for error messages).

* **Trivia included**

    * Preserving trivia enables us to reconstruct the original source if needed.
    * The CST can include spacing and comments, giving testers and users an exact mirror of what they typed.
    * Diagnostics can report issues like “unexpected comment here” or “extra whitespace before token.”

* **Diagnostics impact**

    * By combining `Span` with trivia, error messages can be precise:

      ```
      Error: expected month name, found "2025"
        at line 1, column 14
      ```
    * Even malformed inputs can be tokenized and represented faithfully, which lets the parser recover and keep going.

---

## Key Takeaways for Parser Developers

* **Never assume trivia is gone.** Tokens always carry it forward to the parser.
* **Use spans, not strings, to locate input.** This keeps errors precise and lightweight.
* **The CST layer depends on this design.** We keep everything so that later phases can decide what to ignore.
