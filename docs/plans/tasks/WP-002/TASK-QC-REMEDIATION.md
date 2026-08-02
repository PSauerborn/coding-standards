# Task: QC Remediation — Record the CI Build Deviation and Give the Release Build a Makefile Entry Point

Work Plan ID: WP-002
Task ID: TASK-QC-REMEDIATION
Created Date: 2026-08-02
Description: Remediate two SHOULD-level GENERAL.md violations found in the WP-002 changeset — an unrecorded GEN-003 deviation for the new CI build path, and the GEN-004 absence of a Makefile entry point for the release cross-builds.

## Implementation Content

Two independent violations, both anchored in `.github/workflows/release-indexer.yaml`:

1. **[GEN-003] `.github/workflows/release-indexer.yaml:68`** — the release
   cross-builds run on bare GitHub runners via `actions/setup-go`, a third
   non-containerized build path for the indexer. The repository's recorded GEN-003
   deviation (`indexer/README.md`, "Build environment"; `indexer/Makefile` header
   comment) covers only "the Makefile targets run against the host Go toolchain"
   and ends with "Revisit this if the module ever ships as a deployed artifact
   rather than as a checkout-local tool." The Release Indexer workflow makes the
   module exactly that, so the deviation must be revisited and re-stated to cover
   the CI release build. This is a documentation fix: do **not** containerize the
   CI build — GEN-003 is a SHOULD and the deviation is accepted, it just has to be
   noted at its current scope. Extend the "Build environment" section of
   `indexer/README.md` (and the corresponding `indexer/Makefile` header comment,
   which repeats the now-fired revisit trigger) to state that the release
   cross-builds in `.github/workflows/release-indexer.yaml` also run against a
   toolchain pinned by `go-version-file: indexer/go.mod` on ephemeral GitHub
   runners rather than in a container, and why that is acceptable.

2. **[GEN-004] `.github/workflows/release-indexer.yaml:73`, `:101`, `:130`** — the
   release build command is defined inline in the workflow and repeated three
   times, while `indexer/Makefile` exposes only the local `build` target. Add a
   parameterized release-build target to `indexer/Makefile` (for example
   `release-build` driven by `GOOS`/`GOARCH`/`CGO_ENABLED` variables with an
   `OUTPUT` path) and have the three `Build indexer` steps invoke it instead of
   calling `go build` directly. Keep the emitted binary names
   (`indexer-linux-amd64`, `indexer-linux-arm64`, `indexer-darwin-arm64`) and
   their location at the repository root, and keep `CGO_ENABLED=0`, exactly as
   they are today — `download-artifact`, `chmod +x bin/indexer-*`,
   `git add bin/indexer-*` and `bin/indexer-linux-amd64` in
   `release-standards.yaml:79` all depend on those names. If, after attempting it,
   a Makefile target measurably increases entropy relative to the three inline
   invocations, record a GEN-004 deviation with its reason in `indexer/README.md`
   instead — but do not leave the situation unaddressed and undocumented.

Out of scope by explicit user ruling — do not "fix" any of these: the duplicated
semver-validation and commit/tag shell across the two workflows (no composite
action, `.github/scripts/` or reusable workflow may be introduced); the absent
`concurrency` guard; the absent missing-binary diagnostic in the standards
workflow; skip-if-unchanged for the indexer binary commit; the root `Makefile`
`index` target's `./bin/indexer` path; major-version action pins; and the
`github-actions[bot]` commit identity.

## Scope Restriction (user decision, 2026-08-02)

The user explicitly restricted this task to finding 1 (GEN-003) only. Finding 2
(GEN-004) is **not** to be implemented: no release-build target is added to
`indexer/Makefile` and the three `Build indexer` steps in
`.github/workflows/release-indexer.yaml` are left untouched. The task file's own
escape hatch is taken instead — a GEN-004 deviation is recorded in
`indexer/README.md`, with the reason that this work was scoped to exclude
changes to the `indexer/` build tooling (`SPEC-002` §3.2 puts the indexer out of
scope). Per `GENERAL.md` meta rules, a user request outranks a **SHOULD**.

Items below marked *superseded by user decision* are deliberately not done.

## Target Files

- [x] `indexer/README.md` — "Build environment" section (GEN-003 deviation extended, GEN-004 deviation recorded)
- [x] `indexer/Makefile` — header deviation comment only (GEN-003)
- [ ] ~~`.github/workflows/release-indexer.yaml` — the three `Build indexer` steps (GEN-004)~~ — superseded by user decision; not modified

## Investigation Targets

- `docs/plans/quality/WP-002/WP-002-quality-report.md` — the findings being remediated
- `indexer/README.md` ("Build environment", lines 48-71) — the existing recorded GEN-003 deviation and its revisit trigger
- `indexer/Makefile` (lines 1-20) — the header deviation comment and the existing `build` target
- `.github/workflows/release-indexer.yaml` (the three build jobs, lines 60-143) — the inline cross-build commands and artifact names
- `.github/workflows/release-standards.yaml` (line 79) — consumes `bin/indexer-linux-amd64`, so binary naming is a contract
- `/etc/coding-standards/GENERAL.md` — GEN-003 and GEN-004 as written, plus the meta rules on deviating from a SHOULD

## Investigation Notes

