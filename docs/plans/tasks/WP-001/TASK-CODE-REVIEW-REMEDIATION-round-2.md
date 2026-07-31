# Task: Code Review Remediation (Round 2) — Standards Indexer

Work Plan ID: WP-001
Task ID: TASK-CODE-REVIEW-REMEDIATION
Created Date: 2026-07-31
Description: Fix two peer-review findings surviving remediation round 1 in `indexer/` — a `--source` naming a symbolic link to the corpus directory silently produces an empty tree and exit 0, and a single malformed example file cited by two standards is reported twice.
Acceptance Criteria Covered: AC-1, AC-7

Round 1's task file is preserved alongside this one as
`TASK-CODE-REVIEW-REMEDIATION-round-1.md`.

## Implementation Content

Round 1's six findings were re-verified against the current sources and the
built binary and are all genuinely closed (evidence in Investigation Notes).
The two findings below are additional defects, each reproduced against the
built binary. Each entry gives the file and line it anchors to, the concrete
input that triggers it, and the fix required.

### Finding 1 — a symlinked `--source` silently yields an empty tree and exit 0

- **File / line**: `indexer/discovery.go:211` (`WalkMarkdownFiles`, the
  `filepath.WalkDir(root, ...)` call and its root handling at lines 215-217).
- **Failure scenario**: `WalkDir` lstats its root. When the final component of
  `--source` is a symbolic link to the corpus directory, the root is handed to
  the callback as a **non-directory** entry, so the `if path == root` branch is
  never reached, the extension test drops it, and the walk descends nowhere.
  `WalkMarkdownFiles` returns zero paths **and a nil error**, so discovery finds
  no documents, `ValidateCorpus` finds no corpus to validate, the collector
  stays empty, and the run writes `nodes: []` and exits 0. Reproduced:

  ```text
  $ ln -s $PWD/corpus $PWD/link
  $ indexer --source $PWD/link  --output tree.yaml ; echo $?   # nodes: []   0
  $ indexer --source $PWD/link/ --output tree.yaml ; echo $?   # 1 node      0
  ```

  A trailing slash on the same path indexes the corpus correctly, so the output
  of the tool depends on how the operator spelled the path. This is the
  deployment shape the README describes at line 21 and lines 85-94 (a synced or
  deployed copy of the corpus addressed through `--prefix`, living under a
  dot-prefixed directory such as `.claude/standards`), which is exactly where a
  symbolic link is likely. `validateConfig` does not catch it: `os.Stat` follows
  the link, so `requireDirectory("--source", ...)` passes.

  The consequence is silent: consumers of the tree (per `CLAUDE.md`) read an
  index that declares the corpus empty, and nothing in the run says so. This
  also contradicts AC-1, which requires 18 nodes for this repository.
- **Required fix**: resolve the source root before it is walked, so the walk
  descends into the directory the link names. The security remediation already
  put `filepath.EvalSymlinks(root)` in `containedFile`
  (`indexer/discovery.go:141`); make `WalkMarkdownFiles` walk that same resolved
  root, so both halves of the corpus decision agree on what the root is, and so
  the paths it returns stay relative to the same root `containedFile` re-derives.
  Keep returning the paths in root-relative slash form: nothing the tree emits
  may change. If the root cannot be resolved, return the error rather than an
  empty result — an unwalkable corpus must not read as an empty one.
- **Tests required**: a `WalkMarkdownFiles` case over a corpus reached through a
  symlinked root asserting the same paths as the direct root; and a pipeline
  case (`run`) over a symlinked source asserting the same node count and the
  same emitted bytes as the unsymlinked run. Both must assert a non-empty
  result so they cannot pass vacuously.

### Finding 2 — one malformed example cited by two standards is reported twice

- **File / line**: `indexer/examples.go:83` (the `if len(diagnostics) > 0`
  block of `ResolveExamples`).
