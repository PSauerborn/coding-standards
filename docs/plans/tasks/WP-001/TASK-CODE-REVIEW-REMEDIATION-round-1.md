# Task: Code Review Remediation — Standards Indexer

Work Plan ID: WP-001
Task ID: TASK-CODE-REVIEW-REMEDIATION
Created Date: 2026-07-31
Description: Fix six peer-review findings in `indexer/` — a false "uncited example" diagnostic, a fence-tracking hole that lets a code sample satisfy REQ-9, a CRLF-corrupted example title, silently discarded citation diagnostics for detached documents, silent truncation of a wrapped `Statements:` line, and the missing pipeline-level fixture that hid the first finding.
Acceptance Criteria Covered: AC-4, AC-7, AC-8, AC-9

## Implementation Content

Every finding below was reproduced against the built binary. Each entry gives the
file and line it anchors to, the concrete input that triggers it, and the fix
required. Findings 1, 2, 3 and 5 all produce a *plausible but wrong* result
rather than a crash, which is the failure mode this changeset was written to
avoid.

### Finding 1 (correctness, medium) — a cited example that fails to parse is additionally reported as "not cited by any standards document"

- File: `indexer/validate.go`, line 153 (`reportUncitedExamples`, condition `len(citedBy[path]) == 0`)
- Failure scenario: a corpus whose `GENERAL.md` declares `examples: [examples/GENERAL/e.md]`, where `e.md` has no `# [<ID>] <Title>` heading. `ResolveExamples` (`indexer/examples.go:58-65`) drops the entry and calls `MarkFailed`, so `citations` carries no entry for the file and `citedBy` has no key for it. `reportUncitedExamples` then reads the absence as "nobody cites it". Reproduced output:

  ```text
  examples/GENERAL/e.md: example file has no title heading
  examples/GENERAL/e.md: example file is not cited by any standards document
  exit=1
  ```

  The second line is false — the file *is* cited — and it sends the author to add
  a citation that already exists. It also contradicts the manifest's RISK-006
  claim that "one corrupted document yields exactly one diagnostic".
- Required fix: derive the "is cited" set from the *declared* citations rather
  than from the successfully resolved entries, or suppress `UncitedExampleError`
  for any path for which `collector.IsFailed(path)` is true. `citingDocuments`
  (`indexer/validate.go:105`) is the natural place to record declared-but-unresolved
  citations.

### Finding 2 (correctness, medium) — nested code fences invert fence tracking, so an identifier that appears only inside a code sample counts as a defined statement

- File: `indexer/discovery.go`, line 180 (`ExtractStatements` fence toggle); same defect at `indexer/examples.go`, line 154 (`parseExampleStatements`); shared cause is `codeFenceRegexp` at `indexer/discovery.go:49`
- Failure scenario: `codeFenceRegexp` matches any line opening with ``` or `~~~` and the toggle is a naive parity flip, so it does not implement the CommonMark rule that a closing fence must use the same character and be at least as long as the opener. A standards document containing a four-backtick block that wraps a three-backtick block:

  ````text
  ````markdown
  ```go
  `[XX-003]` **MUST**: sample only.
  ```
  ````
  ````

  flips `fenced` four times, leaving the sample line classified as *prose*.
  `XX-003` is therefore collected as a defined statement, an example citing it
  on its `Statements:` line passes REQ-9, and the run exits 0 emitting an index
  that advertises coverage which does not exist. Verified: with the nested fence
  the run exits 0; with a single fence the same corpus correctly reports
  `examples/ok.md: unknown statement ID "XX-003" cited by GENERAL.md`. This is
  the exact defeat of the mitigation the manifest records under RISK-005.
- Required fix: track the opening fence's character and run length and treat a
  line as a closing fence only when it uses the same character and is at least as
  long; ignore fence-looking lines while already inside a fence opened by a longer
  or different delimiter. Apply the same helper in both `ExtractStatements` and
  `parseExampleStatements`.

### Finding 3 (edge-case, medium) — a CRLF example file emits a title with a trailing carriage return

- File: `indexer/examples.go`, line 22 (`exampleHeadingRegexp`)
- Failure scenario: an example file written with CRLF line endings, e.g.
  `# [XX-001] My Title\r\n`. The content is split on `"\n"`, so the line still
  ends in `\r`; `[ \t]*$` does not absorb it and `(\S.*?)` does. The run exits 0
  and emits `title: "My Title\r"` — a quoted control character in the index that
  every consumer then carries. `splitFrontmatter` (`indexer/frontmatter.go:120`)
  deliberately tolerates CRLF, so example parsing is inconsistent with the
  codebase's own convention. Verified against the built binary.
