---
title: Javascript Code Standards
description: General standards for writing Javascript applications.
scope: '*.js, *.ts, *.vue'
topics:
- javascript
- node
- vue
---

# 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.
1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

# 2. Versions and Tooling

**MUST**: Node version 20 or higher must be used.

**MUST**: Applications must be written using Vue3.

**MUST**: Applications must use the Quasar component framework.

**SHOULD**: Applications should be written using Javascript.

**SHOULD**: All code should be formatted using `prettier`.

**SHOULD**: All code should be linted using `eslint`.

# 3. Syntax, Naming & Style

**MUST**: All file and functions names must be PascalCase.

**MUST**: All functions must have a JSDoc comment that clearly describes the purpose of the function, its parameters, and its return values. The first word of every JSDoc comment should be the name of the function. This ensures that the JSDoc comment is easily searchable.

**MUST**: Code must follow DRY (Don't Repeat Yourself) principles. This ensures that code is maintainable and easy to understand.

**MUST**: Code must implement components where possible, inline with DRY principles. This minimizes the amount of code that needs to be written and maintained.
