# SPEC-001: Standards Indexer

Spec ID: SPEC-001
Spec Date: 2026-07-31

## 1. Spec Statement

As an agent operator I want to be able to index standards files and examples So that agents can find relevant standards documentation without reading all standards files

## 2. Context and Background

The standards repository holds 18 standards documents and 28 single-example
files (~3,200 lines total). Agents must not read this corpus wholesale: they
consult a generated `standards-tree.yaml` index and use its scope/topic metadata to
decide which standards documents and example files to open and read. This significantly reduces context bloat: agents working on a JS application do not need to read Golang standards, and agents working on a Golang API do not need to read examples for CLI applications.

The aim of this spec is to create an indexer at `indexer/` that traverses the standards repository and indexes all standards documents and examples using the provided frontmatter to generate a single index file that agents can reference.

## 3. Scope Definitions

### 3.1 In Scope

- A CLI that scans a standards repository and generates the standards tree
  YAML from document frontmatter and example files.
- Extraction of per-example metadata (title, covered statement IDs) into the
  tree.
- Validation of the corpus with actionable diagnostics and a non-zero exit on
  failure.

### 3.2 Out of Scope

- Authoring or modifying standards documents and example files.

## 4. Requirements

- **REQ-1**: The indexer MUST discover every standards document in the source
  directory: any `*.md` file with parseable YAML frontmatter containing a
  `title` key, excluding files under `examples/` and hidden (dot-prefixed)
  directories. Files without frontmatter, and frontmatter without a `title`
  key (e.g. `.claude/commands/*.md`), MUST be ignored.
- **REQ-2**: The indexer MUST build the tree from `parent:` frontmatter: each
  document is placed as a child of the document its `parent:` names (path
  relative to the source root); documents without `parent:` become root nodes.
- **REQ-3**: Each node MUST carry `path`, `title`, `description`, `scope`,
  `topics`, and `children` from frontmatter, plus `aliases` and `examples`
  when the document declares them. `scope` MUST always be emitted as a YAML
  list of distinct patterns, never a comma-joined string; a scalar `scope` in
  frontmatter is normalized to a single-element list.
- **REQ-4**: For each entry in a document's `examples:` frontmatter, the
  indexer MUST read the example file and emit `path`, `title` (from the
  `# [<ID>] <Title>` heading, ID stripped), and `statements` (every ID on the
  `Statements:` line, in order).
- **REQ-5**: All emitted paths (node and example) MUST be relative to the
  `--source` root; resolving them against that root is the consumer's
  responsibility. When `--prefix` is supplied, the deployment prefix MUST be
  prepended to every emitted path, with any trailing slash on the prefix
  normalized away so that exactly one `/` separates prefix from path.
- **REQ-6**: Output MUST be deterministic: two runs over identical input
  produce byte-identical files. Sibling nodes are ordered by path; examples
  keep their frontmatter order.
- **REQ-7**: The indexer MUST fail (non-zero exit, diagnostic naming file and
  cause) on: unparseable frontmatter, a `parent:` that names no discovered
  document, a parent cycle, an `examples:` entry whose file is missing, or an
  example file lacking the `# [<ID>] <Title>` heading or `Statements:` line.
- **REQ-8**: The indexer MUST fail if any `*.md` file under an `examples/`
  directory is not cited by any discovered document's `examples:` frontmatter.
- **REQ-9**: The indexer MUST fail if an example's heading ID is absent from
  its `Statements:` line, or if any ID on the `Statements:` line does not
  occur as a `` `[<ID>]` `` statement in at least one of the standards
  citing that example.
- **REQ-10**: The generated tree, run against this repository, MUST be
  semantically identical to the reference at `docs/standards-tree.yaml`
  (same nodes, hierarchy, field values, and example metadata; sibling
  ordering and comments may differ).

## 5. Acceptance Criteria

- **AC-1** (REQ-1): Given the current repository, when the indexer runs, then
  the tree contains exactly 18 nodes; any `*.md` file lacking parseable
  frontmatter with a `title` key, and every file under `examples/`, yields
  no node.
- **AC-2** (REQ-2): Given the current repository, when the indexer runs, then
  the tree has the single root `GENERAL.md`, `python/DOCKER.md` is a child of
  `general/DOCKER.md`, and `golang/API.md` is a child of `golang/GENERAL.md`.
