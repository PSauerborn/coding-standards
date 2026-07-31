# Task: Data models and aggregate error types

Work Plan ID: WP-001
Task ID: TASK-002
Created Date: 2026-07-31
Description: Define the indexer's data models in `types.go` (Document, Frontmatter, Node, ExampleEntry, Tree, Config — including scope scalar-to-list normalization and order-preserving `aliases`) and the typed-error vocabulary plus collect-all aggregate error in `errors.go`.
Acceptance Criteria Covered: AC-3

## Implementation Content

Two cohesive source/test pairs that every later task consumes.

### `indexer/types.go` (GO-010)

- `Config` — the run configuration: `Source`, `Output`, `Prefix` (flag values, populated by TASK-009).
- `Frontmatter` — parsed standards-document frontmatter:
  - required: `title`, `description`, `scope`, `topics`
  - optional: `parent` (path relative to the source root), `examples` (paths relative to the *document's own directory*), `aliases`
  - **`scope` normalization**: frontmatter may carry `scope` as a YAML scalar (`scope: '*.go'`) or a list (`scope: ['*.js','*.ts','*.vue']`). Normalize a scalar to a one-element list at unmarshal time — this is NOT an error. `scope` is always a list downstream and in the output.
  - **`aliases` must preserve declaration order**. The output tree emits the root's 8-key alias map in frontmatter order — `go, js, ts, py, postgres, pg, deploy, containerization` — which is neither sorted nor Go-map order. Model aliases as an ordered slice of key/value pairs (e.g. `type Alias struct{ Key, Value string }` with a custom `UnmarshalYAML(*yaml.Node)` walking the mapping node's content pairwise), NOT as `map[string]string`. TASK-008 asserts this exact key sequence.
- `Document` — a discovered standards document: source-root-relative path (slash form), absolute path, `Frontmatter`, and the statement IDs defined in its body (populated by TASK-004).
- `ExampleEntry` — `Path`, `Title`, `Statements []string` (ordered).
- `Node` — output node: `Path`, `Title`, `Description`, `Scope []string`, `Topics []string`, ordered `Aliases`, `Examples []ExampleEntry`, `Children []*Node`. Absent vs empty must remain distinguishable: a document that declares no `aliases`/`examples` must be representable such that TASK-008 omits those keys entirely while still emitting `children: []`.
- `Tree` — the top-level document: the ordered list of root nodes.

Use `go.yaml.in/yaml/v3` for any YAML struct tags / custom unmarshalling. Never import `gopkg.in/yaml.v3`.

### `indexer/errors.go` (GO-016, GO-017)

- Typed errors, one per failure class, each carrying the offending path(s) and rendering a diagnostic that names file and cause: unparseable frontmatter, missing/empty required key, unknown `parent`, parent cycle, missing example file (names citing standard AND missing path), example path escaping the source root, missing example heading, missing `Statements:` line, duplicate statement ID, uncited example file, unknown statement ID (names example, ID, citing standard(s)).
- An **aggregate error** implementing collect-all-then-fail (spec §7): validation NEVER returns fail-fast on the first problem. It must support:
  - `Add(err)` accumulation and an `ErrorOrNil()`-style terminal check
  - **deterministic ordering**: accumulated diagnostics are sorted by path (then message) so repeated runs print identical output (RISK-006)
  - **cascade suppression**: mark a document path as failed, and query whether a path is failed, so later stages can skip dependent checks for it. One corrupted document must ultimately yield exactly ONE diagnostic (asserted end-to-end in TASK-009) — not an unknown-parent error for each of its children plus an uncited-example error for each example it cites.
  - rendering all collected diagnostics as stderr-ready lines

### Tests

`types_test.go`: scalar scope normalizes to a one-element list; list scope round-trips as `['*.js','*.ts','*.vue']` unchanged; an 8-key `aliases` mapping unmarshals preserving declaration order `go, js, ts, py, postgres, pg, deploy, containerization`; a document with no `aliases`/`examples` is distinguishable from one declaring empty ones.

`errors_test.go`: each typed error's message names the file and the cause; the aggregate collects multiple errors without returning early; ordering is deterministic across repeated construction with shuffled insertion order; a path marked failed reports as failed and an unmarked one does not.

## Target Files

