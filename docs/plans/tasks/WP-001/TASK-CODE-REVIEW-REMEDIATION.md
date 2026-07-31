# Task: Code Review Remediation (Round 3)

Work Plan ID: WP-001
Task ID: TASK-CODE-REVIEW-REMEDIATION
Created Date: 2026-07-31
Description: Close the empty-corpus guard for a non-corpus directory that holds any markdown file. The containment-relativity fix is deferred by the user.
Acceptance Criteria Covered: AC-1, AC-6, AC-8

Status: **executed with a reduced scope directed by the user.** Only Finding 2
was fixed. Findings 1 and 3 are deferred follow-ups at the user's explicit
direction, not oversights — see the status note on each. Their analysis and
suggested fixes are preserved below verbatim so a later task can pick them up
without re-deriving them.

## Implementation Content

Two verified defects from the re-review of remediation round 2. Both are in
`indexer/discovery.go`. The two fixes of round 2 (symlinked `--source`, and the
duplicate report of a malformed example cited twice) were verified to hold and
are not reopened.

Finding 3 below is a low-severity observation recorded for completeness; it is a
follow-up, not a required change.

### Finding 1 (high, correctness) — containment depends on the spelling of `--source`

**Status: DEFERRED by the user — not fixed in this task, deliberately.** The
user directed that only Finding 2 be addressed now and that the containment
logic be left entirely untouched, including any partial change. The analysis and
the required fix below are preserved unchanged as the specification of the
follow-up. The defect is still present in `indexer/discovery.go` as described,
and the completion criteria and implementation steps covering it are marked
deferred rather than done.

`indexer/discovery.go:164` — `filepath.Rel(resolvedRoot, resolved)`.

`filepath.EvalSymlinks` preserves the relativity of its input, so a relative
`--source` produces a relative `resolvedRoot` (`resolveRoot`,
`indexer/discovery.go:125-131`). A symbolic link inside the corpus whose target
is spelled **absolutely** makes `resolved` absolute. `filepath.Rel` cannot
relate a relative base to an absolute target, returns an error, and line 166
maps that error to `errEscapesRoot` — so a file that is plainly inside the root
is judged an escape.

The consequence is two different results for one corpus, decided only by the
command line:

- In `WalkMarkdownFiles` (`indexer/discovery.go:263`) the file is dropped from
  the walk **silently**, so a standards document reached through such a link
  yields no node under a relative `--source` and a node under an absolute one.
  This breaks REQ-6 determinism and the RISK-008 claim that "relative vs
  absolute `--source` from different working directories produce identical
  output", and it publishes a wrong tree with exit 0.
- In `ResolveExamples` (`indexer/examples.go:59-61`) the same file is reported
  as `example path escapes the source root`, which is a false diagnostic.

Failure scenario, reproduced against the built binary:

```text
corpus/GENERAL.md          (title: General, no parent)
corpus/REAL.md             (title: Real Doc, parent: GENERAL.md)
corpus/LINKED.md    -> /abs/path/to/corpus/REAL.md   (absolute symlink target)

indexer --source /abs/path/to/corpus --output a.yaml
  # exit 0, nodes: GENERAL.md, LINKED.md, REAL.md
cd /abs/path/to && indexer --source corpus --output b.yaml
  # exit 0, nodes: GENERAL.md, REAL.md
cmp a.yaml b.yaml   # differ
```

And for the example path:

```text
corpus/A.md                        (examples: [examples/abs-link.md])
corpus/examples/real.md            (valid example)
corpus/examples/abs-link.md -> /abs/path/to/corpus/examples/real.md

indexer --source /abs/path/to/corpus   # abs-link.md accepted
cd /abs/path/to && indexer --source corpus
  # A.md: example path escapes the source root: examples/abs-link.md   <- false
```

Required fix: make both sides of the `filepath.Rel` comparison absolute, so the
comparison never depends on the spelling of `--source`. Resolve the root to an
absolute path in `resolveRoot` (`filepath.Abs` around the `EvalSymlinks` call),
and make the candidate absolute in `containedFile` before relating it. The
emitted paths are derived from `filepath.Rel` and therefore stay relative and
unchanged — this must be re-proven, see Verification.

