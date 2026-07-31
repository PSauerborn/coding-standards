# Task: QC Remediation — Coding Standards Violations (WP-001)

Work Plan ID: WP-001
Task ID: TASK-QC-REMEDIATION
Created Date: 2026-07-31
Description: Remediate the 13 coding-standards violations reported in the WP-001 quality report — GO-017, GO-028, GO-010, GO-005 (nine test files) and GEN-003.
Acceptance Criteria Covered: none (standards conformance only — no acceptance criterion changes)

## Implementation Content

Bring the `indexer/` changeset into conformance with `golang/GENERAL.md` and `GENERAL.md` without
changing any observable behaviour of the tool. The emitted tree, the diagnostic wording, the exit
codes and the fixture corpora must all be byte-for-byte unchanged when this task completes.

The five remediations, in the order they should be applied:

1. **GO-017 — relocate the misplaced typed error.** Move `MissingHeadingStatementError`, together
   with its `Error()` and `DocumentPath()` methods, from `indexer/validate.go:18-33` into
   `indexer/errors.go`. Place it next to the other example-file diagnostics (after
   `MissingStatementsLineError`, `errors.go:161`) so the file's grouping stays intact. Drop the
   file-local rationale paragraph from its doc comment — it no longer applies — but keep the
   per-field comments. The message string must not change: `UncitedExampleError` and
   `UnknownStatementError` are raised only by the same validation stage and already live in
   `errors.go`, so no rationale for the split survives the move.

2. **GO-028 — cover the relocated error.** Add `MissingHeadingStatementError` to the table in
   `indexer/errors_test.go:11 TestTypedErrorMessages` so its `Error()` and `DocumentPath()` are
   exercised, matching the shape of the eleven entries already there. Both methods currently report
   0.0% statement coverage. Additionally, extend the `"example heading identifier missing from its
   statements line"` subtest in `indexer/validate_test.go:122` with a `collector.Diagnostics()`
   assertion on the rendered line, mirroring the `"identifier only mentioned by the citing
   document is unknown"` subtest at `validate_test.go:99-111`, which does pin its wording.

3. **GO-010 — relocate the misplaced data model.** Move the `field` struct from
   `indexer/render.go:34-39` into `indexer/types.go`, preserving its doc comment and both field
   comments. It stays unexported. If moving it is judged to hurt the locality of the renderer more
   than it helps navigability, do **not** silently leave it in place: stop and raise it with the
   user, because GO-010 is a **MUST** and the meta rules require confirmation before a MUST is
   deviated from.

4. **GO-005 — document the test functions.** Add a doc comment whose first word is the function
   name to every test function listed below. Match the style already used in
   `indexer/compare_test.go:100`, `indexer/integration_test.go:464` and
   `indexer/failure_paths_test.go:176`, which state what the test covers rather than restating the
   signature. The affected functions:

   - `indexer/types_test.go` — `:22`, `:63`, `:102`, `:119`, `:173`, `:231`, `:273`
   - `indexer/errors_test.go` — `:11`, `:107`, `:138`, `:161`, `:209`, `:241`
   - `indexer/discovery_test.go` — `:59`, `:96`, `:202`, `:234`
   - `indexer/main_test.go` — `:122`, `:213`, `:340`, `:366`
   - `indexer/frontmatter_test.go` — `:42`, `:176`, `:242`
   - `indexer/examples_test.go` — `:52`, `:172`
   - `indexer/tree_test.go` — `:112`, `:148`
   - `indexer/validate_test.go` — `:60`, `:92`
   - `indexer/render_test.go` — `:147`

   Line numbers are as of this report and will shift once earlier edits land; re-derive them with
   the verification command rather than trusting the offsets.

