# Quality Report: Standards Tree Indexer

Work Plan ID: WP-001
Created Date: 2026-07-31
Revision: 3 — re-review after remediation round 2 (final iteration under the flow's 2-iteration cap)

## Coding Standards Review

Applicable Standards:

- `GENERAL.md` — Cross-language general standards (scope `*`)
- `golang/GENERAL.md` — Golang general standards (scope `*.go`, topic `golang`)
- `general/LOGGING.md` — Logging general standards (scope `*`, topic `logging`) — see Adjudicated Deviations
- `testing/ACCEPTANCE.md` — Acceptance testing standards (scope `*`, topic `testing`) — see Adjudicated Deviations

Not applicable to this changeset: `general/DOCKER.md` and `golang/DOCKER.md` (no Dockerfile in the
changeset — see the GEN-003 confirmation below), `general/API.md` / `golang/API.md` (no HTTP
surface), `golang/WORKER.md` (no message broker), `databases/*` (no persistence layer — GO-034
through GO-045 have no subject), `javascript/*`, `python/*`, `terraform/*`.

Scope of this revision: the single GO-017 finding of revision 2, plus a full sweep of the code the
round-2 remediations added — `resolveRoot` (`discovery.go:125`), the empty-corpus guard in
`DiscoverDocuments` (`discovery.go:303-306`), the `IsFailed` suppression in `ResolveExamples`
(`examples.go:83-103`), the relocated sentinel block (`errors.go:10-31`), and every test function
accompanying them.

### Coding Standards Violations

Total Violations: 0

No violation requiring remediation was found in this revision. The one item still open is a
**SHOULD** deviation whose severity is adjudicated below rather than remediated.

## Revision 2 Finding — Closure Verification

| Rule | Revision 2 location | Status |
|------|--------------------|--------|
| GO-017 | `discovery.go:115,119` — `errEscapesRoot`, `errNotRegularFile` | **CLOSED, and wider than reported.** All four sentinels now sit in one documented block at `errors.go:10-31`: `errEscapesRoot` (`:17`), `errNotRegularFile` (`:21`), `errUnterminatedFrontmatter` (`:25`) and `errArgumentsReported` (`:31`). The remediation also took the two the revision-2 report named only as out-of-changeset observations, so the package's sentinel vocabulary is no longer split three ways. `grep -n "errors.New" *.go` over the non-test sources returns those four declarations and nothing else outside `errors.go` apart from the two inline `main.go` flag messages (`:129`, `:132`), which are argument-validation strings returned once each and never matched on. |

**The relocation is verified as a pure move.** The four messages are byte-identical to the strings
the revision-2 report and the pre-move call sites carried, and every consumer still resolves:
`errUnterminatedFrontmatter` is wrapped by `ParseFrontmatter` (`frontmatter.go:46`),
`errArgumentsReported` is produced at `main.go:115` and matched at `main.go:78` (`report` stays
silent for it), and `errEscapesRoot` / `errNotRegularFile` are produced by `containedFile`
(`discovery.go:166,170,180`) and matched with `errors.Is` at `examples.go:59,62`. `make test`,
`make lint` (0 issues) and `gofmt -l .` (clean) all pass on the moved tree.

**The wording is now pinned.** `TestSentinelErrors` (`errors_test.go:130-147`) asserts each of the
four messages with `assert.EqualError` in its own `t.Run` block, carries a doc comment stating why
the messages are output rather than an implementation detail (GO-005), and satisfies GO-028 and
GO-029 for the block it covers.

## Round-2 Code Sweep — New Code

All four additions were reviewed against the full Go and general rule sets. No violation found.

| Addition | Verification |
|----------|--------------|
| `resolveRoot` (`discovery.go:125-131`) | GO-005: doc comment leads with the function name and states why both halves of the corpus decision go through it. GO-007: pure function of its argument. GO-025/GO-028: **100.0%** of statements in `go tool cover -func`, both branches pinned by named subtests — `"a symlinked root walks the corpus it names"` (`discovery_test.go:201`) and `"a root that cannot be resolved is reported"` (`discovery_test.go:223`), the latter asserting the root name appears in the message. No dedicated `TestResolveRoot`: consistent with the reading revision 2 recorded and the orchestrator accepted, that unexported helpers are covered through their callers' named subtests. The deliberate `%v` rather than `%w` is correct and documented — wrapping would let an unresolvable root satisfy the `errors.Is(err, fs.ErrNotExist)` arm at `examples.go:66` and be reported as a missing example file. |
| Empty-corpus guard (`discovery.go:303-306`) | GO-005: the reason is written into the `DiscoverDocuments` doc comment, including why the check is on walked files rather than on produced documents. GO-025/GO-029: pinned by `"reports a source root holding no markdown file"` (`discovery_test.go:399`), which asserts the error, the nil result and that the message names the offending directory. GEN-001/GEN-002: four lines, no new type or structure. |
| `IsFailed` suppression in `ResolveExamples` (`examples.go:83-103`) | GO-005: the eleven-line comment states why the marking and the drop stay unconditional while the reporting does not. GO-029: both directions pinned — `"a malformed example resolved a second time adds no diagnostic"` (`examples_test.go:246`, which also asserts the failure marking survives) and its complement `"two citers of the same missing example are both reported"` (`:273`), which guards against the suppression being widened into citation faults. GO-032: all assertions through testify. |
| Accompanying tests | GO-005: every one of the 22 `.go` files' top-level functions carries a doc comment whose first word is the function name — re-verified mechanically over source and tests, methods included. GO-026: every test file imports `testing`. GO-029: subtests throughout. GO-032: no `t.Errorf` / `t.Fatalf` comparison assertion anywhere in the suite; the single `fmt.Errorf` in a test file (`integration_test.go:141`) is a helper's return value, not an assertion. GO-031: fixtures remain under `indexer/tests/data/.corpora/`, with the symlink corpora built under `t.TempDir` for the recorded portability reason. |

## Outstanding GO-016 Cluster — Severity Judgement

**Verdict: a legitimate follow-up, not a blocker. This does not have to be fixed before shipping.**

The sites are confirmed present. They divide into two categories that the manifest and the round-2
executor conflated, and only the first is the follow-up's real subject.

**Category 1 — five untyped `collector.Add(fmt.Errorf(...))` diagnostics.** These bypass the
`DiagnosticError` interface that the whole collector is built around:

- `examples.go:63` — `example %s cited by %s is not a regular file`
- `examples.go:70` — `failed to resolve example %s cited by %s: %w`
- `examples.go:77` — `failed to read example %s cited by %s: %w`
- `validate.go:231` and `validate.go:237` — `failed to read example %s: %w`

The observable consequence is bounded and precisely identifiable: `documentPathOf` (`errors.go:370`)
cannot attribute them, so `ErrorCollector.Errors` (`errors.go:340`) sorts them under the empty path
and they appear ahead of every path-attributed diagnostic instead of beside the document they
concern. Determinism is **not** affected — the sort key `(documentPath, message)` stays total and
the sort stable, which is why `go test -count=2 -shuffle=on` and the byte-identical rerun check both
still pass. Each message names both the example file and the citing document, so a reader still gets
a repairable instruction; what differs is the `path: message` prefix shape of the typed diagnostics.
Three of the five branches are additionally hard to reach (a file that passes containment and then
fails to read — a permission change or a race), and the most reachable one, `examples.go:63`, has
its exact wording pinned by `examples_test.go:235-237` even though it is untyped.

So the defect class is diagnostic *attribution and stderr formatting on rare failure branches*, not
correctness, not security, not determinism, and not a user-facing behaviour of any passing run. GO-016
is a **SHOULD**, and `GENERAL.md`'s meta rules permit a recorded deviation from a SHOULD. Shipping
with it costs a slightly inconsistent stderr layout in failure modes the corpus does not currently
produce; forcing it into a third remediation round at the iteration cap costs a touch of every read
site in `examples.go` and `validate.go` plus five new error types, for no change in what any current
run prints. The follow-up work plan should hold `errors.go` + `examples.go` + `validate.go` in a
single write set, which is the exact reason the item persisted across three tasks.

**Category 2 — the two `discovery.go` sites the root-resolution change added** (`resolveRoot`'s
`failed to resolve the source root %s: %v` at `:128`, and the empty-corpus guard at `:304`). These
are **not** part of the same finding and should not be pulled into the follow-up. They are *returned
fatal errors*, not collected diagnostics: they travel up through `run` to `report` (`main.go:77`),
which prints one line and exits 1. They sit in the same category as the pre-existing
`discovery.go:271` walk failure, the `main.go` flag and directory errors (`:147`, `:149`, `:151`,
`:272`, `:285`, `:288`), the `render.go` encode errors (`:56`, `:59`) and the `types.go` unmarshal
errors (`:55`, `:84`) — a category for which the package deliberately has no typed vocabulary,
because none of them is attributed to a corpus document or matched on. Introducing types for them
would add eleven error types to serve no `errors.Is` call and no ordering decision, which is the
entropy GEN-001 and GEN-002 exist to prevent. **No violation; nothing to carry forward here.**

## Manifest "Deliberate" Items — Confirmed, Not Re-raised

Each was re-checked on its merits against the current file contents.

- **GEN-003, no Dockerfile — CONFIRMED as a documented deviation.** GEN-003 is a SHOULD. The
  rationale is recorded in two places, both naming the rule: `indexer/Makefile:1-13` (header comment,
  stating the approved binary-only boundary, the bind-mount argument, and the condition under which
  to revisit) and `indexer/README.md`, "Build environment". `general/DOCKER.md` and
  `golang/DOCKER.md` correctly stay outside the applicable set.

- **`fenceTracker` and `ExampleDirectoryPolicy` outside `types.go` — CONFIRMED, agreed on the
  merits.** GO-010 governs data models — here `Config`, `Frontmatter`, `Document`, `ExampleEntry`,
  `Node`, `Tree`, `Alias`, `ScopeList`, `AliasList` and `field`, all present in `types.go`.
  `fenceTracker` (`discovery.go:64`) is a stateful scanner whose only member is the `fenced` method
  that advances it, and `ExampleDirectoryPolicy` (`discovery.go:187`) is an enumerated parameter of
  `WalkMarkdownFiles`. Neither carries corpus data. No violation.

- **The "Accepted standards deviations" table — CONFIRMED as user-adjudicated.** GO-018, GO-020,
  GO-022, GO-023, GO-033 and ACPT-002 / ACPT-004 are taken as accepted per the manifest and are not
  re-litigated.

## Adjudicated Deviations (not counted as violations)

Unchanged from revisions 1 and 2.

| Rule | Status |
|------|--------|
| GO-018, GO-020, GO-022, GO-023, GO-033 | Accepted per manifest |
| ACPT-002, ACPT-004 | Accepted per manifest |
| GO-012 | Not reported — mandates `go-playground/validator`, the dependency the GO-022 adjudication forbids. |
| GO-021, GO-024 | Not reported — both presuppose the YAML configuration files the GO-018 adjudication removes. |
| GO-019 | Satisfied vacuously — the tool holds no sensitive configuration. |
| ACPT-001, ACPT-003, ACPT-005 | Not reported — all presuppose the Gherkin/godog runner the ACPT-002 adjudication forbids. |
| LOG-001 through LOG-006 | Not reported — the GO-033 adjudication settles the logging mechanism as plain stderr lines per the CLI contract. |

## Rules Checked and Found Conforming

Re-verified against the post-round-2 tree.

| Rule | Evidence |
|------|----------|
| GO-001 | `indexer/go.mod` — `go 1.26.5`, above the 1.25 floor. |
| GO-002 | `indexer/go.mod` present; module `github.com/PSauerborn/standards/indexer`. |
| GO-003 | `gofmt -l .` over `indexer/` reports no files. |
| GO-004 | `golangci-lint run` reports **0 issues** on the current tree; `make lint` at `indexer/Makefile:27`. |
| GO-005 | Every top-level function across the 22 `.go` files carries a doc comment whose first word is the function name, methods included — re-verified mechanically, no exceptions. |
| GO-006 | All 22 `.go` filenames are lowercase/snake_case. |
| GO-007 | Pipeline stages are pure functions over values; `fenceTracker` remains the one stateful helper, confined to a single document scan. |
| GO-008 | Single binary, flat layout directly under `indexer/`. |
| GO-010 | All ten data models plus `field` in `types.go`; `render.go`, `compare.go` and `validate.go` declare no type. The three structs outside it are test-local (`containmentCorpus`, `cliTreeDocument`/`cliTreeNode`, `integrationResult`) and model test fixtures, not application data. |
| GO-011 | `Config`, `Frontmatter`, `Document`, `ExampleEntry`, `Node`, `Tree` carry the data through every stage. |
| GO-015 | No `panic(` anywhere in `indexer/*.go`; `os.Exit` confined to `main.go`. |
| GO-016 | Twelve typed errors behind `DiagnosticError` (`errors.go:36`) plus four sentinels. The five untyped collector sites are the adjudicated SHOULD deviation above. |
| GO-017 | **All four sentinels and all twelve typed errors in `errors.go`.** No error declaration outside it. |
| GO-025 | `go test -cover` reports **91.6%** of statements, above the 80% guideline. `main` is the only function at 0.0%; `execute` exists so everything below it is driven in process. |
| GO-026 | Every test file imports `testing`. |
| GO-027 | All ten source files paired with a `_test.go`. |
| GO-028 | Every exported pipeline entry point has a `TestFunctionName` test, `TestSentinelErrors` now included. Unexported helpers are exercised through their callers: `resolveRoot` at 100.0%, `containedFile` at 83.3%, with the escape, not-a-regular-file and not-found branches each pinned by a named `t.Run`. No function apart from `main` sits at 0.0%. |
| GO-029 | Paths grouped in `t.Run` blocks within one test function per subject; the round-2 additions each contribute their own named subtests. |
| GO-030 | No live service or database connection in any test; fixtures are on-disk corpora and `t.TempDir` trees. |
| GO-031 | Fixtures under `indexer/tests/data/.corpora/`; the symlink corpora are built at run time under `t.TempDir` with the reason recorded at `discovery_test.go:73-80`. |
| GO-032 | All assertions go through `testify`; no comparison assertion via `t.Errorf` / `t.Fatalf` anywhere. |
| GEN-001, GEN-002 | Round 2 removed structure rather than adding it: the sentinel block collapsed three declaration sites into one, and `resolveRoot` collapsed the root decision — previously implicit in two places — into a single named function that both the walk and the containment check call. |
| GEN-004 | `indexer/Makefile` provides `build`, `test`, `lint`, `fmt`, `fmt-check`. |
| GEN-006 | `indexer/integration_test.go` — 8 integration test functions. |

## Observations Outside the Changeset (not counted)

- **GEN-005, language-specific hooks.** Unchanged across all three revisions. The repository
  `.pre-commit-config.yaml` carries every hook GEN-005 lists as a minimum (`check-yaml`,
  `end-of-file-fixer`, `trailing-whitespace`, `check-added-large-files`, `check-case-conflict`,
  `check-json`, `detect-secrets` with a baseline) plus markdownlint, but no `gofmt` or
  `golangci-lint` hook, so the Go module this work plan introduced is not covered by pre-commit.
  The file is not in the manifest changeset. Worth raising with the user as a follow-up work plan,
  alongside the GO-016 cluster.
