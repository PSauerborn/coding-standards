# Changeset: CI Release Pipelines for Indexer and Standards Tree

Work Plan ID: WP-002
Created Date: 2026-08-02
Author: documenter

## Summary

Adds two manually-triggered GitHub Actions release pipelines — one that
cross-builds and releases the `indexer` binary, one that regenerates and
releases `standards-tree.yaml` — plus the pre-commit and documentation changes
needed to accommodate committing built binaries into `bin/`. The work was
reviewed across a 2-iteration loop (security review, code review x2, quality
control) that hardened the original design with several controls beyond the
initial spec.

## Changes

### Added

- `.github/workflows/release-indexer.yaml` ("Release Indexer"): manually
  dispatched via `workflow_dispatch` with a required `semver` input
  (`major.minor.patch`, no `v` prefix/suffix). Validates the version, cross-builds
  `indexer-linux-amd64`, `indexer-linux-arm64` and `indexer-darwin-arm64`,
  commits them under `bin/` on `main` as `github-actions[bot]`, and tags
  `v{semver}-indexer`.
- `.github/workflows/release-standards.yaml` ("Release Standards"): manually
  dispatched the same way. Regenerates `standards-tree.yaml` using the
  committed `bin/indexer-linux-amd64`, commits the result on `main` as
  `github-actions[bot]` (skipping the commit if regeneration reproduces the
  committed tree byte-for-byte), and tags `v{semver}-standards`.

### Changed

- `.pre-commit-config.yaml`: `check-added-large-files` now carries
  `exclude: ^bin/`, since the release workflow commits binaries there that
  would otherwise trip the size hook.
- `CLAUDE.md`: documents both release workflows under "Commands" (trigger,
  inputs, what each does, and the shared validation gate), and adds a bullet
  under "Discovery hazards" noting that `.github/workflows/` is dot-prefixed
  and therefore never joins the standards tree, plus a note under "Changing
  the corpus" clarifying that the Release Standards workflow is the release
  path for the tree, not a substitute for the manual hand-merge/snapshot-bump
  procedure.
- `indexer/README.md` ("Build environment"): the existing recorded GEN-003
  deviation (no containerized build) is revisited and extended to explicitly
  cover the release cross-build path — the three build jobs run `go build`
  directly on ephemeral GitHub runners via `actions/setup-go` pinned to
  `indexer/go.mod`, with `CGO_ENABLED=0`, giving the same toolchain pin and
  isolation a build container would have bought. A new deviation is recorded
  against GEN-004 ("Makefiles should define common build/test targets"): the
  release cross-build command is inlined in the workflow rather than exposed
  as an `indexer/Makefile` target, because `SPEC-002` §3.2 scopes changes to
  `indexer/` build tooling out of this work.
- `indexer/Makefile` (header comment only): extends the existing GEN-003
  deviation comment to reference the release workflow, and cross-references
  `indexer/README.md` for the GEN-004 deviation. No targets were added,
  removed or modified — the Makefile still has no release-build target, by
  design.

### Security

- Both workflows start from `permissions: {}` at the workflow level; only the
  `release` job in each (the one that pushes) opts back into
  `permissions: contents: write`.
- A `Check release ref is main` step gates every dispatch before checkout, so
  a dispatch from any non-`main` ref is rejected before it can reach
  repository content.
- The three build jobs in `release-indexer.yaml` check out `${{ github.sha }}`
  rather than a branch name, pinning all three builds to the exact commit the
  run was dispatched against; the `release` job then asserts `main` has not
  advanced past that commit before committing, aborting and asking for a
  re-dispatch otherwise.
- `persist-credentials: false` is set on every checkout step that does not
  push (`validate` and the three build jobs), so the `GITHUB_TOKEN` is not
  left in `.git/config` for jobs that never need to push.
- `actions/download-artifact` in the release job of `release-indexer.yaml`
  uses `pattern: indexer-*` to restrict which artifacts are pulled into
  `bin/`, followed by an explicit assertion that exactly the three expected
  binary filenames are present before any of them is made executable or
  committed.
- The semver validation regex rejects components with leading zeros
  (`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`), so `01.2.3` cannot
  be dispatched as a second spelling of `1.2.3`.
- The `semver` dispatch input is passed through `env:` rather than
  interpolated into shell script bodies, so an arbitrary dispatch input
  cannot inject shell.
- `release-standards.yaml`'s commit step stages `standards-tree.yaml` first
  and tests the staged diff (`git diff --cached --quiet`) rather than the
  working-tree diff, so a first release where `standards-tree.yaml` is
  untracked on `main` is committed rather than silently skipped (an untracked
  path reports no change under a plain `git diff`).

## File Diffs

### `CLAUDE.md`

Documents the two release workflows, notes `.github/workflows/` as
dot-prefixed (never joins the standards tree), and clarifies that Release
Standards is the release path for the tree, not a substitute for the
hand-merge/snapshot-bump procedure.

