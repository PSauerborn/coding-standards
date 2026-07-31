# Code Standards

There are a million ways to do things, and every developer has their own preferences and habits. Like anybody working in tech for a prolonged period of time, I'm very opinionated about how to do things. Any AI that I introduced into the process almost certainly did things differently, and when you know exactly what you want, an AI coding agent that does things differently is just annoying.

In my experience, coding with AI is like writting a book with a very talented, very imaginative 6 year old. It can do some great things, but will inevitably go off the rails unless sufficient guidance is provided. Coding with AI turned a page for me once I treated it like any other developer that was working on one of my projects. Just like a junior dev, code produce by AI became dramatically better once I provided:

1. Clear, consistent coding standards and best practices.
2. Clear examples of good code and why it is good.
3. Clear examples of bad code and why it should be avoided.

Defining a clear set of guidelines and standards not only helped AI produce more consistent code, production-quality code, but it also helped it to produce code in a way that was more aligned with my preferences and habits.

This repository changed the way I develop with AI. I hope it helps you too.

**DISCLAIMER**: The standards are intended to produce code the way I like it. Often there are no black and white decisions on good quality code. Adapt it to your own needs.

## What are `code-standards.md` Files?

`code-standards.md` files are markdown files that contain coding standards, best practices and examples of good code. They outline a series of **MUST** and **SHOULD** statements. **MUST** statements are mandatory and must be followed. `code-standards.md` files are intended to provide AI coding agents and tools with a clear set of instructions to enforce coding standards, best practices, and a consistent code style.

`code-standards.md` are organized into a YAML hierarchy that can easily be parsed by agents using the information contained within each file header using [stdidx](https://github.com/psauerborn/stdidx). This allows agents to
selectively apply standards based on the language, domain, and specific patterns they are working with rather than loading a single monolithic standards file into context.

### Standards Frontmatter

Each `code-standards.md` file contains a YAML frontmatter header with metadata that is used to generate a standards index using [stdidx](https://github.com/psauerborn/stdidx). Each file contains the following metadata:

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

The `title` and `description` fields provide a human-readable summary. The `parent` and `scope` fields allow agents to match standards files to the current task context. See [stdidx](https://github.com/psauerborn/stdidx) for more information on how standards are indexed.

### Examples

`code-standards.md` files are rules-only. Code examples live in companion files under an `examples/` directory alongside the standard, **one example per file**, so that `<dir>/<NAME>.md` has its examples in `<dir>/examples/<NAME>/<topic>.md`:

```text
golang/GENERAL.md
golang/examples/GENERAL/config.md
golang/examples/GENERAL/unittests.md
golang/examples/GENERAL/persistence-postgres.md
```

This keeps the standards themselves small, and keeps the examples selective: an agent loads only the examples for the statements it is actually implementing, rather than every example belonging to the standard. Because each example is its own file, reading the cited file *is* the minimal read — no partial-file discipline is required.

Filenames are named after the **topic**, never a statement ID, so that renumbering statement IDs never invalidates a path.

Companion files have **no frontmatter**. They are not nodes in the standards tree; they are payload that is copied alongside the standard it belongs to. Each one opens with an ID-keyed heading and a `Statements:` line listing every statement it illustrates, so that a search for any covered ID finds the file:

```markdown
# [GO-022] Configuration Loading

Statements: `[GO-018]` `[GO-020]` `[GO-021]` `[GO-022]` `[GO-023]` `[GO-024]`
```

A standard declares its companion files via an `examples:` frontmatter key:

```yaml
examples:
- examples/GENERAL/config.md
- examples/GENERAL/unittests.md
```

In-text references use a path relative to the standard's own directory:

```markdown
See `examples/GENERAL/config.md` for an illustration.
```

Because all linkage is directory-relative, references resolve both in this repository and in a synced copy of the standards tree.

## Licence

Everything in this repository is licensed under the [MIT License](https://choosealicense.com/licenses/mit/). Feel free to use it, modify it, and distribute it as you see fit.

If it was useful to you, please consider starring this repository. If you have any questions, suggestions, or feedback, feel free to get in touch.
