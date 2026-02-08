# Golden snapshot test harness for the AST.

The harness writes/reads JSON snapshots in `testdata/` and supports `-update` to regenerate.

## File: `ast/golden_test.go`

## Create initial goldens

Run once to generate:

```bash
cd ast
go test -run TestAST_Golden -update
```

This will create:

```
ast/
  golden_test.go
  testdata/
    happy_current_turn.golden.json
    unknown_month_and_bad_turn.golden.json
    year_out_of_range.golden.json
```

## Notes & tips

* Keep the **snapshot structs** stable; evolve them only when AST shape changes.
* Always run with `-race` in CI (`go test -race ./...`).
* Prefer **small, focused inputs**; add new cases per rule you change.
* If you add new diagnostics, it’s fine to reorder them—the harness sorts by `Code` then `Span.Start` for determinism.