```diff
+ Releases have no local command — both are GitHub Actions workflows,
+ triggered manually via `workflow_dispatch` with a required `semver` input
+ (`major.minor.patch`, no `v` prefix or suffix):
+
+ - `.github/workflows/release-indexer.yaml` ("Release Indexer") builds
+   `indexer-linux-amd64`, `indexer-linux-arm64` and `indexer-darwin-arm64`,
+   commits them under `bin/` on `main` as `github-actions[bot]`, and tags
+   `v{semver}-indexer`.
+ - `.github/workflows/release-standards.yaml` ("Release Standards") runs
+   `bin/indexer-linux-amd64 --source . --output standards-tree.yaml` from the
+   repository root, commits `standards-tree.yaml` on `main` as
+   `github-actions[bot]`, and tags `v{semver}-standards`. A regeneration that
+   reproduces the committed tree has nothing to commit, so the commit is
+   skipped and the tag is still pushed.
+
+ Both workflows validate the version first and abort before building,
+ committing or tagging anything if the `semver` is malformed or its tag
+ already exists. The binaries released into `bin/` are why
+ `check-added-large-files` carries `exclude: ^bin/` in `.pre-commit-config.yaml`.
```

### `.pre-commit-config.yaml`

Excludes `bin/` from the large-file hook so committed release binaries do not
trip it.

```diff
       - id: check-added-large-files    # prevent large blobs from being committed
         args: [--maxkb=1024]
+        exclude: ^bin/                 # released indexer binaries are committed here
```

### `indexer/README.md`

Extends the recorded GEN-003 deviation to the release cross-build path and
records a new GEN-004 deviation for the inlined release build command.

```diff
+ The deviation used to close by asking for a revisit "if the module ever
+ ships as a deployed artifact rather than as a checkout-local tool". That has
+ now happened: `.github/workflows/release-indexer.yaml` cross-compiles the
+ binary, commits it to `bin/` on `main` and tags the result
+ `v{semver}-indexer`. Revisited, the deviation stands, extended to cover the
+ release build path: [...]
+
+ **The release cross-build command itself is inlined in
+ `.github/workflows/release-indexer.yaml` rather than exposed as an
+ `indexer/Makefile` target. This is a recorded deviation from `GEN-004`**
+ [...] The reason is scope: this work was scoped to exclude changes to the
+ `indexer/` build tooling — `SPEC-002` §3.2 puts the indexer out of scope —
+ so the three `Build indexer` steps keep their inline `go build` invocations.
```

### `indexer/Makefile`

Header comment only: extends the GEN-003 deviation note to the release
workflow and cross-references the GEN-004 deviation recorded in the README.
No targets changed.

```diff
+ # The same deviation covers the release cross-builds in
+ # .github/workflows/release-indexer.yaml, which is where this module ships as a
+ # deployed artifact: those jobs also run go build outside a container, on
+ # ephemeral GitHub runners, against a toolchain pinned by
+ # "go-version-file: indexer/go.mod" — the same go directive that pins a build
+ # here — and with CGO_ENABLED=0, so the released binaries link nothing from the
+ # runner image. Revisit if the released binary ever needs cgo, a system library
+ # or a code generator, since the runner image would then decide what ships.
+ #
+ # That release build command is inlined in the workflow rather than exposed as a
+ # target below; see indexer/README.md, "Build environment", for the full
+ # reasoning and for the GEN-004 deviation that records it.
```

## New Files

| File Path | Description |
| --------- | ----------- |
| `.github/workflows/release-indexer.yaml` | Manually-dispatched pipeline that validates a semver, cross-builds the indexer for linux/amd64, linux/arm64 and darwin/arm64, commits the binaries to `bin/` on `main`, and tags `v{semver}-indexer`. |
| `.github/workflows/release-standards.yaml` | Manually-dispatched pipeline that validates a semver, regenerates `standards-tree.yaml` with the committed `bin/indexer-linux-amd64`, commits the tree on `main` if it changed, and tags `v{semver}-standards`. |

## Accepted Gaps and Environment Notes

Recorded here so they are not mistaken for oversights on a later pass:

- **No `concurrency` guard** on either workflow — concurrent dispatches are
  not deduplicated or queued.
- **No missing-binary diagnostic** in `release-standards.yaml` if
  `bin/indexer-linux-amd64` is absent — the regeneration step simply fails
  with the shell's own "not found" error rather than a purpose-written check.
- **No skip-if-unchanged short-circuit** for the indexer binary commit in
  `release-indexer.yaml` (unlike the standards tree commit, which does skip
  when unchanged) — a re-release of identical binaries still produces a
  commit.
- The root `Makefile`'s `index` target still references `./bin/indexer`,
  which will not resolve once the indexer ships only as the versioned
  `bin/indexer-<os>-<arch>` binaries produced by the release workflow. Left
  as-is; out of scope per user instruction (root `Makefile` excluded from
  this changeset).
- **GEN-004** (Makefile targets for common build/test commands) is accepted
  as a recorded, scoped-out deviation for the inlined release cross-build
  command — see `indexer/README.md`, "Build environment".
- GitHub Actions workflows could not be executed in this environment (no
  `go`, `python3`, `pre-commit`, or GitHub Actions runner available), so all
  acceptance evidence for this changeset is static: workflow YAML inspection,
  scratchpad harnesses that parse and replay the extracted `run:` bodies
  against throwaway git repositories, and — for the cross-builds — an
  actually-downloaded Go 1.26.5 toolchain reproducing all three builds
  locally. `pre-commit run --all-files` likewise could not be executed against
  `.pre-commit-config.yaml`'s changes, for the same reason.
