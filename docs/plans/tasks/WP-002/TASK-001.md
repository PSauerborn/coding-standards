# Task: Indexer Release Workflow (`release-indexer.yaml`)

Work Plan ID: WP-002
Task ID: TASK-001
Created Date: 2026-08-02
Description: Add `.github/workflows/release-indexer.yaml` — a manually-triggered workflow that gates on a validated, unused semver, cross-builds the three indexer binaries, commits them into `bin/` on `main` as `github-actions[bot]`, and tags `v{semver}-indexer`.
Acceptance Criteria Covered: AC-1, AC-2, AC-3

## Implementation Content

Create the new file `.github/workflows/release-indexer.yaml` (the `.github/` directory does not yet exist — create it). The workflow must implement:

1. **Trigger**: `workflow_dispatch` with a **required** input `semver`.
2. **Top-level `permissions: contents: write`** (needed for push + tag with the default token).
3. **Validation job** (runs first; every other job `needs` it, directly or transitively):
   - Checkout with tags available (`fetch-depth: 0` or `fetch-tags: true`) — `actions/checkout` does not fetch tags by default.
   - **Semver format check** per §6.2: accept exactly `major.minor.patch` with no prefix or suffix. The regex must be anchored (`^[0-9]+\.[0-9]+\.[0-9]+$`) so `v0.2.0`, `0.2.0-rc1`, and `0.2` are all rejected. On failure: print a message naming the offending input as not a valid semver (e.g. `Error: '0.2' is not a valid semver (expected major.minor.patch)`) and exit non-zero.
   - **Tag non-existence check** for `v{semver}-indexer`. On failure: print a message naming the version as already existing (e.g. `Error: version 0.2.0 already exists (tag v0.2.0-indexer)`) and exit non-zero.
   - **Injection safety**: never interpolate `${{ inputs.semver }}` directly into a `run:` script. Pass it through `env:` and reference it as a shell variable (`"$SEMVER"`), especially in the step that runs *before* the format gate.
4. **Build jobs** (all `needs: <validation job>`), each with Go from `indexer/go.mod` via `actions/setup-go` with `go-version-file: indexer/go.mod` (never a hardcoded version), building from the `indexer/` directory with `CGO_ENABLED=0` and explicit `GOOS`/`GOARCH`:
   - `ubuntu-latest`: `linux/amd64` -> `indexer-linux-amd64`
   - `ubuntu-latest`: `linux/arm64` -> `indexer-linux-arm64`
   - `macos-latest`: `darwin/arm64` -> `indexer-darwin-arm64` (macOS runner is mandatory per REQ-2.1)
   - Each uploads its executable as a pipeline artifact named per §6.1 (`actions/upload-artifact`).
5. **Release job** (`needs` all build jobs):
   - Checkout `main`; download all artifacts into `bin/`.
   - `chmod +x bin/indexer-*` **before** `git add` — `upload-artifact` drops POSIX modes, and a mode-100644 `bin/indexer-linux-amd64` breaks the standards workflow.
   - Configure the bot identity per §6.3 exactly:
     `git config --global user.name "github-actions[bot]"` and
     `git config --global user.email "github-actions[bot]@users.noreply.github.com"`.
   - `git add bin/indexer-*` (no `-f` needed: `bin/` is not gitignored), commit, push to `main`.
   - Create and push tag `v{semver}-indexer` **after** the push, pointing at the release commit.
6. **Action pinning**: third-party actions pinned to **major version tags only** (e.g. `actions/checkout@v5`) — not commit SHAs, not exact release tags. Verify each action's current major at implementation time.

Explicitly **out of scope / waived — do not add**: skip-if-unchanged for the binary commit, any `concurrency:` key, any composite action / `.github/scripts/` helper / reusable workflow (ruling 6: the validation and commit/tag shell stays inline).

## Target Files

- [x] .github/workflows/release-indexer.yaml

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- /home/agent/workspace/indexer/go.mod (the `go` directive — the version `setup-go` will resolve via `go-version-file`)
- /home/agent/workspace/indexer/Makefile (the `build` target — the canonical `go build` invocation and output name to cross-compile from)
- /home/agent/workspace/.pre-commit-config.yaml (`check-yaml`, `end-of-file-fixer`, `mixed-line-ending` hooks the new YAML file must pass)

## Investigation Notes

- `indexer/go.mod`: module `github.com/PSauerborn/standards/indexer`, `go 1.26.5`
  (a full patch version, so `setup-go`'s `go-version-file: indexer/go.mod`
  resolves an exact toolchain — no hardcoded version needed in the workflow).
- `indexer/Makefile`: canonical build is `go build -o indexer .` run from
  `indexer/`. No flags beyond `-o`, so the cross-builds only add
  `CGO_ENABLED=0` + `GOOS`/`GOARCH` and change `-o` to the release filename.
  Every target runs on the host toolchain (recorded GEN-003 deviation), so the
  workflow uses `actions/setup-go` rather than a container.