- Required fix: strip a trailing `\r` when splitting lines in `parseExampleTitle`,
  `parseExampleStatements`, `exampleHeadingStatement` (`indexer/validate.go:243`)
  and `ExtractStatements`, or add `\r?` before `$` in `exampleHeadingRegexp`.
  Prefer normalising line endings once at the split site so every reader benefits.

### Finding 4 (correctness, low) — citation diagnostics of a detached document are discarded entirely

- File: `indexer/main.go`, line 217 (`resolveCitations`, `ResolveExamples(source, document, discarded)`)
- Failure scenario: `ORPHAN.md` declares `parent: nowhere/MISSING.md` and
  `examples: [examples/absent.md]`, where `absent.md` does not exist.
  `BuildTree` never places `ORPHAN.md`, so it never resolves its citations into
  the real collector; `resolveCitations` resolves them into a throwaway collector.
  The run reports only `ORPHAN.md: unknown parent "nowhere/MISSING.md"` and never
  reports the missing example file. REQ-7 requires a diagnostic for an `examples:`
  entry whose file is missing, and SPEC-001 §7 ("Partial failure") requires all
  validation errors in a run to be collected and reported together. Combined with
  Finding 1, the same corpus with a *malformed* rather than missing example file
  reports `examples/broken.md: example file is not cited by any standards
  document` and never reports the real fault. Both reproduced.
- Required fix: resolve every discovered document's citations exactly once into
  the real collector before tree assembly, and have `BuildTree` consume the
  already-resolved entries instead of calling `ResolveExamples` again. That
  removes the double-resolution the throwaway collector exists to paper over and
  makes detached documents report their citation faults like any other document.

### Finding 5 (edge-case, low) — a wrapped `Statements:` line is silently truncated

- File: `indexer/examples.go`, line 170 (`parseExampleStatements` returns at the first matching line)
- Failure scenario: an example file whose statements are wrapped across lines:

  ```text
  Statements: `[XX-001]`
    `[XX-002]`
  ```

  Only `XX-001` is emitted; `XX-002` is dropped, the run exits 0, and the index
  under-reports the example's coverage so agents never select it for `XX-002`.
  Nothing reports the discrepancy. Reproduced against the built binary.
- Required fix: either consume continuation lines until a blank line, or reject an
  example file whose `Statements:` block continues past the first line with a
  diagnostic naming the file. Silent truncation is the one outcome to remove.

### Finding 6 (test-coverage, medium) — no pipeline-level fixture covers a cited-but-malformed example file

- File: `indexer/failure_paths_test.go`, line 260 (`TestFailurePathFixtureCorpora` case table)
- Failure scenario: REQ-7's clause "an example file lacking the `# [<ID>] <Title>`
  heading or `Statements:` line" and SPEC-001 §7's "duplicate IDs on a
  `Statements:` line" are exercised only by `TestParseExampleFile` and
  `TestResolveExamples` in `indexer/examples_test.go`, which call
  `ParseExampleFile`/`ResolveExamples` directly and never reach `ValidateCorpus`.
  `.corpora/failure/` holds no corpus for them. That is precisely why Finding 1's
  spurious second diagnostic went unnoticed: no test asserts the *diagnostic
  count* for a corpus whose cited example is malformed.
- Required fix: add `.corpora/failure/malformed-example` (heading missing) and
  `.corpora/failure/duplicate-statement` corpora, and add their cases to the
  `TestFailurePathFixtureCorpora` table asserting the exact diagnostic count and
  wording — which must be 1 once Finding 1 is fixed.

## Target Files

- [x] `indexer/validate.go`
- [x] `indexer/discovery.go`
- [x] `indexer/examples.go`
- [x] `indexer/main.go`
- [x] `indexer/validate_test.go`
- [x] `indexer/discovery_test.go`
- [x] `indexer/examples_test.go`
- [x] `indexer/main_test.go`
- [x] `indexer/failure_paths_test.go`
- [x] `indexer/tests/data/.corpora/failure/malformed-example/` (new fixture corpus)
- [x] `indexer/tests/data/.corpora/failure/duplicate-statement/` (new fixture corpus)
- [x] `indexer/tests/data/.corpora/discovery/` (new nested-fence fixture, `discovery/fences/nested.md`)
- [x] `indexer/tests/data/.corpora/examples/` (new wrapped-statements fixture; the CRLF
  content is derived from the LF fixture in the test, see Investigation Notes)
