# Task: Module scaffolding and tooling

Work Plan ID: WP-001
Task ID: TASK-001
Created Date: 2026-07-31
Description: Pin the Go module, dependencies and toolchain for the `indexer` CLI — `go.mod`/`go.sum`, `Makefile` (build/test/lint/fmt), `.golangci.yml` with a `depguard` deny rule for the archived `gopkg.in/yaml.v3` — and prove `make lint` and `make test` run green on the bare scaffold.
Acceptance Criteria Covered: none (foundation task)

## Implementation Content

Establish the build/test/lint gate that every later task in WP-001 depends on.

1. `indexer/go.mod`: module `github.com/PSauerborn/standards/indexer`, `go` directive >= 1.25 (GO-001). Add `go.yaml.in/yaml/v3` as a runtime dependency and `stretchr/testify` as a test-only dependency, both at pinned versions, with a committed `indexer/go.sum`.
2. Record in `indexer/go.mod` comments that `gopkg.in/yaml.v3` is a **test-only transitive requirement of testify**, never imported by indexer code. It WILL appear in `go.sum`; that is expected and is not a dependency-constraint violation.
3. `indexer/.golangci.yml` (GO-004): enable a sensible linter set plus a `depguard` rule that DENIES `gopkg.in/yaml.v3` with the message that `go.yaml.in/yaml/v3` must be used instead.
4. `indexer/Makefile` (GEN-004, MAKE-001) with `.PHONY` targets:
   - `build` — builds the binary named `indexer` (`go build -o indexer .`)
   - `test` — `go test ./...`
   - `lint` — `golangci-lint run` (binary at `/Users/pascal/go/bin/golangci-lint`)
   - `fmt` — `gofmt -w .` (GO-003), and a way to verify a clean diff (e.g. `gofmt -l .` producing no output)
5. Verify `go list -deps .` for the binary does NOT include `gopkg.in/yaml.v3`.
6. `indexer/.gitignore` ignoring the built `indexer` binary so `make build` never dirties the tree.

Toolchain note (RISK-022): the local toolchain is go1.26.5 darwin/arm64. If `golangci-lint` cannot parse a go1.26 module, pin the `go` directive in `go.mod` down to the highest version golangci-lint supports that still satisfies GO-001 (>= 1.25), and record why in a `go.mod` comment.

Placeholder note: a module with no `.go` files may make `go build ./...`/`golangci-lint` refuse to run. If so, add a minimal placeholder `indexer/main.go` containing `package main` and an empty `func main() {}` — TASK-009 replaces it with the real entrypoint. Do not put any logic in it.

## Target Files

- [x] indexer/go.mod
- [x] indexer/go.sum
- [x] indexer/Makefile
- [x] indexer/.golangci.yml
- [x] indexer/.gitignore
- [x] indexer/main.go (placeholder only, and only if the toolchain requires at least one Go file)

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- indexer/go.mod (current content — module path and `go` directive already present)
- /Users/pascal/Github/psauerborn/standards/Makefile (repo Makefile conventions: `.PHONY` per target, `@echo` progress lines)
- /Users/pascal/Github/psauerborn/standards/general/MAKEFILES.md (`MAKE-001` — what a Makefile must automate)

## Investigation Notes

