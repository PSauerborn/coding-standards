# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repository is

A corpus of coding-standards documents (`*.md` with YAML frontmatter) meant to be fed to AI coding agents, plus a Go CLI (`indexer/`) that walks the corpus and generates `standards-tree.yaml` — the checked-in index agents read to decide which standards apply to a task.

The standards in this repository govern work *on* this repository too. `GENERAL.md` defines the meta rules: every statement is tagged **MUST** (do not violate without confirming with the user) or **SHOULD** (deviate only with a noted reason).

## Commands

Indexer (run from `indexer/`):

```sh
make build       # go build -o indexer .
make test        # go test ./... (unit + integration)
make lint        # golangci-lint run
make fmt         # gofmt -w .
make fmt-check   # fail if any file is unformatted

go test -run TestName ./...   # single test
```

The indexer deliberately builds against the host Go toolchain with no Dockerfile — a recorded deviation from GEN-003; see "Build environment" in `indexer/README.md` before "fixing" this.

Repository root:

```sh
make index                 # regenerate standards-tree.yaml via ./bin/indexer
make scan-secrets          # detect-secrets scan + audit
pre-commit run --all-files # markdownlint, detect-secrets, file hygiene
```

Slash commands: `/review-standards` (consistency review of the corpus), `/sync-ids` (regenerate and align statement IDs).

Releases have no local command — both are GitHub Actions workflows, triggered manually via `workflow_dispatch` with a required `semver` input (`major.minor.patch`, no `v` prefix or suffix):

- `.github/workflows/release-indexer.yaml` ("Release Indexer") builds `indexer-linux-amd64`, `indexer-linux-arm64` and `indexer-darwin-arm64`, commits them under `bin/` on `main` as `github-actions[bot]`, and tags `v{semver}-indexer`.
- `.github/workflows/release-standards.yaml` ("Release Standards") runs `bin/indexer-linux-amd64 --source . --output standards-tree.yaml` from the repository root, commits `standards-tree.yaml` on `main` as `github-actions[bot]`, and tags `v{semver}-standards`. A regeneration that reproduces the committed tree has nothing to commit, so the commit is skipped and the tag is still pushed.

Both workflows validate the version first and abort before building, committing or tagging anything if the `semver` is malformed or its tag already exists. The binaries released into `bin/` are why `check-added-large-files` carries `exclude: ^bin/` in `.pre-commit-config.yaml`.

## Architecture

**Standards documents.** A markdown file is a standards document if and only if its *first line* opens a YAML frontmatter block that parses and declares a `title`. Frontmatter carries `title`, `description`, `scope` (globs), `topics`, optional `parent`, `aliases`, and `examples`. The `parent:` keys link every document into a single tree rooted at `GENERAL.md`; language/domain directories (`golang/`, `python/`, `databases/`, `general/`, ...) hold the children. Statements inside documents are ID-tagged like `` `[GO-001]` **MUST**: ... ``.

**Example companion files.** Standards are rules-only. Code examples live in `<dir>/examples/<NAME>/<topic>.md` beside the standard `<dir>/<NAME>.md`, one example per file, declared in the standard's `examples:` frontmatter and referenced in-text by directory-relative path. Companion files have **no frontmatter** (they must not become tree nodes) and open with an ID-keyed heading plus a `Statements:` line listing every statement they illustrate. Filenames are named after the topic, never a statement ID, so renumbering never breaks paths.

**Discovery hazards.** The walk skips `examples/` and all dot-prefixed directories, nothing else. Consequences:

- Any markdown file that gains `title:` frontmatter silently joins the standards tree — keep templates and generated documents under `docs/` frontmatter-free, or the run fails on their missing fields.
- Test fixture corpora live in `indexer/tests/data/.corpora/<case>/` and must stay under that dot-prefixed name; they are deliberately broken corpora sitting inside the real one. `TestIntegrationFixtureCorporaAreNotIndexed` guards this.
- The workflow files under `.github/workflows/` sit in a dot-prefixed directory, so they can never join the tree — adding or changing one needs no regeneration and no snapshot bump.

**Changing the corpus.** Adding, removing, or re-parenting a standards document or example file requires, in the same commit: regenerating the tree (`cd indexer && make build`, then run against the repository root), hand-merging changed nodes into `standards-tree.yaml` (it keeps a header comment and flow-style `statements` lists the raw output lacks — merge, don't overwrite), and bumping `nodeCountSnapshot` / `exampleCountSnapshot` in `indexer/integration_test.go`. The integration suite pins the tree's shape and fails otherwise. The Release Standards workflow does not replace any of this: it is the release path, regenerating and committing the tree from the corpus already on `main`, so a corpus change that skipped the hand-merge or the snapshot bump reaches the release as-is.

**Indexer contract.** Output is deterministic (byte-identical across runs and working directories), writes are atomic via temp-file rename, all diagnostics of a run are reported at once to stderr, and a failed run leaves the previous tree untouched. Exit 0 = validated and written; exit 1 = anything else. Don't weaken these properties — the integration suite compares separate process invocations byte for byte.
