# Task: AC-10 structural tree comparator

Work Plan ID: WP-001
Task ID: TASK-011
Created Date: 2026-07-31
Description: Implement `indexer/compare.go` — a production-grade structural comparator that normalizes sibling node order and NOTHING else — plus `indexer/compare_test.go` containing its negative self-tests and the AC-10 comparison of the generated tree against `docs/standards-tree.yaml`.
Acceptance Criteria Covered: AC-10

## Implementation Content

### Comparator (`indexer/compare.go`)

This is production logic in its own source file with its own test file (GO-027), not a helper buried in the integration suite.

- Input: two parsed trees (the generated one and the parsed reference `docs/standards-tree.yaml`). Output: a structural difference report — a list of human-readable differences with the node path and field named, empty when the trees are structurally identical.
- **Normalize sibling node order ONLY.** The reference tree's sibling order genuinely differs from the sort-by-path order the indexer emits (e.g. `general/MAKEFILES.md` is followed by `databases/POSTGRES.md`, `databases/DYNAMODB.md`, `testing/ACCEPTANCE.md`), so comparing children as an ordered sequence would fail against a correct implementation.
- **Normalize NOTHING else.** `examples`, `statements`, `topics` and `scope` are compared as **ORDERED sequences**. Do not sort them, do not deduplicate them.
- **Distinguish an absent key from an empty one**: a node declaring no `examples` differs from a node declaring `examples: []`; likewise `aliases`. `aliases` mappings compare by key set and per-key values (mapping order is not a structural difference, but key presence is).
- Never implement AC-10 as a golden byte diff: the reference uses flow style (`statements: [API-002, API-004]`) and leading comments that block-style marshalling will not reproduce. AC-6 is the byte-level determinism check; AC-10 is structural equality.

### Tests (`indexer/compare_test.go`)

- **Negative self-tests** — parse the reference, produce mutated copies, and assert the comparator reports a difference for EACH:
  - swap two `statements` IDs within one example
  - reorder two `examples` within one node
  - drop one `topics` entry
  - add `examples: []` to a node that declares none
  - change one `scope` pattern
  - (positive control) the unmutated reference compared against itself reports no difference
  - sibling-order normalization: reordering two sibling nodes reports NO difference
- **AC-10** — generate the tree from the real repository (via the pipeline entrypoint, using `repoRoot()` from `indexer/integration_test.go`) and compare it structurally to `docs/standards-tree.yaml`; assert no differences. The failure message MUST print the reference refresh procedure: run `indexer --source . --output /tmp/tree.yaml`, hand-merge into `docs/standards-tree.yaml` preserving comments and flow style, and update the node-count snapshot constant (RISK-014).

## Target Files

- [x] indexer/compare.go
- [x] indexer/compare_test.go

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- /Users/pascal/Github/psauerborn/standards/docs/standards-tree.yaml (whole file — sibling ordering that differs from sort-by-path, flow-style `statements: [...]`, leading comments, nodes with and without `aliases`/`examples`)
- indexer/types.go (`Tree`, `Node`, `ExampleEntry`, ordered `Aliases`, absent-vs-empty modelling)
- indexer/integration_test.go (`repoRoot()` helper to reuse — do not add a second root-resolution mechanism, and never hardcode an absolute path)
- indexer/main.go (the testable `run(cfg)`/pipeline entrypoint used to generate the tree under test)

## Investigation Notes

- `docs/standards-tree.yaml`: 18 nodes, single-rooted at `GENERAL.md`. Sibling order genuinely
  differs from sort-by-path (`general/MAKEFILES.md` -> `databases/POSTGRES.md` ->
  `databases/DYNAMODB.md` -> `testing/ACCEPTANCE.md`, and `python/DOCKER.md` before
  `golang/DOCKER.md`), so sibling normalization is mandatory. `statements:` are flow style and the
  file opens with a comment block, which is why a byte/golden diff is impossible. `aliases` is
  declared on the root only; `examples` is absent on `general/DOCKER.md`, `general/MAKEFILES.md`,
  `databases/POSTGRES.md`, `javascript/GENERAL.md`.
