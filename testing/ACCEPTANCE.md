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
- examples/ACCEPTANCE/feature-tagging.md
---

# Acceptance Testing Standards

## 1. General

`[ACPT-001]` **MUST**: Acceptance tests must be implemented using Gherkin `Given, When, Then` syntax. The specific Gherkin-compatible BDD framework is chosen per language (see `[ACPT-002]`).

`[ACPT-002]` **SHOULD**: Acceptance tests should be implemented in the same language as the application. Golang projects should use the `github.com/cucumber/godog` package. Python projects should use the `behave` package. See `examples/ACCEPTANCE/godog-scaffold.md` for a reference `main_test.go` scaffold.

`[ACPT-003]` **SHOULD**: Acceptance tests should be ran against live DEV environments to ensure that live systems are comprehensively tested.

`[ACPT-004]` **SHOULD**: Acceptance tests should be placed in the `acceptance` directory. `.feature` files should be placed in the `acceptance/features` directory. Step definitions should be placed in the root `acceptance` directory.

`[ACPT-005]` **SHOULD**: A dockerfile and Makefile should be provided to run the acceptance tests so that tests can be ran locally and as part of CI/CD processes interchangeably.

## 2. Feature Files

See `examples/ACCEPTANCE/feature-tagging.md` for reference good and bad feature files covering the statements in this section and section 3.

`[ACPT-006]` **SHOULD**: Each `.feature` file should cover a single functional domain (e.g. `blog.feature`, `contacts.feature`), opening with a `Feature:` name and a one-line, plain-language description of the value it provides.

`[ACPT-007]` **SHOULD**: Scenarios should be tagged with the component or surface they exercise (e.g. `@api`, `@ui`) so that subsets can be run selectively. A tag shared by all scenarios in a file may be applied once at the `Feature:` level instead.

`[ACPT-008]` **SHOULD**: Scenarios should carry a tag that maps them back to a source requirement — a spec, ticket, or user story (e.g. `@spec-002`, `@JIRA-123`) — so that every scenario is traceable to the requirement it verifies. The tag format follows project convention.

## 3. Writing Acceptance Criteria

`[ACPT-009]` **MUST**: Acceptance criteria must be written in plain, non-technical language that is understandable by all stakeholders. Steps must not reference implementation details such as database tables, function names, or internal services. Protocol-level observations (e.g. response status codes, content types) are acceptable where the API itself is the product surface.

`[ACPT-010]` **MUST**: Acceptance criteria must assert on observable behavior — what a user or client can see through the system's public surface (responses, rendered pages, emitted notifications). Where no public read path exists, assertions on externally-meaningful side effects (persisted records, published events, sent notifications) are acceptable; assertions on pure internals (intermediate state, function calls) are not.

`[ACPT-011]` **SHOULD**: Use `Scenario Outline` with `Examples` tables to cover input variations rather than duplicating near-identical scenarios.
