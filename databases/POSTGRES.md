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

# 1. Data Modeling

`[PG-001]` **MUST**: PostgreSQL tables must exist under a dedicated schema. This ensures that the database can have multiple, segregated schemas for different components of the sample application.

`[PG-002]` **MUST**: PostgreSQL table names and column names must be in snake_case.

`[PG-003]` **MUST**: Timestamps must be stored in UTC.

`[PG-004]` **SHOULD**: Persistence layers should favor transactions over auto-commit. This ensures data consistency and prevents partial updates.

`[PG-005]` **SHOULD**: ID fields should consist of high-entropy values, such as UUIDs. Avoid using incrementing integers as IDs.

`[PG-006]` **SHOULD**: Columns should have foreign key constraints to other tables where possible to enforce referential integrity.

`[PG-007]` **SHOULD**: `DELETE` statements should make use of `CASCADE` where applicable to ensure that all related rows are deleted.

`[PG-008]` **SHOULD**: PostgreSQL tables should have a `created_at` and `updated_at` column to track when the row was created and last updated.
