# Parser Documentation: Goals and Structure

## Goals of the Parser (and Lexer)

Our parser is responsible for turning **source text** (the input files written in our DSL) into structured data that the rest of the system can use.

* **Lexer (sometimes called a scanner):**
  The lexer breaks raw text into **tokens**. Tokens are the smallest meaningful pieces of the language — like words in a sentence.

    * Example: the line

      ```
      Current Turn August 2025 (#123)
      ```

      could be split into tokens:

      ```
      "Current", "Turn", "August", "2025", "(#123)"
      ```

* **Parser:**
  The parser reads these tokens and organizes them into a **tree structure** that reflects how the language is built. This structure lets us reason about the input in a reliable way.

The ultimate goal:

* **Provide clear, helpful error messages** when input is wrong.
* **Preserve details** from the original text (via the CST).
* **Create a clean, simplified version** for use by the rest of the system (via the AST).

---

## CST vs. AST: Two Views of the Input

We deliberately build **two different trees** from the input:

### 1. Concrete Syntax Tree (CST)

* A detailed tree that keeps **every piece of the original text**, including punctuation, spacing, and keywords.
* Useful for:

    * **Error reporting** (we can point exactly to where something went wrong).
    * **Testing/debugging** (we can see exactly what the parser saw).

**Example:**
Input:

```
Current Turn August 2025 (#123)
```

CST (simplified view):

```
Header
 ├── "Current"
 ├── "Turn"
 ├── Month("August")
 ├── Year("2025")
 └── TurnNumber("#123")
```

Notice how every word is kept in the tree.

---

### 2. Abstract Syntax Tree (AST)

* A **cleaned-up, simplified version** of the CST.
* Keeps only the information that matters for program logic.
* Easier to use in later steps of the system.

**Example (same input):**

```
Header {
    Turn: 123,
    Month: 8,
    Year: 2025
}
```

Here, we dropped unnecessary words (“Current Turn”), converted “August” to `8`, and stripped the “#” from the number. The AST focuses only on meaning.

---

## Why Two Trees?

* **Testers** get better feedback: error messages can show the exact spot in the source where a problem occurs.
* **Managers** can trust that the tool will be friendlier for users (not just “syntax error,” but “unexpected word ‘Currnet’, did you mean ‘Current’?”).
* **Developers** get a simpler AST for building features, without worrying about formatting details.

---

## Key Takeaways

* The **lexer** breaks text into tokens.
* The **parser** organizes tokens into a **CST** (detailed) and then into an **AST** (simplified).
* The CST is for **diagnostics**; the AST is for **execution/logic**.
* This two-stage design helps us deliver:

    * **Better user experience** (clear error messages).
    * **Cleaner code downstream** (easier to work with AST).
    * **Confidence in correctness** (testers can check CST vs. AST transformations).
