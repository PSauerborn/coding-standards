# Quality Report: Release Pipelines for Indexer and Standards Tree

Work Plan ID: WP-002
Created Date: 2026-08-02
Last Updated: 2026-08-02 (review loop iteration 2)

## Coding Standards Review

Applicable Standards:

- GENERAL.md (scope `*`; topics `docker`, `makefiles`, `pre-commit`, `integration-tests`)
- general/MAKEFILES.md (`indexer/Makefile` entered the changeset in iteration 1 remediation)

Standards resolved from `/etc/coding-standards/standards-tree.yaml`. No other node
in the tree matches this changeset: it contains no `*.go`, `*.py`, `*.js`/`*.ts` or
`*.tf` source, no Dockerfile, no API surface and no database access, so
`general/DOCKER.md`, `general/API.md`, `general/LOGGING.md`, `golang/*`,
`python/*`, `javascript/*`, `terraform/*`, `databases/*` and
`testing/ACCEPTANCE.md` are out of context. The corpus contains no CI/CD or
GitHub Actions standard, so the workflow files are governed by `GENERAL.md` alone.

Files reviewed (manifest changeset only):

- `.github/workflows/release-indexer.yaml` (new; modified since iteration 1 by the
  security and code-review remediations)
- `.github/workflows/release-standards.yaml` (new; likewise modified)
- `.pre-commit-config.yaml`
- `CLAUDE.md`
- `indexer/README.md` (iteration-1 QC remediation)
- `indexer/Makefile` (iteration-1 QC remediation, header comment only)

`Dockerfile.claude` and the root `Makefile` are modified in the working tree but
are not in the manifest changeset, and were not reviewed.

### Coding Standards Violations

Total Violations: 0

| Standards File | Rule ID | File Path | Description |
|----------------|---------|-----------|-------------|
| — | — | — | No violations found in this iteration. |

### Iteration-1 Findings: Disposition

- **GEN-003 — remediated; fix verified as adequate.** `indexer/README.md:50-101`
  now opens the deviation with "Neither build path for this module is
  containerized", explicitly records that the module's own revisit trigger
  ("if the module ever ships as a deployed artifact") has fired, revisits it, and
  gives three reasons the deviation stands for the release path: the toolchain is
  pinned by the same `go` directive via
  `go-version-file: indexer/go.mod`, each job builds `CGO_ENABLED=0` on a fresh
  disposable runner so nothing from the runner image is linked in, and a
  release-only Dockerfile would add a third build path rather than replace one. It
  closes with a new, narrower revisit trigger (cgo, a system library or a code
  generator). `indexer/Makefile:12-24` carries the matching header note. The
  deviation is recorded with a reason, which is what the `GENERAL.md` meta rules
  require of a SHOULD.
- **GEN-004 — closed by user decision; deviation confirmed recorded with a
  reason.** `indexer/README.md:96-101` states plainly that the release cross-build
  command is inlined in `.github/workflows/release-indexer.yaml` rather than
  exposed as an `indexer/Makefile` target, names GEN-004 and its SHOULD level, and
  gives the reason (scope: `SPEC-002` §3.2 lists "Updating the indexer source
  code" as out of scope, verified at `docs/specs/SPEC-002.md:26-29`).
  `indexer/Makefile:22-24` cross-references it. The three inline `go build`
  invocations at `release-indexer.yaml:103`, `:138` and `:174` are covered by this
  recorded deviation and are not raised.

### Conformance Confirmed

New content added since iteration 1 (workflow-level `permissions: {}` with
`contents: write` scoped to the release jobs, the "Check release ref is main"
validate gate, `ref: main` on build checkouts, `persist-credentials: false` on
non-pushing checkouts, `pattern: indexer-*` plus the file-set assertion on the
download step, the leading-zero-rejecting semver regex, and the staged-diff
commit detection in the standards workflow) was reviewed and conforms:

- **GEN-001 / GEN-002 (KISS, YAGNI).** The job graph is still linear
  (validate → build×3 → release; validate → release) with no indirection added.
  Each new construct is a single declarative key or a short guarded shell block
  sited where it acts, and each carries an inline comment stating why it exists
  (`release-indexer.yaml:16-18`, `:27-29`, `:40-44`, `:57-59`, `:82-85`,
  `:199-201`, `:209-211`, `:222-224`, `:241`; `release-standards.yaml:17-20`,
  `:33-36`, `:108-110`, `:115-117`, `:126-128`). The download-pattern plus
  file-set assertion is two overlapping checks on the same property, but it is a
  four-line literal comparison annotated as belt-and-braces on a step that decides
  what gets committed to `main` — proportionate, not entropy.
- **GEN-003 (containerized build/test/run).** Covered by the recorded deviation
  above for both workflows. `release-standards.yaml:101` runs the committed
  `bin/indexer-linux-amd64` directly on the runner rather than in a container;
  that is the same execution the recorded rationale addresses ("the tool indexes
  the repository it lives in, so a containerized run would have to bind-mount the
  repository root back in to reach its own corpus"), so applying the rationale as
  `GENERAL.md` §1 directs, it is within the deviation and is not a separate
  finding.
- **GEN-004 / MAKE-001 (Makefile targets).** `indexer/Makefile` still exposes
  `build`, `test`, `lint`, `fmt` and `fmt-check`; the iteration-1 remediation
  changed comments only, so no target was added, renamed or removed. The one
  build command with no target is the release cross-build, closed above as a
  recorded deviation.
- **GEN-005 (pre-commit).** Unchanged since iteration 1 and still conformant: the
  GEN-005 minimum set (`check-yaml`, `end-of-file-fixer`, `trailing-whitespace`,
  `check-added-large-files`, `check-case-conflict`, `check-json`, and
  `detect-secrets` with `--baseline`) is all present. The `exclude: ^bin/` at
  `.pre-commit-config.yaml:18` narrows a SHOULD-level hook rather than removing
  it and carries an inline reason, which the meta rules accept for a SHOULD.
  Recorded as a scoped deviation, not a violation. It removes the size ceiling on
  `bin/` entirely rather than raising it.
- **GEN-006 (integration tests).** No regression: the workflows add no in-repo
  testable surface and `indexer/`'s integration suite is untouched.
- **File hygiene.** All six changeset files were re-checked directly for trailing
  whitespace, CRLF line endings, stray tabs and a missing or duplicated final
  newline. All clean, so `end-of-file-fixer`, `trailing-whitespace` and
  `mixed-line-ending` would pass on them.
- **Corpus-discovery invariants.** The workflows live under the dot-prefixed
  `.github/`, so they cannot become standards-tree nodes and require no tree
  regeneration or `nodeCountSnapshot` bump. Neither `CLAUDE.md` nor
  `indexer/README.md` gained `title:` frontmatter, so neither joins the tree.

### Checks That Could Not Be Run

The container has no `pre-commit`, `python3`, `go` or `node` (only `perl`), and
GitHub Actions workflows cannot be executed here at all. Consequently:

- `pre-commit run --all-files` was not run; `check-yaml` validity of the two
  workflow files and `markdownlint` conformance of the `CLAUDE.md` and
  `indexer/README.md` additions are asserted by inspection against
  `.markdownlint.yaml` (MD013 disabled, lists blank-line separated), not by
  execution.
- The `exclude: ^bin/` behaviour of `check-added-large-files` was not reproduced.
- The workflow job graph, shell bodies and cross-builds were not executed in this
  review; they are covered by the executors' scratchpad harnesses recorded in the
  manifest.
- `make build` / `make test` in `indexer/` were not run (no Go toolchain); the
  `indexer/Makefile` change is comment-only, so no target behaviour changed.
