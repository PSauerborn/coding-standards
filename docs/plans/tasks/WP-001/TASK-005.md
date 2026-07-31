# Task: Example file parsing and citation resolution

Work Plan ID: WP-001
Task ID: TASK-005
Created Date: 2026-07-31
Description: Implement `indexer/examples.go` — resolve each `examples:` entry against the citing document's directory with proper containment checking, read the example file, and extract its title (from the `# [<ID>] <Title>` heading) and ordered `Statements:` IDs.
Acceptance Criteria Covered: AC-4, AC-7

## Implementation Content

### Resolution and containment (RISK-010)

- Each `examples:` frontmatter entry is a path relative to the **citing document's own directory** (e.g. `python/GENERAL.md` declares `examples/GENERAL/config.md` → `python/examples/GENERAL/config.md`).
- Containment check: resolve the entry against the citing document's directory, `Clean` the result, then require `filepath.Rel(sourceRoot, resolved)` NOT to begin with `..`; reject absolute entries outright. **Never substring-match `..`** — that false-positives on legitimate paths and misses absolute ones.
- Required fixtures proving the three behaviours:
  - `../../etc/passwd` → MUST fail (escapes the source root)
  - `examples/../examples/config.md` → MUST **PASS** (cleans to a contained path)
  - `/abs/path.md` → MUST fail (absolute)
- Missing example file → typed error naming **both** the citing standard and the missing path (this is the AC-7 diagnostic contract).

### Example file parsing (REQ-4)

- First line: `# [<ID>] <Title>`. The emitted title has the `[<ID>]` stripped — e.g. `# [PY-023] Configuration Loading` → title `Configuration Loading`.
- A `Statements:` line lists every covered ID as `` `[<ID>]` `` — e.g. ``Statements: `[PY-019]` `[PY-020]` `[PY-021]` `[PY-022]` `[PY-023]` `[PY-024]` ``. Emit the IDs **in order**.
- Parse IDs with the SHARED exported ID regex declared in `indexer/discovery.go` (TASK-004). Do not declare a second pattern: the corpus has three-segment families (`GO-API`, `GO-DOCKER`, `GO-WRK`, `JS-DOCKER`, `PY-API`, `PY-DOCKER`) and a truncating pattern applied on both sides of the cross-check passes validation while emitting wrong statement values.
- Failures: missing/malformed `# [<ID>] <Title>` heading; missing `Statements:` line; duplicate IDs on a `Statements:` line (error names the example file). Example files have NO frontmatter.
- Produce `ExampleEntry` values (`Path`, `Title`, `Statements`) in the citing document's frontmatter order. Path emission/prefixing is NOT this task's job — TASK-006's `emitPath()` handles it; store the source-root-relative path.
- Errors accumulate into the aggregate error (no fail-fast).

### Tests (`examples_test.go`)

- Real corpus round-trip: `python/examples/GENERAL/config.md` → title `Configuration Loading`, statements `PY-019, PY-020, PY-021, PY-022, PY-023, PY-024` in that order (AC-4). Resolve the repo root relatively (`runtime.Caller` + relative join), never a hardcoded absolute path.
- The three escape fixtures above (fail / PASS / fail).
- Missing example file → error naming the citing standard AND the missing path.
- Missing heading; missing `Statements:` line; duplicate ID on a `Statements:` line → each fails, naming the example file.
- Three-segment IDs (e.g. a `Statements:` line with `PY-DOCKER-001`, `GO-WRK-011`) parse as full tokens.
- Multiple entries preserve frontmatter order.

Fixtures for this task live under `indexer/tests/data/.corpora/examples/`.

## Target Files

