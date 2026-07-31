# Task: QC Remediation Round 2 — GO-017 sentinel error placement

Work Plan ID: WP-001
Task ID: TASK-QC-REMEDIATION
Created Date: 2026-07-31
Description: Relocate the package's sentinel errors into `errors.go` so every custom error of the package is declared in one place, per GO-017.
Acceptance Criteria Covered: none (standards conformance only — no acceptance criterion changes)

> **Round 2.** The five findings of round 1 are closed and verified — see
> `docs/plans/quality/WP-001/WP-001-quality-report.md`, "Revision 1 Findings — Closure Verification".
> Nothing from round 1 needs redoing. The round-1 task file, with its investigation notes, is kept
> alongside this one as `TASK-QC-REMEDIATION-round-1.md`.

## Implementation Content

`golang/GENERAL.md` GO-017 requires custom errors to be defined in a dedicated `errors.go` file. The
security remediation added two sentinel errors in `discovery.go`:

- `indexer/discovery.go:115` — `var errEscapesRoot = errors.New("path escapes the source root")`
- `indexer/discovery.go:119` — `var errNotRegularFile = errors.New("path is not a regular file")`

Both are package vocabulary rather than file-local details: they are produced by `containedFile` in
`discovery.go` and matched with `errors.Is` from `examples.go:59` and `examples.go:62`. A reader who
goes to `errors.go` for the package's error vocabulary does not find them. This is the same finding
round 1 raised against `MissingHeadingStatementError`, which was accepted and relocated to
`errors.go`.

Move both declarations, with their doc comments unchanged, into `indexer/errors.go`. Place them above
the `DiagnosticError` interface declaration (`errors.go:13`), under a short comment saying that these
are the errors the package matches on internally, as distinct from the `DiagnosticError`
implementations below them that are rendered to the user.

**Also move the two pre-existing sentinels in the same commit**, so the package ends consistent
rather than split three ways:

- `indexer/frontmatter.go:15` — `errUnterminatedFrontmatter`
- `indexer/main.go:43` — `errArgumentsReported`

These two are not counted as violations (they sit in regions no remediation touched), but relocating
only the new pair would leave `errors.go` plus two other files holding sentinels — a worse end state
than either consistent option, and against GEN-001/GEN-002.

This is a pure relocation. No message text, no wrapping, no `errors.Is` call site and no behaviour
changes. Identifiers stay unexported and keep their names.

## Target Files

- [x] `indexer/errors.go` — receives the four sentinel declarations
- [x] `indexer/discovery.go` — `errEscapesRoot`, `errNotRegularFile` removed
- [x] `indexer/frontmatter.go` — `errUnterminatedFrontmatter` removed
- [x] `indexer/main.go` — `errArgumentsReported` removed
- [x] `indexer/errors_test.go` — sentinel wording test (see Red Phase)

## Investigation Targets

- `indexer/errors.go` (lines 1-30 — the file header, the import block and the `DiagnosticError`
  interface, which is where the sentinel block goes)
- `indexer/discovery.go` (lines 113-173 — the two sentinels and the `containedFile` doc comment,
  which names both by name and must keep reading correctly after the move)
- `indexer/examples.go` (lines 57-73 — the `switch` matching both sentinels with `errors.Is`)
- `indexer/frontmatter.go` (lines 13-20) and `indexer/main.go` (lines 40-50)
- Check the `import` block of each of the four source files after the move: `errors` may become an
  unused import in `frontmatter.go`, and must stay in `discovery.go`, `examples.go` and `main.go`,
  which all still call `errors.Is`. `make lint` fails on an unused import.

## Investigation Notes

Observations recorded before implementation (round 2, line numbers re-derived from the current tree
after the round-2 code-review remediation landed):

- `indexer/errors.go` — package `main`; imports `errors`, `fmt`, `sort`, `strings`. The
  `DiagnosticError` interface is declared at line 13 after a doc comment starting at line 10. No
  package-level `var` currently exists in the file, so the sentinel block is new. `errors` is already
  imported (used by `errors.As` in `documentPathOf`), so adding `errors.New` calls needs no import
  change here.
- `indexer/discovery.go` — `errEscapesRoot` (line 115) and `errNotRegularFile` (line 119) sit between
  `splitLines` and the new `resolveRoot` helper. `grep -n "errors\."` shows these are the *only* two
  uses of the `errors` package in the file, so `errors` becomes an unused import after the move and
  must be dropped, or `make lint` fails. The `containedFile` doc comment names both sentinels by
  name; the names do not change, so it keeps reading correctly.
- `indexer/examples.go` — `ResolveExamples` matches both sentinels with `errors.Is` at lines 59 and
  62, and `fs.ErrNotExist` at line 66. `errors` stays imported. No call site changes.