- `.pre-commit-config.yaml`: `check-yaml` (must parse), `end-of-file-fixer`
  (single trailing newline), `mixed-line-ending --fix=lf` (LF endings) and
  `trailing-whitespace` all apply to the new YAML file. `markdownlint` does
  not (markdown only). No `.github/` exclusion exists in any hook.
- Coding standards (`/etc/coding-standards/standards-tree.yaml`): no CI/CD or
  GitHub Actions node exists; only the root `GENERAL.md` (`scope: ['*']`)
  applies. GEN-001/GEN-002 (KISS/YAGNI, minimize entropy) back the waivers in
  Implementation Content — inline shell, no composite actions, no
  `concurrency:` key.

## Task Dependencies

(None.)

## Implementation Steps (TDD: Red-Green-Refactor)

Workflows cannot be executed here (no runner, no GitHub API). The TDD cycle is
run against **locally extracted shell snippets and local cross-builds**, driven
by scripts written to the session scratchpad. Never write reproduction output
into the repository — `bin/` and `standards-tree.yaml` must not change.

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations (Go version, build invocation)
- [x] Write a scratchpad verification script that asserts the behaviors below, and confirm it fails while `.github/workflows/release-indexer.yaml` does not exist:
  - the semver-format snippet extracted from the workflow rejects `0.2`, `v0.2.0`, `0.2.0-rc1` (non-zero exit, message naming the input as invalid) and accepts `0.2.0`
  - the tag-existence snippet, run in a scratch git repo where `v0.2.0-indexer` exists, exits non-zero and prints a message naming version `0.2.0` as existing; and exits zero when the tag is absent
  - the workflow file parses as YAML
- [x] Run the script and confirm failure

### 2. Green Phase

- [x] Author `.github/workflows/release-indexer.yaml` as specified in Implementation Content
- [x] Re-run the scratchpad script against the snippets extracted from the authored file and confirm it passes
- [x] Reproduce the three cross-builds locally, writing to the scratchpad (not `bin/`):
  `cd /home/agent/workspace/indexer && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o <scratch>/indexer-linux-amd64 .` plus the `linux/arm64` and `darwin/arm64` equivalents — all three must succeed and produce the §6.1 filenames

### 3. Refactor Phase

- [x] Review the workflow against the contract literals: binary filenames, tag literal `v${SEMVER}-indexer`, branch `main`, bot name/email, `permissions: contents: write`, `go-version-file: indexer/go.mod`, major-tag action pins
- [x] Confirm the `needs` graph makes validation gate every other job, and re-run the scratchpad script

## Completion Criteria

- [x] `.github/workflows/release-indexer.yaml` exists, parses as valid YAML, has LF line endings and a trailing newline
- [x] `workflow_dispatch` declares `semver` as a required input; `permissions: contents: write` is set at workflow level
- [x] The extracted semver-format snippet exits non-zero and prints a message naming `0.2` as not a valid semver, and rejects `v0.2.0` and `0.2.0-rc1`; it accepts `0.2.0` (AC-3)
- [x] The extracted tag-existence snippet exits non-zero and prints a message naming version `0.2.0` as already existing when `v0.2.0-indexer` is present (AC-2), and checkout fetches tags before that check
- [x] Every build and release job transitively `needs` the validation job (AC-2/AC-3 gating)
- [x] Local cross-builds produce `indexer-linux-amd64`, `indexer-linux-arm64`, `indexer-darwin-arm64` in the scratchpad; the workflow builds them on `ubuntu-latest`, `ubuntu-latest`, and `macos-latest` respectively with `CGO_ENABLED=0` and explicit `GOOS`/`GOARCH` (AC-1, REQ-2.1)
- [x] Artifacts are uploaded under the §6.1 names, downloaded into `bin/`, `chmod +x`-ed before `git add`, committed and pushed to `main` as `github-actions[bot]`, and only then tagged `v{semver}-indexer` (AC-1)
- [x] `${{ inputs.semver }}` is never interpolated into a `run:` body — it is passed via `env:` and used as a shell variable
- [x] No `concurrency:` key, no skip-if-unchanged on the binary commit, no shared action/script abstraction was introduced
- [x] All added tests pass

## Notes

- Impact scope: new `.github/` directory only. The indexer walk skips dot-prefixed directories, so this file cannot join the standards tree — no `standards-tree.yaml` regeneration and no `indexer/integration_test.go` snapshot bump is required.
- Scope boundary: do not modify `bin/`, `standards-tree.yaml`, the root `Makefile` (ruling 3), `README.md`, `.pre-commit-config.yaml` (TASK-003 owns it), any standards document, or any indexer source/test file.
- Local reproduction artifacts go to the session scratchpad only.