- **Failure scenario**: spec §7 explicitly permits an example file cited by two
  standards. `ResolveExamples` runs once per citing document and re-parses the
  example file each time, so every diagnostic `ParseExampleFile` attributes to
  the **example file** — `MissingExampleHeadingError`,
  `MissingStatementsLineError`, `DuplicateStatementError` — is added once per
  citing document. `ErrorCollector.Add` (`indexer/errors.go:288`) does not
  deduplicate. Reproduced with a corpus whose `GENERAL.md` and `SECOND.md` both
  cite one heading-less `examples/broken.md`:

  ```text
  examples/broken.md: example file has no title heading
  examples/broken.md: example file has no title heading
  ```

  The same doubling occurs when one citer is placed and the other is detached:
  assembly reports it for the placed document and `resolveCitations`
  (`indexer/main.go:234`) reports it again for the detached one through the run's
  real collector, so remediation round 1's finding-4 fix widened the reach of
  this defect.

  A second diagnostic derived from the first is precisely what
  `TestFailurePathFixtureCorpora` (`indexer/failure_paths_test.go:256-260`)
  declares unacceptable — "a false repair instruction" — and what the
  `malformed-example` corpus asserts against for the single-citer case. The
  two-citer case is simply not covered.
- **Required fix**: suppress the repeat, not the first report. In
  `ResolveExamples`, before adding the diagnostics of `ParseExampleFile`, skip
  the add when `collector.IsFailed(relative)` is already true; still call
  `MarkFailed` and still `continue`, so the entry stays dropped and cascade
  suppression is unchanged. Do **not** extend the suppression to the
  diagnostics attributed to the *citing* document — `MissingExampleError`,
  `ExampleEscapesRootError`, and the not-a-regular-file diagnostic each name a
  different citer and each one is a distinct repair.
- **Tests required**: a fixture corpus under
  `indexer/tests/data/.corpora/failure/` whose two documents cite one malformed
  example, added to the `TestFailurePathFixtureCorpora` table with
  `Reported: 1`; plus a `ResolveExamples`-level test asserting that the second
  resolution of an already-failed example adds no diagnostic while still
  returning no entry.

## Target Files

- [x] `indexer/discovery.go`
- [x] `indexer/examples.go`
- [x] `indexer/discovery_test.go`
- [x] `indexer/examples_test.go`
- [x] `indexer/main_test.go`
- [x] `indexer/failure_paths_test.go`
- [x] `indexer/tests/data/.corpora/failure/shared-malformed-example/` (fixture for finding 2)

## Investigation Targets

- `indexer/discovery.go` (`WalkMarkdownFiles`, `containedFile`) — the root
  resolution the fix has to share
- `indexer/main.go` (`run`, `resolveCitations`) — the stage order and which
  collector each document's citations reach
- `indexer/examples.go` (`ResolveExamples`) — the diagnostic classes and which
  are attributed to the example file rather than to the citer
- `indexer/errors.go` (`ErrorCollector.Add`, `MarkFailed`, `IsFailed`) — no
  deduplication exists; the failure marking is the only handle
- `indexer/failure_paths_test.go` (`TestFailurePathFixtureCorpora`) — the table
  the new corpus is added to, and its exact-count contract
- `indexer/tests/data/.corpora/failure/README.md` — the fixture placement and
  do-not-repair constraint (RISK-001)

## Change Category

`Change Category: bug-fix, boundary-change`

The boundary is the filesystem read path: sweep every site that decides what
the corpus root is (`WalkMarkdownFiles`, `containedFile`, `DiscoverDocuments`,
`ValidateCorpus`/`corpusFiles`, `exampleHeadingStatement`) for the same
class of defect, and confirm each agrees on the resolved root after the fix.

## Investigation Notes

Round 1 re-verification, against the current sources and the built binary:

- **Finding 1 (false uncited)** — closed. `reportUncitedExamples`
  (`validate.go:138`) skips a path the collector marked failed; a corpus citing
  a malformed example reports only the parse fault, never an uncited one.
