---
title: Python REST API Standards
description: Standards for writing REST APIs in Python.
parent: python/GENERAL.md
scope:
- '*.py'
topics:
- python
- api
- rest
- fastapi
examples:
- examples/API/endpoint-structure.md
- examples/API/api-anti-pattern.md
---

# Python REST API Standards

## 1. REST API Guidelines

`[PY-API-001]` **SHOULD**: REST APIs should be implemented using the `FastAPI` package.

`[PY-API-002]` **SHOULD**: REST APIs should be run using `uvicorn`.

`[PY-API-003]` **SHOULD**: Database clients and other dependencies should be initialized within the endpoint handler, when the endpoint is invoked. Prefer a new database/client/service connection for each request as this avoids long-lived connections and reduces the risk of connection leaks. See `examples/API/endpoint-structure.md` for an illustration and `examples/API/api-anti-pattern.md` for an anti-pattern to avoid.

`[PY-API-004]` **SHOULD**: DTO datamodels should be defined separately from domain models using the `pydantic` package, and should include validation for all fields.

`[PY-API-005]` **SHOULD**: Each endpoint should have a `endpoint_name` function. It should return a `JSONResponse` instance that contains the HTTP response code, and body. See `examples/API/endpoint-structure.md` for an illustration and `examples/API/api-anti-pattern.md` for an anti-pattern to avoid.

`[PY-API-006]` **SHOULD**: `Depends` should be used to inject dependencies into endpoint handlers. This ensures that dependencies are initialized only once per request.

`[PY-API-007]` **SHOULD**: CORS middleware should use the `fastapi.middleware.cors` package.

`[PY-API-008]` **SHOULD**: When returning `pydantic` models, use the `model_dump` method with the `mode="json"` argument to ensure that the model is serialized to a JSON object.
