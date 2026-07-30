---
title: General Code Standards
description: Cross-language general coding standards and best practices.
scope:
- '*'
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
  deploy: deployment
  containerization: container
---

# General Code Standards

## 1. Meta Rules

Each rule is tagged **MUST** or **SHOULD**:

- **MUST** — hard requirement. Do not violate without confirming with the user.
- **SHOULD** — strong default. Deviate if the situation genuinely calls for it; briefly note the reason when you do.

Rules often include a short rationale. When a case isn't explicitly covered, apply the rationale rather than pattern-matching the literal wording.

**Examples**: each section may include `Good` and `Bad` code samples. The `Bad` samples show real failure modes — not just style preferences. Match the `Good` shape.

**Conflict Resolution**:

- User request vs. **SHOULD** : follow the user.
- User request vs. **MUST** : stop and confirm before proceeding.

## 2. General Guidelines

`[GEN-001]` **MUST**: When making design and implementation choices, follow KISS (Keep It Simple, Stupid) and YAGNI (You Ain't Gonna Need It) principles. Submit any architectural decisions for review before implementing them to ensure that high-level decisions are made by users.

`[GEN-002]` **MUST**: Complexity introduces significant cost, engineering debt and risk. Prefer solutions and implementations that minimize entropy. If a situation calls for a more complex architecture, whether for security reasons or otherwise, submit your suggestions for explicit review before making any changes.

`[GEN-003]` **SHOULD**: Code should be built, tested and ran in a containerized environment, preferably using Docker. This ensures consistency and reproducibility.

`[GEN-004]` **SHOULD**: Makefiles should be used to define common build and test targets, as well as linting and formatting.

`[GEN-005]` **SHOULD**: `pre-commit` hooks should be used to enforce coding standards and best practices. At minimum, include the following hooks:

```yaml
# File: .pre-commit-config.yaml
# GOOD: include standard hooks that apply to any project
repos:
- repo: https://github.com/pre-commit/pre-commit-hooks
  rev: v2.3.0
  hooks:
  - id: check-yaml
  - id: end-of-file-fixer
  - id: trailing-whitespace
  - id: check-added-large-files
  - id: check-case-conflict
  - id: check-json

- repo: https://github.com/yelp/detect-secrets
  rev: v1.5.0
  hooks:
  - id: detect-secrets
    args: ['--baseline', '.secrets.baseline']

```

Add any language-specific hooks for formatting/linting on project-by-project basis.

`[GEN-006]` **SHOULD**: Projects should have dedicated integration tests.
