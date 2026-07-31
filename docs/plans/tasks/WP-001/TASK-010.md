# Task: Integration harness and happy-path acceptance (AC-1..AC-6)

Work Plan ID: WP-001
Task ID: TASK-010
Created Date: 2026-07-31
Description: Build the shared integration harness (`runtime.Caller`-based repo-root resolution, one `copyCorpus` helper, clean-baseline assertion) in `indexer/integration_test.go`, prove AC-1 through AC-6 against the real corpus, add the fixtures-on-disk regression test, and document the discovery rule and refresh procedure in `indexer/README.md`.
Acceptance Criteria Covered: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6

## Implementation Content

### Shared harness (RISK-013) — consumed by TASK-011 and TASK-012

Define in `indexer/integration_test.go`, exported to the package's other test files:

- `repoRoot(t)` — resolve the repository root via `runtime.Caller(0)` plus a relative join (the test file lives in `<repo>/indexer/`). **Never a hardcoded absolute path** such as `/Users/pascal/...`; the suite must be machine-independent.
- `copyCorpus(t, dstDir)` — copy the REAL repository corpus into a caller-supplied `t.TempDir()`. The copy must be faithful: do not skip dot-directories, empty directories, or symlinks, since a lossy copy weakens every failure-path assertion built on it.
- `cleanBaseline(t, dir)` — before any mutation, assert the copy produces a clean run: exit 0, **18 nodes**, **28 example entries**. A defective copy then fails as itself rather than as a false acceptance-test failure.

### Acceptance coverage against the real corpus

- **AC-1**: exactly 18 nodes; no node for any `*.md` lacking parseable frontmatter with a `title`, and none for any file under `examples/`. Derive counts structurally where possible; keep `18` as ONE clearly-labelled snapshot constant whose failure message states that the corpus changed and the constant must be updated together with `docs/standards-tree.yaml` (RISK-014).
- **AC-2**: single root `GENERAL.md`; `python/DOCKER.md` is a child of `general/DOCKER.md`; `golang/API.md` is a child of `golang/GENERAL.md`. Assert these hierarchy facts BOTH without a prefix and under `--prefix` (RISK-016).
- **AC-3**: `javascript/GENERAL.md` node's scope is the list `['*.js', '*.ts', '*.vue']`; the root node carries the 8-key `aliases` map (`go, js, ts, py, postgres, pg, deploy, containerization`).
- **AC-4**: the `python/GENERAL.md` node lists `python/examples/GENERAL/config.md` with title `Configuration Loading` and statements `PY-019` through `PY-024`.
- **AC-5**: with `--prefix {prefix}`, EVERY `path` in the output — node and example — starts with `{prefix}/`.
- **AC-6**: **two separate PROCESS invocations** of the binary built by `make build`, outputs byte-compared. An in-process double render exercises far less nondeterminism (RISK-007). Also assert a relative `--source` and an absolute `--source` (from different working directories) produce byte-identical output (RISK-008).

### Fixture-poisoning regression test (RISK-001, CRITICAL)

Run the full pipeline with `--source <repo root>` **while the fixture corpora exist on disk under `indexer/tests/data/.corpora/`**, asserting exit 0 and exactly 18 nodes. This is the test that proves the tool still works on its own repository.

### `indexer/README.md`

Document, briefly: the CLI contract (`indexer --source <dir> --output <file> [--prefix <p>]`); the discovery rule (a `*.md` is a standards document iff its FIRST line opens YAML frontmatter containing `title`; `examples/` and dot-prefixed directories are skipped) with a warning that any document template under `docs/` gaining `title:` frontmatter silently becomes a node; the fixture-placement rule (`indexer/tests/data/.corpora/` is dot-prefixed so the walker skips it — never move fixtures out of it); that `docs/standards-tree.yaml` is the authoritative output-schema contract; and the refresh procedure when the corpus legitimately changes (`indexer --source . --output /tmp/tree.yaml`, then hand-merge into `docs/standards-tree.yaml` preserving its comments and flow style, and update the node-count constant).

## Target Files

