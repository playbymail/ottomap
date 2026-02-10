# Agent

This project implements OttoMap, an application that reads TribeNet turn reports and creates WXX files for Worldographer (a map viewer).

## Status
We are in the middle of a rewrite of the code that parses turn reports and creates the WXX files.
This new code is in `internal/reports`, `internal/runners`, and `internal/walkers`.

The "legacy" code was written when we were first learning about the reports and Worldographer files.
It contains a lot of duplicated or unused code from experiments and also a couple of tries at implementing the parser.

The legacy code may contain references to a web server.
That server was moved to a separate project (OttoWeb).

The legacy parser uses Pigeon and a PEG grammar.
It reports the first error it finds in the input and quits.
The error reporting is terrible.

## Turn Report Documentation
There is no official documentation for TribeNet turn reports.
The team has created `internal/reports/grammar.txt` to capture what we know about the format of these files.
That file uses a PEG-like grammar that is meant to be easy to read.

Note that the TribeNet game-master updates turn reports by hand before sending them to the players.
He makes corrections and provides edits to make the reports easier to understand.
Unfortunately, this also introduces errors that the parser must detect and report.

## Go data structures
We favor a naming style with endings to indicate struct, interface, and enums:
* `_t` for structs (example Map_t)
* `_i` for interfaces (example `Node_i`)
* `_e` for enums (example `SortOrder_e`)

We like to implement the Stringer interface for our enums.

## Data files for testing
The `testdata` directory contains turn reports for testing.
Do not ever commit any file from this path.

We use tribe `0987` for testing.

You must run tests from the `testdata/0987` folder (you must make that the working directory).
Sub-folders for `input`, `output`, and `logs` are used by OttoMap.

## Git
Do not commit changes unless explicitly told to.

When commiting changes, always update the version information first.

We use semantic versioning and store our version numbers in a variable defined in `main.go`.
That variable, `version`, uses the `github/maloquacious/semver` package for version numbering.
Please note that the `Build` member contains the Git commit information and status.

When bumping the major version, reset the minor and patch levels back to 0.
When bumping the minor version, reset the patch level back to 0.

## Sprint Plan
Plans for upcoming tasks are detailed in `docs/SPRINT_PLAN.md`.
That file tracks all completed and planned sprints, including the current
Phase 2 refactoring work (separating parser and render pipelines via
`internal/domain/`).

## Building and Deploying
We are switching from a `build` folder to `dist` to make OttoMap be consistent with other projects.

Use the `dist/local` folder for testing.
Use the `dist/linux` folder for deployments (the servers are Linux).

Do not build for deployment if there are uncommited changes in the repository.
When building for deployment, attach the version name to the binary (example, `dist/linux/ottomap-0.32.1`).
 
* Build: `go build -o dist/local/ottomap ./cmd/version`
* Version: `go run . --version`
