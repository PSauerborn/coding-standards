---
title: Javascript Code Standards
description: General standards for writing Javascript applications.
parent: GENERAL.md
scope:
- '*.js'
- '*.ts'
- '*.vue'
topics:
- javascript
- quasar
- typescript
- node
- vue
---

# Javascript Code Standards

## 1. Versions and Tooling

`[JS-001]` **MUST**: Node version 22 or higher must be used.

`[JS-002]` **MUST**: Applications must be written using Vue3.

`[JS-003]` **MUST**: Applications must use the Quasar component framework.

`[JS-004]` **SHOULD**: Applications should be written using Javascript.

`[JS-005]` **SHOULD**: All code should be formatted using `prettier`.

`[JS-006]` **SHOULD**: All code should be linted using `eslint`.

`[JS-007]` **SHOULD**: Applications should be styled with SCSS.

## 2. Syntax, Naming & Style

`[JS-008]` **MUST**: All file and function names must be camelCase.

`[JS-009]` **MUST**: All functions must have a JSDoc comment that clearly describes the purpose of the function, its parameters, and its return values. The first word of every JSDoc comment should be the name of the function. This ensures that the JSDoc comment is easily searchable.

`[JS-010]` **MUST**: Code must follow DRY (Don't Repeat Yourself) principles. This ensures that code is maintainable and easy to understand.

`[JS-011]` **MUST**: Code must implement components where possible, inline with DRY principles. This minimizes the amount of code that needs to be written and maintained.

## 3. Deployment

`[JS-012]` **MUST**: Frontend applications must be deployed using docker.

`[JS-013]` **MUST**: SPA applications should be built and served via NGINX.

`[JS-014]` **MUST**: SSR applications should be built and served via a node server.
