# Task: Deterministic YAML rendering

Work Plan ID: WP-001
Task ID: TASK-008
Created Date: 2026-07-31
Description: Implement `indexer/render.go` — marshal the assembled tree through the `yaml.Node` API with explicit, code-declared key ordering for every mapping, correct absent-vs-empty field handling, and safe special-scalar emission.
Acceptance Criteria Covered: AC-3, AC-6

## Implementation Content

### Determinism mechanism (RISK-007)

- Build the output document through the `yaml.Node` API of `go.yaml.in/yaml/v3` with **EXPLICIT, code-declared key ordering for every mapping**. Never rely on incidental encoder map-key sorting — that is undocumented behaviour a later refactor silently drops, and the failure surfaces intermittently.
- Node key order follows the reference schema: `path`, `title`, `description`, `scope`, `topics`, `aliases` (when declared), `examples` (when declared), `children`. Example entry key order: `path`, `title`, `statements`.
- The root node's `aliases` mapping is emitted in **frontmatter declaration order**: `go, js, ts, py, postgres, pg, deploy, containerization`. Assert this exact key sequence in `render_test.go`.
- Top level is `nodes:` — the list of root nodes.

### Scalars (RISK-011)

- Marshal exclusively through the YAML library with **default scalar styles**. No hand-rolled string formatting, and no explicitly forced style/tag that suppresses quoting: every node's `scope` contains patterns like `'*'`, and `*` is the YAML alias indicator.
- Include an emit → re-parse structural round trip asserting an emitted `'*'` scope re-parses to the string `*`.

### Field presence (RISK-015)

- `scope` is ALWAYS a list. `topics` is a list. `children` is ALWAYS present, `[]` when empty.
- `aliases` and `examples` keys are **omitted entirely** when the document declares none — never `aliases: {}` or `examples: []`.
- `examples` are emitted in frontmatter order; sibling nodes in the order the tree assembly produced (by path).

### Content constraints (RISK-008)

- No timestamps, tool versions, hostnames, or absolute filesystem paths anywhere in the emitted document.

### Output write ordering

Rendering produces bytes; the file is written only after full validation passes — the write itself is TASK-009's responsibility. Do not write output from this file.

### Tests (`render_test.go`)

- Exact `aliases` key sequence `go, js, ts, py, postgres, pg, deploy, containerization` on the root node.
- A node with neither declared `aliases` nor `examples` omits BOTH keys while still carrying `children: []`; the root carries `aliases`.
- Emit → re-parse → structural equality round trip, including the explicit assertion that an emitted `'*'` scope re-parses to the string `*`.
- Repeated renders of the same tree are byte-identical; run render/determinism tests under `-count=10 -shuffle=on` (document the invocation in the Makefile-independent test comment or a `t.Log`).
- Render fixtures include an **example-bearing node** (path/title/statements) so the render↔example coupling is exercised here rather than downstream (RISK-021).
- Emitted output contains no absolute filesystem path.
- Scope emitted as a list even when the source frontmatter carried a scalar.

Build `Tree`/`Node`/`ExampleEntry` values in code for these tests — no filesystem fixtures are needed.

## Target Files

- [x] indexer/render.go
- [x] indexer/render_test.go

## Investigation Targets

Files to read before starting implementation (file path, with optional search hint):

- indexer/types.go (`Tree`, `Node`, `ExampleEntry`, ordered `Aliases` representation, absent-vs-empty modelling)
- indexer/tree.go (assembled tree shape and sibling ordering; `emitPath()` output form)
- indexer/examples.go (`ExampleEntry` fields to render)
- /Users/pascal/Github/psauerborn/standards/docs/standards-tree.yaml (lines 1-60 — authoritative output schema: exact node key order, `aliases` order, `children: []` on leaves, `scope: - '*'` quoting, `examples` entry shape)

## Investigation Notes

- `indexer/types.go`: `Node` carries `Path`, `Title`, `Description`, `Scope []string`,
  `Topics []string`, `Aliases AliasList`, `Examples []ExampleEntry`, `Children []*Node`;
  `Tree` is `{Nodes []*Node}`. `AliasList` is an ordered `[]Alias` (Key/Value) whose
  `UnmarshalYAML` walks the mapping node pairwise, so declaration order survives a
  decode; a map would not. `ScopeList.UnmarshalYAML` normalizes a scalar scope into a
  one-element list, so `Node.Scope` is already a list by the time rendering sees it.
  Undeclared `aliases`/`examples` stay nil.
- `indexer/tree.go`: `BuildTree` already applies `--prefix` through the single
  `emitPath()` helper at node and example construction; rendering must NOT re-apply it.
  Roots and sibling lists are sorted by path, example entries stay in frontmatter order,
  leaves get a non-nil empty `Children` slice, and `emitExamples()` collapses an empty
  example list back to nil. Rendering therefore only walks ordered slices — no map
  iteration anywhere, which is what makes the output deterministic.
