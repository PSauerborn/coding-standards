---
title: Fixture Nested Fence Standards
description: Fixture document whose code samples nest one fenced block inside another.
scope:
- '*'
topics:
- fixtures
---

# Fixture Nested Fence Standards

## 1. Definitions

`[NEST-001]` **MUST**: A definition outside every fence is collected.

## 2. Nested fences

The block below opens with four backticks and closes with four backticks. The
three-backtick block inside it is sample text, not a fence of the document, so
everything between the outer delimiters is a code sample.

````markdown
```go
`[NEST-101]` **MUST**: sample only, defined by nothing.
```
````

`[NEST-002]` **MUST**: A definition after the nested block is collected, which
it is only when the outer block closed where it says it does.

## 3. Mixed delimiters

A tilde fence is closed by tildes alone: the backtick line inside it is sample
text.

~~~text
```
`[NEST-102]` **MUST**: sample only, defined by nothing.
```
~~~

`[NEST-003]` **SHOULD**: A definition after the tilde block is collected.

## 4. Longer closing delimiter

A closing fence may be longer than the one that opened the block.

```text
`[NEST-103]` **MUST**: sample only, defined by nothing.
`````

`[NEST-004]` **SHOULD**: A definition after a block closed by a longer
delimiter is collected.
