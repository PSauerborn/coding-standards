---
title: Fixture General Standards
description: Fixture document used to exercise discovery and statement extraction.
scope:
- '*'
topics:
- fixtures
examples:
- examples/GENERAL/config.md
---

# Fixture General Standards

## 1. Definitions

`[FIX-001]` **MUST**: A backticked identifier anchored at the start of a line and
followed by a marker is a definition.

- `[FIX-002]` **SHOULD**: A definition behind a list marker is a definition too.

## 2. Non-definitions

The statement `[FIX-003]` is only mentioned in prose and must never be collected.

`[FIX-004]` is anchored at the start of a line but carries no marker, so it is
not a definition either.

```go
// `[FIX-005]` **MUST**: this definition lives inside a fenced code block.
func Example() {}
```
