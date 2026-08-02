# Task: Standards Release Workflow (`release-standards.yaml`)

Work Plan ID: WP-002
Task ID: TASK-002
Created Date: 2026-08-02
Description: Add `.github/workflows/release-standards.yaml` — a manually-triggered workflow that gates on a validated, unused semver (inline mirror of TASK-001), regenerates `standards-tree.yaml` with `bin/indexer-linux-amd64`, commits it to `main` as `github-actions[bot]` unless unchanged, and always tags `v{semver}-standards`.
Acceptance Criteria Covered: AC-4, AC-5, AC-6, AC-7

## Implementation Content

Create the new file `.github/workflows/release-standards.yaml`. It mirrors TASK-001's gate/commit/tag pattern **inline** and adds the standards-specific regeneration and skip-if-unchanged behavior.

1. **Trigger**: `workflow_dispatch` with a **required** input `semver`.
2. **Top-level `permissions: contents: write`**.
3. **Validation, duplicated inline from `release-indexer.yaml`** and running before all other steps:
   - Checkout with tags available (`fetch-depth: 0` or `fetch-tags: true`).
   - Anchored semver-format check per §6.2 (`^[0-9]+\.[0-9]+\.[0-9]+$`); on failure print a message naming the input as not a valid semver and exit non-zero.
   - Tag non-existence check for **`v{semver}-standards`**; on failure print a message naming the version as already existing and exit non-zero.
   - Messages must match TASK-001's **modulo the tag suffix**. Pass `${{ inputs.semver }}` via `env:` and use it as a shell variable — never interpolate into the `run:` body.