Do **not** "fix" this by treating a `filepath.Rel` error as containment. The
error must stop being reachable for a contained file; a `Rel` error between two
absolute paths is still an escape.

### Finding 2 (medium, edge case) — the empty-corpus guard misses the common case

**Status: ADDRESSED in this task.** `DiscoverDocuments` now carries the
second guard (`len(documents) == 0 && collector.Len() == 0`) after the discovery
loop, alongside the walked-files guard it keeps. Reproduced red against the
built binary and against three new tests, then fixed; see Verification Record.

`indexer/discovery.go:303-306` — the guard fires only when the walk found no
markdown file at all.

Its own doc comment (`indexer/discovery.go:289-297`) states the hazard it
exists to close: a `--source` naming something that is not a standards corpus
"publishes an index declaring the corpus empty -- over whatever tree was
deployed before -- and exits zero while doing it". Keying the guard on walked
files rather than discovered documents leaves that hazard open for any
directory holding at least one markdown file, which is nearly every directory,
since a bare `README.md` satisfies it.

Failure scenario, reproduced against the built binary:

```text
mkdir nonCorpus && printf '# Some project\n\nHello.\n' > nonCorpus/README.md
printf 'nodes:\n- path: PRECIOUS.md\n' > deployed.yaml     # a real deployed tree
indexer --source nonCorpus --output deployed.yaml
# exit 0; deployed.yaml is now exactly "nodes: []"
```

Required fix: keep the existing walked-files guard (it is what preserves the
more useful diagnostic when the only document has unparseable frontmatter), and
add a second guard after the discovery loop: when no document was discovered
**and** the collector accumulated no diagnostic, return the same
"does not name a standards corpus" error. The `collector.Len() == 0` conjunct is
what preserves the round-2 behaviour — a corpus whose only document has
unparseable frontmatter has a non-empty collector, so it still gets the
diagnostic naming that document rather than this error.

### Finding 3 (low, edge case) — follow-up only, no change required

**Status: DEFERRED by the user — not fixed in this task, deliberately.** It was
already recorded as a follow-up rather than a required change, and the user
confirmed that the suppression key space is to be left untouched. The suggested
repair, preserved for the follow-up: give the example-diagnostic suppression its
own key space (or key it on the citing document rather than the cited path), so
a file that is both a standards document and a cited example does not have its
example-parse diagnostics suppressed by its own frontmatter failure.

`indexer/examples.go:94` — the `IsFailed(relative)` suppression is keyed on a
path, and `DiscoverDocuments` marks failed in the same key space. A file that is
both a discovered standards document and a cited `examples:` entry (possible
only outside an `examples/` directory) and that fails frontmatter validation is
marked failed before `ResolveExamples` runs, so its example-parse diagnostics
are suppressed although nothing else reported them.

Verified: a corpus where `SHARED.md` omits `topics` and is cited as an example
by `A.md` reports only `SHARED.md: missing required frontmatter key "topics"`;
with `topics` present the two example-parse diagnostics appear. The run still
exits 1, and the suppressed diagnostics surface on the next run after the first
repair, which is how the cascade suppression behaves elsewhere
(`indexer/validate.go:286-296`). Recorded, not required.

## Target Files

- [x] `indexer/discovery.go`
- [x] `indexer/discovery_test.go`
- [ ] `indexer/examples_test.go` — untouched: it carried the Finding 1 coverage
  only, and Finding 1 is deferred
- [x] `indexer/main_test.go`

## Investigation Targets

- `indexer/discovery.go` (`resolveRoot`, `containedFile`, `WalkMarkdownFiles`, `DiscoverDocuments`)
- `indexer/examples.go` (`ResolveExamples` error switch, `resolveExamplePath`)
- `indexer/validate.go` (`exampleHeadingStatement` — the second `containedFile` caller)
- `indexer/discovery_test.go:201-272` (existing symlink walk tests and the `newContainmentCorpus` / `linkCorpusEntry` helpers to extend)
- `indexer/main_test.go:403-454` (existing symlinked-root and empty-corpus CLI tests)

