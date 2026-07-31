# Task: Tree assembly, hierarchy and path emission

Work Plan ID: WP-001
Task ID: TASK-006
Created Date: 2026-07-31
Description: Implement `indexer/tree.go` — build the node hierarchy from `parent:` frontmatter with unknown-parent and cycle diagnostics, order siblings by path, and funnel every emitted path (node AND example) through a single `emitPath()` helper applying the `--prefix`.
Acceptance Criteria Covered: AC-2, AC-5

## Implementation Content

### Hierarchy (REQ-2)

- Each discovered document's `parent:` value is a path **relative to the source root** (e.g. `python/DOCKER.md` declares `parent: general/DOCKER.md`). Place the document as a child of that document's node.
- Documents without `parent:` become root nodes. **Multiple roots are permitted, not an error** (today's corpus happens to have the single root `GENERAL.md`).
- `parent:` naming no discovered document → typed error naming **both child and parent**.
- Parent cycle (A→B→A) → typed error **listing the cycle members**.
- Sibling nodes are ordered **by path** (REQ-6). Examples keep their frontmatter order — never reorder them.
- Errors accumulate into the aggregate error (no fail-fast). Documents already marked failed by earlier stages must not additionally produce unknown-parent errors for their children (cascade suppression).

### Path emission — one helper only (RISK-009, RISK-016)

- Every emitted path — **node paths AND example paths** — funnels through a single `emitPath()` helper that applies `filepath.ToSlash`, `path.Clean`, and the prefix join, producing exactly one `/` between prefix and path.
- Prefix spellings to handle: `p`, `p/`, `p//`, `./p`, and empty (no prefix).
- The `--prefix` is applied **strictly as the final emission step** — never during parsing, parent lookup, or example resolution. Applying it earlier breaks parent lookup (spurious unknown-parent failures) or silently resolves the wrong example file.

### Tests (`tree_test.go`)

Build `Document` values in code for the hierarchy cases (no filesystem fixtures needed for most of them):

- Hierarchy facts against the real corpus or equivalent in-code documents: single root `GENERAL.md`; `python/DOCKER.md` is a child of `general/DOCKER.md`; `golang/API.md` is a child of `golang/GENERAL.md` (AC-2).
- Unknown parent → error naming child and parent. Cycle A→B→A → error listing both members. Two parentless documents → two roots, no error.
- Siblings ordered by path.
- Table-driven `emitPath()` test over prefix spellings (`p`, `p/`, `p//`, `./p`, empty) asserting exactly one separator and slash-form cleaned output.
- Example paths are prefixed too, not just node paths.
- `--prefix x` yields a tree of identical SHAPE to no prefix (same hierarchy, same node ordering; only the path strings differ) and every path starts with `x/` (AC-5).

If any case genuinely needs on-disk documents, put fixtures under `indexer/tests/data/.corpora/tree/`.

## Target Files

- [x] indexer/tree.go
- [x] indexer/tree_test.go
- [x] indexer/tests/data/.corpora/tree/ (fixture files, only if a case needs on-disk documents)

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- indexer/discovery.go (discovered `Document` values: source-root-relative slash-form paths)
- indexer/examples.go (`ExampleEntry` values whose paths must also pass through `emitPath()`)
- indexer/types.go (`Node`, `Tree`, `Config.Prefix`)
- indexer/errors.go (unknown-parent and cycle error types; aggregate accumulation and failed-document marking)
- /Users/pascal/Github/psauerborn/standards/python/DOCKER.md (lines 1-12 — `parent: general/DOCKER.md`, a cross-directory parent link)
- /Users/pascal/Github/psauerborn/standards/docs/standards-tree.yaml (lines 30-60 — nesting shape of `children` and where `examples` sit relative to `children`)

## Investigation Notes

- `types.go`: `Document{Path, AbsolutePath, Frontmatter, Statements}` carries source-root-relative slash-form paths. `Node{Path, Title, Description, Scope []string, Topics, Aliases, Examples, Children []*Node}` — `aliases`/`examples` are `omitempty`, `children` is always emitted, so a leaf needs a non-nil empty `Children` slice to render `children: []`. `Tree{Nodes []*Node}`. `Config{Source, Output, Prefix}` supplies both the corpus root and the prefix, so `BuildTree(config, documents, collector)` needs no extra parameters.
- `discovery.go`: `DiscoverDocuments` returns documents in lexical path order and already calls `collector.MarkFailed(path)` for documents whose frontmatter is unparseable (which contribute NO document) and for documents failing validation (which DO contribute a document). Cascade suppression therefore means: skip `UnknownParentError` when the unresolved parent path is `collector.IsFailed`.
- `examples.go`: `ResolveExamples(source, document, collector)` returns entries in frontmatter order with source-root-relative, unprefixed paths, and never fails fast. Prefixing must be applied to `ExampleEntry.Path` at node construction only.
- `errors.go`: reuse `UnknownParentError{Path, Parent}` (message names both) and `ParentCycleError{Cycle []string}` (message joins members with ` -> `, `DocumentPath()` is `Cycle[0]`). Cycle is recorded in traversal order ending with the path it returns to, so an A→B→A cycle is `["A","B","A"]`; starting the walk from the lexically first document makes the recorded representation deterministic.
- `python/DOCKER.md` declares `parent: general/DOCKER.md` — a cross-directory parent link, confirming `parent:` is source-root-relative and NOT relative to the declaring document's directory.
- `docs/standards-tree.yaml` (lines 30-70): `examples:` sits before `children:` inside a node; every node carries `children` even when empty (`children: []`), and example paths are full source-root-relative paths (`python/examples/DOCKER/two-stage-build.md`), not relative to the citing document.
- Documents whose parent is unknown or that take part in a cycle cannot be attached anywhere; they are excluded from the emitted roots (together with their subtrees) while the diagnostic is collected, so the run reports the problem instead of silently promoting a broken node to a root.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-004 | Document discovery and statement-ID extraction | blocks | Discovered `Document` values with source-root-relative paths and `parent:` frontmatter |
| TASK-005 | Example file parsing and citation resolution | blocks | `ExampleEntry` values whose paths must flow through `emitPath()` |

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review dependency deliverables (TASK-004 documents, TASK-005 example entries)
- [x] Write failing tests in `tree_test.go` (hierarchy facts, unknown parent, cycle, multi-root, sibling ordering, `emitPath()` prefix table, prefixed example paths, prefix shape-invariance)
- [x] Run `make test` and confirm failure

### 2. Green Phase

- [x] Add minimal implementation in `tree.go` to pass tests
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Improve code (doc comments per GO-005), maintain passing tests
- [x] Confirm added tests still pass; `make lint` and `make fmt` clean

## Completion Criteria

- [x] The assembled tree has the single root `GENERAL.md`; `python/DOCKER.md` is a child of `general/DOCKER.md`; `golang/API.md` is a child of `golang/GENERAL.md`
- [x] An unknown `parent:` fails naming both child and parent; an A→B→A cycle fails listing the cycle members; multiple parentless documents produce multiple roots without error
- [x] Sibling nodes are ordered by path; example entries retain frontmatter order
- [x] Every emitted node path and example path passes through the single `emitPath()` helper; prefixes `p`, `p/`, `p//`, `./p` all produce exactly one `/` separator, and empty prefix leaves paths unchanged
- [x] With `--prefix x`, every node AND example path starts with `x/` and the tree shape is identical to the unprefixed run
- [x] All added tests pass; `make test`, `make lint`, `make fmt` (no diff) pass in `indexer/`

## Notes

- Impact scope: TASK-008 renders this tree; TASK-010 asserts hierarchy facts under `--prefix`; TASK-011 compares the assembled tree against the reference.
- Scope boundary:
  - **Fixture placement (load-bearing, CRITICAL)**: any synthetic fixture corpus MUST live under `indexer/tests/data/.corpora/<case>/`. `indexer/tests/data/` sits INSIDE the `--source` root used when the tool indexes its own repository, and `indexer/` is neither dot-prefixed nor an `examples/` directory — the dot-prefixed `.corpora` directory is the ONLY thing keeping fixtures out of the walk. A fixture standards document reachable by the walker becomes an extra node and breaks the exactly-18-node criterion.
  - Flat `package main` directly in `indexer/` (GO-008); snake_case filenames (GO-006); one `_test.go` per source file (GO-027); fixture data under `tests/data` (GO-031).
  - Do not modify files owned by earlier tasks (`types.go`, `errors.go`, `frontmatter.go`, `discovery.go`, `examples.go`, and their tests/fixtures, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`).
  - Permitted dependencies: `go.yaml.in/yaml/v3` (runtime), `stretchr/testify` (test-only). `gopkg.in/yaml.v3` MUST NOT be imported.
  - The standards corpus (18 documents, 28 example files) and `docs/standards-tree.yaml` are READ-ONLY; no test may mutate the working tree.
  - The binary is `indexer`; the former name `stdidx` must not appear anywhere in the implementation.