- [x] indexer/integration_test.go
- [x] indexer/README.md

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- indexer/main.go (the testable `run(cfg)` entrypoint and exit-code behaviour)
- indexer/Makefile (the `build` target and the binary name/location used for the two-process AC-6 test)
- indexer/types.go (`Tree`/`Node` shape for structural assertions)
- /Users/pascal/Github/psauerborn/standards/javascript/GENERAL.md (lines 1-15 — the three-pattern `scope` list asserted by AC-3)
- /Users/pascal/Github/psauerborn/standards/python/examples/GENERAL/config.md (lines 1-4 — the title and statement IDs asserted by AC-4)
- /Users/pascal/Github/psauerborn/standards/docs/standards-tree.yaml (lines 1-60 — node/example shape and the 8-key `aliases` order)

## Investigation Notes

- `indexer/main.go` exposes three in-process entrypoints: `execute(arguments []string, stderr io.Writer) int`
  (full CLI including flag parsing, returns the exit code; `os.Exit` is confined to `main`), `parseConfig`
  and `run(config Config) error`. Exit codes are the package constants `exitSuccess` (0) / `exitFailure` (1) —
  the harness compares against those rather than literals. `writeTree` writes through a dot-prefixed temporary
  file and renames it onto `--output`, so the output file only exists after a successful run.
- `indexer/Makefile` `build` target is `go build -o indexer .`, i.e. it writes the binary INTO the working tree.
  The two-process AC-6 test therefore performs the same compilation with `go build -o <t.TempDir()>/indexer .`
  (`cmd.Dir` = the `indexer` directory) so no test mutates the working tree; the binary is byte-identical in
  behaviour to the `make build` artefact and `/indexer` is `.gitignore`d anyway.
- `indexer/types.go`: `Tree{Nodes []*Node}`, `Node{Path,Title,Description,Scope []string,Topics,Aliases AliasList,
  Examples []ExampleEntry,Children []*Node}`, `ExampleEntry{Path,Title,Statements}`. All carry `yaml` tags and
  `AliasList` implements `UnmarshalYAML`, so the rendered tree can be decoded straight back into `Tree` — the
  output schema is asserted structurally instead of by string matching.
- `javascript/GENERAL.md` frontmatter declares `scope: ['*.js', '*.ts', '*.vue']` as a block list (AC-3), and
  `parent: GENERAL.md`.
- `python/examples/GENERAL/config.md` heading is `# [PY-023] Configuration Loading` with
  `Statements: [PY-019] [PY-020] [PY-021] [PY-022] [PY-023] [PY-024]` (AC-4: title `Configuration Loading`,
  statements `PY-019`..`PY-024`).
- `docs/standards-tree.yaml` root `GENERAL.md` carries the 8-key `aliases` map in the order
  `go, js, ts, py, postgres, pg, deploy, containerization`; node and example entries both key on `path`.
- Discovery rule verified empirically against the working tree: 43 `*.md` files open with `---` frontmatter,
  but only 18 sit outside a dot-prefixed directory — `.claude/commands/*.md` (2) and
  `indexer/tests/data/.corpora/**` (23) are excluded solely by the dot-prefix rule. This is exactly the
  RISK-001 exposure, so the structural derivation in the AC-1 test re-derives the 18 independently of the
  indexer's own walker.
- Helper names already taken in `package main` tests: `repositoryRoot`, `exampleRepositoryRoot`,
  `validationRepositoryRoot`, `cliRepositoryRoot`, `fixtureRoot`, `readFixture`, `readDiscoveryFixture`,
  `treeFindNode`, `treeNodePaths`, `documentPaths`, plus the `cli*`/`render*`/`tree*` families. The new file
  therefore uses `repoRoot`/`copyCorpus`/`cleanBaseline` (mandated) and prefixes every other helper with
  `integration`.
- `.golangci.yml` enables `revive` with no exclusion preset, so every helper needs a doc comment starting with
  its own name (GO-005), and `depguard` denies `gopkg.in/yaml.v3` (use `go.yaml.in/yaml/v3`).

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-009 | CLI entrypoint and pipeline wiring | blocks | The `indexer` CLI and its testable `run(cfg)` pipeline entrypoint; `make build` binary |

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review dependency deliverables (TASK-009 CLI entrypoint, exit codes, output write semantics)
- [x] Write the harness (`repoRoot`, `copyCorpus`, `cleanBaseline`) and the failing AC-1..AC-6 assertions plus the fixtures-on-disk regression test
- [x] Run `make test` and confirm failure

  The pipeline was already implemented, so the red phase was run as a perturbation
  sweep instead: each expectation was inverted in turn and the suite re-run to prove
  the assertion bites. `nodeCountSnapshot` 18 -> 19 failed AC-1, AC-5, the RISK-001
  regression and `cleanBaseline`; the AC-4 title, the AC-3 scope list, the AC-2
  `python/DOCKER.md` parentage (both prefixed and unprefixed) and the AC-6
  relative-source comparison each failed on their own. The RISK-001 condition was
  reproduced directly: a copy of the repository with `.corpora` renamed to `corpora`
  exits 1 with 15 fixture diagnostics, which is exactly what that test asserts against.

