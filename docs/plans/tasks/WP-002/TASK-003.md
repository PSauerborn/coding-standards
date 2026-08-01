# Task: Exclude `bin/` from the `check-added-large-files` Pre-commit Hook

Work Plan ID: WP-002
Task ID: TASK-003
Created Date: 2026-08-02
Description: Add an `exclude` pattern for `bin/` to the `check-added-large-files` hook only in `.pre-commit-config.yaml`, so released indexer binaries can be committed without raising the global 1024kb limit.
Acceptance Criteria Covered: AC-1

## Implementation Content

The indexer release workflow (TASK-001) commits cross-compiled Go binaries into `bin/`. These exceed the `check-added-large-files` `--maxkb=1024` limit and would trip pre-commit.

Per ruling 7, make exactly one change to `/home/agent/workspace/.pre-commit-config.yaml`:

- Add an `exclude:` regex scoped to the `check-added-large-files` hook entry that matches files under `bin/` (e.g. `^bin/`).

Constraints:

- Do **not** change `args: [--maxkb=1024]` — the global limit stays.
- Do **not** add an `exclude` at the repo or `repos:` level.
- Do **not** touch any other hook (`end-of-file-fixer`, `trailing-whitespace`, `mixed-line-ending`, `check-merge-conflict`, `check-case-conflict`, `check-yaml`, `check-json`, `markdownlint`, `detect-secrets`).
- Keep the existing inline-comment style of the file intact.

## Target Files

- [x] .pre-commit-config.yaml

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- /home/agent/workspace/.pre-commit-config.yaml (the `check-added-large-files` hook entry and the `exclude:` style already used by the `markdownlint` hook)

## Investigation Notes

- `.pre-commit-config.yaml` has three `repos:` entries. The `check-added-large-files` entry
  (lines 16-17 before the change) was `- id: check-added-large-files    # prevent large blobs from being committed`
  followed by `args: [--maxkb=1024]` at 8-space indentation.
- The only existing `exclude:` in the file is on the `markdownlint` hook:
  `exclude: (docs/.*|indexer/tests/data/.corpora/.*)` — hook-scoped, unquoted regex, same 8-space indentation.
  The new `exclude: ^bin/` follows that style, with an inline comment to match the file's comment convention.
- `bin/` currently contains only an empty `.gitignore`; nothing in `bin/` was touched by this change.
- **Environment limitation**: this container has no `pre-commit`, no `python3`/`pip`, no `go`, and
  `apt-get install` fails (exit 100, no network/root). The oversized-file reproduction and
  `pre-commit run --all-files` steps could therefore not be executed here; they remain to be run in an
  environment with pre-commit installed. No YAML parser (python, go, node, perl YAML module) was available
  to machine-validate the file either; the added line is a plain scalar mapping key at the same indentation
  as the sibling `args:` key.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-001 | Indexer Release Workflow (`release-indexer.yaml`) | informs | The `bin/indexer-*` binaries the workflow commits are what trip the size hook; their paths determine the exclude pattern |

## Implementation Steps (TDD: Red-Green-Refactor)

This change has no unit-testable behavior; the executable check is pre-commit itself.

### 1. Red Phase

- [x] Read the Investigation Target and record the current `check-added-large-files` entry
- [ ] Reproduce the failure: stage an oversized dummy file at `bin/indexer-linux-amd64` (build one into the scratchpad and copy it in, or generate a >1024kb file), run `pre-commit run check-added-large-files --all-files` from `/home/agent/workspace`, and confirm the hook fails on it
- [x] Keep the dummy file out of any commit — unstage and delete it after each check; `bin/` must not change in this changeset

### 2. Green Phase

- [x] Add the `exclude` pattern to the `check-added-large-files` hook only
- [ ] Re-run the reproduction with the oversized `bin/` file staged and confirm `check-added-large-files` now passes
- [x] Remove the dummy file and restore `bin/` to its original state

### 3. Refactor Phase

- [ ] Run `pre-commit run --all-files` from `/home/agent/workspace` and confirm it passes
- [x] Confirm the diff is exactly one added `exclude` line under `check-added-large-files`

## Completion Criteria

- [ ] `.pre-commit-config.yaml` parses as valid YAML and `pre-commit run --all-files` passes
- [ ] An oversized file under `bin/` is accepted by `check-added-large-files`, while an oversized file outside `bin/` is still rejected
- [x] `args: [--maxkb=1024]` is unchanged and no global/`repos:`-level exclude was added
- [x] No hook other than `check-added-large-files` was modified
- [x] `bin/` contents are unchanged in the working tree at task completion
- [ ] All added tests pass

## Notes

- Impact scope: pre-commit behavior for files under `bin/` only.
- Scope boundary: do not modify `bin/` contents, `standards-tree.yaml`, the root `Makefile`, `README.md`, `.markdownlint.yaml`, `.secrets.baseline`, or the workflow files owned by TASK-001/TASK-002.