- `indexer/README.md` "Build environment" (lines 48-71 before the change) scoped
  the GEN-003 deviation to "the Makefile targets run against the host Go
  toolchain" and closed with the now-fired revisit trigger. Confirmed.
- `indexer/Makefile` lines 1-13 repeated the same deviation and the same revisit
  trigger. Confirmed.
- `.github/workflows/release-indexer.yaml`: three build jobs
  (`build-linux-amd64`, `build-linux-arm64`, `build-darwin-arm64`), each
  `actions/checkout@v7` with `ref: main` and `persist-credentials: false`, then
  `actions/setup-go@v7` with `go-version-file: indexer/go.mod`, then
  `working-directory: indexer` + `CGO_ENABLED: "0"` + `GOOS`/`GOARCH` and
  `run: go build -o ../indexer-{os}-{arch} .` (lines 103, 138, 174 as the file
  now stands), then `upload-artifact@v7` with `if-no-files-found: error`. No
  container anywhere on that path. Confirmed.
- Binary-name contract confirmed: `indexer-linux-amd64`, `indexer-linux-arm64`,
  `indexer-darwin-arm64` at the repository root, consumed downstream by
  `chmod +x bin/indexer-*` / `git add bin/indexer-*` and by
  `release-standards.yaml`'s use of `bin/indexer-linux-amd64`. Unchanged by this
  task — no build command was touched.
- `GENERAL.md`: GEN-003 and GEN-004 are both **SHOULD**; the meta rules require a
  noted reason for a deviation and resolve "user request vs. SHOULD" in favour of
  the user, which is what the scope restriction above relies on.
- Environment: this container has no `go`, `python3`, `node` or `pre-commit`, so
  `make -C indexer build`, `make -C indexer test` and `pre-commit run
  --all-files` are **unrunnable here**. They are also unaffected by this change:
  the only edits are a markdown section and a Makefile comment block (lines 1-23,
  all `#`-prefixed, terminated by a blank line before `GOLANGCI_LINT ?=`), so no
  target, recipe or variable changed. Verified by inspection.

## Remediation Context

- Source: quality-controller
- Finding / failing command: `[GEN-003] .github/workflows/release-indexer.yaml:68` and `[GEN-004] .github/workflows/release-indexer.yaml:73` — see `docs/plans/quality/WP-002/WP-002-quality-report.md`
- Evidence: `indexer/Makefile:12-13` — "See indexer/README.md, 'Build environment'. Revisit if this module ever ships as a deployed artifact rather than a checkout-local tool." The new workflow commits built binaries to `bin/` on `main` and tags `v{semver}-indexer`, i.e. the module now ships as a deployed artifact, and neither the README nor the Makefile comment was revisited. Separately, `go build -o ../indexer-linux-amd64 .` and its two siblings appear only in the workflow, not in `indexer/Makefile`.
- Verification: `indexer/README.md` and `indexer/Makefile` state the GEN-003 deviation for the CI release build path; `grep -n 'go build' .github/workflows/release-indexer.yaml` returns no direct `go build` invocation (or a GEN-004 deviation is recorded in `indexer/README.md`); `make -C indexer build` and `make -C indexer test` still succeed where a Go toolchain is available.

For remediation without a testable behavior change (documentation and build-entry-point changes), replace the TDD cycle below with: reproduce the finding, apply the fix, re-run the Verification checks until they pass.

## Implementation Steps

### 1. Reproduce

- [x] Read all Investigation Targets and confirm both findings against the current file contents
- [x] Confirm the release binary names and paths that downstream steps depend on

### 2. Fix

- [x] GEN-003: extend the recorded deviation in `indexer/README.md` and the `indexer/Makefile` header to cover the CI release build path
- [ ] ~~GEN-004: add the parameterized release-build target to `indexer/Makefile` and call it from the three `Build indexer` steps~~ — superseded by user decision; the GEN-004 deviation is recorded in `indexer/README.md` instead

### 3. Verify

- [x] Run the Verification checks from Remediation Context (as restricted; see Investigation Notes for the checks that are unrunnable in this container)
- [x] Confirm the two workflow files still parse as YAML and the job graph is unchanged — neither workflow file was modified

## Completion Criteria

- [x] The GEN-003 deviation covering the CI release build path is recorded in `indexer/README.md` and `indexer/Makefile`
- [x] A GEN-004 deviation is recorded in `indexer/README.md` with its reason (the Makefile-target alternative is superseded by user decision)
- [x] Released binary names, output locations and `CGO_ENABLED=0` are unchanged — no build command was touched
- [x] Verification checks from Remediation Context pass, except the toolchain-dependent ones, which are unrunnable in this container and unaffected by a comment-only Makefile change

## Notes

- Impact scope: `indexer/` build documentation and Makefile; the three build jobs of the Release Indexer workflow. No change to the release job, the tagging steps, or `release-standards.yaml`.
- Scope boundary: `.pre-commit-config.yaml` and `CLAUDE.md` — reviewed and conformant, leave unchanged. `.github/workflows/release-standards.yaml` — no violation found; leave unchanged. Root `Makefile`, `Dockerfile.claude` — out of scope by user ruling / not in the changeset.
- Environment: this container has no `pre-commit`, `python3`, `go` or `node`; the `make -C indexer` verification steps must be run where a Go toolchain exists, or explicitly reported as unrunnable.