- `indexer/frontmatter.go` — `errUnterminatedFrontmatter` at line 15 is the only use of the `errors`
  package in the file, so `errors` must be dropped from its import block too.
- `indexer/main.go` — `errArgumentsReported` at line 43. `errors` stays imported: `errors.Is` at
  lines 61, 83, 117, 151 and `errors.New` at lines 134 and 137.
- Round-2 remediation check: the round-2 changes (`resolveRoot` in `discovery.go`, the empty-corpus
  guard in `DiscoverDocuments`) introduced no new `var err...` sentinel. They introduced two inline
  `fmt.Errorf` diagnostics (`failed to resolve the source root %s: %v`, `no markdown files found
  under %s: ...`). Those are untyped-diagnostic sites of the GO-016 carry-forward cluster, not
  sentinel declarations, and this task's Notes explicitly exclude that cluster.
- Baseline for the byte-identical proof: `./indexer --source .. --output <tmp>` over the repository
  root emits a tree with SHA-256 `093bf70f1aac3e66d2dbd07569427cf8e3cdf4d637b1124a01a4e6be5b0793eb`
  before any change.

## Remediation Context

- Source: `quality-controller`
- Finding: GO-017 (`golang/GENERAL.md` §4) — `indexer/discovery.go:115` and `:119`. Custom errors
  should be defined in the dedicated `errors.go` file.
- Evidence: `grep -n "^var err" indexer/*.go` reports four package-level sentinels, none of them in
  `errors.go`: `discovery.go:115`, `discovery.go:119`, `main.go:43`, `frontmatter.go:15`. Meanwhile
  all 13 `Error() string` methods, and the 12 `DiagnosticError` implementations behind them, are in
  `errors.go` — including `MissingHeadingStatementError`, relocated there by round 1 of this same
  remediation, for this same rule.
- Verification: `grep -n "^var err" indexer/*.go` returns matches only in `indexer/errors.go`, and
  `make test && make lint && make fmt-check` all pass from `indexer/`.

This remediation has no testable behaviour change: reproduce with the grep above, apply the move,
re-run the verification command until it passes. The Red Phase below is limited to the one assertion
that makes the relocation load-bearing.

## Implementation Steps

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations in Investigation Notes
- [x] Add to `indexer/errors_test.go` a `TestSentinelErrors` function with a doc comment whose first
      word is `TestSentinelErrors` (GO-005), holding one `t.Run` per sentinel (GO-029) that asserts
      the sentinel's message with `assert.EqualError` (GO-032). This pins the wording the relocation
      has to carry across, exactly as `TestTypedErrorMessages` pins the wording of the rendered
      diagnostics.
- [x] Run `go test ./...` and confirm the new test passes against the *current* placement — it is a
      wording pin, not a failing test. The move itself is verified by the grep in Verification.

### 2. Green Phase

- [x] Move the four `var` declarations and their doc comments into `indexer/errors.go`
- [x] Fix the import blocks of the four source files
- [x] Run `go test ./...` and confirm every test still passes

### 3. Refactor Phase

- [x] Run `make fmt` then `make lint` and confirm 0 issues
- [x] Confirm `git diff` / the working copy shows only moved lines plus import adjustments — no
      message text changed

## Completion Criteria

- [x] `grep -n "^var err" indexer/*.go` reports matches only in `indexer/errors.go` (lines 17, 21,
      25, 31)
- [x] Every sentinel's message text is byte-identical to its pre-move text — pinned by
      `TestSentinelErrors` with `assert.EqualError`, and the tree the indexer emits over the
      repository root is byte-identical before and after the move
      (SHA-256 `093bf70f1aac3e66d2dbd07569427cf8e3cdf4d637b1124a01a4e6be5b0793eb` both times)
- [x] All added tests pass
- [x] Verification command from Remediation Context passes: `make test`, `make lint` (0 issues) and
      `make fmt-check` all pass from `indexer/`, with no test assertion altered

## Notes

- Impact scope: the `indexer` package only. Every consumer is in the same package, so there are no
  import changes outside the four files and no call-site edits.
- Scope boundary: do **not** touch the standards corpus, `docs/standards-tree.yaml`, or anything under
  `indexer/tests/data/.corpora/`. Re-run the indexer against the repository root afterwards and
  confirm the emitted tree is byte-identical — that is the check that proves this was a pure
  relocation.
- Do **not** fold the GO-016 untyped-diagnostic cluster into this task. It is a confirmed,
  rationale-recorded carry-forward spanning `errors.go` + `examples.go` + `validate.go`. The quality
  report records that it covers five sites (`examples.go:63`, `:70`, `:77`, `validate.go:231`,
  `:237`) rather than the two the manifest names. It wants its own task.