4. **Single `ubuntu-latest` job** (Linux AMD64 runner per REQ-3):
   - From the repository root, run exactly `bin/indexer-linux-amd64 --source . --output standards-tree.yaml` (REQ-3.1). Exit 0 = validated and written; exit 1 = anything else — a non-zero exit must fail the job.
   - Configure the bot identity per §6.3 exactly (ruling 1 supersedes REQ-3.2's `ci-bot`):
     `git config --global user.name "github-actions[bot]"` and
     `git config --global user.email "github-actions[bot]@users.noreply.github.com"`.
   - **Skip-if-unchanged** (§7 / AC-7): if `standards-tree.yaml` is byte-identical to the committed version, skip the commit and push without failing the job (`git commit` on an empty diff exits non-zero — guard with e.g. `git diff --quiet -- standards-tree.yaml`). Otherwise commit and push to `main`.
   - **Tag `v{semver}-standards` unconditionally** after the commit/push step — the tag step must **not** be conditioned on the commit having run. The tag points at the pushed release commit, or at current HEAD when the commit was legitimately skipped.
5. **Action pinning**: major version tags only (e.g. `actions/checkout@v5`); verify current majors at implementation time.

Explicitly **out of scope / waived — do not add**: any diagnostic for a missing `bin/indexer-linux-amd64` on a first run; any `concurrency:` key; any composite action, `.github/scripts/` helper, or reusable workflow. Ruling 6 requires the validation and commit/tag shell to stay **duplicated inline** in this file even though it mirrors TASK-001 — do not factor it out, and do not "improve" it into a shared unit.

## Target Files

- [x] .github/workflows/release-standards.yaml

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- /home/agent/workspace/.github/workflows/release-indexer.yaml (TASK-001 deliverable — the validation job, bot-identity config, and commit/tag steps to mirror inline; read-only, do not edit)
- /home/agent/workspace/indexer/main.go (CLI flag definitions `--source` / `--output` and the exit-code contract: 0 = written, 1 = anything else)
- /home/agent/workspace/.pre-commit-config.yaml (`check-yaml`, `end-of-file-fixer`, `mixed-line-ending` hooks the new YAML file must pass)

## Investigation Notes

- `.github/workflows/release-indexer.yaml` (TASK-001): gate lives in a `validate`
  job on `ubuntu-latest` whose first step is `actions/checkout@v7` with
  `fetch-depth: 0`; the two checks are steps `id: format` and `id: tag`, each
  passing `SEMVER: ${{ inputs.semver }}` via `env:`. Messages are
  `Error: '$SEMVER' is not a valid semver (expected major.minor.patch)` and
  `Error: version $SEMVER already exists (tag $TAG)`, both to stderr with
  `exit 1`, plus a success line each. The release job checks out with
  `ref: main`, configures the bot identity globally, commits and pushes, then
  tags. Actions in that file are pinned at `@v7` (`checkout`, `setup-go`,
  `upload-artifact`) and `@v8` (`download-artifact`) — major tags only.
- `indexer/main.go`: flags are exactly `--source`, `--output` and optional
  `--prefix`; `exitSuccess = 0` on a validated corpus that was written and
  `exitFailure = 1` for every other outcome, so a bare `run:` of the binary
  fails the job on any problem.
- `.pre-commit-config.yaml`: `check-yaml`, `end-of-file-fixer`,
  `mixed-line-ending --fix=lf`, `trailing-whitespace` apply to the new file
  (markdownlint excludes it by extension). The workflow is written with LF
  endings, a single trailing newline and no trailing whitespace.
- Environment: this container has no `python3`, no `yq` and no perl YAML module,
  so the "parses as YAML" check runs through the scratchpad YAML-subset parser
  (`yamlmini.pl`) written for TASK-001 rather than through `check-yaml`. Snippet
  extraction, local execution against throwaway git repositories (with a bare
  `origin` so the pushes are exercised) and the literal assertions live in
  `<scratchpad>/verify-standards.sh`.
- The regenerated tree was reproduced locally with a host-native build of the
  indexer: `<scratchpad>/indexer-host --source . --output <scratchpad>/standards-tree.yaml`
  from the repository root exits 0 and produces a file byte-identical to the
  committed `standards-tree.yaml` — i.e. the skip-if-unchanged path is the one a
  release of the current corpus actually takes. No repository file was written.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-001 | Indexer Release Workflow (`release-indexer.yaml`) | blocks | The inline validation and commit/tag shell authored in `.github/workflows/release-indexer.yaml`, mirrored here verbatim modulo the tag suffix |

## Implementation Steps (TDD: Red-Green-Refactor)

Workflows cannot be executed here. The TDD cycle runs against locally extracted
shell snippets, driven by scripts written to the session scratchpad. Never write
reproduction output into the repository — `standards-tree.yaml` must not change.

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review TASK-001's `release-indexer.yaml` and note the exact validation/commit/tag shell to mirror
- [x] Write a scratchpad verification script asserting the behaviors below, and confirm it fails while `.github/workflows/release-standards.yaml` does not exist:
  - the extracted semver-format snippet rejects `0.2`, `v0.2.0`, `0.2.0-rc1` (non-zero exit, message naming the input as invalid) and accepts `0.2.0`
  - the extracted tag-existence snippet, in a scratch git repo where `v0.2.0-standards` exists, exits non-zero and prints a message naming version `0.2.0` as existing; exits zero when absent
  - the extracted skip-if-unchanged snippet, run against an unchanged `standards-tree.yaml` in a scratch repo, exits zero without committing; and commits when the file differs
  - the workflow file parses as YAML
- [x] Run the script and confirm failure

### 2. Green Phase

- [x] Author `.github/workflows/release-standards.yaml` as specified in Implementation Content
- [x] Re-run the scratchpad script against the snippets extracted from the authored file and confirm it passes
- [x] Reproduce the indexer invocation locally: build a host-native indexer into the scratchpad (`cd /home/agent/workspace/indexer && go build -o <scratch>/indexer .`), then from `/home/agent/workspace` run `<scratch>/indexer --source . --output <scratch>/standards-tree.yaml` and verify exit 0

### 3. Refactor Phase

- [x] Diff every literal against the contracts: `bin/indexer-linux-amd64 --source . --output standards-tree.yaml`, tag literal `v${SEMVER}-standards`, branch `main`, bot name/email, `permissions: contents: write`, `runs-on: ubuntu-latest`, major-tag action pins
- [x] Verify the inline duplication is faithful to TASK-001 (same structure and messages modulo the tag suffix) and that no shared abstraction was introduced
- [x] Re-run the scratchpad script

## Completion Criteria

- [x] `.github/workflows/release-standards.yaml` exists, parses as valid YAML, has LF line endings and a trailing newline
- [x] `workflow_dispatch` declares `semver` as a required input; `permissions: contents: write` is set at workflow level
- [x] The extracted semver-format snippet exits non-zero and prints a message naming `0.2` as not a valid semver (AC-6); `v0.2.0` and `0.2.0-rc1` are also rejected
- [x] The extracted tag-existence snippet exits non-zero and prints a message naming version `0.2.0` as already existing when `v0.2.0-standards` is present (AC-5), and checkout fetches tags before that check
- [x] Validation runs before every other step, so no regeneration, commit, push, or tag can occur on an invalid or already-used semver
- [x] The job runs on `ubuntu-latest` and invokes exactly `bin/indexer-linux-amd64 --source . --output standards-tree.yaml` from the repository root, failing the job on non-zero exit (AC-4)
- [x] The extracted skip-if-unchanged snippet exits zero and performs no commit when `standards-tree.yaml` is unchanged, and commits when it differs (AC-7)
- [x] The tag step is not conditioned on the commit step having run, so `v{semver}-standards` is created in both the changed and unchanged cases (AC-4, AC-7)
- [x] Commit/push targets `main` with `user.name "github-actions[bot]"` and `user.email "github-actions[bot]@users.noreply.github.com"` (ruling 1 — not `ci-bot`)
- [x] `${{ inputs.semver }}` is never interpolated into a `run:` body — it is passed via `env:` and used as a shell variable
- [x] Validation and commit/tag shell is inline in this file; no composite action, `.github/scripts/` helper, or reusable workflow exists; no `concurrency:` key; no missing-binary diagnostic
- [x] All added tests pass

## Notes

- Impact scope: `.github/workflows/` only. The file cannot join the standards tree (dot-prefixed directories are skipped), so no `standards-tree.yaml` or `indexer/integration_test.go` snapshot changes are needed.
- Scope boundary: do not modify `.github/workflows/release-indexer.yaml` (TASK-001 deliverable — read only), `standards-tree.yaml`, `bin/`, the root `Makefile` (ruling 3), `README.md`, `.pre-commit-config.yaml` (TASK-003 owns it), or any standards document / indexer source file.
- Local reproduction outputs go to the session scratchpad only.