### 2. Green Phase

- [x] Make the assertions pass (implementation defects found here are fixed in the owning source file only if strictly required; otherwise record them in Investigation Notes)
- [x] Run only added tests and confirm they pass, including the two-process AC-6 comparison

  No implementation defect surfaced: all eight new test functions pass against the
  real corpus without a change to any source file.

### 3. Refactor Phase

- [x] Improve the harness (single `copyCorpus`, clear failure messages), maintain passing tests
- [x] Write `indexer/README.md`; confirm added tests still pass; `make lint` and `make fmt` clean

## Completion Criteria

- [x] The harness resolves the repository root via `runtime.Caller` plus a relative join — no absolute path literal appears in any test file
- [x] `copyCorpus` produces a copy that passes `cleanBaseline`: exit 0, 18 nodes, 28 example entries, asserted before any mutation
- [x] AC-1: a real-corpus run yields exactly 18 nodes; the `18` snapshot constant is labelled and its failure message states the corpus changed and names the refresh procedure
- [x] AC-2: single root `GENERAL.md`, `python/DOCKER.md` under `general/DOCKER.md`, `golang/API.md` under `golang/GENERAL.md` — asserted both with and without `--prefix`
- [x] AC-3: `javascript/GENERAL.md` scope is `['*.js', '*.ts', '*.vue']`; the root carries the 8-key `aliases` map
- [x] AC-4: `python/GENERAL.md` lists `python/examples/GENERAL/config.md` with title `Configuration Loading` and statements `PY-019`..`PY-024`
- [x] AC-5: with `--prefix`, every node AND example path starts with `{prefix}/`
- [x] AC-6: two separate process invocations of the `make build` binary produce byte-identical output; relative vs absolute `--source` also byte-identical
- [x] RISK-001 regression: running the pipeline with `--source <repo root>` while fixture corpora exist under `indexer/tests/data/.corpora/` exits 0 with exactly 18 nodes
- [x] `indexer/README.md` documents the CLI contract, discovery rule, fixture-placement rule, schema authority, and reference-refresh procedure
- [x] All added tests pass; `make test`, `make lint`, `make fmt` (no diff) pass in `indexer/`

## Notes

- Impact scope: TASK-011 (AC-10 comparator) and TASK-012 (AC-7/8/9) both consume `repoRoot`, `copyCorpus` and `cleanBaseline` from this file — keep their signatures stable and package-visible.
- Scope boundary:
  - **Fixture placement (load-bearing, CRITICAL)**: all synthetic fixture corpora live under `indexer/tests/data/.corpora/<case>/`. `indexer/tests/data/` sits INSIDE the `--source` root this task indexes; the dot-prefixed `.corpora` directory is the ONLY thing keeping fixtures out of the walk. Do not create fixture standards documents or fixture `examples/` directories anywhere else — this task's regression test exists precisely to catch that.
  - **No test may write to the working tree or mutate the corpus in place.** All copies and outputs go to `t.TempDir()`. The 18 standards documents, the 28 example files, and `docs/standards-tree.yaml` are READ-ONLY.
  - Flat `package main` directly in `indexer/` (GO-008); snake_case filenames (GO-006).
  - Do not modify files owned by earlier tasks (`types.go`, `errors.go`, `frontmatter.go`, `discovery.go`, `examples.go`, `tree.go`, `validate.go`, `render.go`, `main.go`, their `_test.go` files and fixtures, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`).
  - Permitted dependencies: `go.yaml.in/yaml/v3` (runtime), `stretchr/testify` (test-only). `gopkg.in/yaml.v3` MUST NOT be imported.
  - The binary is `indexer`; the former name `stdidx` must not appear anywhere in the implementation, tests, or `indexer/README.md`.
