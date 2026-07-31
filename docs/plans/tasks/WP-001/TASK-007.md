# Task: Corpus-level validation (uncited examples and statement coverage)

Work Plan ID: WP-001
Task ID: TASK-007
Created Date: 2026-07-31
Description: Implement `indexer/validate.go` — REQ-8 (every `*.md` under an `examples/` directory must be cited) reusing the discovery walker, and REQ-9 (heading ID present on the `Statements:` line; every `Statements:` ID defined in the union of citing standards), with cascade suppression for already-failed documents.
Acceptance Criteria Covered: AC-8, AC-9

## Implementation Content

### REQ-8 — uncited example files (AC-8)

- Every `*.md` file under any `examples/` directory must be cited by some discovered document's `examples:` frontmatter. Otherwise fail with a diagnostic naming the uncited file.
- **The scan MUST reuse the same walker and exclusion set as discovery** (`indexer/discovery.go`). A second, unfiltered walk would treat fixture example files under `indexer/tests/data/.corpora/` as uncited real-corpus files and make the tool fail on its own repository. Dot-prefixed directories must be excluded by exactly the same logic that discovery uses.
- Citations from documents marked failed by an earlier stage still count as "cited" (cascade suppression).

### REQ-9 — statement coverage (AC-9)

- The example's own heading ID must appear on its `Statements:` line; otherwise fail.
- Every ID on an example's `Statements:` line must occur as a statement **DEFINITION** (per TASK-004's definition-only extraction: line-anchored backticked ID followed by a MUST/SHOULD marker, fenced code blocks skipped) in the **UNION** of the standards citing that example. An ID satisfied by any one citing standard is satisfied.
- Failure diagnostic names the example, the offending ID, and the citing standard(s).

### Cascade suppression (RISK-006)

Documents marked as failed by earlier stages (unparseable frontmatter, missing required key) suppress dependent REQ-8/REQ-9 checks: their citations still count as "cited", and their children do not additionally report unknown-parent. One corrupted document must ultimately yield exactly ONE diagnostic. All findings accumulate into the aggregate error — never fail fast.

### Tests (`validate_test.go`)

Mandatory cases:

- **Dual-citation union fixture**: two standards both cite the same example; an ID on that example's `Statements:` line is defined ONLY in the second citing standard → validation PASSES. This path is unexercised by the real corpus (all 28 example files are cited exactly once — `golang/examples/API/api-anti-pattern.md` and `python/examples/API/api-anti-pattern.md` are distinct files, not a shared citation), so the synthetic fixture is required.
- **Mention-not-definition fixture** (RISK-005): a standard mentions a foreign ID both in prose AND inside a fenced code block; an example cites that ID → REQ-9 still FAILS, naming example, ID and citing standard.
- Uncited example file under a fixture `examples/` directory → fails naming the file.
- An example whose heading ID is absent from its own `Statements:` line → fails.
- Cascade suppression: a corpus where one document fails frontmatter parsing and cites two examples and has one child → exactly one diagnostic, naming the corrupted document.
- Real-corpus run: all 28 example files are cited and every `Statements:` ID resolves — validation produces zero findings.
- REQ-8 scan does NOT report files under dot-prefixed directories (assert by running the scan over a directory tree containing `.corpora/`).

Fixtures for this task live under `indexer/tests/data/.corpora/validation/`.

## Target Files

- [x] indexer/validate.go
- [x] indexer/validate_test.go
- [x] indexer/tests/data/.corpora/validation/ (fixture files created by this task)

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- indexer/discovery.go (the walker + exclusion set to REUSE; definition-only statement extraction; shared ID regex)
- indexer/examples.go (`ExampleEntry` values: heading ID, ordered `Statements:` IDs, citing document linkage)
- indexer/errors.go (uncited-example and unknown-statement error types; aggregate accumulation; failed-document marking/query)
- indexer/types.go (`Document`, `ExampleEntry`)
- /Users/pascal/Github/psauerborn/standards/python/GENERAL.md (lines 12-19 and 20-50 — an `examples:` citation list plus the statement-definition shape the union check resolves against)

## Investigation Notes

- `discovery.go` exports `WalkMarkdownFiles(root, policy)` returning slash-form,
  root-relative paths in lexical order; `isExcludedDirectory` always drops
  dot-prefixed directories and drops `examples` only under
  `SkipExampleDirectories`. The REQ-8 scan therefore calls
  `WalkMarkdownFiles(source, IncludeExampleDirectories)` ONCE and partitions its
  result into example files (a path segment equal to `examples`) and candidate
  documents; no second walk exists, so the exclusion set cannot drift (RISK-001).
- `ExtractStatements` is definition-only (line-anchored backticked ID plus a
  MUST/SHOULD marker, fenced blocks skipped) and is already stored on
  `Document.Statements` by `DiscoverDocuments`, so REQ-9 reads
  `Document.Statements` and never re-scans document bodies (RISK-004/RISK-005).