- [x] indexer/types.go
- [x] indexer/types_test.go
- [x] indexer/errors.go
- [x] indexer/errors_test.go

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- /Users/pascal/Github/psauerborn/standards/GENERAL.md (lines 1-20 — the root document's frontmatter: the 8-key `aliases` map and its exact declaration order, list-form `scope: ['*']`)
- /Users/pascal/Github/psauerborn/standards/python/GENERAL.md (lines 1-19 — `parent:`, `examples:` entries relative to the document's directory, list `scope`)
- /Users/pascal/Github/psauerborn/standards/docs/standards-tree.yaml (lines 1-60 — authoritative output schema: node key set and order, `aliases` emitted in declaration order, `children: []` on leaves, `examples` entries `{path, title, statements}`)
- indexer/go.mod (module path and permitted dependencies)

## Investigation Notes

- `GENERAL.md` (root corpus document) frontmatter: `scope` is declared in *list* form (`- '*'`), `topics` is a list, and `aliases` is a block mapping whose declaration order is exactly `go, js, ts, py, postgres, pg, deploy, containerization` — neither sorted nor stable under `map[string]string`. It declares no `parent` and no `examples`.
- `python/GENERAL.md`: declares `parent: GENERAL.md` (source-root-relative) and six `examples:` entries written relative to the document's own directory (`examples/GENERAL/data-models.md` → `python/examples/GENERAL/data-models.md`). Confirms `parent`/`examples` are optional keys.
- `docs/standards-tree.yaml` (authoritative output schema): node key order is `path, title, description, scope, topics, [aliases], [examples], children`. `aliases` is emitted as a block mapping in declaration order; `examples` entries are `{path, title, statements}` with `statements` in *flow* style (`[API-002, API-004]`); leaves still emit `children: []`. Nodes without aliases/examples omit those keys entirely — so the model must distinguish "absent" (nil) from "declared but empty" (non-nil, len 0).
- `indexer/go.mod`: module path `github.com/PSauerborn/standards/indexer`, go 1.25.0, requires only `go.yaml.in/yaml/v3 v3.0.5` (runtime) and `github.com/stretchr/testify v1.11.1` (test-only). `.golangci.yml` has a depguard deny rule for `gopkg.in/yaml.v3` and disables the default exclusions so revive's missing doc-comment reports surface (GO-005). Existing `main.go` is flat `package main`.
- Standards applied: GEN-001/GEN-002 (KISS/YAGNI), GO-003/GO-004 (gofmt/golangci-lint), GO-005 (doc comment starting with the identifier name — required on every exported type/method because revive default exclusions are off), GO-006 (snake_case filenames), GO-008 (flat layout), GO-010 (models in `types.go`), GO-016/GO-017 (custom errors in `errors.go`), GO-026/GO-027/GO-028/GO-029/GO-032 (`testing` + one `_test.go` per source file + `TestFunctionName` with `t.Run` subtests + testify assertions).
- GO-012 (validator tags) is intentionally not applied: `github.com/go-playground/validator/v10` is not a permitted dependency for this module, and required-key validation is expressed through the typed `MissingFrontmatterKeyError` vocabulary instead.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-001 | Module scaffolding and tooling | blocks | `go.mod`/`go.sum` with `go.yaml.in/yaml/v3` and testify pinned; green `make lint`/`make test` gate |

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review dependency deliverables (TASK-001 scaffold: module path, permitted deps, Makefile targets)
- [x] Verify/create contract definitions (type names and error vocabulary above)
- [x] Write failing tests in `types_test.go` and `errors_test.go`
- [x] Run `make test` and confirm failure (undefined `AliasList`, `DiagnosticError`, every typed error)

### 2. Green Phase

- [x] Add minimal implementation in `types.go` and `errors.go` to pass tests
- [x] Run only added tests and confirm they pass (`ok github.com/PSauerborn/standards/indexer`)

### 3. Refactor Phase

- [x] Improve code (doc comments per GO-005: first word is the identifier name), maintain passing tests
- [x] Confirm added tests still pass; `make lint` and `make fmt` clean

## Completion Criteria

- [x] A scalar `scope: '*.go'` unmarshals to the one-element list `['*.go']` and is not an error; a list `scope` is preserved in declared order
- [x] An 8-key `aliases` mapping unmarshals preserving declaration order `go, js, ts, py, postgres, pg, deploy, containerization` (ordered pairs, not a Go map)
- [x] Undeclared `aliases`/`examples` are distinguishable in the model from declared-but-empty ones
- [x] The aggregate error collects every added error (no fail-fast), renders diagnostics sorted by path, and supports marking/querying failed documents for cascade suppression
- [x] Each typed error's message names the offending file and the cause
- [x] All added tests pass; `make test`, `make lint`, `make fmt` (no diff) pass in `indexer/`

## Tooling Resolution (TASK-001-owned files, modified under explicit orchestrator authorization)

Writing the first testify-importing tests surfaced two defects in the TASK-001 scaffold.
The orchestrator extended this task's write set to `indexer/go.mod`, `indexer/go.sum` and
`indexer/.golangci.yml` (only those three; `Makefile`, `.gitignore` and `main.go` stayed
untouched) so they could be fixed here.

1. **`go.sum` was missing testify's transitive entries**, because TASK-001 shipped no test
   importing testify, so `make test` failed with
   `missing go.sum entry for module providing package github.com/davecgh/go-spew/spew`.
   Ran `go mod tidy` in `indexer/`: `go-spew`, `go-difflib` and the test-only
   `gopkg.in/yaml.v3` sums were added, and testify/`go.yaml.in/yaml/v3` were promoted to
   direct requirements (the `// indirect` markers are gone, the explanatory comments in
   `go.mod` survived tidy). `go list -deps .` confirms the binary's dependency graph still
   contains `go.yaml.in/yaml/v3` and zero occurrences of `gopkg.in/yaml.v3`.
2. **`golangci-lint` was upgraded v1.64.8 → v2.12.2 (built with go1.26.5)** by the
   orchestrator, which removes the go1.25-vs-go1.26 stdlib typecheck failure that
   testify's `net/http` import triggered. v2 uses a breaking config schema, so
   `.golangci.yml` was migrated with `golangci-lint migrate` (its `.golangci.bck.yml`
   backup was removed) and then hand-tightened to preserve the original intent:
   - The depguard `archived-yaml` deny rule for `gopkg.in/yaml.v3` (RISK-012) carried
     over and was re-proven by temporarily importing `gopkg.in/yaml.v3` in `types.go`:
     lint failed with `import 'gopkg.in/yaml.v3' is not allowed from list 'archived-yaml'`
     (exit 2). The probe was reverted and `types.go` verified byte-identical by checksum.
   - Linter coverage is preserved: `gosimple` folded into `staticcheck` in v2, and
     `gofmt`/`goimports` moved to the new `formatters` section. `errcheck`, `govet`,
     `ineffassign`, `staticcheck` and `unused` are v2 defaults but are listed explicitly
     so the enabled set stays visible. `misspell`, `revive`, `unconvert` and `depguard`
     carry over unchanged.
   - v1's `issues.exclude-use-default: false` becomes "no `exclusions.presets` enabled" in
     v2, which keeps revive's missing-doc-comment reports visible (GO-005).
   - v1's `run.timeout` is dropped: v2 disables the timeout by default.
3. **The `go` directive was raised `1.25.0` → `1.26.5`** to match the local toolchain, now
   that the linter no longer refuses a newer language version; the obsolete comment about
   the v1.64.8 constraint was rewritten. `make lint`, `make test`, `make fmt-check` and
   `make build` all pass at that version.

Final gate in `indexer/`: `make test` → `ok github.com/PSauerborn/standards/indexer`,
`make lint` → `0 issues.`, `make fmt` → no diff (`gofmt -l .` empty, `make fmt-check`
exit 0).

## Notes

- Impact scope: every later task imports these types and errors; the aggregate-error contract determines TASK-007's suppression logic and TASK-009's diagnostic output.
- Scope boundary:
  - Flat `package main` directly in `indexer/` (GO-008) — no `cmd/`, no `internal/`, no subpackages. Filenames snake_case (GO-006). One `_test.go` per source file (GO-027).
  - Do not modify `indexer/go.mod`, `go.sum`, `Makefile` or `.golangci.yml` (owned by TASK-001).
  - Permitted dependencies: `go.yaml.in/yaml/v3` (runtime), `stretchr/testify` (test-only). `gopkg.in/yaml.v3` MUST NOT be imported.
  - The standards corpus (18 documents, 28 example files) and `docs/standards-tree.yaml` are READ-ONLY; no test may mutate the working tree.
  - Any fixture data added later belongs under `indexer/tests/data/.corpora/<case>/` (dot-prefixed). This task needs no filesystem fixtures — build values in code.
  - The binary is `indexer`; the former name `stdidx` must not appear anywhere in the implementation.
