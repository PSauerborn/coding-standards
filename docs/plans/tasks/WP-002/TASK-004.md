# Task: Document the Release Workflows in Root `CLAUDE.md`

Work Plan ID: WP-002
Task ID: TASK-004
Created Date: 2026-08-02
Description: Update `/home/agent/workspace/CLAUDE.md` to record that `bin/` indexer binaries and `standards-tree.yaml` are released via the two manually-triggered workflows, and reconcile that with the retained manual regenerate-and-hand-merge flow for corpus changes.
Acceptance Criteria Covered: (none — supporting documentation)

## Implementation Content

Update the root `CLAUDE.md` so a future agent understands the new release model:

1. Record the two manually-triggered release workflows:
   - `.github/workflows/release-indexer.yaml` — dispatch with a required `semver` (`major.minor.patch`); cross-builds `indexer-linux-amd64`, `indexer-linux-arm64`, `indexer-darwin-arm64` into `bin/`, commits them to `main` as `github-actions[bot]`, and tags `v{semver}-indexer`.
   - `.github/workflows/release-standards.yaml` — dispatch with a required `semver`; runs `bin/indexer-linux-amd64 --source . --output standards-tree.yaml` from the repo root, commits `standards-tree.yaml` to `main` as `github-actions[bot]` (skipped when the tree is unchanged), and tags `v{semver}-standards`.
   - Both abort before doing anything else if the semver is malformed or its tag already exists.
2. Reconcile with the existing **"Changing the corpus"** paragraph: the local regenerate + hand-merge into `standards-tree.yaml` + `nodeCountSnapshot`/`exampleCountSnapshot` bump flow **remains mandatory** for corpus changes in the same commit; the workflow is the *release* path, not a replacement for that flow.
3. Note that the `.github/workflows/*.yaml` files cannot join the standards tree (the walk skips dot-prefixed directories), so they require no tree regeneration or snapshot bump.
4. Note that released binaries under `bin/` are excluded from the `check-added-large-files` pre-commit hook (TASK-003).

Write only what is accurate against the final state of the three preceding tasks — read the actual files, do not paraphrase this task file. Keep the existing document structure, tone, and heading style; add to the relevant existing sections rather than appending a disconnected block. Do not describe the waived behaviors (no concurrency guard, no missing-binary diagnostic, no skip-if-unchanged for binaries) as if they exist.

Out of scope per rulings 3 and 8: `README.md` and the root `Makefile` `index` target are **not** touched, even though `./bin/indexer` will no longer resolve after a release.

## Target Files

- [x] CLAUDE.md

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- /home/agent/workspace/.github/workflows/release-indexer.yaml (TASK-001 deliverable — actual input name, binary filenames, tag literal, bot identity; read-only)
- /home/agent/workspace/.github/workflows/release-standards.yaml (TASK-002 deliverable — actual indexer invocation, skip-if-unchanged behavior, tag literal; read-only)
- /home/agent/workspace/.pre-commit-config.yaml (TASK-003 deliverable — the `check-added-large-files` exclude, to describe it accurately)
- /home/agent/workspace/CLAUDE.md ("Commands" and "Changing the corpus" sections — where the new content belongs and what must stay true)

## Investigation Notes

Literals read from the delivered files:

- `.github/workflows/release-indexer.yaml` — `name: Release Indexer`; `workflow_dispatch` with one required string input `semver` (described as `major.minor.patch`, no `v` prefix/suffix); `validate` job checks `^[0-9]+\.[0-9]+\.[0-9]+$` and that `refs/tags/v$SEMVER-indexer` does not exist, and gates every other job; three build jobs produce `indexer-linux-amd64`, `indexer-linux-arm64` (ubuntu-latest) and `indexer-darwin-arm64` (macos-latest); the `release` job checks out `ref: main`, downloads artifacts into `bin`, `chmod +x bin/indexer-*`, configures `user.name "github-actions[bot]"` / `user.email "github-actions[bot]@users.noreply.github.com"`, commits `bin/indexer-*` as `Release: indexer v$SEMVER`, pushes `main`, then tags `v$SEMVER-indexer`. No skip-if-unchanged for binaries, no concurrency guard.
- `.github/workflows/release-standards.yaml` — `name: Release Standards`; same required `semver` input and same `validate` gate but against tag `v$SEMVER-standards`; the `release` job checks out `ref: main`, runs `bin/indexer-linux-amd64 --source . --output standards-tree.yaml` (no missing-binary diagnostic), configures the same `github-actions[bot]` identity, and skips the commit (`git diff --quiet -- standards-tree.yaml`, `exit 0`) when the tree is unchanged; the tag `v$SEMVER-standards` is pushed unconditionally after the commit step.
- `.pre-commit-config.yaml` — `check-added-large-files` has `args: [--maxkb=1024]` and `exclude: ^bin/`.
- `CLAUDE.md` — "Commands" ends with the repository-root code block plus the slash-commands line; "Architecture" is a sequence of `**Bold lead-in.**` paragraphs, with "Discovery hazards" carrying a bullet list. New prose is added to those existing sections.

