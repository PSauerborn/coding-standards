# Execution Manifest: WP-001

Work Plan ID: WP-001
Last Updated: 2026-07-31

Maintained by the orchestrator: append a row to Task Results after each `task-executor` completion (from the executor's JSON response) and keep the Changeset section deduplicated. Downstream reviewers and the documenter treat this file as the definitive changeset for the work plan — they do not re-derive it from task files.

**Execution phase complete: all 12 tasks landed.** Gate verified by the orchestrator at completion: `make test` pass, `make lint` 0 issues, `make fmt-check` clean, `go test -count=2 -shuffle=on` pass, every source file paired with a `_test.go` (GO-027), standards corpus untouched.

## Task Results

| Task ID | Status | Files Modified | Tests Added |
| --- | --- | --- | --- |
| TASK-001 | completed | `indexer/go.mod`, `go.sum`, `Makefile`, `.golangci.yml`, `.gitignore`, `main.go` (placeholder) | _(none — scaffolding)_ |
| TASK-002 | completed | `indexer/types.go`, `errors.go`, `go.mod`, `go.sum`, `.golangci.yml` | `indexer/types_test.go`, `errors_test.go` |
| TASK-003 | completed | `indexer/frontmatter.go`, 2 fixtures under `.corpora/frontmatter/` | `indexer/frontmatter_test.go` |
| TASK-004 | completed | `indexer/discovery.go`, 7 fixtures under `.corpora/discovery/` | `indexer/discovery_test.go` |
| TASK-005 | completed | `indexer/examples.go`, 9 fixtures under `.corpora/examples/` | `indexer/examples_test.go` |
| TASK-006 | completed | `indexer/tree.go`, 2 fixtures under `.corpora/tree/` | `indexer/tree_test.go` |
| TASK-007 | completed | `indexer/validate.go`, 12 fixtures under `.corpora/validation/` | `indexer/validate_test.go` |
| TASK-008 | completed | `indexer/render.go` | `indexer/render_test.go` |
| TASK-009 | completed | `indexer/main.go` (replaced placeholder), 12 fixtures under `.corpora/cli/` | `indexer/main_test.go` |
| TASK-010 | completed | `indexer/README.md` | `indexer/integration_test.go` |
| TASK-011 | completed | `indexer/compare.go` | `indexer/compare_test.go` |
| TASK-012 | completed | 19 fixtures under `.corpora/failure/` | `indexer/failure_paths_test.go` |
| TASK-CODE-REVIEW-REMEDIATION | completed | `indexer/discovery.go`, `examples.go`, `validate.go`, `main.go` | `discovery_test.go`, `examples_test.go`, `validate_test.go`, `main_test.go`, `failure_paths_test.go`, 9 new fixtures under `.corpora/` |
| TASK-SEC-REMEDIATION | completed | `indexer/discovery.go`, `examples.go`, `validate.go`, `main.go` | `discovery_test.go`, `examples_test.go`, `validate_test.go`, `main_test.go` |
| TASK-QC-REMEDIATION | completed | `indexer/validate.go`, `errors.go`, `render.go`, `types.go`, `Makefile`, `README.md` | `errors_test.go`, `validate_test.go`, + doc comments across 9 `_test.go` files |
| TASK-CODE-REVIEW-REMEDIATION (round 2) | completed | `indexer/discovery.go`, `examples.go` | `discovery_test.go`, `examples_test.go`, `main_test.go`, `failure_paths_test.go`, 3 fixtures under `.corpora/failure/shared-malformed-example/` |
| TASK-QC-REMEDIATION (round 2) | completed | `indexer/errors.go`, `discovery.go`, `frontmatter.go`, `main.go` | `errors_test.go` (`TestSentinelErrors`) |
| TASK-CODE-REVIEW-REMEDIATION (round 3, finding 2 only) | completed | `indexer/discovery.go` | `discovery_test.go`, `main_test.go` |

Each task also appended investigation notes to its own file under `docs/plans/tasks/WP-001/`.

Gate re-verified by the orchestrator after all three remediations: `make test` pass, `make lint` 0 issues, `make fmt-check` clean, `go test -count=2 -shuffle=on` pass, self-index exit 0 with byte-identical reruns and 46 path entries, standards corpus and `docs/standards-tree.yaml` untouched.

## Final status

- **Acceptance: all 10 criteria met**, independently verified. The validator built its own binary and wrote its own structural comparator rather than relying on `compare.go` (which is itself under test), confirmed the temp-copy corpus produces a byte-identical tree before mutating it, and proved the AC-10 comparison non-vacuous with six negative controls. `docs/standards-tree.yaml` verified unmodified by SHA-1 before and after.
- **Final gate**: `make test` pass, `make lint` 0 issues, `make fmt-check` clean, `go test -count=2 -shuffle=on` pass.
- **Documentation**: changeset at `docs/plans/changesets/WP-001/WP-001-changeset.md`; `indexer/README.md` updated for the post-authoring remediations (source-root symlink resolution, two-part empty-corpus guard); one cross-reference paragraph added to the repository-root `README.md`, with its three external `stdidx` links deliberately preserved per user decision.
- **Ready for the user to commit. Nothing was staged, committed or pushed.**

## Changeset

Deduplicated union of all files modified across tasks, with the tasks that touched each:

**Source** (`package main`, flat layout in `indexer/` per GO-008)

- `indexer/main.go` — TASK-001 (placeholder), TASK-009 (full CLI + pipeline wiring)
- `indexer/types.go` — TASK-002
- `indexer/errors.go` — TASK-002
- `indexer/frontmatter.go` — TASK-003
- `indexer/discovery.go` — TASK-004
- `indexer/examples.go` — TASK-005
- `indexer/tree.go` — TASK-006
- `indexer/validate.go` — TASK-007
- `indexer/render.go` — TASK-008
- `indexer/compare.go` — TASK-011

**Tests** (one per source file, GO-027)

- `indexer/types_test.go`, `errors_test.go` — TASK-002
- `indexer/frontmatter_test.go` — TASK-003
- `indexer/discovery_test.go` — TASK-004
- `indexer/examples_test.go` — TASK-005
- `indexer/tree_test.go` — TASK-006
- `indexer/validate_test.go` — TASK-007
- `indexer/render_test.go` — TASK-008
- `indexer/main_test.go` — TASK-009
- `indexer/integration_test.go` — TASK-010 (shared harness: `repoRoot`, `copyCorpus`, `cleanBaseline`, `integrationIndex`)
- `indexer/compare_test.go` — TASK-011
- `indexer/failure_paths_test.go` — TASK-012

**Documentation**

- `indexer/README.md` — TASK-010

**Fixtures** — all under the dot-prefixed `indexer/tests/data/.corpora/`, per the RISK-001 placement constraint. Orchestrator-verified: no fixture exists outside `.corpora/`.

- `.corpora/frontmatter/` (2) — TASK-003
- `.corpora/discovery/` (7) — TASK-004
- `.corpora/examples/` (9) — TASK-005
- `.corpora/tree/` (2) — TASK-006
- `.corpora/validation/` (12: union, mention, uncited, heading, cascade) — TASK-007
- `.corpora/cli/` (12: valid, cascade, failures) — TASK-009
- `.corpora/failure/` (19 across 8 corpora + a README warning against repairing or relocating them) — TASK-012

**Tooling**

- `indexer/go.mod`, `indexer/go.sum` — TASK-001, TASK-002
- `indexer/.golangci.yml` — TASK-001, TASK-002
- `indexer/Makefile`, `indexer/.gitignore` — TASK-001

**Task files** (investigation notes appended by their executors)

- `docs/plans/tasks/WP-001/TASK-001.md` … `TASK-012.md`

## Remediation Round 1 — what changed

**Code review** (6 findings, all closed):

1. A cited-but-malformed example was also falsely reported as **uncited** — `ResolveExamples` drops the failed entry, so the uncited scan read absence as "nobody cites it". Fixed by skipping paths the collector marked failed. This contradicted the manifest's own earlier RISK-006 claim.
2. **The nested code-fence bug defeated RISK-005.** The fence tracker was a naive parity flip, so a four-backtick block wrapping a three-backtick block inverted the state and reclassified sample code as prose — an ID shown only in a code sample was collected as a *defined* statement, and a citing example passed REQ-9 with exit 0. Replaced with a single shared `fenceTracker` recording the opening fence character and run length, closing only on a same-character run at least as long. Driven by both `ExtractStatements` and `parseExampleStatements` so they cannot drift. The `.corpora/failure/nested-fence` corpus now exits 1 where it previously exited 0.
3. CRLF example files emitted `title: "My Title\r"`. Fixed with a shared `splitLines` stripping the carriage return at the split site, aligning example parsing with `splitFrontmatter`'s existing CRLF tolerance.
4. Citations of documents **detached** by an unknown parent or cycle were never resolved, so their example faults went unreported (REQ-7, §7). Fixed narrowly — see Outstanding below.
5. A wrapped `Statements:` line silently dropped continuation IDs. The line is now read as the markdown paragraph it opens.
6. Added the missing pipeline-level fixtures (malformed-example, duplicate-statement, nested-fence, detached-citation) with exact diagnostic-count assertions.

**Security** (3 findings, all closed): containment was decided **lexically, before link resolution**, so a symlink committed inside the corpus passed the `filepath.Rel` check and was read. Pre-fix runs demonstrably published an out-of-root file's heading as an emitted entry title, and admitted a symlinked `*.md`'s frontmatter as a corpus document. Fixed with a single `containedFile(root, relative)` helper resolving both root and candidate via `filepath.EvalSymlinks`, re-verifying containment between the *resolved* paths, and requiring a regular file — while preserving `fs.ErrNotExist` so AC-7's missing-example diagnostic does not regress. All four read sites route through it. The output temp file's `os.Chmod` by path (a TOCTOU) became a `Chmod` on the open descriptor.

**Quality** (5 findings, all closed): `MissingHeadingStatementError` relocated to `errors.go` (GO-017) and its wording pinned in the typed-error table (GO-028, 0% → 100% coverage on both methods); the `field` struct moved to `types.go` (GO-010); 36 test functions documented (GO-005) — five more than the report listed, because the two prior remediations had added functions. Behaviour was proven unchanged by building a reconstructed pre-change binary and diffing the emitted tree byte-for-byte against the post-change binary (identical).

## Remediation Round 2 — what changed

Round-2 review found two new code-review defects and one quality finding; all three were remediated.

- **Symlinked `--source` silently emitted `nodes: []` and exited 0** (high). `filepath.WalkDir` lstats its root, so a `--source` whose final component is a symlink arrived as a non-directory entry and nothing was descended — returning zero paths *and a nil error*. A trailing slash changed the result entirely. Fixed by extracting `resolveRoot(root)` and resolving before the walk; an unresolvable root now returns that error rather than an empty result. `DiscoverDocuments` additionally reports a source under which no markdown file was found.
- **A malformed example cited by two standards was reported twice** (medium). `ResolveExamples` re-parses per citing document and the collector does not deduplicate. Fixed by skipping the add when the path is already marked failed, with suppression deliberately confined to file-attributed diagnostics so citer-attributed ones still name each distinct citer.
- **GO-017**: the four package sentinels (`errEscapesRoot`, `errNotRegularFile`, `errUnterminatedFrontmatter`, `errArgumentsReported`) were relocated into `errors.go` as a pure move, with `TestSentinelErrors` pinning their wording.

Both remediations preserved byte-identical output for this repository, verified by building pre- and post-change binaries and diffing.

## Review status at the 2-iteration cap

| Reviewer | Round 1 | Round 2 | Final |
| --- | --- | --- | --- |
| validation-runner | pass | pass | **clean** |
| security-reviewer | 3 findings | 0 findings | **clean** |
| quality-controller | 13 violations | 1 violation | **clean** |
| risk-reviewer | 0 deviations | _(not re-run — raised nothing)_ | **clean** |
| code-reviewer | 6 findings | 2 findings | **3 open** (below) |

## Code-review findings at the iteration cap — user-directed disposition

The flow's Review & Remediation Loop caps at 2 iterations. Three findings were raised by the third code-review pass. **The user directed that finding 2 be fixed and findings 1 and 3 be deferred**, so they are open by decision rather than oversight.

**Finding 2 — RESOLVED.** A second guard was added after the discovery loop in `DiscoverDocuments`: `len(documents) == 0 && collector.Len() == 0` now returns "does not name a standards corpus". The `collector.Len() == 0` conjunct is load-bearing, preserving the round-2 behaviour where a corpus whose only document has unparseable frontmatter still gets the more useful diagnostic naming that document. Proven red first (the pre-fix binary replaced a populated `deployed.yaml` with `nodes: []` at exit 0; post-fix it exits 1 and leaves the file byte-for-byte intact). Every existing fixture corpus used as a `--source` root was checked to hold either a titled document or a non-empty collector, so the guard cannot fire spuriously. Emitted output remains byte-identical across all four `--source` spellings.

### DEFERRED by user decision — follow-up owed

1. **`discovery.go` — containment relativity (high).** `filepath.EvalSymlinks` preserves its input's relativity, so a relative `--source` yields a relative resolved root while a symlink inside the corpus with an absolute target yields an absolute resolved path. `filepath.Rel` cannot relate a relative base to an absolute target, returns an error, and that error is mapped to `errEscapesRoot` — so a file plainly inside the root is judged an escape, and the judgement flips purely on how `--source` was spelled. In the walk the file is dropped silently (different tree, still exit 0); in `ResolveExamples` it draws a false escape diagnostic. Contradicts REQ-6 determinism and the RISK-008 claim that relative and absolute `--source` produce identical output. **Not reachable in this repository, which contains no symlinks.** Suggested fix: make both operands absolute before comparing, keeping emitted paths relative and unchanged.
2. **`discovery.go` — empty-corpus guard too narrow (medium).** The guard fires only when the walk finds *no* markdown file at all, but its own doc comment states the hazard it exists to close: a `--source` that is not a standards corpus "publishes an index declaring the corpus empty — over whatever tree was deployed before — and exits zero while doing it." A directory containing any markdown file (a bare `README.md` suffices) still takes that path. Reproduced: pointing `--source` at a non-corpus directory replaced a populated `deployed.yaml` with `nodes: []` at exit 0. Suggested fix: add a second guard for `len(documents) == 0 && collector.Len() == 0`, which preserves the round-2 behaviour for a corpus whose only document has unparseable frontmatter.
3. **`examples.go` — suppression key space (low, reviewer-recorded as follow-up).** The `IsFailed` suppression is keyed on a path in the same key space `DiscoverDocuments` marks, so a file that is both a discovered document and a cited example has its example-parse diagnostics suppressed by a document-level failure. Cannot cause a false exit 0 — every `MarkFailed` is paired with an `Add`, so the run still fails — and the suppressed diagnostics surface on the next run once the first fault is repaired.

## Outstanding items (carried forward, not defects blocking completion)

- **Code-review finding 4 was fixed narrowly.** The reviewer suggested resolving every document's citations once into the real collector and having `BuildTree` consume the results. That requires editing `tree.go`, which was outside the remediation task's write set, so the executor instead routes detached documents' citations into the real collector while placed documents keep using a throwaway one. **The defect is closed** (detached documents' faults are now reported, pinned by the `detached-citation` corpus and a prefixed-vs-unprefixed regression test), but the structural double-resolution the finding wanted removed remains.
- **Two untyped diagnostics remain** and want a single follow-up covering `errors.go` + `examples.go` + `validate.go`: `example %s cited by %s is not a regular file` (`examples.go`, added by the security remediation) and `failed to read example %s: %w` (`validate.go`). Each remediation task's write set excluded one of the pair, and typing one alone would preserve the inconsistency the change is meant to remove.
- **GEN-003 recorded as a documented deviation, not fixed.** The `Makefile` targets invoke the host toolchain with no containerized build path. Per the user's approved binary-only boundary — and because no other module in this repository ships a Dockerfile — no Dockerfile was added; the rationale is recorded in the `Makefile` header and `indexer/README.md`. Adding one would pull `general/DOCKER.md` and `golang/DOCKER.md` in as newly applicable standards for a checkout-local tool that would have to bind-mount the repository back in to reach its own corpus.
- **RISK-018, RISK-019 and RISK-020** were omitted from this manifest's earlier mitigation list. The risk reviewer verified all three are in fact honoured in the implementation (README discovery-rule warning naming `docs/` templates as the hazard; README naming `docs/standards-tree.yaml` the authoritative schema contract; Task 10 split into TASK-010/011/012 with the comparator in its own source+test pair). The gap was in this record, not the code.
- **RISK-007's `-count=10 -shuffle=on` clause** is realized as an in-process 10-iteration loop plus a `t.Log` hint rather than a `make` target — the form sanctioned at decomposition. Every substantive element of the countermeasure is implemented. A `test-determinism` target would close it if CI is added.