- **Finding 2 (nested fences)** — closed and correct. `fenceTracker`
  (`discovery.go:65-95`) records the opening character and run length and
  closes only on a run of the same character, at least as long, with an empty
  info string, which is the CommonMark closing rule. `ExtractStatements` and
  `parseExampleStatements` each drive their own instance over their own file,
  which is right — a shared instance across two files would be wrong — and both
  go through the one type, so they cannot drift. `declaredStatements` receives
  the live tracker, and every line is consumed exactly once (the `fenced` call
  is the left operand of the `||`, so short-circuiting cannot skip it). The
  regex tolerates any leading indentation rather than CommonMark's three-space
  limit; that is the right trade for this corpus (it keeps fences inside list
  items working) and a constructed nested-indented-fence document did not
  mis-collect.
- **Finding 3 (CRLF)** — closed. A CRLF example file emits `title: My Title`
  with no carriage return.
- **Finding 4 (detached citations)** — closed on every path. Detached
  documents (unknown parent, cycle) and their would-be descendants are all
  absent from `placedPaths`, so their citations reach the run's real collector;
  every document `buildNode` resolved is in the tree, so no document is
  double-reported. The placed-document lookup is correct under `--prefix`:
  `resolveCitations` (`main.go:231`) and `buildNode` (`tree.go:85`) call the
  same `emitPath(config.Prefix, document.Path)`, and `emitPath` is injective
  over distinct document paths for any prefix.
- **Finding 5 (wrapped `Statements:`)** — closed. A continuation line's
  identifiers are collected in order.
- **Finding 6 (fixtures)** — closed; the four corpora exist and assert exact
  diagnostic counts.

Security remediation interaction with AC-7 — verified, no regression:

```text
deleted example      -> GENERAL.md: example file not found: examples/a.md      (exit 1)
dangling symlink     -> GENERAL.md: example file not found: examples/a.md      (exit 1)
symlink out of root  -> GENERAL.md: example path escapes the source root: ...  (exit 1)
example is a dir     -> example examples/a.md cited by GENERAL.md is not a ... (exit 1)
```

`filepath.EvalSymlinks` returns a `*PathError` wrapping `ENOENT`, so
`errors.Is(err, fs.ErrNotExist)` still selects `MissingExampleError`
(`examples.go:66`) and the diagnostic still names both the citing standard and
the missing path.

Round 2 investigation, against the current sources:

- **Root resolution sites** — `containedFile` (`discovery.go:141`) is the only
  site that resolves the root today, and it does so per call. `WalkMarkdownFiles`
  walked the raw root and derived its relative paths from it, so the two halves
  disagreed on the root exactly when the raw root was a link. Every other site
  that reads a corpus file (`DiscoverDocuments`, `ResolveExamples`,
  `exampleHeadingStatement`, `corpusFiles`) goes through `containedFile` or
  through `WalkMarkdownFiles`, so resolving once in each is sufficient and the
  resolution is now shared as `resolveRoot`. `Document.AbsolutePath` keeps
  joining the raw source: it is an internal field, never emitted (asserted by
  `render_test.go:291`), and `discovery_test.go:252` pins it to the raw join.
- **Empty results** — `WalkMarkdownFiles` returning zero paths with a nil error
  was the shape the symlinked root produced. `DiscoverDocuments` now reports a
  corpus that holds no markdown file at all, rather than letting the run write
  `nodes: []` and exit 0. The check is on walked files rather than on discovered
  documents on purpose: a corpus whose only document has unparseable frontmatter
  yields zero documents and one diagnostic naming the offending file
  (`discovery_test.go:305`), and a hard error there would replace a usable
  repair instruction with a vaguer one.
- **Diagnostic attribution in `ResolveExamples`** — the diagnostics of
  `ParseExampleFile` name the example file (`MissingExampleHeadingError`,
  `MissingStatementsLineError`, `DuplicateStatementError`); every other
  diagnostic of the loop names the citing document and is left alone. The
  citer-attributed ones never call `MarkFailed`, so no path of the loop could
  suppress them even accidentally.
