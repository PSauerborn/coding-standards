# Task: Security remediation — symlink-aware corpus containment and fd-based temp file permissions

Work Plan ID: WP-001
Task ID: TASK-SEC-REMEDIATION
Created Date: 2026-07-31
Description: Close three security findings in the indexer CLI: the example-path containment check is purely lexical and is bypassed by a symlinked directory inside the corpus, the document walk reads symlinked `*.md` files whose targets live outside the source root, and the atomic-write sequence chmods the temporary file by path rather than through its open descriptor.
Acceptance Criteria Covered: (none — remediation of review findings, no new acceptance criteria)

## Implementation Content

The indexer resolves attacker-influenceable relative paths (`examples:` frontmatter entries) and walks attacker-influenceable directory trees, then reads the resulting paths with `os.ReadFile`. Containment is decided lexically only, so a symbolic link stored inside the corpus makes a path that is lexically inside the source root resolve to a target outside it. This task makes containment symlink-aware at every point where the tool decides that a path belongs to the corpus, and removes a path-based `chmod` from the output write.

Three findings, all in the WP-001 changeset:

### Finding 1 — `indexer/examples.go:47` (path traversal, medium)

`resolveExamplePath` (`indexer/examples.go:107-129`) decides containment with `filepath.Join` + `filepath.Rel` on the *lexical* path and rejects absolute entries. It never resolves symbolic links, so a link stored inside the corpus defeats it. The resolved path is then read at line 47 with `os.ReadFile`.

Attack scenario: a contributor opens a pull request against an indexed repository adding `docs/examples/root` as a symbolic link to `/` (or to `$HOME`), plus a standards document whose frontmatter declares `examples: ["examples/root/etc/passwd"]`. `resolveExamplePath` computes the relative path `docs/examples/root/etc/passwd`, which does not start with `../`, so containment passes and CI reads a file outside the corpus. Two consequences follow. First, partial content disclosure: if the target's first non-empty line matches the example heading grammar, `parseExampleTitle` (`indexer/examples.go:134-146`) copies that line's text into the `title` of the emitted tree, which is committed and read by downstream agents; `parseExampleStatements` likewise copies any bracketed identifier from a `Statements:` line. Second, resource exhaustion: citing `examples/root/dev/zero` makes `os.ReadFile` allocate until the CI runner is out of memory, since no size bound is applied. The same link also yields a file-existence oracle, because a missing target reports `MissingExampleError` while an existing unparseable one reports heading/statements diagnostics.

### Finding 2 — `indexer/discovery.go:135` (path traversal, low)

`WalkMarkdownFiles` (`indexer/discovery.go:79-109`) uses `filepath.WalkDir`, which does not descend into symlinked directories but *does* yield symlinked files. The extension test at line 94 accepts a symlink named `*.md`, and line 135 reads it with `os.ReadFile`, which follows the link.

