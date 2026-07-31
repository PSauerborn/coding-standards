---
title: Acceptance Testing Standards
description: Standards for acceptance testing using Gherkin and Cucumber.
parent: GENERAL.md
scope:
- '*'
topics:
- testing
- acceptance-testing
- gherkin
- cucumber
examples:
- examples/ACCEPTANCE/godog-scaffold.md
---

# Acceptance Testing Standards

## 1. General

`[ACPT-001]` **MUST**: Acceptance tests must be implemented using Gherkin `Given, When, Then` syntax. The specific Gherkin-compatible BDD framework is chosen per language (see `[ACPT-002]`).

`[ACPT-002]` **SHOULD**: Acceptance tests should be implemented in the same language as the application. Golang projects should use the `github.com/cucumber/godog` package. Python projects should use the `behave` package. See `examples/ACCEPTANCE/godog-scaffold.md` for a reference `main_test.go` scaffold.

`[ACPT-003]` **SHOULD**: Acceptance tests should be ran against live DEV environments to ensure that live systems are comprehensively tested.

`[ACPT-004]` **SHOULD**: Acceptance tests should be placed in the `acceptance` directory. `.feature` files should be placed in the `acceptance/features` directory. Step definitions should be placed in the root `acceptance` directory.

`[ACPT-005]` **SHOULD**: A dockerfile and Makefile should be provided to run the acceptance tests so that tests can be ran locally and as part of CI/CD processes interchangeably.
