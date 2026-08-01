# Task: Security remediation for the WP-002 release workflows

Work Plan ID: WP-002
Task ID: TASK-SEC-REMEDIATION
Created Date: 2026-08-02
Description: Harden the two manually-triggered release workflows: scope `permissions` per job, stop persisting the write-capable `GITHUB_TOKEN` in jobs that do not push, pin release builds to `main`, and constrain artifact download to the expected artifact names.

## Implementation Content

Five verified findings from the security review of the WP-002 changeset. Each entry below states
the file, line, vulnerability class, attack scenario, and the required fix. No behavioural change
to the release contract (validate -> build -> commit -> tag) is intended; the fixes narrow the
blast radius of the release credentials and pin the released artifact to reviewed code.

### SEC-1 (high) — Release builds are taken from the dispatch ref, not `main`

- File: `.github/workflows/release-indexer.yaml`, lines 65-66 (also 93-94 and 121-124)
- Vulnerability class: authnz (privilege escalation / branch-protection bypass)
- Detail: the three build jobs check out the ref the workflow was dispatched from
  (`actions/checkout@v7` with no `ref:`), while the release job checks out `ref: main`
  (line 156) and commits the resulting binaries there.
- Attack scenario: a collaborator with plain write access (enough to dispatch a
  `workflow_dispatch` workflow on their own branch, not enough to merge to `main`) pushes a
  branch containing a backdoored `indexer/`, dispatches "Release Indexer" against that branch,
  and the pipeline compiles the backdoored source and commits it to `bin/` on `main` under the
  `github-actions[bot]` identity, bypassing pull-request review. The poisoned binary is then
  executed with `contents: write` by `.github/workflows/release-standards.yaml` line 79 on the
  next standards release.
- Required fix: pin the build jobs' checkout to the same ref the release job publishes —
  add `with: { ref: main }` to the checkout in `build-linux-amd64`, `build-linux-arm64` and
  `build-darwin-arm64` — or add an explicit gate in the `validate` job that fails the run when
  `github.ref != 'refs/heads/main'`. Apply the same gate to `.github/workflows/release-standards.yaml`.

### SEC-2 (medium) — Workflow-level `contents: write` grants build jobs a push-capable token

- File: `.github/workflows/release-indexer.yaml`, line 19
- Vulnerability class: authnz (least privilege)
- Detail: `permissions: contents: write` is declared at workflow scope, so the `validate` job
  and all three build jobs receive a `GITHUB_TOKEN` that can push commits and tags, although
  only the `release` job pushes.
- Attack scenario: the build jobs execute repository-controlled code and the Go module graph
  (`go build` in `indexer/`). Any compromise there — a malicious module, a poisoned
  `go.mod`/toolchain directive, or the SEC-1 branch path — runs on a runner that holds a token
  able to write to `main` and create tags, turning a build-time compromise into a repository-write
  compromise.
- Required fix: set `permissions: {}` at workflow level and declare `permissions: contents: write`
  only on the `release` job.

### SEC-3 (low) — Same over-broad `permissions` scope in the standards workflow

- File: `.github/workflows/release-standards.yaml`, line 20
- Vulnerability class: authnz (least privilege)
- Detail: `contents: write` is workflow-scoped, so the `validate` job holds a push-capable token
  it never uses.
- Attack scenario: an exploit of anything running in `validate` (checkout of full history plus
  the version checks) has an unnecessary write-capable token available to exfiltrate or use.
- Required fix: set `permissions: {}` at workflow level and `permissions: contents: write` on the
  `release` job only.

### SEC-4 (medium) — `actions/checkout` persists the write-capable token in non-pushing jobs

- File: `.github/workflows/release-indexer.yaml`, lines 31, 66, 94, 124; `.github/workflows/release-standards.yaml`, line 36
- Vulnerability class: secrets (credential persistence)
- Detail: `actions/checkout` defaults to `persist-credentials: true`, writing the `GITHUB_TOKEN`
  into `.git/config` as an `http.extraheader` on disk for the whole job. Only the `release` jobs
  need pushed credentials.
- Attack scenario: any code executed later in a build job (`go build`, a Go toolchain download, a
  malicious module) can read `.git/config` and exfiltrate a token that can push to `main` and
  create release tags. The same exposure exists in both `validate` jobs, which only need read
  access to refs and tags.
- Required fix: add `persist-credentials: false` to every `actions/checkout` step except the
  `release` jobs' checkout, which pushes.

### SEC-5 (low) — Unfiltered artifact download into the commit path

- File: `.github/workflows/release-indexer.yaml`, lines 158-162
- Vulnerability class: injection (artifact/supply-chain poisoning)
- Detail: `actions/download-artifact@v8` is used with `merge-multiple: true`, `path: bin` and no
  `pattern:`, so it retrieves every artifact produced anywhere in the run and flattens it into
  `bin/`, which is then `chmod +x`-ed and `git add`-ed by name glob.
- Attack scenario: any future or injected job in the same run that uploads an artifact whose name
  starts with `indexer-` (e.g. a debug or test job) has its file flattened into `bin/`, made
  executable and committed to `main` as a released binary; the last writer of a colliding filename
  wins silently under `merge-multiple`.
- Required fix: add `pattern: indexer-*` to the `download-artifact` step so only the three expected
  build artifacts are retrieved, and assert the expected file set (exactly the three names) before
  `chmod +x` / `git add`.

