# PostgreSQL Code Standards

This document contains coding standards for applications that work with PostgreSQL. The document outlines a series of **MUST** and **SHOULD**. **MUST**s** are mandatory and must be followed. **SHOULD** are best practices and should be implemented where reasonable. Examples should be treated as **SHOULD**.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

## General

**MUST**: PostgreSQL tables must exist under a dedicated schema. This ensures that the database can have multiple, segregated schemas for different components of the sample application.

**MUST**: PostgreSQL table names and column names must be in snake_case.

**MUST**: Timestamps must be stored in UTC.

**SHOULD**: Persistence layers should favor transactions over auto-commit. This ensures data consistency and prevents partial updates.

**SHOULD**: ID fields should consist of high-entropy values, such as UUIDs. Avoid using incrementing integers as IDs.

**SHOULD**: Columns should have foreign key constraints to other tables where possible to enforce referential integrity.

**SHOULD**: `DELETE` statements should make use of `CASCADE` where applicable to ensure that all related rows are deleted.

**SHOULD**: PostgreSQL tables should have a `created_at` and `updated_at` column to track when the row was created and last updated.