- [x] `indexer/tests/data/.corpora/failure/nested-fence/` (new fixture corpus, added:
  Finding 2's completion criterion is a pipeline outcome, "an example citing it fails
  the run", which no unit-level fixture can assert)
- [x] `indexer/tests/data/.corpora/failure/detached-citation/` (new fixture corpus, added:
  no existing fixture has a detached document that cites anything, so Finding 4 was
  unobservable without one)

## Investigation Targets

- `indexer/validate.go` (`ValidateCorpus`, `citingDocuments`, `reportUncitedExamples`, `validateStatementCoverage`) — the citation bookkeeping Finding 1 turns on
- `indexer/main.go` (`run`, `resolveCitations`) — the stage ordering Finding 4 turns on; the doc comment on `run` records why assembly precedes validation
- `indexer/tree.go` (`BuildTree`, `emitExamples`) — the other half of the double resolution
- `indexer/discovery.go` (`codeFenceRegexp`, `ExtractStatements`) and `indexer/examples.go` (`parseExampleStatements`) — the shared fence toggle
- `indexer/examples.go` (`exampleHeadingRegexp`, `parseExampleTitle`) and `indexer/frontmatter.go` (`isDelimiterLine`) — the CRLF convention the example parser departs from
- `indexer/tests/data/.corpora/failure/README.md` — the placement constraint every new fixture must respect (fixtures stay under a dot-prefixed directory)

## Change Category

`Change Category: bug-fix, boundary-change`

The sweep for Findings 2 and 3 is the whole class of line-oriented readers in the
package: every `strings.Split(string(content), "\n")` loop in `discovery.go`,
`examples.go` and `validate.go` shares the same fence and CRLF handling and must
be fixed together, not one call site at a time.

## Investigation Notes

- Line-oriented readers of the package are exactly four: `ExtractStatements`
  (`discovery.go:179`), `parseExampleTitle` (`examples.go:135`),
  `parseExampleStatements` (`examples.go:153`) and `exampleHeadingStatement`
  (`validate.go:243`). All four split on `"\n"` and none strips the `\r` of a
  CRLF ending; two of them additionally flip `fenced` on `codeFenceRegexp`.
  `frontmatter.go:120` is the convention they depart from (`isDelimiterLine`
  trims the `\r`). One `splitLines` helper and one `fenceTracker` therefore fix
  every site at once.
- Finding 1: `ResolveExamples` calls `collector.MarkFailed(relative)` for an
  example file that fails to parse, so `reportUncitedExamples` can suppress
  `UncitedExampleError` on `collector.IsFailed(path)` without any new plumbing
  through `ValidateCorpus`'s signature. That is the smaller of the two fixes the
  finding offers and it needs no change to the citations map.
- Finding 4: the restructure the finding proposes (`BuildTree` consuming
  already-resolved entries) requires changing `BuildTree`'s signature and
  `emitExamples` in `indexer/tree.go`, which is an Investigation Target and NOT
  a Target File of this task. The narrower fix stays inside `main.go`: assembly
  places a subset of the discovered documents, so `resolveCitations` can resolve
  the citations of the documents assembly placed into a throwaway collector (as
  today, they were already collected) and those of the documents it detached
  into the real collector. Diagnostics and the `MarkFailed` marks of a detached
  document then land where every other document's do, and the double resolution
  keeps the load-bearing stage order (`BuildTree` before `ValidateCorpus`) and
  the cascade suppression exactly as they are.
- No detached fixture document (`cli/failures/golang/GENERAL.md`,
  `cli/cascade/golang/GENERAL.md`, `validation/cascade/child.md`,
  `failure/cycle/*`, `failure/unknown-parent/orphan.md`) cites any example, so
  the Finding 4 fix moves no diagnostic in an existing fixture corpus. A fixture
  that does cite one is needed to observe it.
- Finding 5: no example file of the corpus, and none of the fixture corpora, has
  a non-blank line directly below its `Statements:` line, so reading the whole
  paragraph is safe. It is also the correct reading: markdown renders
  consecutive non-blank lines as one paragraph, so the wrapped continuation is
  part of the same rendered `Statements:` line. The alternative (rejecting a
  wrapped line) would need a new diagnostic type and would fail a file that
  merely wraps.
- Fixture placement: `mixed-line-ending --fix=lf` in `.pre-commit-config.yaml`
  rewrites any committed CRLF file to LF, which would silently make a stored
  CRLF fixture prove nothing. The CRLF cases therefore derive their content from
  the LF fixture inside the test rather than storing a CRLF file on disk.

## Task Dependencies

(None — this task remediates the completed WP-001 changeset.)

## Remediation Context

- Source: code-reviewer
- Finding / failing command: six findings above, each reproduced against a binary
  built with `go build -o /tmp/idx .` from `indexer/` and run over a scratch corpus.
  Finding 1: cited example with no heading yields two diagnostics, the second
  false. Finding 2: nested-fence corpus exits 0 where the single-fence control
  exits 1. Finding 3: CRLF example emits `title: "My Title\r"`. Finding 4:
  detached document's missing-example citation is never reported. Finding 5:
  wrapped `Statements:` line emits `statements: [XX-001]` only.
- Evidence:

  ```text
  # Finding 1
  examples/GENERAL/e.md: example file has no title heading
  examples/GENERAL/e.md: example file is not cited by any standards document
  exit=1

  # Finding 2 (nested fence)   -> exit=0   (XX-003 accepted as defined)
  # Finding 2 (single fence)   -> exit=1   examples/ok.md: unknown statement ID "XX-003" cited by GENERAL.md

  # Finding 3
  title: "My Title\r"

  # Finding 4
  ORPHAN.md: unknown parent "nowhere/MISSING.md"
  (no diagnostic for the cited-but-absent examples/absent.md)

  # Finding 5
  statements: [XX-001]        (XX-002 declared on the continuation line, silently dropped)
  ```

- Verification: from `indexer/`, `make test`, `make lint` and
  `go test -count=2 -shuffle=on ./...` all pass, and the new fixture cases in
  `TestFailurePathFixtureCorpora` assert exactly one diagnostic for a
  cited-but-malformed example. The tree generated over this repository must remain
  unchanged: `TestCompareGeneratedTreeAgainstReference` and
  `TestIntegrationDeterminism` must still pass, and
  `docs/standards-tree.yaml` must not be edited.

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Sweep every line-oriented reader in `discovery.go`, `examples.go` and `validate.go` for the shared fence-parity and CRLF defects (Findings 2 and 3) and fold each site into the failing tests
- [x] Add the fixture corpora listed under Target Files, all under `indexer/tests/data/.corpora/`
- [x] Write failing tests: one diagnostic (not two) for a cited-but-malformed example; a nested-fence document that must NOT define its code-sample identifier; a CRLF example whose title carries no `\r`; a detached document whose missing cited example is reported; a wrapped `Statements:` line that is either fully read or reported
- [x] Run tests and confirm each fails for the stated reason

### 2. Green Phase

- [x] Apply the fixes described per finding
- [x] Run only the added tests and confirm they pass

### 3. Refactor Phase

- [x] Extract the fence tracking and the line splitting into single shared helpers so the two call sites cannot drift again
- [x] Confirm added tests and the full suite still pass

## Completion Criteria

- [x] A corpus whose cited example file has no heading reports exactly one diagnostic, naming the heading
- [x] An identifier appearing only inside a nested fenced code block is not a defined statement, and an example citing it fails the run
- [x] An example file with CRLF line endings emits a title with no trailing carriage return
- [x] A document with an unknown parent still reports the citation faults of the examples it declares
- [x] A wrapped `Statements:` line is either read in full or reported as a fault; it is never silently truncated
- [x] `TestFailurePathFixtureCorpora` covers the malformed-example and duplicate-statement corpora
- [x] All added tests pass
- [x] Verification commands from Remediation Context pass

## Deviations

- Finding 4 was fixed narrowly, inside `main.go` alone. The restructure the
  finding describes (`BuildTree` consuming already-resolved entries) requires
  changing `BuildTree` and `emitExamples` in `indexer/tree.go`, which is an
  Investigation Target of this task and not a Target File, so it is out of this
  task's write set. `resolveCitations` now resolves the citations of the
  documents assembly placed into a throwaway collector, exactly as before, and
  those of the documents it detached into the run's own collector. The defect is
  closed -- a detached document reports its citation faults, marks its broken
  example files failed, and both are asserted at pipeline level -- while the
  load-bearing stage order and the cascade suppression are untouched. The double
  resolution the finding wanted removed remains.

## Notes

- Impact scope: `ValidateCorpus` diagnostic set, statement extraction, example
  parsing and the pipeline stage order in `run`. Finding 4's fix moves citation
  resolution ahead of `BuildTree`, which changes which collector receives example
  diagnostics; the cascade-suppression assertions in `main_test.go`
  ("corrupted document yields exactly one diagnostic") and
  `validate_test.go` ("cascade") pin the behaviour that must survive.
- Scope boundary: `docs/standards-tree.yaml` (read-only reference, never written
  by the tool); the standards corpus itself (`*.md` outside `indexer/`) must stay
  byte-identical — the generated tree over this repository must not change. New
  fixtures must stay under the dot-prefixed `indexer/tests/data/.corpora/`, or the
  indexer will report its own repository as broken.