- `indexer/types.go`: `Node` marks `aliases` and `examples` `omitempty` but `scope`, `topics` and
  `children` unconditional, so an absent key decodes to a nil slice and `examples: []` decodes to a
  non-nil empty one; that distinction is what the comparator must preserve. `AliasList` is an
  ordered `[]Alias` with its own (un)marshaller, so alias declaration order survives decoding —
  order is asserted by AC-3 in `integration_test.go`, and is deliberately NOT a structural
  difference here.
- `indexer/integration_test.go`: reusable in-package harness — `repoRoot(t)` (via `runtime.Caller`),
  `integrationIndex(t, source, extra...)` (runs `execute` in process, writes into `t.TempDir()`),
  `integrationDecode`, `integrationFind`, `integrationNodePaths`, and the `nodeCountSnapshot` = 18 /
  `exampleCountSnapshot` = 28 constants plus the `corpusChanged` refresh wording to mirror.
- `indexer/main.go`: `execute(arguments, stderr) int` is the testable entrypoint; `run(config)`
  drives discovery -> assembly -> validation -> render -> atomic write. Generating the tree under
  test therefore needs no new plumbing.
- Standards applied: GO-005 (doc comments naming the function), GO-006 (snake_case files), GO-008
  (flat package), GO-010 (data models live in `types.go`, which is out of this task's write set —
  the comparator therefore returns plain difference messages rather than introducing a new model
  type), GO-027/028/029 (one `_test.go` per source file, `TestFunctionName`, one `t.Run` per path),
  GO-032 (testify).

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-010 | Integration harness and happy-path acceptance | blocks | `repoRoot()` harness helper; a verified end-to-end pipeline producing the real-corpus tree |

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review dependency deliverables (TASK-010 harness helpers and verified pipeline)
- [x] Write the failing negative self-tests (each mutation must be reported) and the failing AC-10 comparison
- [x] Run `make test` and confirm failure

### 2. Green Phase

- [x] Add minimal implementation in `compare.go` to pass tests
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Improve code (doc comments per GO-005, clear difference messages naming node path and field), maintain passing tests
- [x] Confirm added tests still pass; `make lint` and `make fmt` clean

## Completion Criteria

- [x] The comparator normalizes sibling node order only; reordering siblings reports no difference while the reference compared against itself also reports none
- [x] Swapping two statements, reordering two examples, dropping a topic, changing a scope pattern, and adding `examples: []` to a node that declares none are EACH reported as a difference
- [x] Absent and empty keys are distinguished for both `examples` and `aliases`
- [x] The generated tree for the real repository is structurally identical to `docs/standards-tree.yaml` (no byte/golden diff anywhere in the comparison)
- [x] The AC-10 failure message prints the reference refresh procedure (regenerate, hand-merge preserving comments and flow style, update the node-count constant)
- [x] All added tests pass; `make test`, `make lint`, `make fmt` (no diff) pass in `indexer/`

## Notes

- Impact scope: AC-10 is the only end-to-end net catching a plausible-but-wrong tree; over-normalization here silently disables the strongest correctness check in the plan.
- Scope boundary:
  - `docs/standards-tree.yaml` is READ-ONLY — the tool never rewrites it and no test may modify it. Mutations for the negative self-tests are made on **in-memory parsed copies**, never on the file.
  - No test may write to the working tree or mutate the corpus in place; all generated output goes to `t.TempDir()`.
  - Flat `package main` directly in `indexer/` (GO-008); snake_case filenames (GO-006); one `_test.go` per source file (GO-027).
  - Do not modify `indexer/integration_test.go` (owned by TASK-010) or any earlier task's source/test/fixture files, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`.
  - Any fixture data (not expected for this task) belongs under `indexer/tests/data/.corpora/<case>/`; the dot prefix is what keeps fixtures out of the walked source root, and a fixture standards document reachable by the walker breaks both AC-1 and AC-10.
  - Permitted dependencies: `go.yaml.in/yaml/v3` (runtime), `stretchr/testify` (test-only). `gopkg.in/yaml.v3` MUST NOT be imported.
  - The binary is `indexer`; the former name `stdidx` must not appear anywhere in the implementation.
