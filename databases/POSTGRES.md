---
title: PostgreSQL Code Standards
description: Standards for PostgreSQL data modeling and schema design.
parent: GENERAL.md
scope: '*'
topics:
- postgresql
- data-modeling
- schema-design
---

# 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.
1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

# 2. Data Modelling

`[PG-001]` **MUST**: PostgreSQL tables must exist under a dedicated schema. This ensures that the database can have multiple, segregated schemas for different components of the sample application.

`[PG-002]` **MUST**: PostgreSQL table names and column names must be in snake_case.

`[PG-003]` **MUST**: Timestamps must be stored in UTC.

`[PG-004]` **SHOULD**: Persistence layers should favor transactions over auto-commit. This ensures data consistency and prevents partial updates.

`[PG-005]` **SHOULD**: ID fields should consist of high-entropy values, such as UUIDs. Avoid using incrementing integers as IDs.

`[PG-006]` **SHOULD**: Columns should have foreign key constraints to other tables where possible to enforce referential integrity.

`[PG-007]` **SHOULD**: `DELETE` statements should make use of `CASCADE` where applicable to ensure that all related rows are deleted.

`[PG-008]` **SHOULD**: PostgreSQL tables should have a `created_at` and `updated_at` column to track when the row was created and last updated.
