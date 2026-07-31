# [PY-API-003] Endpoint Structure

Statements: `[PY-API-001]` `[PY-API-002]` `[PY-API-003]` `[PY-API-004]` `[PY-API-005]` `[PY-API-006]` `[PY-API-007]` `[PY-API-008]`

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
