# Task: Code Review Remediation (iteration 2) — WP-002 Release Workflows

Work Plan ID: WP-002
Task ID: TASK-CODE-REVIEW-REMEDIATION
Created Date: 2026-08-02
Description: Pin the three `release-indexer.yaml` build jobs to the single immutable dispatch commit instead of the moving `main` tip, and abort the release job if `main` advanced past that commit, so a release never ships binaries built from more than one revision.
Acceptance Criteria Covered: AC-1

## Implementation Content

Iteration 1's findings CR-1 through CR-4 are re-verified as fixed against the
current file contents and must not be re-applied. This task carries a single new
finding introduced by the CR-1 remediation.

### Finding 5 — `ref: main` resolves independently in four jobs, so a release can mix revisions

`.github/workflows/release-indexer.yaml:89`, `:124`, `:160` (the three build
jobs) and `:197` (the release job) each check out `ref: main`. A branch name is
resolved separately by each job, at the moment that job starts. Before the CR-1
remediation the build jobs used the default checkout ref, which for
`workflow_dispatch` is `github.sha` — one immutable commit shared by every job —
so the fix traded a cross-job-consistent revision for a moving target.

Failure scenario: a maintainer dispatches Release Indexer for `0.3.0` from
`main` at commit `A`. `build-linux-amd64` and `build-linux-arm64` start
immediately and compile `A`. `build-darwin-arm64` queues behind macOS runner
availability (routinely minutes, sometimes longer). Meanwhile a pull request
merges to `main`, producing commit `B` that changes `indexer/`. The darwin job
then checks out `B` and ships a binary built from different source than the two
linux binaries. The release job checks out `B` too, commits all three binaries
and tags `v0.3.0-indexer` at a tree whose `indexer/` source produced only one of
the three shipped executables. Nothing in the workflow detects this: the
"Verify downloaded release binaries" step only asserts the file *names*. The
released tag then attests to binaries it did not produce, which is exactly the
guarantee the CR-1 fix was meant to establish.

Note the release job is not affected by the same danger in
`release-standards.yaml`: there the checkout, regeneration, commit and tag all
happen in one job against one checkout, so the committed tree always matches the
corpus it was generated from. Leave that file unchanged.

Required fix, in `.github/workflows/release-indexer.yaml` only:

1. In `build-linux-amd64`, `build-linux-arm64` and `build-darwin-arm64`, replace
   `ref: main` with `ref: ${{ github.sha }}`. For a `workflow_dispatch` this is
   the tip of the dispatched ref at dispatch time, and the "Check release ref is
   main" gate in `validate` already guarantees that ref is `main` — so the
   commit is both a `main` commit and identical across the three jobs. Update the
   `# ref: main pins the build ...` comment above each checkout to say the build
   is pinned to the dispatched commit, and why that commit is known to be on
   `main`.

2. Keep the release job's checkout at `ref: main` — it needs a local `main`
   branch to fast-forward and push, which a detached SHA checkout does not give
   it — and add a step immediately after that checkout which fails the job when
   `main` has moved:

   ```yaml
         # The build jobs compiled ${{ github.sha }}. If main advanced while they
         # ran, committing their output here would tag a tree that did not
         # produce the released binaries, so the release aborts and asks for a
         # re-dispatch rather than shipping a mixed revision.
         - name: Check main is still the released commit
           env:
             RELEASE_SHA: ${{ github.sha }}
           run: |
             HEAD_SHA="$(git rev-parse HEAD)"
             if [[ "$HEAD_SHA" != "$RELEASE_SHA" ]]; then
               echo "Error: main advanced from $RELEASE_SHA to $HEAD_SHA during the release; re-dispatch the workflow" >&2
               exit 1
             fi
             echo "main is still at the released commit $RELEASE_SHA"
   ```

   Place it before "Download release binaries" so the run stops before any
   artifact is written into `bin/`.

No other change: do not touch `release-standards.yaml`, the validate jobs, the
`permissions:` blocks, the `persist-credentials:` keys, the artifact
`pattern:`/verification step, or the action pins.

## Target Files

- [x] `/home/agent/workspace/.github/workflows/release-indexer.yaml`

## Investigation Targets

