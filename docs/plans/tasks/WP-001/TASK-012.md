# Task: Failure-path fixture corpora and AC-7/AC-8/AC-9 acceptance

Work Plan ID: WP-001
Task ID: TASK-012
Created Date: 2026-07-31
Description: Add the failure-path fixture corpora under `indexer/tests/data/.corpora/failure/` and the end-to-end failure suite in `indexer/failure_paths_test.go`, driving AC-7, AC-8 and AC-9 through mutated `t.TempDir()` copies of the real corpus via the shared `copyCorpus` harness.
Acceptance Criteria Covered: AC-7, AC-8, AC-9

## Implementation Content

### Failure-path fixture corpora

Under `indexer/tests/data/.corpora/failure/<case>/`, one small self-contained corpus per case:

- `missing-example` — an `examples:` entry pointing at a nonexistent file
- `uncited-example` — an `*.md` under an `examples/` directory that no document cites
- `unknown-statement` — an example whose `Statements:` line carries an ID no citing standard defines
- `invalid-frontmatter-unparseable` — syntactically broken YAML in a correctly delimited block
- `invalid-frontmatter-missing-key` — a discovered document missing `description`/`scope`, and one with a present-but-empty `topics`
- `unknown-parent` — `parent:` naming no discovered document
- `cycle` — A→B→A parent cycle
- `escape` — an `examples:` entry escaping the source root (`../../etc/passwd`) and an absolute entry

Each case runs the CLI pipeline over its own corpus root and asserts a non-zero exit plus the expected diagnostic text.

### AC-7 / AC-8 / AC-9 against mutated real-corpus copies

Use the shared harness from `indexer/integration_test.go` (`repoRoot`, `copyCorpus`, `cleanBaseline`). For each: copy the REAL corpus into `t.TempDir()`, assert the clean baseline (exit 0, 18 nodes, 28 example entries), THEN mutate the copy and re-run.

- **AC-7**: delete a cited example file (e.g. `python/examples/GENERAL/config.md`) from the copy → non-zero exit; the diagnostic names BOTH the citing standard (`python/GENERAL.md`) and the missing path.
- **AC-8**: add an uncited `*.md` under `golang/examples/` in the copy → non-zero exit; the diagnostic names the uncited file.
- **AC-9**: edit an example's `Statements:` line in the copy to include an ID no citing standard defines → non-zero exit; the diagnostic names the example, the ID, and the citing standard(s).

Also assert, for each, that no output file is produced by the failing run.

## Target Files

- [x] indexer/failure_paths_test.go
- [x] indexer/tests/data/.corpora/failure/ (fixture corpora created by this task)

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- indexer/integration_test.go (`repoRoot`, `copyCorpus`, `cleanBaseline` signatures — reuse them, do not duplicate the harness)
- indexer/main.go (pipeline entrypoint, exit-code and stderr diagnostic behaviour; no-output-on-failure semantics)
- indexer/errors.go (exact diagnostic wording for missing-example, uncited-example, unknown-statement, unknown-parent, cycle and escape errors)
- /Users/pascal/Github/psauerborn/standards/python/GENERAL.md (lines 12-19 — the `examples:` citation of `examples/GENERAL/config.md` mutated by AC-7)
- /Users/pascal/Github/psauerborn/standards/python/examples/GENERAL/config.md (lines 1-4 — the `Statements:` line mutated by AC-9)

## Investigation Notes

**Harness (TASK-010, `indexer/integration_test.go`)** — reused unchanged, nothing duplicated:

- `repoRoot(t)` resolves the corpus root from `runtime.Caller(0)`.
- `copyCorpus(t, destination)` copies faithfully (dot-directories, empty directories, permission bits, symlinks).
- `cleanBaseline(t, directory)` asserts exit 0, 18 nodes, 28 example entries.
- `integrationIndex(t, source, extra...)` drives `execute` in process, writes the tree into a temporary
  directory of its own and returns `{ExitCode, Stderr, Output, Content}`. That per-run output directory is what
  makes the "no output file, no leftover temporary" assertion cheap: the directory must be empty afterwards.
- Constants reused: `exitSuccess`, `exitFailure`, `integrationModuleDirectory`.

**Pipeline (`indexer/main.go`)** — `run` collects diagnostics and returns before `RenderTree`/`writeTree`, so a
corpus that does not validate never reaches the write. `writeTree` publishes by renaming a temporary
(`.indexer-tree-*.yaml`, created in the output directory) and removes it on any error — both halves of RISK-017
are asserted by `failureWroteNothing`.

**Diagnostic wording (`indexer/errors.go`)** — verified against real runs of every fixture corpus:

