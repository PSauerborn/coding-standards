# Task: CLI entrypoint and pipeline wiring

Work Plan ID: WP-001
Task ID: TASK-009
Created Date: 2026-07-31
Description: Implement `indexer/main.go` — stdlib flag parsing for `--source`/`--output`/`--prefix`, the discover→parse→assemble→validate→render pipeline, aggregate diagnostics on stderr sorted by path, atomic output write only on a fully clean run, and exit codes.
Acceptance Criteria Covered: none directly (REQ-7 CLI contract; enables AC-1..AC-10)

## Implementation Content

### CLI contract

- `indexer --source <dir> --output <file> [--prefix <p>]`, parsed with the **stdlib `flag` package** (no `viper`/`cobra` — dependencies are restricted).
- Validate flag presence and paths at startup: `--source` and `--output` are required; `--source` must exist and be a directory. Clear diagnostics on stderr for missing/invalid flags.
- Exit 0 on success; non-zero on any validation failure. All diagnostics go to **stderr**; nothing but diagnostics is printed there.

### Pipeline

Wire the existing stages: discover → parse frontmatter/examples → assemble tree → validate corpus → render → write.

- **Collect-all, never fail-fast** (spec §7): every validation error in a run is collected into the aggregate error and reported together on stderr, sorted by path for reproducibility, before exiting non-zero.
- **Cascade suppression** (RISK-006): one corrupted document yields exactly ONE diagnostic naming that file — not an unknown-parent error per child plus an uncited-example error per citation. Assert this in `main_test.go`.

### Output write (RISK-017)

- Write to a **temporary file in the output file's own directory**, then `rename` atomically, and only once the aggregate error is empty.
- A failing corpus must leave **no new output file** and must **not truncate a pre-existing** output file.
- A missing output directory yields a clear diagnostic (and no partial write).

### Structure

Keep `main()` thin: parse flags into the `Config` type, call a testable `run(cfg) error`-style function so `main_test.go` can exercise the pipeline in-process without spawning the binary. `main()` may exit non-zero on a returned error (GO-015 permits top-level handling here).

### Tests (`main_test.go`)

- Happy path over a small synthetic corpus: exit 0, output file written, content parses as YAML with a `nodes:` key.
- Flag validation: missing `--source`, missing `--output`, `--source` pointing at a nonexistent path → non-zero with a clear diagnostic naming the problem.
- Failing corpus → non-zero, diagnostics on stderr, **no output file created**; and with a pre-existing output file present, its contents are unchanged (not truncated).
- Missing output directory → clear diagnostic, non-zero, no file created.
- Multiple independent failures in one corpus are all reported in a single run (no fail-fast), sorted by path — assert identical diagnostic ordering across repeated runs.
- One corrupted document (unparseable frontmatter) that has a child and cites two examples → exactly ONE diagnostic, naming that document.
- `--prefix` accepted and applied end-to-end (every emitted path prefixed).

Fixtures for this task live under `indexer/tests/data/.corpora/cli/`. All output writes in tests go to `t.TempDir()`.

## Target Files

- [x] indexer/main.go
- [x] indexer/main_test.go
- [x] indexer/tests/data/.corpora/cli/ (fixture files created by this task)

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- indexer/types.go (`Config`: `Source`, `Output`, `Prefix`)
- indexer/errors.go (aggregate error: accumulation, path-sorted rendering, failed-document marking)
- indexer/discovery.go (walk entrypoint signature)
- indexer/validate.go (corpus validation entrypoint and cascade suppression)
- indexer/render.go (render entrypoint producing the output bytes)
- indexer/tree.go (tree assembly entrypoint and prefix application at emission)

## Investigation Notes

