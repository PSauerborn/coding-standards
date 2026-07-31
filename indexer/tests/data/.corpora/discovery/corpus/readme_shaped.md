# Fixture Readme

This document imitates the repository README: it documents the frontmatter and
example conventions by showing them in fenced blocks, and must therefore yield
no node, no citation and no statement identifier.

## Standards Frontmatter

```yaml
---
title: Golang REST API Standards
description: Standards for writing REST APIs in Go.
parent: golang/GENERAL.md
scope: '*.go'
topics:
- golang
- api
examples:
- examples/GENERAL/config.md
---
```

## Examples

```markdown
# [GO-022] Configuration Loading

Statements: `[GO-018]` `[GO-020]` `[GO-021]` `[GO-022]`

`[GO-023]` **SHOULD**: Even a definition-shaped line inside a fenced sample is
never a definition.
```

A standard declares its companion files via an `examples:` frontmatter key.