- `/home/agent/workspace/.github/workflows/release-indexer.yaml` (the four `Check out ...` steps and the `release` job's step order)
- `/home/agent/workspace/docs/specs/SPEC-002.md` (REQ-2.1, REQ-2.3, REQ-2.4, AC-1)

## Change Category

`Change Category: bug-fix, regression`

## Remediation Context

- Source: code-reviewer
- Finding / failing command: finding 5 against the WP-002 changeset; GitHub Actions cannot run in this container, so the defect is established by inspection of the ref-resolution semantics rather than by a repository command.
- Evidence: `.github/workflows/release-indexer.yaml:89`, `:124`, `:160`, `:197` all read `ref: main`. `actions/checkout` resolves a branch name at job start, and the four jobs start at different times; `github.sha` is fixed for the whole run. `grep -c 'ref: main' .github/workflows/release-indexer.yaml` currently returns 4.
- Verification:
  - (a) `grep -c 'ref: ${{ github.sha }}' .github/workflows/release-indexer.yaml` returns 3, and `grep -c 'ref: main' .github/workflows/release-indexer.yaml` returns 1 (the release job).
  - (b) The release job contains a `Check main is still the released commit` step positioned after `Check out main` and before `Download release binaries`.
  - (c) Extract that step's `run:` body into a throwaway git repository and assert it exits 0 when `RELEASE_SHA` equals `git rev-parse HEAD` and exits 1 with the error message after an additional commit.
  - (d) `git diff -- .github/workflows/release-standards.yaml` is empty.

## Investigation Notes

- Confirmed the four checkouts before the fix: `build-linux-amd64` (:89),
  `build-linux-arm64` (:124), `build-darwin-arm64` (:160) and `release` (:197)
  all read `ref: main`; `grep -c 'ref: main'` returned 4. The three build jobs
  carry an identical four-line comment starting `# ref: main pins the build ...`.
- The `validate` job's first step, "Check release ref is main", rejects any
  dispatch whose `github.ref` is not `refs/heads/main`, and every other job
  `needs: validate` directly. So for any run that reaches a build job,
  `github.sha` is the tip of `main` as of dispatch — a `main` commit — which
  keeps the SEC-1/CR-1 protection intact while making the revision immutable
  and shared across the three build jobs.
- The `release` job checks out with credentials persisted (it is the only job
  with `contents: write`) and later runs `git push origin main`, so it needs a
  real local `main` branch; a detached `${{ github.sha }}` checkout would break
  the push. Hence the guard step rather than a ref change there.
- Step order in `release` is: Check out main -> Download release binaries ->
  Verify downloaded release binaries -> Restore executable permissions ->
  Configure git identity -> Commit and push -> Tag release. The new guard goes
  between the first two so nothing is written into `bin/` on a stale run.
- Standards: only `GENERAL.md` (scope `*`) applies; no CI/workflow-specific
  node exists in the tree. GEN-001/GEN-002 favour the minimal guard step over a
  larger redesign (no `concurrency:` block, no composite action).
- `release-standards.yaml` performs checkout, regeneration, commit and tag in a
  single job against a single checkout, so it has no cross-job revision skew and
  is left untouched.

Verification results (GitHub Actions cannot run here; the workflow is checked by
extracting its `run:` bodies and by parsing the YAML):

- (a) `grep -c 'ref: ${{ github.sha }}'` returns 3 and `grep -c 'ref: main'`
  returns 1 on `release-indexer.yaml`. The prose in the new comments deliberately
  avoids the literal strings `ref: main` / `ref: ${{ github.sha }}` so these
  counts stay assertions about the YAML and not about the commentary.
- (b) The release job's step order is now `Check out main` ->
  `Check main is still the released commit` -> `Download release binaries`.
- (c) The guard's `run:` body, extracted with `scratchpad/extract.pl`, exits 0
  with `main is still at the released commit <sha>` when `RELEASE_SHA` equals
  `git rev-parse HEAD` in a throwaway repository, and exits 1 with
  `Error: main advanced from <sha> to <sha2> during the release; re-dispatch the
  workflow` after one further commit.
- (d) `release-standards.yaml` is byte-unchanged (untouched mtime; `.github/` is
  still untracked as a whole, so `git diff` on it is trivially empty).
- All three scratchpad harnesses (`verify.sh`, `verify-standards.sh`,
  `verify-remediation.sh`) report `RESULT: SUCCESS`. Two stale assertions in
  them were updated for this change: the SEC-1 build-ref expectation, and the
  `download-artifact` pattern lookup, which used a hard-coded step index that the
  new guard step shifted — it now locates the step by its action.
- `pre-commit`, `go`, `python3` and `node` are absent from this container, so
  markdownlint/`check-yaml`/`go test` are unrunnable. Checked instead by
  inspection: no trailing whitespace or tabs, file ends in a newline, and the
  workflow still parses with `scratchpad/yamlmini.pl`.

## Implementation Steps

This remediation has no in-repo test suite to drive (workflows are not executable
here). Replace the TDD cycle with: reproduce the finding by inspection, apply the
fix, then run the Verification checks until all pass.

- [x] Read the four checkout steps and confirm the current `ref:` values
- [x] Apply fix 1 (three build jobs pinned to `${{ github.sha }}`, comments updated)
- [x] Apply fix 2 (release-job guard step)
- [x] Run Verification checks (a)–(d)

## Completion Criteria

- [x] All three build jobs check out the same immutable commit for a given run
- [x] The release job aborts with a clear message when `main` no longer points at the released commit
- [x] `release-standards.yaml` is unchanged
- [x] Verification checks (a)–(d) pass

## Notes

- Impact scope: `.github/workflows/release-indexer.yaml` only. No corpus change, so no tree regeneration and no `nodeCountSnapshot` / `exampleCountSnapshot` bump.
- Scope boundary: do not add a `concurrency:` guard, a missing-binary diagnostic, skip-if-unchanged handling for the indexer binary commit, a composite action, `.github/scripts/`, or a reusable workflow; do not change the `github-actions[bot]` identity, the `main` target branch, the `permissions:` model, or the action pins; do not touch the root `Makefile`, `release-standards.yaml`, or `indexer/`.
