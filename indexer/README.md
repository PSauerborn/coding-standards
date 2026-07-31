# indexer

`indexer` builds the standards index for this repository. It walks a corpus of
standards documents, resolves the example files they cite, assembles them into a
single-rooted tree that mirrors the `parent:` frontmatter, validates the corpus,
and writes the result as YAML.

The tree is what an agent reads to decide which standards apply to a task, so
the output is a contract: see [Output schema](#output-schema).

## CLI contract

```text
indexer --source <dir> --output <file> [--prefix <p>]
```

| Flag | Required | Meaning |
| --- | --- | --- |
| `--source` | yes | Root directory of the standards corpus to index. Paths in the output are relative to it. |
| `--output` | yes | Path of the tree file to write. Its directory must already exist. |
| `--prefix` | no | Prepended to every emitted path, node and example alike, so a deployed copy can be addressed through its sync directory (for example `--prefix .indexer`). |

Behaviour:

- **Exit code `0`** — the corpus validated and the tree was written.
- **Exit code `1`** — the command line was invalid, the corpus could not be
  read, the corpus did not validate, or the output could not be written. Every
  diagnostic goes to stderr, one per line.
- **All diagnostics of a run are reported at once**, in a deterministic order,
  rather than one failure at a time.
- **The write is atomic and all-or-nothing.** The tree goes through a temporary
  file in the output directory that is renamed onto `--output` once the bytes
  are on disk, and nothing is written at all by a run that had anything to
  report. A failed run therefore leaves the previous tree intact.
- **The output is a function of the corpus alone.** Two runs over the same
  corpus produce byte-identical output; so do a relative and an absolute
  `--source` naming the same directory from different working directories.

Build, test and lint from this directory:

```sh
make build   # go build -o indexer .
make test    # unit and integration tests
make lint    # golangci-lint
make fmt     # gofmt -w .
```

## Build environment

**The Makefile targets run against the host Go toolchain, and there is no
Dockerfile. This is a recorded deviation from `GEN-003`, not an oversight.**

`GEN-003` ("code should be built, tested and ran in a containerized
environment") is a **SHOULD**, and the meta rules of `GENERAL.md` permit a
deviation whose reason is noted. The reasons here:

- The approved scope of this module is *binary only: ship `indexer/` plus its own
  Makefile*. No other module in this repository ships a Dockerfile, so adding one
  here would introduce a build path that nothing else in the repository follows.
- The tool indexes the repository it lives in. A containerized run would have to
  bind-mount the repository root back into the container to reach its own corpus,
  so the container would buy reproducibility for a single-binary developer tool
  at the cost of a build path that no longer matches how the tool is invoked —
  from the checkout, against that checkout.

Reproducibility is instead pinned by the `go` directive in `go.mod` and by the
determinism contract above: the output is a function of the corpus alone, and the
integration suite compares two separate process invocations byte for byte.

Revisit this if the module ever ships as a deployed artifact rather than as a
checkout-local tool.

## Discovery rule

A `*.md` file is a standards document **if and only if its very first line opens
a YAML frontmatter block that parses and declares a `title`.** Frontmatter that
starts on a later line, never terminates, does not parse, or carries no `title`
means the file is not a document and no node is emitted for it.

Two kinds of directory are never descended into:

- **`examples/`** — the files under it are cited by the document that owns them
  and appear in the tree as example entries of that document's node, never as
  nodes of their own.
- **Any dot-prefixed directory** (`.git/`, `.claude/`, `.corpora/`, a deployed
  `.indexer/`, and so on).

A `--source` whose final path component is itself a symbolic link is resolved
before the walk begins, so a corpus addressed through a deployed symlink is
indexed as the directory it names rather than descending nowhere. The same
resolution applies to every file the walk visits: a file reached through a
symbolic link is only part of the corpus when its target, once resolved, is
still inside the resolved source root — a link that escapes it is left out of
the corpus silently rather than followed. Emitted paths are unaffected by any
of this: they stay relative to `--source` exactly as it was spelled.

A `--source` that does not name a standards corpus at all fails the run
rather than emitting an empty tree over whatever was deployed before. This is
checked in two ways: once when the walk finds no markdown file at all under
`--source`, and once when markdown files exist but none of them parses into a
standards document — a project checkout or a plain documentation folder, for
example. A corpus whose only document has unparseable frontmatter is exempt
from the second check, since the frontmatter diagnostic already names the
repair needed.

> **Warning.** The rule is about frontmatter, not about location. The repository
> is indexed from its own root, so *any* markdown file anywhere outside those two
> exclusions silently becomes a node the moment it gains a `title:` frontmatter
> key. This is a live hazard for templates and generated documents under `docs/`:
> a plan or task template that grows a frontmatter block joins the standards tree
> and, if it is missing `description`, `scope` or `topics`, fails the whole run.
> Keep such documents free of frontmatter, or place them under a dot-prefixed
> directory.

## Fixture placement

**All test fixture corpora live under `indexer/tests/data/.corpora/<case>/`, and
they must stay there.**

The fixtures are themselves standards corpora — deliberately broken ones, with
unparseable frontmatter, unknown parents and missing example files — and they sit
*inside* the corpus this tool indexes. The single thing keeping them out of the
walk is the leading dot on `.corpora`. Move them to `tests/data/corpora/`, or add
a fixture document anywhere outside a dot-prefixed directory, and indexing this
repository fails with a page of diagnostics about files that are not standards at
all.

`TestIntegrationFixtureCorporaAreNotIndexed` in `integration_test.go` exists for
exactly this: it indexes the repository root while the fixtures are on disk and
asserts a clean run with the expected node count. Do not weaken it.

## Output schema

`docs/standards-tree.yaml` is the **authoritative reference output** and the
schema contract. It is a hand-maintained, commented copy of what indexing this
repository produces, and the integration suite pins its shape: node and example
counts, the single root, the parent relationships, scope lists, the alias map and
its declaration order.

A node carries `path`, `title`, `description`, `scope`, `topics` and `children`;
`aliases` and `examples` are emitted only when the document declares them. An
example entry carries `path`, `title` and the `statements` it illustrates, so an
agent can select individual example files by statement identifier without opening
them. All paths are relative to the corpus root, prefixed when `--prefix` is
given.

## Refreshing the reference tree

When the corpus legitimately changes — a standards document or an example file is
added, removed or re-parented — refresh the reference in the same commit:

1. Regenerate the tree from the repository root:

   ```sh
   cd indexer && make build
   ./indexer --source .. --output /tmp/tree.yaml
   ```

2. Hand-merge `/tmp/tree.yaml` into `docs/standards-tree.yaml`. Do not overwrite
   the file: it keeps a header comment and the flow style of the `statements`
   lists, neither of which the generated output carries. Merge the changed nodes
   only.

3. Update the snapshot constants in `indexer/integration_test.go`
   (`nodeCountSnapshot`, `exampleCountSnapshot`) to the new totals. The
   assertions that use them fail with a message pointing back at this section, so
   a corpus change that forgets the reference tree is caught in CI rather than
   discovered by an agent reading a stale index.

4. Run `make test`, `make lint` and `make fmt` from `indexer/`.
