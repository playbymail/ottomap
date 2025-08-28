# Copyright (c) 2025 Michael D Henderson. All rights reserved.
# Notes:
#
# golden runs golden tests without changing files; golden-update rewrites snapshots (your old test-update now aliases to this).
#
# Use GOTESTFLAGS=" -run TestName " to selectively run tests, e.g.:
#   make test GOTESTFLAGS='-run TestHeaderCrosswalk'
#
# You can still run specific tests using GOTESTFLAGS, e.g.:
#   make test-ast GOTESTFLAGS='-run TestHeaderCrosswalk'
#
# Target names match the test function names used in the templates we wrote:
#   Lexer: TestLexer_Golden
#   CST  : TestCST_Golden
#   AST  : TestAST_Golden

# ---- Package paths ----
LEXERS := ./internal/reports/lexers
CST    := ./internal/reports/cst
AST    := ./internal/reports/ast

# ---- Common flags ----
GO_TEST_FLAGS := -race

.PHONY: all test test-lexer test-cst test-ast \
        golden golden-update \
        test-update coverage clean

# Default: run all tests with race detector
all: test
test:
	go test ./... $(GO_TEST_FLAGS) $(GOTESTFLAGS)

# ---- Per-package tests (convenience) ----
test-lexer:
	go test $(LEXERS) $(GO_TEST_FLAGS) $(GOTESTFLAGS)

test-cst:
	go test $(CST) $(GO_TEST_FLAGS) $(GOTESTFLAGS)

test-ast:
	go test $(AST) $(GO_TEST_FLAGS) $(GOTESTFLAGS)

# ---- Golden snapshots (read-only compare) ----
# Runs the golden tests WITHOUT rewriting snapshots.
golden:
	go test $(LEXERS) $(GO_TEST_FLAGS) -run TestLexer_Golden
	go test $(CST)    $(GO_TEST_FLAGS) -run TestCST_Golden
	go test $(AST)    $(GO_TEST_FLAGS) -run TestASTGolden

# ---- Golden snapshots (update) ----
# Regenerates snapshots (use after intentional output changes).
golden-update:
	go test $(LEXERS) $(GO_TEST_FLAGS) -run TestLexer_Golden -update
	go test $(CST)    $(GO_TEST_FLAGS) -run TestCST_Golden   -update
	go test $(AST)    $(GO_TEST_FLAGS) -run TestASTGolden    -update

# ---- Coverage across repo ----
coverage:
	go test ./... $(GO_TEST_FLAGS) -coverprofile=coverage.out
	@echo "---- Coverage summary ----"
	@go tool cover -func=coverage.out | tail -n 1
	@echo "HTML report: coverage.html"
	@go tool cover -html=coverage.out -o coverage.html

# ---- Clean generated artifacts ----
clean:
	@rm -f coverage.out coverage.html
