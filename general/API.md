---
title: REST API General Standards
description: General standards for REST APIs.
scope:
- '*'
parent: GENERAL.md
topics:
- api
- rest
- http
---

# REST API General Standards

## 1. General Guidelines

`[API-001]` **MUST**: REST APIs must accept and return JSON data. Exceptions can be made for file uploads and responses where binary data is required.

`[API-002]` **MUST**: REST APIs must structure responses in a consistent manner. The same shape must be used for all success responses and the same shape for all error responses.

`[API-003]` **MUST**: REST APIs must have a version prefix in the URL. This ensures that APIs can be versioned and that old APIs can be deprecated without breaking existing clients.

`[API-004]` **MUST**: Error responses must contain an `error` field and an optional `details` field. The `error` field must contain a generic error message (e.g. "Internal Server Error", "Bad Request"). The `details` field should contain additional context where applicable. Error responses __must not__ leak any implementation details. If in doubt, keep error details generic to avoid potential data leaks.

`[API-005]` **MUST**: REST APIs must expose a `/health` endpoint that checks the health of the API and its dependencies. The endpoint should return `200 OK` when all dependencies are healthy. The `/health` endpoint should explicitly check database connections if applicable as part of the health check.

`[API-006]` **MUST**: REST APIs must expose a `/version` endpoint that returns the version of the API.

`[API-007]` **MUST**: REST APIs must have an associated `openapi.yaml` file that defines the API contract. All endpoints must be documented in this file.

`[API-008]` **SHOULD**: REST API endpoints should follow a dependency injection pattern. Prefer initializing per-request dependencies (database connections, service clients) in the endpoint handler rather than inside business logic.

`[API-009]` **SHOULD**: CORS should be enabled for REST APIs by default. The language-specific implementation determines which CORS middleware/package to use.

`[API-010]` **SHOULD**: Registration of endpoints should be kept minimal — routing, dependency creation, and error handling only. Business logic must live outside the endpoint registration.

`[API-011]` **SHOULD**: Each endpoint should have its own unit test.

### Example

The following example illustrates how error responses should be structured.

```json
// GOOD: 500 — server-side failure
{
    "error": "Internal Server Error",
    "details": "failed to connect to database"
}

// GOOD: 400 — validation failure
{
    "error": "Bad Request",
    "details": "invalid email format for field 'email'"
}

// GOOD: 404 — `details` is optional
{
    "error": "Not Found"
}
```

The following example illustrates how error responses should NOT be structured.

```json
// BAD: uses `msg` instead of `error`
{
    "msg": "failed to connect to database"
}

// BAD: leaks implementation details into the generic `error` field
{
    "error": "database connection to host=db.example.com:5432 timed out after 30s"
}

// BAD: not an object — clients cannot parse a consistent shape
"Internal Server Error"

// BAD: alternative envelope; spec calls for `error` + optional `details`
{
    "status": "error",
    "code": 500,
    "data": null
}
```
