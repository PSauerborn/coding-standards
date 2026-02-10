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

`code-standards.md` are organized into a YAML hierarchy that can easily be parsed by agents using the information contained within each file header. This allows agents to 
selectively apply standards based on the language, domain, and specific patterns they are working with rather than loading a single monolithic standards file into context.

## How to Use `code-standards.md` File Without the Hierarchy?

`code-standards.md` work great with AI assisted coding tools, such as Cursor and Antigravity.

1. Instruct the AI coding agent to look for `code-standards.md` files in the repository before writing any code.
2. Instruct the AI coding agent to enforce the coding standards and best practices outlined in the `code-standards.md` files.
3. Include any relevant `code-standards.md` files in the repository, and start coding.

## How to Use `code-standards.md` File With the Hierarchy?

The hierarchy approach is designed for AI agents that can programmatically read and parse files. Instead of loading every standards file into context, the agent reads `hierarchy.yaml` and selectively loads only the files relevant to the current task.

### File Headers

Each `code-standards.md` file contains a YAML frontmatter header with metadata:

```yaml
---
Title: Golang REST API Standards
Description: Standards for writing REST APIs in Go.
Language: golang
Topics:
- golang
- api
- rest
---
```

The `Title` and `Description` fields provide a human-readable summary. The `Language` and `Topics` fields allow agents to match standards files to the current task context.

### Hierarchy Structure

The `hierarchy.yaml` file organizes all standards into a tree. Each node has:

- **`path`**: Relative path to the standards file.
- **`role`**: Either `root` (read first for a given language/domain) or `leaf` (read after the root).
- **`title`**: The title of the standards file, as defined in the file header.
- **`description`**: A short description of what the file covers.
- **`topics`**: A list of topics covered by the file, used for matching.

Parent nodes group related children under a common `path` prefix (e.g. `golang/`).

### Agent Workflow

1. **Read `hierarchy.yaml`** at the start of every task before writing or modifying any code.
2. **Identify relevant nodes** by matching the language, domain, or topics to the current task (e.g. if writing a Go REST API, match `golang` and `api`).
3. **Read root nodes first**. For any matched language/domain, always read the `root` node before any `leaf` nodes. Root nodes contain foundational standards that apply broadly.
4. **Read matched leaf nodes**. After the root, read any `leaf` nodes whose topics match the specific pattern being implemented.
5. **Apply all loaded standards** when writing or reviewing code. **MUST** rules are mandatory; **SHOULD** rules are recommended but can be deviated from with justification.

For example, when building a Go REST API, an agent should:
1. Read `hierarchy.yaml`
2. Load `general.md` (cross-language root standards)
3. Load `golang/GENERAL.md` (Go root standards)
4. Load `golang/API.md` (Go REST API leaf standards)

This ensures the agent has all relevant context without loading unrelated standards (e.g. Python, Terraform).


## Licence

Everything in this repository is licensed under the [MIT License](https://choosealicense.com/licenses/mit/). Feel free to use it, modify it, and distribute it as you see fit.

If it was useful to you, please consider starring this repository. If you have any questions, suggestions, or feedback, feel free to get in touch.