- `types.go`: `Config{Source, Output, Prefix}` is the run configuration; `Tree{Nodes []*Node}` is the render input.
- `errors.go`: `NewErrorCollector()`, `Add`, `MarkFailed`/`IsFailed`, `Errors()` (sorted by document path then message), `Diagnostics()` (stderr-ready lines) and `ErrorOrNil()` returning `*AggregateError`. The aggregate renders one diagnostic per line, so the entrypoint prints it as-is.
- `discovery.go`: `DiscoverDocuments(source, collector) ([]Document, error)` returns an error only when the corpus cannot be read; everything else is collected. It marks unparseable documents failed, which is what the cascade suppression downstream keys off.
- `tree.go`: `BuildTree(config, documents, collector) Tree` applies `--prefix` at `emitPath`, suppresses an `UnknownParentError` whose parent is already marked failed, and internally calls `ResolveExamples` for every document it places — so the citation diagnostics of placed documents are collected here.
- `examples.go`: `ResolveExamples(source, document, collector) []ExampleEntry` returns source-root-relative, unprefixed paths, and marks an unparseable example file failed. Because `BuildTree` already collects these diagnostics, the separate citations pass the validation stage needs must run against a throwaway collector, otherwise every citation diagnostic is reported twice.
- `validate.go`: `ValidateCorpus(source, documents, citations, collector) error` needs the UNPREFIXED citation map keyed by document path (the shape `validate_test.go` builds), returns an error only when the corpus cannot be walked, and skips the uncited scan entirely when a candidate document is marked failed but was not discovered.
- `render.go`: `RenderTree(tree) ([]byte, error)` returns the document bytes and writes nothing; persisting them is this task's job. Ordering is code-declared, so repeated runs are byte identical.
- Ordering constraint: `BuildTree` must run before `ValidateCorpus` because it is what marks broken example files failed, and `DiscoverDocuments` must run before `BuildTree` for the same reason.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-007 | Corpus-level validation | blocks | REQ-8/REQ-9 validation stage and cascade-suppression semantics |
| TASK-008 | Deterministic YAML rendering | blocks | Render entrypoint producing deterministic output bytes |

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review dependency deliverables (validation stage, render stage, aggregate error contract)
- [x] Create fixtures under `indexer/tests/data/.corpora/cli/` (a small valid corpus; a corpus with one corrupted document plus a child and two citations; a corpus with several independent failures)
- [x] Write failing tests in `main_test.go`
- [x] Run `make test` and confirm failure

### 2. Green Phase

- [x] Add minimal implementation in `main.go` (replacing any TASK-001 placeholder) to pass tests
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Improve code (thin `main()`, testable `run(cfg)`; doc comments per GO-005), maintain passing tests
- [x] Confirm added tests still pass; `make lint` and `make fmt` clean

## Completion Criteria

- [x] `indexer --source <dir> --output <file> [--prefix <p>]` parses with the stdlib `flag` package; missing/invalid flags produce a clear stderr diagnostic and a non-zero exit
- [x] A clean corpus exits 0 and writes a parseable YAML file with a top-level `nodes:` key
- [x] All validation errors in a run are reported together on stderr, sorted by path, with identical ordering across repeated runs; the process then exits non-zero
- [x] One corrupted document yields exactly one diagnostic naming that file
- [x] A failing run creates no output file and does not truncate a pre-existing one; the successful write goes through a temp file in the output directory plus an atomic rename
- [x] A missing output directory produces a clear diagnostic and no partial write
- [x] All added tests pass; `make test`, `make lint`, `make fmt` (no diff) pass in `indexer/`

## Notes

- Impact scope: TASK-010/011/012 drive this entrypoint (in-process and as the built binary) for the acceptance suite.
- Scope boundary:
  - **Fixture placement (load-bearing, CRITICAL)**: every synthetic fixture corpus MUST live under `indexer/tests/data/.corpora/<case>/`. `indexer/tests/data/` sits INSIDE the `--source` root used when the tool indexes its own repository, and `indexer/` is neither dot-prefixed nor an `examples/` directory — the dot-prefixed `.corpora` directory is the ONLY thing keeping fixtures out of the walk.
  - Flat `package main` directly in `indexer/` (GO-008); snake_case filenames (GO-006); one `_test.go` per source file (GO-027); fixture data under `tests/data` (GO-031).
  - Accepted standards deviations for this file (user-adjudicated, do not "fix"): GO-018/GO-020/GO-022/GO-023 (flag-only configuration, no config file, no `go-playground/validator`, no `viper`) and GO-033 (no `logrus`; diagnostics are plain stderr lines).
  - Do not modify files owned by earlier tasks (`types.go`, `errors.go`, `frontmatter.go`, `discovery.go`, `examples.go`, `tree.go`, `validate.go`, `render.go`, and their tests/fixtures, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`).
  - Permitted dependencies: `go.yaml.in/yaml/v3` (runtime), `stretchr/testify` (test-only). `gopkg.in/yaml.v3` MUST NOT be imported.
  - The standards corpus (18 documents, 28 example files) and `docs/standards-tree.yaml` are READ-ONLY; no test may write to the working tree — all test output goes to `t.TempDir()`.
  - The binary is `indexer`; the former name `stdidx` must not appear anywhere in the implementation, including flag help text and diagnostics.