## Change Category

`Change Category: bug-fix, boundary-change`

The boundary is the corpus containment decision. Sweep every caller of
`containedFile` (`WalkMarkdownFiles`, `DiscoverDocuments`, `ResolveExamples`,
`exampleHeadingStatement`) for the same class of defect: a decision that changes
with the spelling of `--source`.

## Investigation Notes

- `DiscoverDocuments` (`indexer/discovery.go:298`) walks with
  `SkipExampleDirectories` and guards on `len(paths) == 0` only. `paths` counts
  walked markdown files, so a directory holding a bare `README.md` passes the
  guard, produces zero documents and returns `(documents, nil)`. `run`
  (`indexer/main.go:174-196`) then renders and writes, so the deployed tree is
  replaced with `nodes: []` at exit 0. Reproduced against a binary built from
  the current tree before any change.
- The collector is the discriminator between the two zero-document shapes: an
  unparseable-frontmatter document calls `collector.Add` + `collector.MarkFailed`
  (`indexer/discovery.go:326-329`), so `collector.Len() > 0` there, while a
  non-corpus directory adds nothing. `ErrorCollector.Len` already exists
  (`indexer/errors.go:319`).
- Existing coverage of the zero-document shapes:
  `discovery_test.go` "reports a source root holding no markdown file" (a bare
  `t.TempDir()`, which is why the gap went unnoticed) and "reports unparseable
  frontmatter and yields no document" (fixture `discovery/broken`, the round-2
  behaviour the `collector.Len() == 0` conjunct preserves).
  `main_test.go` "a corpus holding no markdown file is reported rather than
  emitted" is the CLI-level equivalent of the first.
- Every fixture corpus that is used as a `--source` root holds at least one
  titled document, so the new guard cannot fire on any of them:
  `discovery/corpus` (general.md), `discovery/broken` (no document but a
  non-empty collector), `validation/{uncited,heading,cascade,mention,union}`
  (`cascade` has the titled `child.md` beside the unparseable `broken.md`),
  `cli/{valid,failures,cascade}`, every corpus under `failure/`, and the
  repository root itself. `discovery/fences`, `tree` and `frontmatter` are read
  file-by-file and are never a discovery root.
- New corpora for these tests are built at run time under `t.TempDir()`, the
  shape `main_test.go` already uses for the malformed-example CLI case; no new
  committed fixture was needed.

## Remediation Context

- Source: code-reviewer
- Finding / failing command:
  - Finding 1: `indexer --source <abs>` and `indexer --source <rel>` over one corpus containing an absolutely-targeted internal symlink produce different trees, both exit 0.
  - Finding 2: `indexer --source <dir holding only a README.md> --output <deployed tree>` exits 0 and overwrites the deployed tree with `nodes: []`.
- Evidence: the reproductions in Findings 1 and 2 above, both run against a binary built from the current `indexer/` tree.
- Verification:
  - `cd indexer && make test && make lint && make fmt-check`
  - `go test -count=2 -shuffle=on ./...`
  - Byte-identity of the emitted tree must be unchanged by this task. Build the binary and confirm that all of `--source <repo root>`, `--source <repo root>/`, `--source ./standards` from the parent directory, and `--source <symlink to repo root>` produce byte-identical output with exit 0 and 46 `path:` entries, and that the node paths still match `docs/standards-tree.yaml`.

## Implementation Steps (TDD: Red-Green-Refactor)

Steps belonging to Finding 1 are marked DEFERRED: they were not attempted, by
the user's direction, and remain the specification of the follow-up.

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [ ] DEFERRED (Finding 1): Sweep all four `containedFile` callers for the spelling-dependent decision
- [ ] DEFERRED (Finding 1): Add a `discovery_test.go` case: a corpus holding a `*.md` symlink with an **absolute** target inside the root is walked identically under a relative and an absolute `root`, asserting equal path slices rather than non-emptiness
- [ ] DEFERRED (Finding 1): Add an `examples_test.go` case: the same shape cited as an `examples:` entry resolves to an entry under a relative `source`, with no `escapes the source root` diagnostic
- [ ] DEFERRED (Finding 1): Add a `main_test.go` case: the two `--source` spellings of one such corpus write byte-identical trees, with a non-empty-tree guard so the comparison cannot pass vacuously
- [x] Add a `main_test.go` case: a source directory holding a markdown file but no titled document is reported, no tree file is written, and a pre-existing output file is not truncated
- [x] Add a `discovery_test.go` case pinning the round-2 behaviour that must not regress: a corpus whose only markdown file has unparseable frontmatter returns no error from `DiscoverDocuments` and one diagnostic naming that file
- [x] Run tests and confirm failure

