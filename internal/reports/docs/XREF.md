# CST ⇄ AST Crosswalk (Domain Language Cheat Sheet)

## Legend

* **CK** = CST token kind (from `lexers.TokenKind`)
* **T** = `*cst.TokenNode`
* **Span** = use the CST token’s span when emitting AST diagnostics
* **Diag Codes** = AST diagnostic `Code` values (e.g., `E_MONTH`)

---

## Header: `Current Turn <Month> <Year> (#<Turn>)`

### 1) Month

| CST Source         | CK Accepted                 | Example Text             | AST Target            | Transform                       | Diagnostics (Code)   | Notes                                                   |
| ------------------ | --------------------------- | ------------------------ | --------------------- | ------------------------------- | -------------------- | ------------------------------------------------------- |
| `Header.Month` (T) | `MonthName` \| `Identifier` | `August`, `Agust`, `Aug` | `Header.RawMonth`     | `Raw = tokenText()`             | —                    | Always copy raw                                         |
|                    |                             |                          | `Header.Month` (int)  | `normalizeMonth(Raw)` → `1..12` | `E_MONTH` if unknown | Accept full names + 3-letter abbrevs (case-insensitive) |
|                    |                             |                          | `Header.UnknownMonth` | `true` if unknown               | —                    | Set `Month=0` when unknown                              |

**Examples**

* `August` → `Month=8`, `UnknownMonth=false`
* `Agust` → `Month=0`, `UnknownMonth=true`, diag at `Month` span: `unknown month "Agust"` (`E_MONTH`)

---

### 2) Year

| CST Source        | CK Accepted | Example Text | AST Target           | Transform                                   | Diagnostics (Code)             | Notes                             |
| ----------------- | ----------- | ------------ | -------------------- | ------------------------------------------- | ------------------------------ | --------------------------------- |
| `Header.Year` (T) | `Number`    | `2025`       | `Header.RawYear`     | `Raw = tokenText()`                         | —                              | Always copy raw                   |
|                   |             |              | `Header.Year` (int)  | `parseYear(Raw)`                            | `E_YEAR` if non-numeric        | Trim whitespace prior to parse    |
|                   |             |              |                      | `validYearRange(y)` (default `[1900,2200]`) | `E_YEAR_RANGE` if out-of-range | Range is policy; adjust as needed |
|                   |             |              | `Header.InvalidYear` | `true` if parse fails or out-of-range       | —                              | Use the token’s span              |

**Examples**

* `2025` → `Year=2025`, `InvalidYear=false`
* `202X` → `Year=0`, `InvalidYear=true`, diag `invalid year "202X"` (`E_YEAR`)
* `3025` → `Year=3025`, diag `year 3025 out of range` (`E_YEAR_RANGE`)

---

### 3) Turn

| CST Source          | CK Accepted | Example Text | AST Target           | Transform               | Diagnostics (Code)      | Notes                |
| ------------------- | ----------- | ------------ | -------------------- | ----------------------- | ----------------------- | -------------------- |
| `Header.TurnNo` (T) | `Number`    | `123`        | `Header.RawTurn`     | `Raw = tokenText()`     | —                       | Always copy raw      |
|                     |             |              | `Header.Turn` (int)  | `parsePositiveInt(Raw)` | `E_TURN` if non-numeric | Require `Turn ≥ 1`   |
|                     |             |              |                      |                         | `E_TURN` if `< 1`       |                      |
|                     |             |              | `Header.InvalidTurn` | `true` if invalid       | —                       | Use the token’s span |

**Examples**

* `123` → `Turn=123`, `InvalidTurn=false`
* `12a3` → `Turn=0`, `InvalidTurn=true`, diag `invalid turn number "12a3"` (`E_TURN`)
* `0` → `Turn=0`, `InvalidTurn=true`, diag `invalid turn number "0"` (`E_TURN`)

---

### 4) Structural Tokens (Hash/Parens)

| CST Source                        | CK                  | Example Text | AST Impact      | Suggested Policy                | Diagnostics (Code) | Notes                                                            |
| --------------------------------- | ------------------- | ------------ | --------------- | ------------------------------- | ------------------ | ---------------------------------------------------------------- |
| `Header.Hash`                     | `Hash`              | `#`          | None (semantic) | Optional warning if synthesized | `W_MISSING_HASH`   | Only if you want AST to surface structure already handled in CST |
| `Header.LParen` / `Header.RParen` | `LParen` / `RParen` | `(` / `)`    | None            | —                               | —                  | Usually CST-only concern                                         |

**Synthesis Awareness**

* If CST **synthesized** a token (zero-width span), AST may optionally emit a **warning** to keep users informed at a semantic level. This is policy; many teams defer structure-only issues to CST.

---

## Full Row Example

**Input**

```
Current Turn August 2025 (#123)
```

**CST → AST**

| Field | CST Token Text | AST Raw             | AST Value   | Flags                | Diagnostics |
| ----- | -------------- | ------------------- | ----------- | -------------------- | ----------- |
| Month | `August`       | `RawMonth="August"` | `Month=8`   | `UnknownMonth=false` | —           |
| Year  | `2025`         | `RawYear="2025"`    | `Year=2025` | `InvalidYear=false`  | —           |
| Turn  | `123`          | `RawTurn="123"`     | `Turn=123`  | `InvalidTurn=false`  | —           |

---

## Errory Row Example

**Input**

```
Current Turn Agust 202X (#12a3)
```

**CST → AST**

| Field | CST Token Text | AST Raw            | AST Value | Flags               | Diagnostics             |
| ----- | -------------- | ------------------ | --------- | ------------------- | ----------------------- |
| Month | `Agust`        | `RawMonth="Agust"` | `Month=0` | `UnknownMonth=true` | `E_MONTH` at Month span |
| Year  | `202X`         | `RawYear="202X"`   | `Year=0`  | `InvalidYear=true`  | `E_YEAR` at Year span   |
| Turn  | `12a3`         | `RawTurn="12a3"`   | `Turn=0`  | `InvalidTurn=true`  | `E_TURN` at Turn span   |

---

## Cross-Checks & Guardrails

* **Never drop a Header** just because a field is bad. Keep the AST node and set flags; emit diagnostics localized to each field’s span.
* **Case/spacing**: All normalization functions should `strings.ToLower` and `strings.TrimSpace` first.
* **Synthetic tokens**: Detect with `span.Start == span.End`. If you surface them at AST, prefer **warnings** (structure issues) vs **errors** (semantic issues).
* **Multiple headers** (policy): keep all + warn (`W_DUPLICATE_HEADER`) **or** keep first/last and warn. Don’t silently discard.

---

## Minimal Implementation Hooks (names from the starter)

* `tokenText(input []byte, tn *cst.TokenNode) string`
* `tokenSpan(tn *cst.TokenNode) ast.Span`
* `normalizeMonth(string) (int, bool)`
* `parseYear(string) (int, bool)`
* `validYearRange(int) bool` (default `[1900,2200]`)
* `parsePositiveInt(string) (int, bool)`
* `isSynthetic(*cst.TokenNode) bool` (zero-width)

