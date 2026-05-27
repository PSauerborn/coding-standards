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
---

# Python REST API Standards

## 1. REST API Guidelines

`[PY-API-001]` **SHOULD**: REST APIs should be implemented using the `FastAPI` package.

`[PY-API-002]` **SHOULD**: REST APIs should be run using `uvicorn`.

`[PY-API-003]` **SHOULD**: Database clients and other dependencies should be initialized within the endpoint handler, when the endpoint is invoked. Prefer a new database/client/service connection for each request as this avoids long-lived connections and reduces the risk of connection leaks. See `Example 1` for an illustration.

`[PY-API-004]` **SHOULD**: DTO datamodels should be defined separately from domain models using the `pydantic` package, and should include validation for all fields.

`[PY-API-005]` **SHOULD**: Each endpoint should have a `endpoint_name` function. It should return a `JSONResponse` instance that contains the HTTP response code, and body. See `Example 1` for an illustration.

`[PY-API-006]` **SHOULD**: `Depends` should be used to inject dependencies into endpoint handlers. This ensures that dependencies are initialized only once per request.

`[PY-API-007]` **SHOULD**: CORS middleware should use the `fastapi.middleware.cors` package.

`[PY-API-008]` **SHOULD**: When returning `pydantic` models, use the `model_dump` method with the `mode="json"` argument to ensure that the model is serialized to a JSON object.

### Example 1

The following example illustrates how a REST API should be structured.

```python
# GOOD
# main.py
import structlog
import psycopg
from fastapi import FastAPI, Depends, Request
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware

from models import User, NewUserRequest
# GOOD: alias imports to avoid shadowing handler names below
from persistence import (
    new_postgres_connection,
    create_user as persist_create_user,
    get_user_by_id as persist_get_user,
)
from config import CONFIG

LOGGER = structlog.get_logger()


APP = FastAPI()
# GOOD: CORS is enabled by default.
APP.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


def get_current_user(request: Request) -> User:
    """Retrieves the current user from the request.

    Args:
        request (Request): The incoming request.

    Returns:
        User: The current user.
    """

    return request.state.user


@APP.get("/v1/health")
def health() -> JSONResponse: # GOOD: Return type is specified.
    """Handles health checks by pinging the database.

    Returns:
        JSONResponse: A JSON response indicating the health status.
    """

    LOGGER.info("Processing new request.", method="GET", endpoint="/health")
    # GOOD: Use JSONResponse to return a JSON response.
    return JSONResponse(status_code=200, content={"data": "ok"})


@APP.get("/v1/users/me")
def get_user(db: psycopg.Cursor = Depends(new_postgres_connection), user: User = Depends(get_current_user)) -> JSONResponse: # GOOD: Return type is specified.
    """Handles the retrieval of the current user.

    Args:
        db (psycopg.Cursor): The database cursor.
        user (User): The current user.

    Returns:
        JSONResponse: A JSON response containing the user data.
    """

    # GOOD: logging is implemented using structlog.
    LOGGER.info("Processing new request.", method="GET", endpoint="/users/me", user_id=user.user_id)

    # GOOD: aliased persistence function avoids recursive call into this handler
    user = persist_get_user(db, user.user_id)
    if user is None:
        LOGGER.error("User not found.", user_id=user.user_id)
        return JSONResponse(status_code=404, content={"error": "User not found"})

    LOGGER.info("User retrieved.", user=user)

    # GOOD: use model_dump to serialize pydantic model to json
    data = user.model_dump(mode="json")
    return JSONResponse(status_code=200, content={"data": data})


@APP.post("/v1/users")
# GOOD: Dependencies are injected via Depends.
# GOOD: DTO is defined separately from domain model.
def create_user(user: NewUserRequest, db: psycopg.Cursor = Depends(new_postgres_connection)) -> JSONResponse: # GOOD: Return type is specified.
    """Handles the creation of a new user.

    Args:
        user (NewUserRequest): The new user request object.
        db (psycopg.Cursor): The database cursor.

    Returns:
        JSONResponse: A JSON response containing the ID of the newly created user.
    """

    # GOOD: logging is implemented using structlog.
    LOGGER.info("Processing new request.", method="POST", endpoint="/users", user=user)

    # GOOD: aliased persistence function avoids recursive call into this handler
    user_id = persist_create_user(db, user.username, user.email)
    LOGGER.info("User created.", user_id=user_id)

    return JSONResponse(status_code=201, content={"data": user_id})


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(APP, host="0.0.0.0", port=CONFIG.PORT)
```

### Example 2

The following example illustrates how REST APIs should NOT be structured:

```python
# BAD
# main.py
import psycopg
from fastapi import FastAPI, Request

from models import User, NewUserRequest
from persistence import new_postgres_connection, create_user
from config import CONFIG


APP = FastAPI()
# BAD: CORS is not enabled.


def get_current_user(request: Request) -> User:
    """Retrieves the current user from the request.

    Args:
        request (Request): The incoming request.

    Returns:
        User: The current user.
    """

    return request.state.user


@APP.get("/v1/health")
def health(): # BAD: Return type is not specified.
    """Handles health checks by pinging the database."""

    # BAD: logging is not implemented.
    print("Processing new request.", method="GET", endpoint="/health")

    # BAD: Use JSONResponse to return a JSON response.
    # BAD: response does not match defined contract.
    return {"msg": "ok"}


@APP.get("/v1/users/me")
# BAD: handler shadows imported get_user and recurses into itself below
def get_user():
    """Handles the retrieval of the current user.

    Args:
        db (psycopg.Cursor): The database cursor.
        user (User): The current user.

    Returns:
        JSONResponse: A JSON response containing the user data.
    """

    # BAD: Dependencies should be injected via Depends.
    user = get_current_user(request)
    # BAD: logging is not implemented.
    print(f"Processing request to retrieve user: {user}")

    # BAD: Dependencies should be injected via Depends.
    db = new_postgres_connection()

    # BAD: recursive call shadows imported get_user
    user = get_user(db, user.user_id)
    if user is None:
        print("User not found.", user_id=user.user_id)
        return JSONResponse(status_code=404, content={"error": "User not found"})

    print(f"User retrieved: {user}")

    # BAD: pydantic models should be serialized with model_dump(mode="json") before being returned
    # BAD: JSONResponse should be used to return a JSON response.
    # BAD: response does not match defined contract.
    return user


@APP.post("/v1/users")
# BAD: Dependencies should be injected via Depends.
# BAD: DTO is defined separately from domain model.
# BAD: handler shadows imported create_user and recurses into itself below
def create_user(user: User): # BAD: Return type is not specified.
    """Handles the creation of a new user.

    Args:
        user (User): The new user request object.
    """

    # BAD: Dependencies should be injected via Depends.
    db = new_postgres_connection()

    # BAD: logging should be implemented using structlog.
    print(f"Processing request to create new user: {user}")

    # BAD: recursive call shadows imported create_user
    user_id = create_user(db, user.username, user.email)
    print(f"User created with id: {user_id}")

    # BAD: JSONResponse should be used to return a JSON response.
    return {"data": user_id}


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(APP, host="0.0.0.0", port=CONFIG.PORT)
```
