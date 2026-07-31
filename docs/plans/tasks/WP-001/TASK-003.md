# Task: Frontmatter extraction and parsing

Work Plan ID: WP-001
Task ID: TASK-003
Created Date: 2026-07-31
Description: Implement `indexer/frontmatter.go` — delimit YAML frontmatter with the opening `---` anchored at byte offset 0, parse it into `Frontmatter`, and validate the required keys with diagnostics naming file and cause.
Acceptance Criteria Covered: none directly (REQ-7 foundation for AC-1)

## Implementation Content

### Delimiting (RISK-002 — CRITICAL correctness detail)

- The opening `---` MUST be the file's **first line, at byte offset 0**. A file whose first line is anything else has NO frontmatter — full stop.
- The closing delimiter is found by scanning lines forward from line 1 for the next line consisting solely of `---`.
- NEVER scan for the first `---` anywhere in the file, and NEVER inspect the contents of fenced code blocks.
- Why this matters concretely: the repository's `README.md` has no frontmatter but contains a ```` ```yaml ```` fenced block whose body is `---`-delimited and contains `title: Golang REST API Standards`, `parent: golang/GENERAL.md`, `scope`, and `topics`, plus (further down) a `# [GO-022] Configuration Loading` heading, a `Statements:` line, and an `examples:` list. A lenient extractor turns `README.md` into a plausible phantom 19th node duplicating `golang/API.md` — breaking the "exactly 18 nodes" acceptance criterion — and drags fake citations into corpus validation.
- A file with an opening `---` at offset 0 but no closing delimiter is unparseable frontmatter → error.

### Parsing and validation