- **AC-3** (REQ-3): Given `javascript/GENERAL.md` declaring three scope
  patterns, when the indexer runs, then the node's scope is the list
  `['*.js', '*.ts', '*.vue']`, and the root node carries the `aliases:` map.
- **AC-4** (REQ-4): Given `python/examples/GENERAL/config.md`, when the
  indexer runs, then the `python/GENERAL.md` node lists that example with
  title `Configuration Loading` and statements `PY-019`–`PY-024`.
- **AC-5** (REQ-5): Given `--prefix {prefix}`, when the indexer runs, then
  every `path` in the output starts with `{prefix}/`.
- **AC-6** (REQ-6): Given two runs over an unchanged repository, when outputs
  are compared, then they are byte-identical.
- **AC-7** (REQ-7): Given an `examples:` entry pointing at a deleted file,
  when the indexer runs, then it exits non-zero and the diagnostic names both
  the citing standard and the missing path.
- **AC-8** (REQ-8): Given an uncited file added under `golang/examples/`,
  when the indexer runs, then it exits non-zero naming the uncited file.
- **AC-9** (REQ-9): Given an example whose `Statements:` line contains an ID
  not present in any citing standard, when the indexer runs, then it exits
  non-zero naming the example, the ID, and the citing standard(s).
- **AC-10** (REQ-10): Given the current repository, when the generated tree
  and `docs/standards-tree.yaml` are compared structurally (order- and
  comment-insensitive), then they are identical.

## 6. Contracts and Constraints

- **Output contract**: `standards-tree.yaml` with top-level `nodes:` (list of
  root nodes). Node: `path`, `title`, `description`, `scope` (list),
  `topics` (list), optional `aliases` (map), optional `examples` (list of
  `{path, title, statements}`), `children` (list, always present). The
  reference file `docs/standards-tree.yaml` is authoritative for the schema.
- **Input contract — standards documents**: YAML frontmatter with the
  required keys `title`, `description`, `scope` (list or scalar), and
  `topics`; and the optional keys `parent` (path relative to source root),
  `examples` (paths relative to the document's directory), and `aliases`.
- **Input contract — example files**: first line `# [<ID>] <Title>`; a
  `Statements:` line listing every covered ID as `` `[<ID>]` ``; no
  frontmatter.
- **CLI contract**: `indexer --source <dir> --output <file> [--prefix <p>]`;
  exit 0 on success, non-zero on any validation failure; diagnostics to
  stderr.
- **Constraint**: implemented in Go in `indexer/`
  (`github.com/PSauerborn/standards/indexer`), following this repository's
  Go standards; no dependencies beyond a YAML library.
- **Constraint**: the output must remain consumable by the traversal
  algorithm in `CLAUDE.md` without agent-side changes.

## 7. Edge Cases and Error Handling

- **Unparseable frontmatter**: fail fast naming the file and YAML error; no
  output file is written.
- **Scalar `scope`**: normalized to a one-element list; not an error.
- **Missing required frontmatter key** (`title`, `description`, `scope`,
  `topics`): fail naming the file and key; a `topics` key that is present
  but empty is likewise a failure. Only `title` governs discovery (REQ-1):
  a discovered document that lacks `description`, `scope`, or `topics` is a
  validation error, not a silently ignored file.
- **`parent:` names an undiscovered document**: fail naming child and parent.
- **Parent cycle** (A→B→A): fail listing the cycle members.
- **Multiple roots**: permitted (any document without `parent:` is a root);
  not an error.
- **Example path escaping the source root** (`../`): fail naming the entry.
- **Example cited by two standards**: permitted; the file appears under each
  citing node, and REQ-9 is checked against the union of the citing
  standards: each ID on the `Statements:` line must occur in at least one of
  them.
- **`examples:` key absent or empty**: node is emitted without an `examples`
  field; not an error.
- **Duplicate IDs on a `Statements:` line**: fail naming the example file.
- **Partial failure**: all validation errors in a run are collected and
  reported together before exiting; the output file is only written when
  validation passes in full.

## 8. Infrastructure Requirements

None.

## 9. External Resources

| Filepath | Description | When to use |
|----------|-------------|-------------|
| docs/standards-tree.yaml | Hand-authored reference output defining the target tree schema | Authoritative schema for output; structural comparison target for AC-10 |