## Target Files

- [x] `.github/workflows/release-indexer.yaml`
- [x] `.github/workflows/release-standards.yaml`

## Investigation Targets

- `.github/workflows/release-indexer.yaml` (job-level `permissions`, every `actions/checkout` step, the `download-artifact` step)
- `.github/workflows/release-standards.yaml` (job-level `permissions`, `actions/checkout` in `validate`)
- `docs/plans/tasks/WP-002/TASK-001.md`, `TASK-002.md` (release contract the fixes must preserve)

## Change Category

`Change Category: boundary-change`

## Investigation Notes

- Both workflows confirmed at the cited lines: workflow-level `permissions: contents: write`,
  build-job checkouts with no `ref:`/`persist-credentials:`, and `download-artifact` with
  `merge-multiple: true` and no `pattern:`.
- SEC-1 was fixed with **both** offered remedies, and deliberately so: `ref: main` on the three
  build checkouts stops a non-`main` dispatch from ever having its source compiled, and the new
  `Check release ref is main` gate (first step of each `validate` job, before the checkout) makes a
  non-`main` dispatch fail loudly rather than silently releasing `main`'s code. The gate is
  duplicated verbatim in both workflows (ruling 6 keeps the gate shell inline).
- `github.ref` is passed through `env: REF:` and used as a shell variable, matching the existing
  injection-safety convention for `inputs.semver`.
- SEC-5: `bin/` on `main` also contains `bin/.gitignore` and, after the first release, the previous
  binaries, so the assertion is over the `bin/indexer-*` file set (exactly the three released
  names) rather than over the whole directory. `grep` is guarded with `|| true` so the explicit
  error message survives the runner's `bash -e`.
- Environment: this container has no `pre-commit`, `python3`, `go` or `node`, so `check-yaml`,
  `actionlint` and any workflow execution are unrunnable. Verification is static, through the
  scratchpad harnesses reused from TASK-001/TASK-002
  (`verify.sh`, `verify-standards.sh`, `yamlmini.pl`, `extract.pl`), extended for this task with:
  workflow-level `permissions == {}` plus `contents: write` on the `release` jobs only and absent
  on every other job; `ref: main` on all three build checkouts;
  `persist-credentials: false` on every checkout except the two `release` jobs';
  the extracted ref gate executed against `refs/heads/attacker`, `refs/heads/main-2`,
  `refs/tags/v0.1.0` and `""` (all rejected) and `refs/heads/main` (accepted);
  `pattern: indexer-*` on `download-artifact`; and the extracted file-set assertion executed
  against a good `bin/`, an extra `indexer-backdoor` artifact, a missing binary and an empty
  `bin/` (only the good case exits 0). Both harnesses report `RESULT: SUCCESS`, including the
  pre-existing TASK-001/TASK-002 contract checks (gate messages, needs graph, build matrix,
  commit/tag ordering, file hygiene, major-tag pins), so the job graph and release contract are
  unchanged.
- Standards: `/etc/coding-standards/standards-tree.yaml` still has no CI/CD node; only root
  `GENERAL.md` (`scope: ['*']`) applies. GEN-001/GEN-002 shaped the fixes — no composite action,
  no shared helper, no new `concurrency:` key; each fix is a local key or a short inline gate.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-001 | Indexer release workflow | blocks | `.github/workflows/release-indexer.yaml` |
| TASK-002 | Standards release workflow | blocks | `.github/workflows/release-standards.yaml` |

## Remediation Context

- Source: security-reviewer
- Finding / failing command: SEC-1..SEC-5 above, from the WP-002 changeset security review
- Evidence: `permissions: contents: write` at `.github/workflows/release-indexer.yaml:19` and `.github/workflows/release-standards.yaml:20`; build-job checkouts without `ref:` at `release-indexer.yaml:66,94,124` versus `ref: main` at line 156; no `persist-credentials: false` on any checkout; `merge-multiple: true` without `pattern:` at `release-indexer.yaml:159-162`
- Verification: re-read both workflows and confirm — workflow-level `permissions: {}` with `contents: write` only on the `release` jobs; `persist-credentials: false` on all non-pushing checkouts; build jobs pinned to `main` (or a `validate` gate rejecting non-`main` dispatch refs); `download-artifact` restricted by `pattern:` with an explicit file-set assertion. The workflows must remain parseable and the job graph unchanged (this container cannot run GitHub Actions; verify statically as in TASK-001/TASK-002).

## Implementation Steps (TDD: Red-Green-Refactor)

Remediation without a testable behavior change: reproduce each finding by inspection against the
cited line, apply the fix, then re-run the static verification described above until every clause
holds.

## Completion Criteria

- [x] SEC-1: release builds cannot be produced from a non-`main` dispatch ref
- [x] SEC-2 / SEC-3: only the `release` jobs receive `contents: write`
- [x] SEC-4: no non-pushing job leaves a `GITHUB_TOKEN` in `.git/config`
- [x] SEC-5: only the three expected build artifacts can reach `bin/` before the commit
- [x] (Remediation tasks only) Verification checks from Remediation Context pass

## Notes

- Impact scope: both release workflows only; no change to the indexer, the corpus, or `.pre-commit-config.yaml`.
- Scope boundary: do not alter the validation/commit/tag shell duplication (user ruling), the MAJOR-version action pins (user ruling), or the waived `concurrency` guard / missing-binary diagnostic / skip-if-unchanged items.