5. **GEN-003 — containerized build and test.** `indexer/Makefile` runs `go build`, `go test`,
   `golangci-lint` and `gofmt` against the host toolchain only. GEN-003 is a **SHOULD**, so either
   route closes it:
   - add a `Dockerfile` under `indexer/` pinned to the module's Go version plus `docker-build` and
     `docker-test` Makefile targets that bind-mount the repository root (the tool indexes the
     repository it lives in, so the corpus has to be reachable from inside the container); **or**
   - record the deviation and its reason, per the `GENERAL.md` meta rules ("Deviate if the
     situation genuinely calls for it; briefly note the reason when you do"), in
     `indexer/README.md` and in this task's Investigation Notes.

   Pick one and say which. Do not leave the deviation undocumented.

## Target Files

- [x] `indexer/validate.go`
- [x] `indexer/errors.go`
- [x] `indexer/render.go`
- [x] `indexer/types.go`
- [x] `indexer/Makefile`
- [x] `indexer/README.md` — the GEN-003 documented-deviation route was chosen
- [x] `indexer/Dockerfile` — **not created**: the documented-deviation route was chosen, so this file is deliberately absent (see Investigation Notes, GEN-003)
- [x] `indexer/errors_test.go`
- [x] `indexer/validate_test.go`
- [x] `indexer/types_test.go`
- [x] `indexer/discovery_test.go`
- [x] `indexer/main_test.go`
- [x] `indexer/frontmatter_test.go`
- [x] `indexer/examples_test.go`
- [x] `indexer/tree_test.go`
- [x] `indexer/render_test.go`

## Investigation Targets

- `/Users/pascal/Github/psauerborn/standards/docs/plans/quality/WP-001/WP-001-quality-report.md` — the findings this task remediates, and the "Adjudicated Deviations" table listing the rules that must NOT be changed
- `/Users/pascal/Github/psauerborn/standards/docs/plans/manifests/WP-001-manifest.md` — accepted deviations and the RISK-001 fixture placement constraint
- `/Users/Pascal/.stdidx/golang/GENERAL.md` — GO-005, GO-010, GO-017, GO-028 as written
- `/Users/Pascal/.stdidx/GENERAL.md` — GEN-003 and the meta rules on deviating from a SHOULD
- `indexer/errors.go` (the `MissingStatementsLineError` block, `:148-161`) — insertion point and grouping convention for remediation 1
- `indexer/errors_test.go` (`TestTypedErrorMessages` table, `:11-105`) — the entry shape for remediation 2
- `indexer/compare_test.go` (`:100`), `indexer/integration_test.go` (`:464`) — the doc-comment style remediation 4 must match

## Change Category

`Change Category: boundary-change`

The GO-017 and GO-010 moves relocate declarations across file boundaries within `package main`.
Sweep for any other declaration in the changeset sitting in the wrong file for the same reason: the
review found none beyond these two, but the sweep must confirm that after the moves land.

## Investigation Notes

**Line numbers in the report are stale.** Two remediations (code-review, security) landed on the same
files before this one. Every location below was re-derived from the current code.

### Standards as written

- GO-005 (**MUST**): every function needs a doc comment whose first word is the function name.
- GO-010 (**MUST**): data models must live in `types.go`.
- GO-017 (**SHOULD**): custom errors should live in `errors.go`.
- GO-028 (**SHOULD**): each function should have a corresponding unittest.
- GEN-003 (**SHOULD**): code should be built, tested and run in a container. `GENERAL.md` meta rules:
  a SHOULD may be deviated from when the situation calls for it, *provided the reason is noted*.

### Boundary sweep — every `type` declaration in `indexer/*.go` (non-test)

| Declaration | File | Verdict |
| --- | --- | --- |
| `DiagnosticError` + the 12 `*Error` types, `AggregateError`, `ErrorCollector` | `errors.go` | Correct (GO-017) |
| `MissingHeadingStatementError` | `validate.go` → `errors.go` | **Moved** (GO-017) |
| `Config`, `ScopeList`, `Alias`, `AliasList`, `Frontmatter`, `Document`, `ExampleEntry`, `Node`, `Tree` | `types.go` | Correct (GO-010) |
| `field` | `render.go` → `types.go` | **Moved** (GO-010) |
| `fenceTracker` | `discovery.go` | **Left in place.** Not a data model: it is a line-scanner state machine with behaviour (`fenced(line) bool`) whose zero value is meaningful, closer to `bytes.Buffer` than to a DTO. It carries no corpus data, is unexported, and is consumed only by the two parsers in `discovery.go` and `examples.go`. GO-010's subject is data models; GO-011's rationale (grouping related data for navigability) does not bite. |
| `ExampleDirectoryPolicy` (+ its `Skip`/`Include` constants) | `discovery.go` | **Left in place.** An enumerated parameter of `WalkMarkdownFiles`, declared with the constants that name its values and the function that consumes it. Moving the type to `types.go` would split an enum from its constants across files for no gain. |

No other declaration sits in the wrong file after the two moves.

### GO-028 — the coverage gap

`go tool cover -func` on the pre-change tree reports `validate.go:27 Error 0.0%`; `DocumentPath` is a
one-line method the profile folds in and likewise never called. `validate_test.go`'s heading subtest
matches with `errors.As` and asserts `.Path`/`.Statement` but never calls either method, so the
rendered wording of this diagnostic is the only one of the twelve that nothing pins.

**Pre-change coverage baseline: 91.5% of statements** (`go test -coverprofile`). Note this is below
the 91.8% figure the report recorded — the drift predates this task and comes from the two earlier
remediations adding production code. 91.5% is the baseline this task must not lower.

### GO-005 — the actual gap, re-derived

36 test functions (not the 31 the report recorded) have no doc comment, across the same nine files.
The five extra are functions the two earlier remediations added:
`TestFenceTracker`, `TestSplitLines`, `TestExtractStatements` (`discovery_test.go`),
`TestParseExampleFile` (`examples_test.go`), `TestWriteAndClose` and `TestExecute` (`main_test.go`),
`TestDocument` (`types_test.go`), `TestValidateCorpus` (`validate_test.go`). Every non-test function
in the package, and every helper in the test files, already conforms.

### GEN-003 — decision: documented deviation, no Dockerfile

The user's approved boundary for this work is "binary only: ship `indexer/` plus its own Makefile",
and no other module in this repository ships a Dockerfile. Adding one would exceed that boundary and
introduce `general/DOCKER.md` + `golang/DOCKER.md` as newly applicable standards for a single-binary
developer tool that is built and run from the checkout it indexes. GEN-003 is a **SHOULD**, so the
meta rules permit the deviation as long as the reason is recorded: it is recorded in the `Makefile`
header comment and in `indexer/README.md` ("Build environment"), not left silent.

### Left deliberately: the untyped non-regular-file diagnostic

`examples.go:63` raises `fmt.Errorf("example %s cited by %s is not a regular file", ...)` — the one
ad-hoc untyped diagnostic in the package, added by the security remediation because `errors.go` was
outside its write set. Giving it a typed error would mean editing `examples.go`, which is **not in
this task's Target Files**, so it is out of scope here and is flagged for a follow-up rather than
done silently. (`validate.go:254,260` raise a second untyped `failed to read example %s` diagnostic,
in a file that *is* in scope, but typing one of the pair while the other stays untyped would leave
the inconsistency it is meant to remove; both belong in one follow-up.)

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-001 … TASK-012 | WP-001 execution tasks | blocks | The complete `indexer/` changeset this task remediates |

## Remediation Context

- Source: quality-controller
- Finding / failing command: 13 violations in `docs/plans/quality/WP-001/WP-001-quality-report.md` — GO-017 (`indexer/validate.go:18`), GO-028 (`indexer/validate.go:27` and `:33`), GO-010 (`indexer/render.go:34`), GO-005 (nine test files), GEN-003 (`indexer/Makefile:6`)
- Evidence:
  - `indexer/validate.go:18` declares `type MissingHeadingStatementError struct` — a `DiagnosticError` implementation outside `errors.go`, where the other eleven live.
  - `go tool cover -func` reports `validate.go:27 Error 0.0%` and `validate.go:33 DocumentPath 0.0%`; `MissingHeadingStatementError` appears in exactly one test (`validate_test.go:126`), which matches it with `errors.As` and asserts its fields but never calls either method.
  - `indexer/render.go:34` declares `type field struct` — a data model outside `types.go`.
  - 31 test functions across nine test files carry no preceding `//` doc comment, while `compare_test.go`, `failure_paths_test.go` and `integration_test.go` document all of theirs.
  - `indexer/Makefile:6,11,16,21` invoke `go build`, `go test`, `golangci-lint` and `gofmt` on the host; the repository contains no Dockerfile.
- Verification: from `indexer/`, all of the following must hold —
  1. `make fmt-check` clean, `make lint` reports 0 issues, `make test` passes
  2. `go test -count=2 -shuffle=on ./...` passes
  3. `grep -n "MissingHeadingStatementError" validate.go` returns nothing, and `errors.go` declares it
  4. `grep -n "^type field struct" render.go` returns nothing, and `types.go` declares it
  5. `go test -coverprofile=cov.out ./... && go tool cover -func=cov.out` shows `MissingHeadingStatementError`'s `Error` and `DocumentPath` at 100.0%, and the package total no lower than the 91.8% baseline
  6. every `^func Test` line in `indexer/*_test.go` is immediately preceded by a `//` comment line whose first word after `// ` is that function's name

## Implementation Steps (TDD: Red-Green-Refactor)

Remediations 1, 3, 4 and 5 are declaration moves, comments and tooling — no behaviour changes, so
for those follow the non-TDD route the template allows: apply the change, then re-run the
Verification commands until they pass. Remediation 2 is a genuine test addition and follows the
cycle below.

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations in Investigation Notes
- [x] Sweep (boundary-change): list every `type` declaration in `indexer/*.go` and confirm each sits in the file its rule assigns — typed errors in `errors.go` (GO-017), data models in `types.go` (GO-010). Record the sweep result, including any declaration deliberately left in place and why
- [x] Capture the pre-change coverage baseline: `go test -coverprofile=cov.out ./... && go tool cover -func=cov.out | grep -E "MissingHeadingStatement|total"`
- [x] Add the `MissingHeadingStatementError` entry to `TestTypedErrorMessages` and the `Diagnostics()` wording assertion to the `validate_test.go` heading subtest, asserting the exact current message text
- [x] Run the tests and confirm the new assertions fail only where the wording assertion is genuinely absent — if they pass immediately without any production change, the assertions are not pinning anything; strengthen them

### 2. Green Phase

- [x] Apply remediation 1 (move `MissingHeadingStatementError` into `errors.go`)
- [x] Apply remediation 3 (move `field` into `types.go`)
- [x] Run only the added and adjacent tests and confirm they pass

### 3. Refactor Phase

- [x] Apply remediation 4 (doc comments across the nine test files)
- [x] Apply remediation 5 (GEN-003: containerized targets, or the documented deviation)
- [x] Re-run `make fmt-check`, `make lint`, `make test` and `go test -count=2 -shuffle=on ./...`
- [x] Confirm the emitted tree is unchanged: run the built binary over the repository root before and after and diff the two outputs byte-for-byte

## Completion Criteria

- [x] `MissingHeadingStatementError` is declared in `indexer/errors.go` and nowhere else (GO-017)
- [x] `MissingHeadingStatementError.Error` and `.DocumentPath` report 100.0% statement coverage (GO-028)
- [x] `field` is declared in `indexer/types.go` and nowhere else, or the MUST deviation has been raised with the user and their decision is recorded in Investigation Notes (GO-010)
- [x] Every `Test*` function in `indexer/*_test.go` has a doc comment whose first word is the function name (GO-005)
- [x] `indexer/` provides a containerized build/test path, or the GEN-003 deviation is documented with its reason in `indexer/README.md` (GEN-003)
- [x] All added tests pass
- [x] Verification commands 1–6 from Remediation Context pass

## Notes

- Impact scope: `indexer/` only. Declarations move between files inside `package main`, so no import
  or call site changes. No diagnostic message, exit code or emitted-tree byte may change.
- Scope boundary — preserve unchanged:
  - `indexer/tests/data/.corpora/**` — the fixture corpora, including the deliberately broken
    `failure/` corpora and their README. RISK-001 requires them to stay under the dot-prefixed
    `.corpora/` directory; repairing or relocating one silently disables the failure-path tests.
  - The standards corpus at the repository root (`GENERAL.md`, `golang/`, `python/`, `general/`,
    `databases/`, `javascript/`, `terraform/`, `testing/`) — read-only input to the tool.
  - `docs/standards-tree.yaml` — read-only input; never written by the tool.
  - `indexer/go.mod`, `indexer/go.sum`, `indexer/.golangci.yml` — the toolchain settlement recorded
    in the manifest's Toolchain notes. In particular the `depguard` `archived-yaml` deny rule must
    keep firing.
  - The rules in the quality report's "Adjudicated Deviations" table — GO-012, GO-018, GO-019,
    GO-020, GO-021, GO-022, GO-023, GO-024, GO-033, ACPT-001 through ACPT-005 and LOG-001 through
    LOG-006 were adjudicated by the user or follow from that adjudication. Do not "fix" them; in
    particular, do not add `go-playground/validator`, `spf13/viper`, `sirupsen/logrus` or
    `cucumber/godog`.