| Case | Diagnostic |
| --- | --- |
| missing example | `GENERAL.md: example file not found: examples/absent.md` |
| uncited example | `examples/uncited.md: example file is not cited by any standards document` |
| unknown statement | `examples/statements.md: unknown statement ID "FIX-404" cited by GENERAL.md` |
| unparseable frontmatter | `broken.md: unparseable frontmatter: yaml: line 2: ...` |
| missing/empty key | `missing-description.md: missing required frontmatter key "description"` |
| unknown parent | `orphan.md: unknown parent "nowhere/GENERAL.md"` |
| cycle | `ALPHA.md: parent cycle: ALPHA.md -> BETA.md -> ALPHA.md` |
| escape | `GENERAL.md: example path escapes the source root: ../../etc/passwd` |

Observations that shaped the fixtures and the assertions:

- A `topics: []` that is present but empty is reported as a missing required key, as the task expected — the
  empty-value case did not need a separate error type.
- An absolute `examples:` entry (`/etc/passwd`) is reported as escaping the source root rather than joined onto
  the root and reported as not found. The `escape` corpus asserts both entries.
- Every case reports exactly the diagnostics its fault warrants and no cascading ones, so the suite pins the
  diagnostic **count** as well as the wording. The three real-corpus mutations each report exactly one line.
- The unparseable-frontmatter line ends in the YAML library's own message, so only the attribution prefix
  (`broken.md: unparseable frontmatter:`) is asserted; the rest is not this repository's wording to pin.

**No source defect was found.** Nothing outside the Target Files was edited.

**Red was demonstrated, not assumed**: with `tests/data/.corpora/failure/` moved aside, every subtest of
`TestFailurePathFixtureCorpora` fails on the missing corpus; the fixtures were then restored and the suite
passes. The three mutation tests carry their own red/green inside a single run: `cleanBaseline` requires exit 0
on the copy before the mutation, `failureRun` requires exit 1 after it.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-010 | Integration harness and happy-path acceptance | blocks | `repoRoot`, `copyCorpus` and `cleanBaseline` harness helpers; a verified clean-run baseline |

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review dependency deliverables (TASK-010 harness helpers and baseline semantics)
- [x] Create the failure-path fixture corpora under `indexer/tests/data/.corpora/failure/<case>/`
- [x] Write the failing AC-7/AC-8/AC-9 tests and the per-case fixture-corpus tests
- [x] Run `make test` and confirm failure

### 2. Green Phase

- [x] Make the assertions pass (defects found are fixed in the owning source file only if strictly required; otherwise record them in Investigation Notes)
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Improve fixtures and test structure (table-driven where cases are uniform), maintain passing tests
- [x] Confirm added tests still pass; `make lint` and `make fmt` clean

## Completion Criteria

- [x] AC-7: deleting a cited example from a `t.TempDir()` copy yields a non-zero exit and a diagnostic naming both the citing standard and the missing path
- [x] AC-8: adding an uncited `*.md` under `golang/examples/` in the copy yields a non-zero exit naming the uncited file
- [x] AC-9: an unknown ID on an example's `Statements:` line yields a non-zero exit naming the example, the ID and the citing standard(s)
- [x] Each of the eight fixture corpora (missing-example, uncited-example, unknown-statement, unparseable frontmatter, missing/empty required key, unknown parent, cycle, escape) produces a non-zero exit with the expected diagnostic
- [x] Every failing run produces no output file
- [x] Each mutation test asserts the clean baseline (exit 0, 18 nodes, 28 example entries) on the copy BEFORE mutating it
- [x] No test writes to the working tree or mutates the real corpus in place — all copies and outputs live under `t.TempDir()`
- [x] All added tests pass; `make test`, `make lint`, `make fmt` (no diff) pass in `indexer/`

## Notes

- Impact scope: this suite is the only end-to-end proof of the failure contract; weak diagnostics here ship a tool whose errors do not identify the offending file.
- Scope boundary:
  - **Fixture placement (load-bearing, CRITICAL)**: every fixture corpus MUST live under `indexer/tests/data/.corpora/failure/<case>/`. `indexer/tests/data/` sits INSIDE the `--source` root used when the tool indexes its own repository, and `indexer/` is neither dot-prefixed nor an `examples/` directory — the dot-prefixed `.corpora` directory is the ONLY thing keeping these deliberately-broken corpora out of the walk. A fixture standards document or fixture `examples/` directory placed anywhere the walker can reach makes the tool fail on its own repository (TASK-010's regression test asserts exit 0 and 18 nodes with these fixtures on disk — it must stay green).
  - **Read-only corpus**: the 18 standards documents, the 28 example files, and `docs/standards-tree.yaml` are never modified. All mutations happen on `t.TempDir()` copies.
  - Flat `package main` directly in `indexer/` (GO-008); snake_case filenames (GO-006); fixture data under `tests/data` (GO-031).
  - Do not modify `indexer/integration_test.go` or `indexer/compare.go`/`compare_test.go` (owned by TASK-010 and TASK-011) or any earlier task's source/test/fixture files, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`.
  - Permitted dependencies: `go.yaml.in/yaml/v3` (runtime), `stretchr/testify` (test-only). `gopkg.in/yaml.v3` MUST NOT be imported.
  - The binary is `indexer`; the former name `stdidx` must not appear anywhere in the implementation or fixtures.
