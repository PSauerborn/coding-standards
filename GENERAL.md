---
title: General Code Standards
description: Cross-language general coding standards and best practices.
scope: '*'
topics:
- docker
- makefiles
- pre-commit
- integration-tests
aliases:
  go: golang
  js: javascript
  ts: typescript
  py: python
  postgres: postgresql
  pg: postgresql
---

# General Code Standards

# 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.
1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

# 2. General Guidelines

`[GEN-001]` **SHOULD**: All design and implementation choices should follow KISS (Keep It Simple, Stupid) and YAGNI (You Ain't Gonna Need It) principles.

`[GEN-002]` **SHOULD**: Complexity intruduces significant cost, engineering debt and risk. Prefer solutions and implementations that minimize entropy.

`[GEN-003]` **SHOULD**: Code should be built, tested and ran in a containerized environment, preferably using Docker. This ensures consistency and reproducibility.

`[GEN-004]` **SHOULD**: Makefiles should be used to define build and test targets.

`[GEN-005]` **SHOULD**: `pre-commit` hooks should be used to enforce coding standards and best practices.

`[GEN-006]` **SHOULD**: Projects should have dedicated integration tests.
