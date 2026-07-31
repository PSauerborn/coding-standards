# [API-002] Error Response Structure

Statements: `[API-002]` `[API-004]`

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