## Open Items for the Review Stage

Both items raised during execution have been **resolved** in remediation round 1:

- ~~**GO-017 deviation (TASK-007)**: `MissingHeadingStatementError` declared in `validate.go` rather than `errors.go`.~~ **RESOLVED** — relocated to `errors.go` by TASK-QC-REMEDIATION, and its wording pinned in `TestTypedErrorMessages` (GO-028: both methods went from 0% to 100% coverage).
- ~~**GO-010 note (TASK-011)**: `CompareTrees` returns formatted strings rather than a structured type.~~ **REFUTED** by `quality-controller` — `compare.go` declares no data model, so GO-010 has no subject there. A genuine GO-010 violation was found elsewhere (the `field` struct in `render.go`) and relocated to `types.go`.

## Accepted standards deviations (user-adjudicated — NOT defects)

Per the work plan's Notes and the user's decisions, `quality-controller` must treat these as accepted:

| Standard | Deviation and rationale |
| --- | --- |
| GO-018 | Configuration via CLI flags only — single-shot CLI, no config file |
| GO-020 | Reduced to flag presence/path checks in `main.go`; no config schema to validate |
| GO-022 | `go-playground/validator` forbidden by spec §6 |
| GO-023 | `spf13/viper` forbidden by spec §6 |
| GO-033 | `sirupsen/logrus` forbidden by spec §6; diagnostics are plain stderr lines per the CLI contract |
| ACPT-002 / ACPT-004 | Go unit + integration tests cover all 10 ACs; no `cucumber/godog` |

