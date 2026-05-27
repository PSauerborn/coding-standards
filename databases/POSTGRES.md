---
title: PostgreSQL Code Standards
description: Standards for PostgreSQL data modeling and schema design.
parent: GENERAL.md
scope:
- '*'
topics:
- postgresql
- relational
- data-modeling
- schema-design
---

# PostgreSQL Code Standards

## 1. Data Modeling

`[PG-001]` **MUST**: PostgreSQL tables must exist under a dedicated schema. This ensures that the database can have multiple, segregated schemas for different components of the sample application.

`[PG-002]` **MUST**: PostgreSQL table names and column names must be in snake_case.

`[PG-003]` **MUST**: Datetime entries must be stored in either UTC ISO8601 format or as unix timestamps.

`[PG-004]` **MUST**: Persistence layers must implement transactions when executing multiple updates. This facilitates atomic operations, and ensures data consistency by preventing partial updates.

`[PG-005]` **SHOULD**: ID fields should consist of high-entropy values, such as UUIDs. Avoid using incrementing integers as IDs.

`[PG-006]` **SHOULD**: Columns should have foreign key constraints to other tables where possible to enforce referential integrity.

`[PG-007]` **SHOULD**: Link tables should be used to model many-to-many relationships, or where multiple tables reference a shared, generic resource. For example, `users` and `jobs` may each link to a generic `documents` table via link tables, avoiding the need to add user- or job-specific foreign keys to `documents`.

`[PG-008]` **SHOULD**: PostgreSQL tables should have a `created_at` and `updated_at` column to track when the row was created and last updated.

`[PG-009]` **SHOULD**: Avoid using complex data types such arrays and JSONB types in postgres tables.