- [x] indexer/examples.go
- [x] indexer/examples_test.go
- [x] indexer/tests/data/.corpora/examples/ (fixture files created by this task)

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- indexer/discovery.go (the shared exported statement-ID regex — reuse it, do not redeclare)
- indexer/types.go (`ExampleEntry`, `Document`, `Frontmatter.Examples`)
- indexer/errors.go (typed missing-example / escape / malformed-example errors; aggregate accumulation)
- /Users/pascal/Github/psauerborn/standards/python/GENERAL.md (lines 12-19 — `examples:` entries relative to the document's directory)
- /Users/pascal/Github/psauerborn/standards/python/examples/GENERAL/config.md (lines 1-4 — heading `# [PY-023] Configuration Loading` and the `Statements:` line with `PY-019`…`PY-024`)

## Investigation Notes

- `indexer/discovery.go` exports `StatementIDPattern` (`\[([A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*-\d+)\]`) and the compiled `StatementIDRegexp` (identifier in submatch 1). Both the `# [<ID>] <Title>` heading regex and the `Statements:` scan are built from this single pattern, so no second grammar is declared. `codeFenceRegexp` is package-private in the same `package main`, so example parsing reuses it to ignore fenced code blocks.
- `indexer/types.go`: `Frontmatter.Examples []string` holds the entries in declared order; `ExampleEntry{Path, Title, Statements}` is the emitted shape, with `Path` source-root-relative (no prefixing here — TASK-006 owns `emitPath()`). `Document` carries `Path` (source-root-relative, slash form) and `AbsolutePath`.
- `indexer/errors.go`: the vocabulary this task needs already exists — `MissingExampleError{Path, ExamplePath}`, `ExampleEscapesRootError{Path, ExamplePath}` (both attributed to the citing document, so the AC-7 diagnostic names both sides), `MissingExampleHeadingError{Path}`, `MissingStatementsLineError{Path}` and `DuplicateStatementError{Statement, Paths}`. `DuplicateStatementError.Paths` lists every definition site in order, so a duplicate inside one example file lists that file twice, which is literally what happened. Diagnostics accumulate through `ErrorCollector.Add`; `MarkFailed` suppresses cascades for later stages.
- `python/GENERAL.md` declares its `examples:` entries as `examples/GENERAL/<file>.md`, i.e. relative to `python/`, confirming resolution against the citing document's own directory.
- `python/examples/GENERAL/config.md` line 1 is `# [PY-023] Configuration Loading` and line 3 is ``Statements: `[PY-019]` `[PY-020]` `[PY-021]` `[PY-022]` `[PY-023]` `[PY-024]` ``; the heading ID is itself one of the covered statements and the title is the text after the bracketed ID.
- Containment is decided by `filepath.Rel(source, cleanedResolvedPath)` rejecting `..` and `../…`, plus an outright rejection of absolute entries; no substring matching of `..` anywhere.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-002 | Data models and aggregate error types | blocks | `ExampleEntry` model; typed example errors; aggregate accumulation |
| TASK-004 | Document discovery and statement-ID extraction | blocks | The single shared exported statement-ID regex; discovered `Document` values carrying `examples:` frontmatter |

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review dependency deliverables (shared ID regex from TASK-004; error vocabulary from TASK-002)
- [x] Create fixtures under `indexer/tests/data/.corpora/examples/` (escape cases, missing heading, missing `Statements:`, duplicate ID, three-segment IDs)
- [x] Write failing tests in `examples_test.go`
- [x] Run `make test` and confirm failure

### 2. Green Phase

- [x] Add minimal implementation in `examples.go` to pass tests
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Improve code (doc comments per GO-005), maintain passing tests
- [x] Confirm added tests still pass; `make lint` and `make fmt` clean

## Completion Criteria

- [x] `python/examples/GENERAL/config.md` yields title `Configuration Loading` and statements `PY-019, PY-020, PY-021, PY-022, PY-023, PY-024` in order
- [x] `../../etc/passwd` and `/abs/path.md` entries fail; `examples/../examples/config.md` PASSES; containment is decided by `filepath.Rel` on the cleaned path, not substring matching
- [x] A missing example file produces a diagnostic naming both the citing standard and the missing path
- [x] A missing heading, a missing `Statements:` line, and a duplicate ID on a `Statements:` line each fail naming the example file
- [x] Statement IDs are parsed with the shared regex from `discovery.go` (no second pattern declared) and three-segment IDs survive intact
- [x] `ExampleEntry` values preserve the citing document's frontmatter order; no prefixing is applied in this task
- [x] All added tests pass; `make test`, `make lint`, `make fmt` (no diff) pass in `indexer/`

## Notes

- Impact scope: TASK-006 routes example paths through `emitPath()`; TASK-007 checks `Statements:` IDs against the union of citing standards; TASK-008 renders `ExampleEntry` values.
- Scope boundary:
  - **Fixture placement (load-bearing, CRITICAL)**: every synthetic fixture corpus MUST live under `indexer/tests/data/.corpora/<case>/`. `indexer/tests/data/` sits INSIDE the `--source` root used when the tool indexes its own repository, and `indexer/` is neither dot-prefixed nor an `examples/` directory — the dot-prefixed `.corpora` directory is the ONLY thing keeping fixtures out of the walk. A fixture `examples/` directory placed anywhere the walker can reach is by construction uncited and fails corpus validation on the REAL repository.
  - Flat `package main` directly in `indexer/` (GO-008); snake_case filenames (GO-006); one `_test.go` per source file (GO-027); fixture data under `tests/data` (GO-031).
  - Do not modify files owned by earlier tasks (`types.go`, `errors.go`, `frontmatter.go`, `discovery.go`, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`).
  - Permitted dependencies: `go.yaml.in/yaml/v3` (runtime), `stretchr/testify` (test-only). `gopkg.in/yaml.v3` MUST NOT be imported.
  - The standards corpus (18 documents, 28 example files) and `docs/standards-tree.yaml` are READ-ONLY; no test may mutate the working tree.
  - The binary is `indexer`; the former name `stdidx` must not appear anywhere in the implementation.
