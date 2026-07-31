# Code Standards

This fixture imitates the repository README: it carries no frontmatter of its
own, but it documents the frontmatter header that standards documents carry, so
its body contains `---`-delimited YAML inside fenced code blocks.

## Standards Frontmatter

Each standards document opens with a YAML frontmatter header:

```yaml
---
title: Golang REST API Standards
description: Standards for writing REST APIs in Go.
parent: golang/GENERAL.md
scope: '*.go'
topics:
- golang
- api
- rest
---
```

The `title` and `description` fields provide a human-readable summary, while
`parent` and `scope` place the document in the tree.

## Examples

Companion example files carry no frontmatter. Each opens with an ID-keyed
heading and a `Statements:` line listing every statement it illustrates:

```markdown
# [GO-022] Configuration Loading

Statements: `[GO-018]` `[GO-020]` `[GO-021]` `[GO-022]`
```

A standard declares its companion files via an `examples:` frontmatter key:

```yaml
examples:
- examples/GENERAL/config.md
- examples/GENERAL/unittests.md
```
