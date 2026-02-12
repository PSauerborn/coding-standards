---
title: General Code Standards
description: Cross-language general coding standards and best practices.
scope: '*'
topics:
- containerization
- docker
- makefiles
- pre-commit
- integration-tests
---

# General Code Standards

# 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.
1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

# 2. General Guidelines

**SHOULD**: All design and implementation choices should follow KISS (Keep It Simple, Stupid) and YAGNI (You Ain't Gonna Need It) principles.

**SHOULD**: Complexity intruduces significant cost, engineering debt and risk. Prefer solutions and implementations that minimize entropy.

**SHOULD**: Code should be built, tested and ran in a containerized environment, preferably using Docker. This ensures consistency and reproducibility.

**SHOULD**: Makefiles should be used to define build and test targets.

**SHOULD**: `pre-commit` hooks should be used to enforce coding standards and best practices.

**SHOULD**: Projects should have dedicated integration tests.
