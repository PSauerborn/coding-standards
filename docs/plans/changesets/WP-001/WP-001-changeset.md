# Changeset: Standards Indexer CLI (`indexer`)

Work Plan ID: WP-001
Created Date: 2026-07-31
Author: documenter

## Summary

Implements SPEC-001 with a new Go CLI, `indexer/`, that walks the standards
corpus, resolves the example files each document cites, assembles a
single-rooted tree that mirrors `parent:` frontmatter, validates corpus-level
invariants (citation completeness, statement coverage, uniqueness), and
writes the result as deterministic YAML to `docs/standards-tree.yaml`. All
ten acceptance criteria have been independently verified as met. Twelve
execution tasks landed, followed by two remediation rounds addressing code
review, security and quality-controller findings; the changeset closed with
validation, security, quality and risk review all clean, and the sole
remaining open item — three code-review findings at the two-iteration cap —
resolved one and deferred two by explicit user decision (see
[Deferred by user decision](#deferred-by-user-decision--follow-up-owed)
below).

## Changes

### Added

- New `indexer/` Go module (`package main`, flat layout per GO-008) providing
  the `indexer --source <dir> --output <file> [--prefix <p>]` CLI: discovery
  of standards documents (`discovery.go`), frontmatter parsing
  (`frontmatter.go`), example resolution and parsing (`examples.go`), tree
  assembly from `parent:` frontmatter (`tree.go`), corpus-level validation
  (`validate.go`), deterministic YAML rendering (`render.go`), tree
  comparison for tests (`compare.go`), typed diagnostics and an error
  collector with cascade suppression (`errors.go`), core data types
  (`types.go`), and CLI wiring, atomic output writing and pipeline
  orchestration (`main.go`).
- A one-to-one `_test.go` file per source file (GO-027), plus
  `integration_test.go` (shared test harness: `repoRoot`, `copyCorpus`,
  `cleanBaseline`, `integrationIndex`) and `failure_paths_test.go` (pipeline
  failure-path assertions).
- Deterministic-output rendering built entirely through the `yaml.Node` API
  from an ordered `[]field` slice — no Go map is ever iterated when emitting
  YAML — with `!!str` tags pinned on every string scalar to prevent
  re-parsing corruption (e.g. a `'*'` scope re-parsing as a non-string).
- Atomic, all-or-nothing tree writes: content goes through a dot-prefixed
  temporary file in the output directory, is synced and chmod'd on the open
  descriptor (avoiding a path-based TOCTOU), and is renamed onto `--output`
  only once every diagnostic-free byte is on disk. A failing run leaves a
  pre-existing tree file untouched.
- Symlink-aware corpus containment: `resolveRoot` and `containedFile`
  resolve both the source root and every candidate path through
  `filepath.EvalSymlinks` before any containment or regular-file decision,
  applied at every read site in the pipeline (walk, document read, example
  read, statement-coverage read).
- A two-part empty-corpus guard in `DiscoverDocuments`: a `--source` with no
  markdown files at all, and a `--source` with markdown files that produce no
  recognized standards document (and no diagnostic already explaining why),
  both fail the run rather than silently publishing an empty tree over a
  previously deployed one.
- Test fixture corpora (dot-prefixed, under
  `indexer/tests/data/.corpora/<case>/`) covering frontmatter, discovery,
  examples, tree assembly, validation, CLI behavior and deliberately broken
  failure paths (nested code fences, CRLF line endings, detached citations,
  duplicate statements, malformed examples, and more).
- Toolchain/config: `indexer/go.mod`, `indexer/go.sum`, `indexer/Makefile`,
  `indexer/.golangci.yml` (migrated to the golangci-lint v2 schema), and
  `indexer/.gitignore`.
- `indexer/README.md`, documenting the CLI contract, the discovery rule
  (including symlinked-source resolution and the two-part empty-corpus
  guard, added by this session), the fixture-placement rule,
  `docs/standards-tree.yaml` as the authoritative output schema, the
  reference-refresh procedure, and the recorded GEN-003 build-environment
  deviation.
- A minimal cross-reference in the repository-root `README.md` pointing at
  `indexer/` as this repository's local implementation of the indexing
  contract described there, added by this session without touching any of
  the three existing external `stdidx` links (left unchanged by user
  decision).

### Fixed

*(closed during Remediation Round 1 and Round 2 code review — see the
manifest for full detail)*

- A cited-but-malformed example was previously also reported as uncited;
  `ResolveExamples` now skips paths the collector already marked failed.
- A nested-code-fence bug (naive parity flip) could reclassify sample code as
  a real statement definition, letting a citing example pass with exit 0.
  Replaced with a single shared `fenceTracker` used by both statement
  extraction and example-statement parsing.
- CRLF example files emitted a trailing `\r` inside a parsed title; a shared
  `splitLines` now strips it at the single split site.
- Citations of documents detached by an unknown parent or a parent cycle were
  never resolved, so their example faults went unreported; `resolveCitations`
  now routes detached documents' citations into the run's real collector.
- A `Statements:` line wrapped across multiple markdown lines silently
  dropped its continuation IDs; it is now read as the paragraph it opens.
- A symlinked `--source` (final path component a symlink) previously walked
  nothing and exited 0 with `nodes: []`; `resolveRoot` now resolves the root
  before the walk.
- A malformed example cited by two standards documents was reported twice;
  the collector's failure marking now suppresses the repeat.
- A `--source` naming a directory with markdown files but no recognized
  standards document (e.g. a plain documentation folder) previously fell
  through the discovery guard and emitted an empty tree at exit 0; a second,
  narrower guard now closes this case while preserving the useful diagnostic
  for a corpus whose only document has unparseable frontmatter.

### Security

- Corpus file containment was previously decided lexically, before symlink
  resolution, allowing a symlink committed inside the corpus to pass the
  `filepath.Rel` check and be read — demonstrated to publish an out-of-root
  file's heading as an emitted node title and to admit a symlinked `*.md`'s
  frontmatter as a corpus document. Fixed with a single `containedFile`
  helper that resolves both root and candidate via `filepath.EvalSymlinks`,
  re-checks containment on the resolved paths, and requires a regular file;
  every read site in the pipeline routes through it.
- The output temporary file's permission widening moved from a path-based
  `os.Chmod` (a TOCTOU window) to a `Chmod` on the already-open file
  descriptor.

## New Files

Files created by this changeset (no diff block required — nearly the entire
changeset is new; the only pre-existing file modified is `indexer/go.mod`,
which previously contained just a module line and a `go` directive).

| File Path | Description |
| --------- | ----------- |
| `indexer/main.go` | CLI entrypoint, flag parsing, pipeline orchestration, atomic output write |
| `indexer/types.go` | `Config`, `Frontmatter`, `Document`, `Node`, `Tree`, `ExampleEntry` and the ordered-`field` YAML emission type |
| `indexer/errors.go` | Sentinel errors, typed `DiagnosticError` implementations, `ErrorCollector` with cascade suppression, `AggregateError` |
| `indexer/frontmatter.go` | YAML frontmatter block splitting and parsing |
| `indexer/discovery.go` | Corpus walking, symlink-resolved containment (`resolveRoot`, `containedFile`), document discovery, statement extraction |
| `indexer/examples.go` | Example file resolution, parsing (title/statements), and containment-checked reads |
| `indexer/tree.go` | Tree assembly from `parent:` frontmatter, cycle detection, path emission with `--prefix` |
| `indexer/validate.go` | Corpus-level validation: citation completeness, statement coverage, uniqueness |
| `indexer/render.go` | Deterministic YAML rendering via the `yaml.Node` API |
| `indexer/compare.go` | Structural tree comparison used by tests |
| `indexer/types_test.go`, `errors_test.go`, `frontmatter_test.go`, `discovery_test.go`, `examples_test.go`, `tree_test.go`, `validate_test.go`, `render_test.go`, `main_test.go`, `compare_test.go` | One test file per source file (GO-027) |
| `indexer/integration_test.go` | End-to-end integration harness and acceptance-criteria tests, including the byte-identical-output and fixture-non-indexing checks |
| `indexer/failure_paths_test.go` | Pipeline-level failure-path assertions with exact diagnostic-count pinning |
| `indexer/README.md` | CLI contract, discovery rule, fixture-placement rule, output schema, reference-refresh procedure, GEN-003 deviation |
| `indexer/Makefile` | `build`, `test`, `lint`, `fmt` targets against the host Go toolchain |
| `indexer/.golangci.yml` | Lint configuration (v2 schema), including the `archived-yaml` depguard deny rule (RISK-012) |
| `indexer/.gitignore` | Build artifact exclusions |
| `indexer/tests/data/.corpora/**` | Dot-prefixed fixture corpora (frontmatter, discovery, examples, tree, validation, cli, failure) — must stay dot-prefixed per the fixture-placement rule |
| `docs/plans/tasks/WP-001/TASK-001.md` … `TASK-012.md` | Per-task investigation notes appended by their executors |

## Modified Files

| File Path | Description |
| --------- | ----------- |
| `indexer/go.mod` | Previously contained only a module line and `go` directive; now declares the `go.yaml.in/yaml/v3` runtime dependency and the `stretchr/testify` test-only dependency, and pins the `go` directive to `1.26.5` after the toolchain upgrade recorded in the manifest (RISK-022 and its follow-on blocker) |
| `README.md` (repository root) | Added a one-paragraph cross-reference to `indexer/` as this repository's local implementation of the indexing contract; the three existing external `stdidx` links were deliberately left unchanged per user decision |

## Read-only inputs (not part of this changeset)

- The standards corpus itself (18 standards documents, 28 example files) —
  never modified by this work.
- `docs/standards-tree.yaml` — the authoritative reference output and schema
  contract. Only its two leading comment lines changed, as a consequence of
  the `stdidx` → `indexer` rename; no YAML data was touched. The file remains
  hand-maintained and read-only as far as the indexer tool itself is
  concerned.
- `docs/specs/SPEC-001.md`, `docs/plans/2026-07-31-WP-001.md`,
  `docs/plans/risk/WP-001/WP-001-risk-plan.md`, `.markdownlint.yaml` — touched
  by the orchestrator-directed `stdidx` → `indexer` rename or already
  modified before execution began; not part of this changeset's Go
  implementation work.

## Accepted standards deviations (user-adjudicated — not defects)

| Standard | Deviation and rationale |
| --- | --- |
| GO-018 | Configuration via CLI flags only — single-shot CLI, no config file |
| GO-020 | Reduced to flag presence/path checks in `main.go`; no config schema to validate |
| GO-022 | `go-playground/validator` forbidden by SPEC-001 §6 |
| GO-023 | `spf13/viper` forbidden by SPEC-001 §6 |
| GO-033 | `sirupsen/logrus` forbidden by SPEC-001 §6; diagnostics are plain stderr lines per the CLI contract |
| ACPT-002 / ACPT-004 | Go unit + integration tests cover all 10 acceptance criteria; no `cucumber/godog` |
| GEN-003 | No Dockerfile; Makefile targets run against the host Go toolchain. Rationale recorded in the `Makefile` header and `indexer/README.md`'s "Build environment" section |

Permitted dependencies: `go.yaml.in/yaml/v3` (runtime) and `stretchr/testify`
(test-only).

## Deferred by user decision — follow-up owed

The code-review flow caps at 2 iterations. A third review round raised three
findings; the user directed that finding 2 be fixed (done — see the manifest)
and that findings 1 and 3, plus a related GO-016 follow-up cluster, be
deferred. These are open by decision, not oversight, and are carried forward
here so they are not lost:

1. **`discovery.go` — containment relativity (high, not currently
   reachable).** `filepath.EvalSymlinks` preserves the relativity of its
   input, so a relative `--source` yields a relative resolved root while a
   symlink inside the corpus with an absolute target yields an absolute
   resolved path. `filepath.Rel` cannot relate a relative base to an absolute
   target and errors, and that error is currently mapped to `errEscapesRoot`
   — so a file plainly inside the root can be judged an escape purely on how
   `--source` was spelled. Not reachable in this repository (no symlinks
   present). Suggested fix: make both operands absolute before comparing,
   while keeping emitted paths relative and unchanged.
2. **`examples.go` — suppression key space (low).** The `IsFailed`
   suppression is keyed on a path in the same key space `DiscoverDocuments`
   marks, so a file that is both a discovered document and a cited example
   has its example-parse diagnostics suppressed by a document-level failure.
   Cannot cause a false exit 0 (every `MarkFailed` is paired with an `Add`),
   and the suppressed diagnostics surface on the next run once the first
   fault is repaired.
3. **GO-016 follow-up cluster spanning `examples.go` and `validate.go`.**
   Two untyped diagnostics remain and want a single follow-up covering
   `errors.go` + `examples.go` + `validate.go`: `example %s cited by %s is
   not a regular file` (`examples.go`, added by the security remediation)
   and `failed to read example %s: %w` (`validate.go`). Each remediation
   task's write set excluded one of the pair, and typing one alone would
   preserve the inconsistency the change is meant to remove.

Additionally recorded as outstanding-but-not-blocking in the manifest:

- **Code-review finding 4** (citation double-resolution) was fixed narrowly
  rather than structurally: detached documents' citations now route into the
  real collector, but the structural change the reviewer suggested
  (resolving every document's citations once and having `BuildTree` consume
  the result) was outside the remediation task's write set (`tree.go`) and
  remains open.
- **RISK-007's `-count=10 -shuffle=on` clause** is realized as an in-process
  10-iteration loop plus a `t.Log` hint rather than a `make` target. A
  `test-determinism` target would close it if CI is added.

## Review status at completion

| Reviewer | Final |
| --- | --- |
| validation-runner | clean |
| security-reviewer | clean |
| quality-controller | clean |
| risk-reviewer | clean |
| code-reviewer | 3 open findings — 1 resolved, 2 deferred by user decision (above) |