Attack scenario: a pull request adds `docs/notes.md` as a symbolic link to a path outside the corpus (for example `../../private/plan.md` or an absolute path under the CI runner's home directory). If the target opens with a frontmatter block declaring a `title`, the indexer emits a node whose `title`, `description`, `scope`, `topics` and `aliases` are taken verbatim from that out-of-root file and publishes them in the committed tree. `validate.go:237` re-reads example files through the same unresolved path and inherits the same weakness.

### Finding 3 — `indexer/main.go:255` (path traversal / insecure temporary file, low)

`writeTree` creates the temporary file with `os.CreateTemp` (mode 0600, `O_EXCL`) but then widens it with `os.Chmod(name, treeFileMode)` — by path, not through the open descriptor. `os.Chmod` follows symbolic links.

Attack scenario: the indexer is run with `--output` in a directory writable by other local users (a shared build directory, or `/tmp` in the documented `./indexer --source .. --output /tmp/tree.yaml` workflow from `indexer/README.md:113`). A local attacker watching the directory sees `.indexer-tree-*.yaml` appear, unlinks it and recreates the same name as a symbolic link to a private file owned by the invoking user. The subsequent `os.Chmod` follows the link and sets that file to 0644, exposing it to every local user. The rename then moves the attacker's symlink rather than the tree, so the attack also suppresses the output.

## Target Files

- [x] `indexer/examples.go` — symlink-aware containment for cited example paths
- [x] `indexer/discovery.go` — reject or contain symlinked corpus files during the walk
- [x] `indexer/validate.go` — re-read example files through the same contained path
- [x] `indexer/main.go` — chmod the temporary file through its descriptor
- [x] `indexer/examples_test.go`
- [x] `indexer/discovery_test.go`
- [x] `indexer/validate_test.go`
- [x] `indexer/main_test.go`
- [x] Fixtures under `indexer/tests/data/.corpora/` only (see Scope boundary) — no committed fixture was
  needed: the containment corpora are built under `t.TempDir()` at run time (see Investigation Notes), so
  nothing was added to the corpus the tool indexes and no test writes into the working tree

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- `indexer/examples.go` (`resolveExamplePath`, `ResolveExamples`) — the lexical containment check and the read it guards
- `indexer/discovery.go` (`WalkMarkdownFiles`, `DiscoverDocuments`) — the walk's entry filter and the document read
- `indexer/validate.go` (`exampleHeadingStatement`, `corpusFiles`) — the second read of example files, which must use the same contained path
- `indexer/main.go` (`writeTree`, `writeAndClose`) — the temp-file/chmod/rename sequence
- `indexer/integration_test.go` (`copyCorpus`, `integrationCopySymlink`) — the harness already recreates symlinks, so fixture corpora can carry them
- `indexer/errors.go` — `ExampleEscapesRootError` is the existing diagnostic for a path that leaves the root and should be reused

## Change Category

`Change Category: boundary-change`

The corpus/filesystem boundary is what changes: every place the tool decides a path belongs to the corpus must be swept for the same lexical-only defect, not just the one reported at `examples.go:47`.

## Investigation Notes

- `examples.go`: `resolveExamplePath` rejects absolute/rooted entries and then decides containment on the
  lexical join alone. `ResolveExamples` re-joins `source` with the returned relative path at line 47 and
  reads it. The relative path it returns is also the path emitted into the tree, so the fix must keep
  emitting the *lexical* relative path and use the resolved path only for the read.
- `discovery.go`: `WalkMarkdownFiles` filters on `entry.IsDir()`, the dot-prefix/`examples` exclusions and
  the `.md` extension only; a symlinked `*.md` is admitted and `DiscoverDocuments` reads it at line 196.
  Both walkers (discovery and `corpusFiles`) go through this one function, so filtering here contains the
  example-file listing of `validate.go` as well.
- `validate.go`: `exampleHeadingStatement` joins `source` with the example path a second time and reads it,
  which is the drift the single containment helper has to close.
- `main.go`: `writeTree` creates the temp file with `os.CreateTemp`, hands it to `writeAndClose` (which
  closes it) and only then chmods it *by path*, so the descriptor is already gone when the mode is widened.
  Moving the chmod into `writeAndClose` is what makes the descriptor available at the point of the call.
- `errors.go` is not in the write set, so no new error type is introduced: an escape reuses
  `ExampleEscapesRootError`, a missing target keeps `MissingExampleError`, and a cited path resolving to
  something other than a regular file is reported as an untyped diagnostic like the existing read failure.
- `integration_test.go`'s `copyCorpus` recreates symlinks and rejects any corpus entry that is neither a
  directory, a regular file nor a symlink, so a committed FIFO fixture would break the harness. The
  symlink corpora are therefore built at run time under `t.TempDir()`: a committed symlink whose target
  escapes the repository is not portable, and creating one at run time under `.corpora/` would mutate the
  working tree, which no test of this suite is allowed to do.
- macOS `t.TempDir()` paths live under `/var/folders/...`, where `/var` is a symlink, so the root must be
  resolved before the candidate is joined onto it or every fixture path would read as an escape.

## Task Dependencies

(No blocking tasks — all twelve execution tasks have landed.)

## Remediation Context

- Source: security-reviewer
- Finding / failing command: 3 findings — `indexer/examples.go:47` (path traversal, medium), `indexer/discovery.go:135` (path traversal, low), `indexer/main.go:255` (insecure temporary file, low)
- Evidence: `resolveExamplePath` decides containment with `filepath.Join` + `filepath.Rel` alone (`indexer/examples.go:118-127`) and never calls `filepath.EvalSymlinks` or checks the entry type; `filepath.WalkDir` yields symlinked files, which `indexer/discovery.go:94` accepts on extension and `indexer/discovery.go:135` reads through; `indexer/main.go:255` calls `os.Chmod(name, treeFileMode)` on the temp file's *path* while `*os.File.Chmod` is available on the handle already held.
- Verification: `make test`, `make lint` and `make fmt-check` from `indexer/` all pass, and the new tests below fail against the current implementation and pass after the fix.

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Sweep the corpus/filesystem boundary for the same class of defect: every `os.ReadFile` reached from a corpus-derived path (`examples.go:47`, `discovery.go:135`, `validate.go:237`) and every path-based filesystem mutation in `main.go`
- [x] Add a fixture corpus under `indexer/tests/data/.corpora/` containing a symlinked directory whose target is outside the corpus root, cited by a document's `examples:` entry, and a symlinked `*.md` file whose target is outside the corpus root. Create the links in the test at run time if committing them is impractical, but keep every fixture under `.corpora/` (RISK-001) — built at run time under `t.TempDir()` by `newContainmentCorpus`, so nothing new sits inside the indexed corpus
- [x] Write a failing test asserting that an `examples:` entry traversing a symlinked directory out of the source root is reported as `ExampleEscapesRootError` and contributes no entry, and that the target file is never read
- [x] Write a failing test asserting that a symlinked `*.md` file whose target resolves outside the source root contributes no node, and that its target's frontmatter never reaches the emitted tree
- [x] Write a failing test asserting that a symlink *within* the source root still resolves (containment, not a blanket symlink ban) — or, if the chosen fix is to reject non-regular entries outright, assert that rejection explicitly so the decision is pinned
- [x] Write a failing test asserting that the temporary file's permissions are set through its descriptor: replace the temp path with a symlink to a decoy file between creation and publication is impractical to race in a test, so assert instead that `writeTree` never calls a path-based chmod by pinning that the decoy file's mode is unchanged after a run whose temp path was pre-replaced, or restructure `writeTree` so the chmod is applied inside `writeAndClose` and unit-test that function directly — `writeAndClose` takes the mode and is unit-tested with the created name renamed away, which a path-based chmod cannot survive
- [x] Run tests and confirm failure

### 2. Green Phase

- [x] `indexer/examples.go`: after the existing lexical check, resolve the joined path with `filepath.EvalSymlinks` (and resolve the source root once per run the same way) and re-verify with `filepath.Rel` that the *resolved* path is still inside the resolved root; report `ExampleEscapesRootError` when it is not. A path that does not exist must keep reporting `MissingExampleError`, so handle `fs.ErrNotExist` from `EvalSymlinks` as not-found rather than as an escape
- [x] `indexer/examples.go`: reject a cited path whose final target is not a regular file, so a device or FIFO cannot be read
- [x] `indexer/discovery.go`: in the `WalkMarkdownFiles` callback, skip entries whose `entry.Type()` is not a regular file, or apply the same resolved-root containment check before the file is admitted to the corpus. Keep the dot-prefixed-directory and `examples/` exclusions exactly as they are
- [x] `indexer/validate.go`: route `exampleHeadingStatement`'s read through the same contained-path helper rather than joining the path again
- [x] `indexer/main.go`: replace `os.Chmod(name, treeFileMode)` with `temporary.Chmod(treeFileMode)` on the open descriptor, applied before the file is closed, and drop the path-based call
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Factor the resolved-root containment check into a single helper so no caller can drift from it, mirroring the way `WalkMarkdownFiles` is the single walker and `emitPath` the single prefixing point — `containedFile` in `discovery.go`, used by the walk, the document read, example resolution and the validation re-read
- [x] Confirm added tests still pass, and that the full suite — including `TestIntegrationFixtureCorporaAreNotIndexed` and the RISK-008 relative-vs-absolute `--source` determinism test — still passes, since `EvalSymlinks` can change how a `--source` given through a symlinked path is spelled

## Completion Criteria

- [x] An `examples:` entry that traverses a symbolic link out of the source root is reported as escaping the root, contributes no entry, and its target is never read
- [x] A symlinked `*.md` file whose target resolves outside the source root contributes no node and no frontmatter to the emitted tree
- [x] `writeTree` sets the published file's mode through the open descriptor; no path-based `chmod` remains in the write sequence
- [x] Output for a corpus containing no symbolic links is byte-identical to the output before this change (the reference `docs/standards-tree.yaml` is unaffected): `TestCompareGeneratedTreeAgainstReference`, the integration snapshot assertions and the determinism suite all stay green, and a run over the repository still emits the same 18 nodes and 28 example entries
- [x] All added tests pass
- [x] Verification command from Remediation Context passes (`make test`, `make lint`, `make fmt`/`make fmt-check` from `indexer/`)

## Notes

- Impact scope: example resolution, document discovery, corpus validation and the output write. `filepath.EvalSymlinks` on the source root may normalize a `--source` argument given through a symlink (a `/tmp` path on macOS resolves to `/private/tmp`), which the RISK-008 determinism tests and the integration harness's `t.TempDir()` corpora will exercise — resolve the root once and compare resolved against resolved so the two sides stay consistent.
- Scope boundary: the standards corpus itself (`docs/`, `standards/`, `docs/standards-tree.yaml`) must stay unchanged — this task changes tool behaviour only. All new fixtures stay under `indexer/tests/data/.corpora/` (RISK-001): a fixture placed anywhere else joins the indexed corpus and fails the tool on its own repository.