- `indexer/examples.go`: `ExampleEntry` values carry the source-root-relative `Path`, the
  heading `Title` with the statement identifier stripped, and `Statements` in declaration
  order.
- `docs/standards-tree.yaml` (lines 1-70): top level is `nodes:`; node key order is
  `path`, `title`, `description`, `scope`, `topics`, `aliases`, `examples`, `children`;
  example entry key order is `path`, `title`, `statements` with statements in flow style
  (`statements: [API-002, API-004]`); `scope` entries such as `'*'` are quoted; leaves
  carry `children: []`; nodes declaring no aliases/examples omit those keys entirely.
- Encoder behaviour verified against `go.yaml.in/yaml/v3` v3.0.5 before implementing:
  a scalar node with the resolved `!!str` tag and the zero (default) style emits `*` as
  `'*'` and re-parses to the string `*`, and also keeps `123`/`true`/`null`-looking
  titles as strings, whereas an untagged scalar re-parses them as int/bool/null. The
  `!!str` tag constrains the type, not the quoting, so it does not conflict with the
  RISK-011 requirement to leave quoting to the library.
- The emitter indents block sequences under their parent key (`SetIndent(2)` gives a
  two-space dash indent), which differs cosmetically from the hand-written reference
  file; TASK-011 compares structure, not bytes.
- Package-level helper names already taken (collision avoidance): `readFixture`,
  `readDiscoveryFixture`, `readExampleFixture`, `readCorpusDocument`, `fixtureRoot`,
  `fixtureDocument`, `aliasKeys`, `treeShape`, `treeFindNode`, `treeNodePaths`,
  `treeExamplePaths`, `documentPaths`, `sortedKeys`. New helpers here are `render`-prefixed.

## Task Dependencies

| Task ID | Title | Dependency Type | Deliverable Consumed |
| --- | --- | --- | --- |
| TASK-005 | Example file parsing and citation resolution | blocks | `ExampleEntry` data (path/title/statements) rendered inside nodes |
| TASK-006 | Tree assembly, hierarchy and path emission | blocks | The assembled `Tree` with `emitPath()`-normalized node and example paths |

## Implementation Steps (TDD: Red-Green-Refactor)

### 1. Red Phase

- [x] Read all Investigation Targets and record key observations
- [x] Review dependency deliverables (TASK-005 example entries; TASK-006 assembled tree)
- [x] Write failing tests in `render_test.go` (aliases key sequence, absent-vs-empty, `'*'` round trip, repeated-render byte equality, example-bearing node, no absolute paths)
- [x] Run `make test` (and the determinism tests under `-count=10 -shuffle=on`) and confirm failure

### 2. Green Phase

- [x] Add minimal implementation in `render.go` to pass tests
- [x] Run only added tests and confirm they pass

### 3. Refactor Phase

- [x] Improve code (doc comments per GO-005), maintain passing tests
- [x] Confirm added tests still pass; `make lint` and `make fmt` clean

## Completion Criteria

- [x] Every mapping's key order is declared in code (`yaml.Node` construction), not left to encoder behaviour; the root's `aliases` emit in the exact sequence `go, js, ts, py, postgres, pg, deploy, containerization`
- [x] A node with no declared `aliases`/`examples` omits both keys and still emits `children: []`; `scope` is always a list
- [x] An emitted `'*'` scope re-parses to the string `*`; emit → re-parse → structural equality holds
- [x] Repeated renders of the same tree are byte-identical under `go test -count=10 -shuffle=on`
- [x] Render tests cover an example-bearing node emitting `path`, `title`, `statements` in that key order
- [x] The emitted document contains no timestamp, version, hostname, or absolute filesystem path
- [x] All added tests pass; `make test`, `make lint`, `make fmt` (no diff) pass in `indexer/`

## Notes

- Impact scope: TASK-009 writes these bytes atomically; TASK-010 byte-compares two process invocations (AC-6); TASK-011 compares the structure against `docs/standards-tree.yaml`.
- Scope boundary:
  - Flat `package main` directly in `indexer/` (GO-008); snake_case filenames (GO-006); one `_test.go` per source file (GO-027).
  - Do not modify files owned by earlier tasks (`types.go`, `errors.go`, `frontmatter.go`, `discovery.go`, `examples.go`, `tree.go`, `validate.go`, and their tests/fixtures, `go.mod`, `go.sum`, `Makefile`, `.golangci.yml`).
  - Permitted dependencies: `go.yaml.in/yaml/v3` (runtime), `stretchr/testify` (test-only). `gopkg.in/yaml.v3` is archived and MUST NOT be imported (it reaches `go.sum` only transitively via testify).
  - The standards corpus (18 documents, 28 example files) and `docs/standards-tree.yaml` are READ-ONLY — this tool never rewrites the reference tree; no test may mutate the working tree.
  - Any fixture data (not expected for this task) belongs under `indexer/tests/data/.corpora/<case>/`; the dot prefix is what keeps fixtures out of the walked source root.
  - The binary is `indexer`; the former name `stdidx` must not appear anywhere in the implementation.