- `ExampleEntry` carries `Path`, `Title` and ordered `Statements` but NOT the
  heading identifier: `ParseExampleFile` strips it via `exampleHeadingRegexp`.
  The REQ-9 heading check therefore re-reads the cited example file and applies
  the shared `exampleHeadingRegexp` (no new pattern).
- `ExampleEntry` also carries no citing-document linkage, so validation takes the
  resolved entries keyed by citing document path (`map[string][]ExampleEntry`),
  which is exactly what the pipeline holds after calling `ResolveExamples` per
  document. Citing documents per example are deduped and sorted for a
  deterministic `UnknownStatementError.CitedBy`.
- `ErrorCollector` exposes `MarkFailed`/`IsFailed` but cannot enumerate failed
  paths, so "citations are incomplete" is derived from the walk: a candidate
  document path that is marked failed yet absent from the discovered documents is
  a document whose frontmatter did not parse, hence whose `examples:` list is
  unknowable. While such a document exists, REQ-8 reporting is suppressed
  wholesale — the corrupted document already produced its one diagnostic
  (RISK-006).
- `errors.go` has `UncitedExampleError{Path}` and
  `UnknownStatementError{Path, Statement, CitedBy}` but no error for "heading ID
  missing from the example's own Statements: line"; `errors.go` is owned by
  TASK-002 and out of this task's write set, so `MissingHeadingStatementError` is
  declared in `validate.go` and implements `DiagnosticError` for ordering.
- Real corpus shape (`python/GENERAL.md`): `examples:` entries are relative to
  the citing document's directory, and statements are defined as
  `` `[PY-013]` **MUST**: ... `` at the start of a line.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-004 | Document discovery and statement-ID extraction | blocks | Reusable walker + exclusion set; definition-only statement IDs per document |
| TASK-005 | Example file parsing and citation resolution | blocks | Parsed example heading IDs and ordered `Statements:` IDs with their citing documents |

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review dependency deliverables (walker/exclusions from TASK-004; example parse results from TASK-005)
- [x] Create fixtures under `indexer/tests/data/.corpora/validation/` (dual-citation union corpus, prose+code-fence mention corpus, uncited-example corpus, corrupted-document cascade corpus)
- [x] Write failing tests in `validate_test.go`
- [x] Run `make test` and confirm failure

### 2. Green Phase

- [x] Add minimal implementation in `validate.go` to pass tests
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Improve code (doc comments per GO-005), maintain passing tests
- [x] Confirm added tests still pass; `make lint` and `make fmt` clean

## Completion Criteria

- [x] An uncited `*.md` under an `examples/` directory fails with a diagnostic naming the uncited file
- [x] The REQ-8 scan reuses the discovery walker and exclusion set — files under dot-prefixed directories (including `indexer/tests/data/.corpora/`) are never reported as uncited
- [x] An ID satisfied only by the SECOND of two citing standards passes (union rule, proven by the synthetic dual-citation fixture)
- [x] An ID that a citing standard only mentions in prose or inside a fenced code block does NOT satisfy REQ-9; the failure names the example, the ID and the citing standard(s)
- [x] An example whose heading ID is missing from its `Statements:` line fails
- [x] A corpus with one corrupted document yields exactly one diagnostic naming that document (its citations still count as cited; its children do not report unknown-parent)
- [x] Validation over the real corpus produces zero findings (28 example files, all cited, all statement IDs resolved)
- [x] All added tests pass; `make test`, `make lint`, `make fmt` (no diff) pass in `indexer/`

## Notes

- Impact scope: TASK-009 wires this stage into the pipeline and prints the aggregate diagnostics; TASK-012 exercises AC-8/AC-9 end-to-end.
- Scope boundary:
  - **Fixture placement (load-bearing, CRITICAL)**: every synthetic fixture corpus MUST live under `indexer/tests/data/.corpora/<case>/`. `indexer/tests/data/` sits INSIDE the `--source` root used when the tool indexes its own repository, and `indexer/` is neither dot-prefixed nor an `examples/` directory — the dot-prefixed `.corpora` directory is the ONLY thing keeping fixtures out of the walk. Fixture files under a reachable `examples/` directory are by construction uncited and would fail REQ-8 on the REAL repository.
  - Flat `package main` directly in `indexer/` (GO-008); snake_case filenames (GO-006); one `_test.go` per source file (GO-027); fixture data under `tests/data` (GO-031).
  - Do not modify files owned by earlier tasks (`types.go`, `errors.go`, `frontmatter.go`, `discovery.go`, `examples.go`, `tree.go`, and their tests/fixtures, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`).
  - Permitted dependencies: `go.yaml.in/yaml/v3` (runtime), `stretchr/testify` (test-only). `gopkg.in/yaml.v3` MUST NOT be imported.
  - The standards corpus (18 documents, 28 example files) and `docs/standards-tree.yaml` are READ-ONLY; no test may mutate the working tree.
  - The binary is `indexer`; the former name `stdidx` must not appear anywhere in the implementation.