Permitted dependencies: `go.yaml.in/yaml/v3` (runtime) and `stretchr/testify` (test-only).

## Execution Notes

### Toolchain

- **RISK-022 materialized (TASK-001)**: golangci-lint v1.64.8 (built with go1.25) refused a module targeting go1.26.5, so the `go` directive was initially pinned to `1.25.0`.
- **Follow-on blocker (TASK-002)**: once `_test.go` files imported testify, `net/http` pulled go1.26 stdlib sources the v1 linter could not typecheck. **User approved a global upgrade**; the orchestrator installed golangci-lint **v2.12.2 (built with go1.26.5)**. The `go` directive was consequently raised back to `1.26.5`.
- **`.golangci.yml` migrated to the v2 schema**. The automatic migration silently dropped `errcheck`, `govet`, `ineffassign`, `staticcheck` and `unused` because they are v2 defaults; these were re-listed explicitly so the enabled set stays visible. `gosimple` folded into `staticcheck`; `gofmt`/`goimports` moved to `formatters`; v1's `exclude-use-default: false` became "no `exclusions.presets`" so revive's missing-doc-comment reports still surface for GO-005.

### Risk mitigations implemented and verified

- **RISK-001** *(critical)* (TASK-004, TASK-007, TASK-010): `WalkMarkdownFiles(root, policy)` always skips dot-prefixed directories and takes an `ExampleDirectoryPolicy`; TASK-007's REQ-8 scan **reuses that single walker** rather than performing a second unfiltered walk. `TestIntegrationFixtureCorporaAreNotIndexed` runs the pipeline with `--source <repo root>` while fixtures exist on disk, requires `.corpora` to contain at least one titled document so it cannot pass vacuously, and asserts exit 0 with 18 nodes. **The risk was reproduced out-of-band**: a scratch copy with `.corpora` renamed to `corpora` exits 1 with 15 fixture diagnostics.
- **RISK-002** (TASK-003): `splitFrontmatter` anchors the opening `---` at byte offset 0 (CRLF tolerated, padded delimiters rejected, fenced blocks never inspected). The README-shaped fixture yields no frontmatter and no error.
- **RISK-003** (TASK-011): `CompareTrees` normalizes sibling order **only** — siblings matched by path, everything else compared as ordered sequences, absent distinguished from declared-but-empty. 17 negative self-tests prove it discriminates (swapped statements, reordered/dropped example, changed example path/title, dropped topic, changed scope pattern, dropped `scope` key, `examples: []`/`aliases: {}` added where none declared, changed alias value, changed title/description, renamed/missing/added node, duplicated sibling path), each asserted in both directions, plus a sibling-reorder guard that asserts the reordering actually changed order so the positive control cannot pass vacuously. No byte or golden diff anywhere.
- **RISK-004** (TASK-004, TASK-005, TASK-007): one shared grammar `StatementIDPattern` = `\[([A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*-\d+)\]`, table-tested against all eighteen prefix families including three-segment `PY-DOCKER-001`, `GO-WRK-011`. The example heading regex is built from it, so no second pattern exists.
- **RISK-005** (TASK-004, TASK-007): `ExtractStatements` is definition-only — line-anchored backticked ID followed by `**MUST**`/`**SHOULD**`, fenced regions skipped. A `mention/` fixture proves a prose- or code-fence-only ID does not satisfy REQ-9.
- **RISK-006** (TASK-002, TASK-006, TASK-007, TASK-009): `ErrorCollector` orders diagnostics by document path then message; `MarkFailed`/`IsFailed` drive cascade suppression, so one corrupted document yields exactly one diagnostic (pinned by test).
- **RISK-007** (TASK-008, TASK-010): output is built entirely through the `yaml.Node` API from an ordered `[]field` slice per mapping — **no Go map is iterated anywhere**. Root aliases assert the exact sequence `go, js, ts, py, postgres, pg, deploy, containerization`. AC-6 is **two separate process invocations** of the built binary compared byte-for-byte, not two in-process renders.
- **RISK-008** (TASK-009, TASK-010): no timestamp, version, hostname or absolute path in the emitted document; relative vs absolute `--source` from different working directories produce identical output.
- **RISK-009 / RISK-016** (TASK-006): every emitted node and example path funnels through a single `emitPath(prefix, target)` helper (`filepath.ToSlash` + `path.Clean` + `path.Join`), applied strictly at node construction so parent lookup and example resolution stay unprefixed. Prefix spellings `p`, `p/`, `p//`, `./p` all collapse to one separator.
- **RISK-010** (TASK-005): containment decided by `filepath.Rel` on the joined-and-cleaned path plus outright rejection of absolute entries — never substring matching of `..`. `examples/../examples/config.md` correctly passes; an absolute entry is reported as escaping the root rather than as not-found.
- **RISK-011** (TASK-008): all string scalars tagged `!!str` with zero (default) style, leaving quoting to the library. A `'*'` scope re-parses to the string `*`. The executor verified against v3.0.5 that an **untagged** scalar would re-parse `123`/`true`/`null` as int/bool/null — the explicit tag prevents that corruption class.
- **RISK-012** (TASK-001, TASK-002): depguard `archived-yaml` deny rule **re-proven to fire** after the v2 migration by a temporary probe import (lint exit 2), probe reverted and file confirmed byte-identical. `go list -deps .` reports zero occurrences of `gopkg.in/yaml.v3` in the binary's graph.
- **RISK-013** (TASK-004, TASK-006, TASK-010, TASK-012): repository root resolved via `runtime.Caller`; no absolute path literal in any test. `copyCorpus` preserves dot-directories, empty directories, permission bits and symlinks, and errors loudly on a non-regular/non-symlink entry rather than skipping silently. `cleanBaseline` asserts exit 0 / 18 nodes / 28 example entries **before** every mutation.
- **RISK-014** (TASK-010, TASK-011): AC-1 re-derives the expected node set structurally from disk and asserts set equality, with `18`/`28` kept as clearly-labelled snapshot constants whose failure message states the corpus changed and names the refresh procedure. AC-10's failure message prints the refresh procedure too.
- **RISK-015** (TASK-008): `aliases` and `examples` emitted only when non-empty, so neither `aliases: {}` nor `examples: []` can appear; `children` always emitted.
- **RISK-017** (TASK-009, TASK-012): output written to a dot-prefixed temp file in the output directory, synced, chmod 0644, then renamed — only when the collector is empty. Every failing-run test asserts no tree file was written, a pre-existing output file was not truncated, and no `.indexer-tree-*.yaml` temp file was left behind.
- **RISK-021** (TASK-008): render fixtures include an example-bearing leaf, exercising the TASK-005/TASK-006 coupling where it is introduced.
- **RISK-022** (TASK-001): see Toolchain above.