Unrunnable check: this container has no `pre-commit`, `python3`, `go` or `node`, so `pre-commit run --all-files` (markdownlint over `CLAUDE.md`) could not be executed here (confirmed by the orchestrator). Verification was done by reading the delivered files and diffing each documented literal against them by hand; the edit adds only paragraphs and a bullet list in the styles already present in the file.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-001 | Indexer Release Workflow (`release-indexer.yaml`) | blocks | Final workflow file whose behavior is documented |
| TASK-002 | Standards Release Workflow (`release-standards.yaml`) | blocks | Final workflow file whose behavior is documented |
| TASK-003 | Exclude `bin/` from the `check-added-large-files` Pre-commit Hook | blocks | Final `.pre-commit-config.yaml` exclude that is documented |

## Implementation Steps (TDD: Red-Green-Refactor)

Documentation has no unit-testable behavior; the check is a literal-by-literal
diff of the prose against the delivered files, plus pre-commit.

### 1. Red Phase

- [x] Read all Investigation Targets and record the exact literals to document: input name, binary filenames, tag formats, branch, bot identity, indexer CLI invocation, exclude pattern
- [x] List the statements in the current `CLAUDE.md` that the new release path affects (notably "Changing the corpus" and the root-`Makefile` `make index` line) and decide which must be qualified versus left intact
- [ ] Confirm `pre-commit run --all-files` (markdownlint over `CLAUDE.md`) passes before the edit, establishing the baseline — not runnable in this container; see Investigation Notes

### 2. Green Phase

- [x] Apply the documentation edits to `CLAUDE.md`
- [x] Re-check every documented literal against the source files
- [ ] Run `pre-commit run --all-files` from `/home/agent/workspace` and confirm it passes — not runnable in this container; see Investigation Notes

### 3. Refactor Phase

- [x] Re-read the edited sections for consistency with the surrounding tone and heading style; remove redundancy
- [ ] Re-run `pre-commit run --all-files` — not runnable in this container; see Investigation Notes

## Completion Criteria

- [x] `CLAUDE.md` names both workflow files and, for each, the required `semver` input, the target branch `main`, the commit identity `github-actions[bot]`, and the tag format (`v{semver}-indexer` / `v{semver}-standards`)
- [x] Every literal in the added prose (binary filenames, indexer CLI invocation, tag formats, bot identity, exclude pattern) matches the delivered workflow and pre-commit files exactly
- [x] The document states that both workflows abort when the semver is malformed or its tag already exists, and that the standards workflow skips the commit but still tags when `standards-tree.yaml` is unchanged
- [x] The "Changing the corpus" instruction still requires local regeneration, hand-merge into `standards-tree.yaml`, and the `nodeCountSnapshot`/`exampleCountSnapshot` bump in the same commit, now explicitly distinguished from the release path
- [x] No waived behavior is described as implemented
- [x] `README.md` and the root `Makefile` are unmodified
- [ ] `pre-commit run --all-files` passes (markdownlint over `CLAUDE.md`) — not runnable in this container; see Investigation Notes
- [x] All added tests pass

## Notes

- Impact scope: documentation only; no runtime behavior changes.
- Scope boundary: `README.md` (ruling 8) and root `Makefile` (ruling 3) must remain unchanged; do not edit the workflow files or `.pre-commit-config.yaml` — they are the deliverables of TASK-001/002/003 and are read-only here. `standards-tree.yaml` and `indexer/` are untouched.