- YAML-parse the delimited block into the `Frontmatter` type from `indexer/types.go` using `go.yaml.in/yaml/v3`. On a YAML syntax error, return the typed unparseable-frontmatter error naming the file and the underlying YAML error.
- Required-key validation: `title`, `description`, `scope`, `topics`. Missing or empty values are errors naming the file and the key. A `topics` key that is **present but empty** is a failure.
- Only `title` governs discovery (REQ-1, applied in TASK-004). A discovered document (one that HAS `title`) missing `description`, `scope`, or `topics` is a **validation error, not a silently skipped file**. Expose extraction/parse and required-key validation such that TASK-004 can ask "does this file have parseable frontmatter with a title?" without conflating that with the required-key check.
- Optional keys `parent`, `examples`, `aliases` are absent-tolerant; absent must remain distinguishable from empty (TASK-002's model).

### Tests (`frontmatter_test.go`)

- Round-trips a real corpus document: `javascript/GENERAL.md` yields scope `['*.js','*.ts','*.vue']`; `GENERAL.md` yields the 8-key ordered aliases map and no `parent`; `python/GENERAL.md` yields `parent: GENERAL.md` and 6 `examples:` entries.
- README-shaped regression (RISK-002): a fixture whose first line is a heading and which contains a fenced ```` ```yaml ```` block with `---` delimiters, `title:`, `parent:`, `scope:`, `topics:` — assert NO frontmatter is found and no error is produced (it is simply not a standards document).
- Unparseable YAML inside a correctly delimited block → typed error naming the file and the YAML cause.
- Missing `description`; missing `scope`; missing `topics`; present-but-empty `topics` → validation errors naming file and key.
- Scalar `scope` accepted and normalized (delegates to TASK-002's model).
- Unterminated frontmatter (opening `---`, no closing) → error.

Fixtures for these cases go under `indexer/tests/data/.corpora/frontmatter/` (see the fixture-placement constraint in Notes). Prefer in-memory byte slices where the API allows it; use fixture files only for the cases that need on-disk files.

## Target Files

- [x] indexer/frontmatter.go
- [x] indexer/frontmatter_test.go
- [x] indexer/tests/data/.corpora/frontmatter/ (fixture files created by this task)

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- indexer/types.go (`Frontmatter` shape, scope normalization, ordered `Aliases`)
- indexer/errors.go (typed error constructors for unparseable frontmatter and missing/empty required keys)
- /Users/pascal/Github/psauerborn/standards/README.md (lines 24-72 — the fenced YAML block with `---` delimiters, `title:`/`parent:`/`scope:`/`topics:`, the `# [GO-022]` heading, the `Statements:` line and the `examples:` list; this is the exact shape the regression fixture must imitate)
- /Users/pascal/Github/psauerborn/standards/javascript/GENERAL.md (lines 1-15 — three-pattern list `scope`)
- /Users/pascal/Github/psauerborn/standards/GENERAL.md (lines 1-20 — root frontmatter with the 8-key `aliases` map, no `parent`)

## Investigation Notes

- `indexer/types.go`: `Frontmatter` has `Title`, `Description`, `Scope ScopeList`, `Topics []string`,
  `Parent`, `Examples []string`, `Aliases AliasList`. `ScopeList.UnmarshalYAML` already normalizes a
  scalar `scope` into a one-element list and decodes an explicit null to nil; `AliasList.UnmarshalYAML`
  walks the mapping node pairwise so declaration order is preserved. Absent optional keys therefore stay
  nil without any work in the frontmatter parser — parsing delegates normalization entirely to the model.
- `indexer/errors.go`: `FrontmatterParseError{Path, Err}` renders `"<path>: unparseable frontmatter: <err>"`
  and unwraps to the cause; `MissingFrontmatterKeyError{Path, Key}` renders
  `"<path>: missing required frontmatter key \"<key>\""`. Both implement `DiagnosticError`, and
  `ErrorCollector` sorts by document path, so validation returning a slice of diagnostics (rather than a
  single first error) fits the aggregate model.
- `README.md` (lines 24-72): first line is `# Code Standards`; the fenced ```` ```yaml ```` block at lines
  28-39 is `---`-delimited and carries `title`, `description`, `parent: golang/GENERAL.md`, `scope: '*.go'`
  and a `topics` list, followed later by a fenced `# [GO-022] Configuration Loading` heading with a
  `Statements:` line and a second fenced block holding an `examples:` list. Anchoring the opening delimiter
  at byte offset 0 is the whole defence against this file becoming a phantom node.
- `javascript/GENERAL.md`: list `scope` `['*.js','*.ts','*.vue']`, `parent: GENERAL.md`, five topics, no
  `aliases`, no `examples`.
- `GENERAL.md`: root document, no `parent`, `scope: ['*']`, and an 8-key `aliases` map in the order
  go, js, ts, py, postgres, pg, deploy, containerization.
- `python/GENERAL.md`: `parent: GENERAL.md`, `scope: ['*.py']`, six `examples:` entries.
- API split chosen for TASK-004: `ParseFrontmatter(documentPath, content)` returns `(nil, nil)` for a
  document with no frontmatter, a typed `FrontmatterParseError` for an unterminated or unparseable block,
  and never a partial value; `ValidateFrontmatter(documentPath, fm)` separately reports the missing
  required keys. Discovery can therefore ask "parseable frontmatter with a title?" without triggering the
  required-key check.
- `types_test.go` reuse: the package-level `aliasKeys` test helper already exists, so the aliases
  assertions reuse it rather than redeclaring it.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-002 | Data models and aggregate error types | blocks | `Frontmatter` model with scope normalization and ordered aliases; typed frontmatter/required-key errors |

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review dependency deliverables (TASK-002 `types.go`/`errors.go` contracts)
- [x] Create the README-shaped and malformed-frontmatter fixtures under `indexer/tests/data/.corpora/frontmatter/`
- [x] Write failing tests in `frontmatter_test.go`
- [x] Run `make test` and confirm failure

### 2. Green Phase

- [x] Add minimal implementation in `frontmatter.go` to pass tests
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Improve code (doc comments per GO-005), maintain passing tests
- [x] Confirm added tests still pass; `make lint` and `make fmt` clean

## Completion Criteria

- [x] A file whose first line is not `---` yields no frontmatter, even when it contains a `---`-delimited fenced YAML block with `title:`/`parent:` (README-shaped fixture asserts this)
- [x] `javascript/GENERAL.md` parses to scope `['*.js', '*.ts', '*.vue']`; `GENERAL.md` parses to the 8-key ordered `aliases` map with no `parent`
- [x] Unparseable YAML in a delimited block produces a typed error naming both the file and the YAML cause; no partial `Frontmatter` is returned
- [x] Missing `description`/`scope`/`topics`, and a present-but-empty `topics`, each produce a validation error naming the file and the key — never a silent skip
- [x] All added tests pass; `make test`, `make lint`, `make fmt` (no diff) pass in `indexer/`

## Notes

- Impact scope: TASK-004 discovery calls this parser for every candidate `*.md`; a lenient delimiter here corrupts the node count, the hierarchy, and corpus validation.
- Scope boundary:
  - **Fixture placement (load-bearing)**: every synthetic fixture corpus MUST live under `indexer/tests/data/.corpora/<case>/`. The dot-prefixed `.corpora` directory is what keeps fixtures out of the walked source root — `indexer/tests/data/` sits INSIDE the `--source` root when the tool indexes its own repository, and `indexer/` is neither dot-prefixed nor an `examples/` directory. A fixture standards document or fixture `examples/` directory reachable by the walker makes the tool fail on its own repository.
  - Flat `package main` directly in `indexer/` (GO-008); snake_case filenames (GO-006); one `_test.go` per source file (GO-027); fixture data under `tests/data` (GO-031).
  - Do not modify `types.go`, `errors.go`, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml` (owned by earlier tasks).
  - Permitted dependencies: `go.yaml.in/yaml/v3` (runtime), `stretchr/testify` (test-only). `gopkg.in/yaml.v3` MUST NOT be imported.
  - The standards corpus (18 documents, 28 example files) and `docs/standards-tree.yaml` are READ-ONLY; no test may mutate the working tree.
  - The binary is `indexer`; the former name `stdidx` must not appear anywhere in the implementation.