### 2. Green Phase

- [ ] DEFERRED (Finding 1): Absolutize the root in `resolveRoot` and the candidate in `containedFile`
- [x] Add the zero-documents/zero-diagnostics guard to `DiscoverDocuments`
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Update the doc comment of `DiscoverDocuments` to state the two-part corpus guard (`resolveRoot` and `containedFile` untouched: Finding 1 deferred)
- [x] Confirm added tests still pass

## Completion Criteria

- [ ] DEFERRED (Finding 1): A corpus containing an absolutely-targeted internal `*.md` symlink produces the same tree under a relative and an absolute `--source`
- [ ] DEFERRED (Finding 1): Such an entry cited under `examples:` no longer yields `example path escapes the source root`
- [x] A `--source` holding markdown files but no titled document exits non-zero, writes no tree file, and leaves a pre-existing output file intact
- [x] A corpus whose only document has unparseable frontmatter still reports that document rather than the corpus-guard error
- [x] Indexing this repository still exits 0 with 46 path entries, byte-identical across the four `--source` spellings
- [x] All added tests pass
- [x] Verification commands from Remediation Context pass

## Verification Record (Finding 2)

Red, proven before the fix:

- Against a binary built from the unchanged tree: `--source` at a directory
  holding only a frontmatter-less `README.md`, `--output` at a file containing
  `nodes:\n- path: PRECIOUS.md\n`, exited 0 and left the file as exactly
  `nodes: []`.
- The three added tests failed on `An error is expected but got nil`:
  `TestDiscoverDocuments/reports_a_source_root_holding_markdown_but_no_standards_document`,
  `TestRun/a_corpus_holding_markdown_but_no_standards_document_writes_no_tree`,
  `TestRun/a_corpus_holding_markdown_but_no_standards_document_spares_the_deployed_tree`.
- The fourth added case (unparseable-only corpus) passed before and after, which
  is the point: it pins the `collector.Len() == 0` conjunct.

Green:

- Same reproduction after the fix exits 1 with
  `no standards documents found under <dir>: --source does not name a standards corpus`,
  and the deployed tree file is byte-for-byte unchanged. An empty directory
  still reports through the walked-files guard.
- `make test`, `make lint` (0 issues), `make fmt`, and
  `go test -count=2 -shuffle=on ./...` all pass.
- `TestCompareGeneratedTreeAgainstReference` (AC-10),
  `TestIntegrationFixtureCorporaAreNotIndexed` (RISK-001) and every
  `TestFailurePathFixtureCorpora` subtest, `missing-example` (AC-7) included,
  are green.
- Byte-identity: the tree emitted by the pre-change binary and by the
  post-change binary is identical for all four `--source` spellings
  (absolute, trailing slash, symlink, relative from the parent directory),
  46 `path:` entries, exit 0.

## Notes

- Impact scope: `containedFile` is the single containment decision for four read sites, so a change to it moves discovery, example resolution and the REQ-8 scan together. The emitted paths come from `filepath.Rel` and must not move — that is the one thing to re-prove byte-for-byte.
- Scope boundary: do not touch `indexer/tree.go`, `render.go` or `compare.go` — neither finding is in them, and the emitted-path contract they carry must stay unchanged. Do not modify the standards corpus or `docs/standards-tree.yaml`. Do not act on Finding 3.
- Follow-up owed: Findings 1 and 3 remain open by the user's decision, not
  because they were missed or judged invalid. Finding 1 is the higher priority
  of the two — it publishes a wrong tree at exit 0 — and its required fix is
  written out in full above.
