# [PY-API-003] REST API Anti-Pattern

Statements: `[PY-API-003]` `[PY-API-004]` `[PY-API-005]` `[PY-API-006]` `[PY-API-007]` `[PY-API-008]`

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
