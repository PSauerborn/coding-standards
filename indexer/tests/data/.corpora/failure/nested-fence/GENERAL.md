---
title: Nested Fence Fixture
description: A corpus whose document shows an identifier only inside a nested code fence.
scope:
- '*'
topics:
- fixture
examples:
- examples/nested.md
---

# Nested Fence Fixture

`[FIX-005]` **MUST**: Be defined outside every fenced code block.

The block below opens and closes with four backticks and wraps a three-backtick
block. Everything between the outer delimiters is a code sample, so the
identifier shown in it is defined by nothing and no example may cite it.

````markdown
```go
`[FIX-006]` **MUST**: sample only, defined by nothing.
```
````