### Load-bearing design decisions

- **Pipeline order** (TASK-009): `BuildTree` runs *before* `ValidateCorpus` because it is the stage that resolves citations and marks broken example files failed, and that marking drives cascade suppression. Since `BuildTree` internally calls `ResolveExamples`, the citation map `ValidateCorpus` needs is rebuilt by `resolveCitations` against a throwaway collector — otherwise every citation diagnostic would be reported twice. The `failures` fixture test asserting exactly three diagnostics pins this.
- **Empty `topics: []`** is reported as `missing required frontmatter key "topics"`; no separate empty-value error type was needed.

### Scope extensions

- **TASK-002** was explicitly authorized by the orchestrator to modify the TASK-001-owned files `go.mod`, `go.sum` and `.golangci.yml` to resolve the toolchain blockers. `Makefile`, `.gitignore` and `main.go` were left untouched.

### Pre-existing working-tree state (NOT part of this changeset)

- Repo-root `.gitignore`, `.markdownlint.yaml` and `CLAUDE.md` were already modified before execution began (mtimes 04:16–04:20, hours before the first task write). `.markdownlint.yaml` was additionally touched by the orchestrator-directed `stdidx` → `indexer` rename, alongside `docs/specs/SPEC-001.md`, `docs/plans/2026-07-31-WP-001.md`, `docs/plans/risk/WP-001/WP-001-risk-plan.md`, and the two leading comment lines of `docs/standards-tree.yaml` (comments only — no YAML data changed; the file is read-only input and was never written by the tool).
