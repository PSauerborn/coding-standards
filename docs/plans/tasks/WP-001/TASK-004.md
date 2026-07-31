# Task: Document discovery and statement-ID extraction

Work Plan ID: WP-001
Task ID: TASK-004
Created Date: 2026-07-31
Description: Implement `indexer/discovery.go` — the filesystem walk (excluding `examples/` and dot-prefixed directories), the frontmatter-`title` discovery rule, the single shared statement-ID regex, and definition-only statement extraction from document bodies.
Acceptance Criteria Covered: AC-1

## Implementation Content

### Walk and exclusions (REQ-1)

- Walk `--source` for `*.md` files. **Skip entire directories** named `examples` and every dot-prefixed directory (`.git`, `.claude`, `.corpora`, …) — do not descend into them.
- Export the walker and its exclusion set as a single reusable unit: TASK-007's REQ-8 uncited-example scan MUST reuse the same walker and exclusion set (a second, unfiltered walk reintroduces the fixture-poisoning failure described in Notes). Design the walk so a caller can ask for "all `*.md` outside dot-directories, including those under `examples/`" using the SAME dot-directory exclusion logic.
- A file is a standards document iff it has parseable YAML frontmatter (per TASK-003's byte-offset-0 delimiter rule) containing a `title` key. No frontmatter, or frontmatter without `title` → not a node, not an error (e.g. `.claude/commands/*.md` carry `description` but no `title`).
- Discovery MUST parse frontmatter for candidate files. Never grep for citation-shaped or ID-shaped lines.
- Discovered documents carry a source-root-relative, slash-form path (`python/DOCKER.md`), plus their absolute path for later reads.
- Required-key validation (TASK-003) applies to discovered documents: a document with `title` but missing `description`/`scope`/`topics` is a validation error collected into the aggregate, not a skipped file.

### Shared statement-ID regex (RISK-004 — silent-wrongness hazard)

Declare ONE exported ID regex in this file, used by both the document statement extractor here and the example `Statements:` parser in TASK-005:

```
`\[([A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*-\d+)\]`
```

The corpus has eighteen prefix families, six of them three-segment: `GO`, `PY`, `JS`, `API`, `DDB`, `PG`, `TF`, `LOG`, `GEN`, `MAKE`, `ACPT`, `DOCKER`, `GO-API`, `GO-DOCKER`, `GO-WRK`, `JS-DOCKER`, `PY-API`, `PY-DOCKER`. A naive `[A-Z]+-[0-9]+` captures only the `DOCKER-001` tail; because the same pattern is applied on both sides of the cross-check, validation still PASSES while emitted `statements` values are wrong and misroute agents. The full token must be captured.

### Definition-only statement extraction (RISK-005 — false-negative hazard)

Collect only statement **definitions** from a document body:

- a backticked `` `[ID]` `` anchored at the start of a line (allowing leading whitespace and list markers such as `- `),
- followed by a MUST/SHOULD marker (`**MUST**` / `**SHOULD**`, as in `` `[PY-001]` **MUST**: … ``),
- with fenced code blocks (```` ``` ```` regions) skipped entirely.

Never collect every backticked `[ID]` token in the body — that accepts IDs a standard merely mentions in prose or shows in a code sample, and lets corpus rot pass the REQ-9 check that TASK-007 builds on.

### Tests (`discovery_test.go`)

- Real-corpus run: discovery over the repository root yields exactly **18** documents (the early verification point). Resolve the repository root relatively from the test file (`runtime.Caller` + relative join) — never a hardcoded absolute path.
- Title rule asserted independently of the dot-directory filter: a NON-hidden fixture document with valid frontmatter but no `title` yields no node.
- README-shaped regression fixture (fenced YAML block with `---` delimiters, `title:`, `examples:`, a `Statements:` line and an ID heading): asserts **zero nodes, zero citations, zero statement IDs**.
- Table-driven ID regex test enumerating all eighteen real prefix families above, asserting full-token capture (`PY-DOCKER-001` captures as `PY-DOCKER-001`, not `DOCKER-001`).
- Definition-only extraction: a fixture where an ID appears (a) as a definition, (b) in prose without a MUST/SHOULD marker, (c) inside a fenced code block — only (a) is collected.
- Directory exclusions: files under an `examples/` directory and under a dot-prefixed directory are never discovered.
- Statement extraction over a real document (`python/GENERAL.md`) includes `PY-001` … and excludes anything from its fenced samples.

Fixtures for this task live under `indexer/tests/data/.corpora/discovery/`.

## Target Files

- [x] indexer/discovery.go
- [x] indexer/discovery_test.go
- [x] indexer/tests/data/.corpora/discovery/ (fixture files created by this task)

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- indexer/frontmatter.go (extraction + required-key API; the byte-offset-0 delimiter rule)
- indexer/types.go (`Document` shape: relative path, absolute path, `Frontmatter`, statement IDs)
- indexer/errors.go (aggregate error: `Add`, mark-failed for cascade suppression)
- /Users/pascal/Github/psauerborn/standards/python/GENERAL.md (lines 20-50 — statement definition shape: `` `[PY-001]` **MUST**: … `` at line start)
- /Users/pascal/Github/psauerborn/standards/README.md (lines 24-72 — fenced YAML + `Statements:` + ID heading; the shape the regression fixture must imitate)
- /Users/pascal/Github/psauerborn/standards/.claude/commands/sync-ids.md (frontmatter with `description` but no `title` — the ignored-file case)

## Investigation Notes

- `indexer/frontmatter.go`: `ParseFrontmatter(documentPath string, content []byte) (*Frontmatter, error)` returns `(nil, nil)` when the content does not open with `---` at byte offset 0, and a `FrontmatterParseError` for an unterminated or invalid block (never a partial value). `ValidateFrontmatter(documentPath string, fm Frontmatter) []error` returns one `MissingFrontmatterKeyError` per missing/empty key in the order title, description, scope, topics. Discovery therefore composes: parse → `fm != nil && fm.Title != ""` → validate.
- `indexer/types.go`: `Document{Path, AbsolutePath, Frontmatter, Statements}` — `Path` is source-root-relative slash form, `Statements` is in order of appearance. `ScopeList` decodes a scalar or list scope.
- `indexer/errors.go`: `ErrorCollector` with `Add`, `MarkFailed`, `IsFailed`, `Errors`, `ErrorOrNil`; diagnostics sort by document path. No typed error exists for an unreadable file, so a read failure is returned as a fatal wrapped error rather than collected.
- `python/GENERAL.md`: definitions have the shape `` `[PY-001]` **MUST**: … `` anchored at the start of a line. A corpus-wide scan shows only `**MUST**` and `**SHOULD**` markers follow an ID.
- `README.md`: fenced YAML block delimited by `---` plus a fenced markdown sample containing `# [GO-022] Configuration Loading` and a `Statements:` line — the offset-0 delimiter rule keeps it out of discovery, and fence skipping keeps its IDs out of statement extraction.
- `.claude/commands/sync-ids.md`: frontmatter with `description` but no `title` — the silently ignored case (it also lives under a dot-directory, so the title rule is asserted separately on a non-hidden fixture).
- Corpus check: exactly 18 `*.md` files outside dot-directories carry a frontmatter `title`; the eighteen statement prefix families in the corpus match the list in this task exactly.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-003 | Frontmatter extraction and parsing | blocks | Frontmatter extraction/parse/required-key API with the offset-0 delimiter rule |

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review dependency deliverables (TASK-003 frontmatter API)
- [x] Create fixtures under `indexer/tests/data/.corpora/discovery/` (no-title document, README-shaped document, prose/code-fence mention document, an `examples/` subdirectory, a nested dot-directory)
- [x] Write failing tests in `discovery_test.go`, including the 18-document real-corpus assertion and the 18-family ID regex table
- [x] Run `make test` and confirm failure

### 2. Green Phase

- [x] Add minimal implementation in `discovery.go` to pass tests
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Improve code (doc comments per GO-005), maintain passing tests
- [x] Confirm added tests still pass; `make lint` and `make fmt` clean

## Completion Criteria

- [x] Discovery over the repository root yields exactly 18 documents, with the repo root resolved relatively (no hardcoded absolute path)
- [x] A non-hidden `*.md` with frontmatter but no `title` yields no node and no error; files under `examples/` and under dot-prefixed directories are never discovered
- [x] The README-shaped fixture yields zero nodes, zero citations and zero statement IDs
- [x] The shared exported ID regex captures full tokens for all eighteen prefix families, including `GO-API`, `GO-DOCKER`, `GO-WRK`, `JS-DOCKER`, `PY-API`, `PY-DOCKER`
- [x] Statement extraction collects only line-anchored backticked IDs followed by a MUST/SHOULD marker, outside fenced code blocks; prose mentions and code-fence occurrences are excluded
- [x] The walker and its exclusion set are exposed for reuse by TASK-007's REQ-8 scan (no second, unfiltered walk anywhere)
- [x] All added tests pass; `make test`, `make lint`, `make fmt` (no diff) pass in `indexer/`

## Notes

- Impact scope: TASK-005 (example `Statements:` parsing) reuses the shared ID regex; TASK-006 builds the hierarchy from discovered documents; TASK-007 reuses this walker for the REQ-8 scan.
- Scope boundary:
  - **Fixture placement (load-bearing, CRITICAL)**: every synthetic fixture corpus MUST live under `indexer/tests/data/.corpora/<case>/`. `indexer/tests/data/` sits INSIDE the `--source` root used when the tool indexes its own repository, and `indexer/` is neither dot-prefixed nor an `examples/` directory — the dot-prefixed `.corpora` directory is the ONLY thing keeping fixtures out of the walk. A fixture standards document placed where the walker can reach it becomes a 19th node (breaking the exactly-18 criterion); a fixture file under a reachable `examples/` directory is by construction uncited and fails corpus validation on the REAL repository. Never place fixture standards documents or fixture `examples/` directories anywhere else.
  - Flat `package main` directly in `indexer/` (GO-008); snake_case filenames (GO-006); one `_test.go` per source file (GO-027); fixture data under `tests/data` (GO-031).
  - Do not modify files owned by earlier tasks (`types.go`, `errors.go`, `frontmatter.go`, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`).
  - Permitted dependencies: `go.yaml.in/yaml/v3` (runtime), `stretchr/testify` (test-only). `gopkg.in/yaml.v3` MUST NOT be imported.
  - The standards corpus (18 documents, 28 example files) and `docs/standards-tree.yaml` are READ-ONLY; no test may mutate the working tree.
  - The binary is `indexer`; the former name `stdidx` must not appear anywhere in the implementation.