- **Which collector sees the repeat** — `BuildTree` resolves the citations of
  every placed document into the run's collector (`tree.go:100`), and
  `resolveCitations` (`main.go:234`) resolves those of the detached documents
  into the same one, placed documents going to a discarded collector. Assembly
  always runs first, so the surviving report is the first citer's on every
  ordering, and `IsFailed` is consulted on the same collector the add would go
  to.
- **`ErrorCollector`** — `Add` (`errors.go:288`) appends unconditionally and
  `Errors` only sorts, so the failure marking is the only handle; nothing else in
  the collector can tell a repeat from a distinct diagnostic.
- **Fixture placement** — `failureCorpus` (`failure_paths_test.go:65`) resolves
  corpora under `indexer/tests/data/.corpora/failure`, and the README there
  forbids repairing or relocating them, so the new corpus is added beside them
  and no existing one is touched.

New tests are non-vacuous: `cliDiagnostics` (`main_test.go:74`) requires an
error before comparing, so the prefixed-vs-unprefixed regression test cannot
pass on two empty runs; `TestFenceTracker` asserts a per-line expectation
vector and requires it to be the same length as the document.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-CODE-REVIEW-REMEDIATION (round 1) | Code Review Remediation | informs | The `fenceTracker`, `splitLines` and `resolveCitations` changes this round builds on |
| TASK-SEC-REMEDIATION | Security Remediation | blocks | `containedFile`, whose root resolution finding 1 makes the walk share |

## Remediation Context

- Source: code-reviewer
- Finding / failing command:
  - Finding 1: `indexer --source <symlink-to-corpus> --output tree.yaml` writes
    `nodes: []` and exits 0, where `--source <symlink-to-corpus>/` writes the
    full tree and exits 0.
  - Finding 2: a corpus whose two documents cite one heading-less example file
    reports `example file has no title heading` twice.
- Evidence: see the reproductions in Implementation Content above; both were
  run against a binary built from the current sources.
- Verification: `make test`, `make lint`, `make fmt-check`, plus
  `go test -count=2 -shuffle=on ./...`; and the self-index must still exit 0
  with byte-identical reruns and the same node/example counts as before.

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Sweep every site that decides the corpus root, and every site that adds a
      diagnostic attributed to a file other than the one being processed, for
      the same class of defect; fold any found within scope into the failing
      tests
- [x] Write the failing tests named under each finding
- [x] Run tests and confirm failure

### 2. Green Phase

- [x] Add minimal implementation to pass tests
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Improve code (maintain passing tests)
- [x] Confirm added tests still pass

## Completion Criteria

- [x] `indexer --source <symlink-to-corpus>` emits the same tree, byte for byte,
      as `--source <corpus>` and as `--source <symlink-to-corpus>/`
- [x] An unresolvable source root is reported as an error, never as an empty
      corpus
- [x] A corpus whose two documents cite one malformed example reports that fault
      exactly once, while two documents each citing a *different* missing example
      still report one diagnostic each
- [x] The existing failure corpora still report their exact diagnostic counts
- [x] All added tests pass
- [x] Verification commands from Remediation Context pass

## Notes

- Impact scope: the corpus read path (`WalkMarkdownFiles`, `DiscoverDocuments`,
  `ValidateCorpus`) and the citation resolution path (`ResolveExamples`,
  `resolveCitations`, cascade suppression via `MarkFailed`/`IsFailed`).
- Scope boundary: do not change the emitted tree. Node and example paths stay
  relative to the source root as the citing documents spelled them, never the
  paths symbolic links resolve to (`examples.go:38-40`), and `emitPath` remains
  the only place a prefix is applied. Do not touch the standards corpus,
  `docs/standards-tree.yaml`, or any fixture under
  `indexer/tests/data/.corpora/` other than the new corpus for finding 2
  (RISK-001, and the do-not-repair warning in the failure corpora README).
- The two untyped diagnostics already carried forward in the manifest
  (`example %s ... is not a regular file`, `failed to read example %s`) are out
  of scope here and remain a separate follow-up.