- `indexer/go.mod` already exists with module `github.com/PSauerborn/standards/indexer` and `go 1.26.5`. No `require` block, no `go.sum`, and `indexer/` contains no other files.
- Repo `Makefile` conventions: one `.PHONY: <target>` line immediately above each target, `@echo "..."` progress lines, commands prefixed with `@`, multi-line commands escaped with `\`.
- `general/MAKEFILES.md` `MAKE-001` (SHOULD): Makefiles automate tasks inherent to the repository — linting, formatting, unittests and building of application source.
- `GENERAL.md` `GEN-004` (SHOULD): Makefiles define common build/test/lint/format targets. `GEN-001`/`GEN-002`: KISS/YAGNI, minimise entropy — keep the scaffold minimal.
- `golang/GENERAL.md`: `GO-001` Go >= 1.25, `GO-002` go modules, `GO-003` `gofmt`, `GO-004` `golangci-lint`, `GO-006` snake_case filenames, `GO-008` flat layout for a single binary, `GO-032` testify assertions for unittests.

Implementation observations:

- RISK-022 materialised. `golangci-lint` is v1.64.8, built with go1.25.0, and refused the module with `can't load config: the Go language version (go1.25) used to build golangci-lint is lower than the targeted Go version (1.26.5)`. The `go` directive is therefore pinned to `1.25.0` (still satisfies GO-001) with the reason recorded as a `go.mod` comment. `make lint` passes at that directive.
- The toolchain does require at least one `.go` file, so the permitted placeholder `indexer/main.go` (`package main`, empty `func main`) was added.
- **`go mod tidy` was deliberately NOT used.** On the bare scaffold nothing imports either dependency, so `tidy` strips both `require` lines (verified on a throwaway copy) and would drop the explanatory comments with them, breaking the completion criterion that `go.mod` declares both dependencies. `go get` was used instead, which pins both modules and writes `go.sum`. Both are currently marked `// indirect` because no source file imports them yet; the first task that adds real imports and tests will clear those markers. Later tasks re-running `go mod tidy` must preserve the `go.mod` comments.
- `gopkg.in/yaml.v3` is absent from `go.sum` on the bare scaffold — with module-graph pruning it is only recorded once testify is actually imported by a test. It remains a test-only transitive requirement either way.
- The `depguard` deny rule was proved to fire: temporarily importing `gopkg.in/yaml.v3` in `main.go` produced `import 'gopkg.in/yaml.v3' is not allowed from list 'archived-yaml'` and failed `make lint`; the probe was reverted.
- `issues.exclude-use-default: false` is set so revive's missing-doc-comment reports are not filtered out by golangci-lint's default exclusion list (GO-005).
- Makefile resolves the linter as `$(shell go env GOPATH)/bin/golangci-lint` (overridable via `GOLANGCI_LINT`), which resolves to the expected `/Users/pascal/go/bin/golangci-lint` locally without hardcoding a user-specific path.
- Formatting verification is split into `fmt` (`gofmt -w .`) and `fmt-check` (`gofmt -l .`, non-empty output fails), since a single target cannot both rewrite and assert a clean diff.

## Task Dependencies

(None — this is the first task of the work plan.)

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Run `make lint` and `make test` in `indexer/` and confirm they fail (targets missing / dependencies unresolved) — this is the failing gate this task closes
- [x] Confirm `gopkg.in/yaml.v3` is currently absent from `go.sum`

### 2. Green Phase

- [x] Write `go.mod`, resolve and pin dependencies (`go mod tidy`), commit `go.sum`
- [x] Write `Makefile`, `.golangci.yml` (with the `depguard` deny rule), `.gitignore`
- [x] Run `make lint`, `make test`, `make fmt` in `indexer/` and confirm all pass

### 3. Refactor Phase

- [x] Tidy target definitions and lint configuration (keep the gate green)
- [x] Re-run `make lint` and `make test` and confirm still green

## Completion Criteria

- [x] `make lint`, `make test` and `make fmt` (no diff) all exit 0 in `indexer/` on the bare scaffold — this is the explicit exit condition for the whole work plan's foundation (RISK-022)
- [x] `indexer/go.mod` declares module `github.com/PSauerborn/standards/indexer`, `go` >= 1.25, `go.yaml.in/yaml/v3` and `stretchr/testify`, with a committed `go.sum`
- [x] `indexer/.golangci.yml` contains a `depguard` deny rule for `gopkg.in/yaml.v3`
- [x] `go list -deps .` output for the binary does not contain `gopkg.in/yaml.v3`, and a `go.mod` comment records its test-only transitive status
- [x] `make build` produces a binary named `indexer` (not `stdidx`), and the binary is gitignored
- [x] All added tests pass

## Notes

- Impact scope: every later WP-001 task runs behind these Makefile targets; a wrong `go` directive or lint config stalls the whole plan.
- Scope boundary:
  - Flat `package main` directly in `indexer/` (GO-008) — never create `cmd/`, `internal/`, or any subpackage. Filenames snake_case (GO-006).
  - Permitted dependencies are exactly `go.yaml.in/yaml/v3` (runtime) and `stretchr/testify` (test-only). `gopkg.in/yaml.v3` is archived and MUST NOT be imported by any file under `indexer/`.
  - The standards corpus (18 documents, 28 example files) and `docs/standards-tree.yaml` are READ-ONLY. Do not modify anything outside `indexer/`.
  - The tool was formerly called `stdidx`; that name must not appear anywhere in `indexer/`. The binary is `indexer`, invoked as `indexer --source <dir> --output <file> [--prefix <p>]`.
  - Fixture corpora (added by later tasks) will live under `indexer/tests/data/.corpora/<case>/`; do not add any `tests/` content in this task.
