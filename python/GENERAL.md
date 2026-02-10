---
Title: Python Code Standards
Description: General standards for writing Python applications.
Language: python
Topics:
- python
- pydantic
- fastapi
- pytest
- structlog
- persistence
---

# 1. Meta Rules

You are a Senior Software Engineer acting as an autonomous coding agent.
1.  **Strict Adherence**: You MUST follow all **MUST** rules below.
2.  **Pattern Matching**: When writing code, check the "Example" sections. If you are tempted to write code that looks like a "BAD" example, STOP and refactor to match the "GOOD" example.
3.  **Explanation**: If you deviate from a **SHOULD** rule, you must explicitly state why in your reasoning trace.

If a user request contradicts a **SHOULD** statement, follow the user request. If it contradicts a **MUST** statement, ask for confirmation.

# 2. Versions and Tooling

**MUST**: Python version 3.13 or higher must be used.

**MUST**: When in doubt, follow PEP 8.

**SHOULD**: All code should be formatted using `black`.

**SHOULD**: All code should be linted using `flake8`.

# 3. Syntax, Naming & Style

**MUST**: All file and functions names must be snake_case. This ensures adherence to PEP 8.

**MUST**: All functions must have a doc string that clearly describes the purpose of the function, its parameters, and its return values. The first word of every doc string should be the name of the function. This ensures that the doc string is easily searchable.

**MUST**: All doc strings must be formatted using Google style.

**MUST**: Type hints must be provided for all function parameters and return values.

**MUST**: Use spaces for indentation instead of tabs.

**SHOULD**: Prefer a functional approach when possible.

**SHOULD**: Prefer a single return type per function. This ensures that function complexity is kept minimal.

**SHOULD**: Source code should be placed in a `src` directory. A single `main.py` or `app.py` file should serve as the entry point at the root of the `src` directory.

# 4. Data Models and Validation

**MUST**: Data models must be defined in a dedicated `models.py` file.

**MUST**: Data models must be implenented using `pydantic`.

**SHOULD**: Data models should be used throught the application to group related data. This ensures that the code base is more type-safe, and makes the code more readable.

**SHOULD**: An alias should be defined for fields using the `pydantic` `Field` class if applicable. Only add aliases if it is necessary to ensure that the model is compatible with either the API or the database schema.

**SHOULD**: DTOs and domain models should be defined separately. This ensures that API logic is decoupled from database/storage logic.

**SHOULD**: DTOs should have strict validation rules to ensure that they are valid before being used in business logic. This ensures that external data is validated before being used in business logic.

**SHOULD**: Prefer length-constrained strings over just `str` types in datamodels. `None` values should be preferred to empty strings. Empty strings are ambiguous, and it is not clear whether they should be considered as "empty" or "not set".

## Example 1

The following illustrates a data model that follows the above recommendations.

```python
# GOOD
# File: models.py
from datetime import datetime

from pydantic import BaseModel, StringConstraints, Field
from typing_extensions import Annotated


# GOOD: User is a domain model that implements a pydantic model
# GOOD: Adequate constraints are used for string fields
# GOOD: Aliases are used for fields to ensure that the model is compatible with the API
# GOOD: Default values are used for optional fields
class User(BaseModel):
    user_id: Annotated[str, StringConstraints(min_length=1), Field(alias="userId")] 
    username: Annotated[str, StringConstraints(min_length=1)] 
    email: Annotated[str, StringConstraints(min_length=1)] 
    created_at: Annotated[datetime, Field(alias="createdAt")] 
    updated_at: Annotated[datetime | None, Field(alias="updatedAt", default=None)]
    created_by: Annotated[str, StringConstraints(min_length=1), Field(alias="createdBy")] 
    last_updated_by: Annotated[str, StringConstraints(min_length=1), Field(alias="lastUpdatedBy")] 

```

# 5. Configuration

**MUST**: Configuration must be handled via environment variables. This ensures that configuration is decoupled from code and can be easily changed without modifying code.

**MUST**: Configuration must be validated at application startup to ensure that all required variables are set. This ensures that the application fails fast if a required variable is not set.

**MUST**: Configuration settings must be defined in a dedicated `config.py` file.

**MUST**: Configuration must be parsed and validated using the `pydantic_settings` package. A `Config` class must be defined in the `config.py` file that implements the `BaseSettings` class.

**MUST**: A global instance of the `Config` class must be created and made available to the application. This ensures that configuration settings can be easily accessed from anywhere in the application, and that configuration settings are validated at application startup.

### Example 2

The following example illustrates how application configuration should be handled.

```python
# GOOD
# File: config.py
from pydantic import StringConstraints, Field, SecretStr
from pydantic_settings import BaseSettings
from typing_extensions import Annotated


# GOOD: Config is a pydantic model that defines the configuration for the application
# GOOD: use adequate constraints for string fields
class Config(BaseSettings):
    LOG_LEVEL: Annotated[str, StringConstraints(min_length=1), Field(default="INFO")]
    PG_PORT: Annotated[int, Field(default=5432)]
    PG_HOST: Annotated[str, StringConstraints(min_length=1), Field(default="localhost")]
    PG_USER: Annotated[str, StringConstraints(min_length=1)]
    PG_PASSWORD: Annotated[SecretStr, StringConstraints(min_length=1)]
    PG_DB: Annotated[str, StringConstraints(min_length=1), Field(default="postgres")]


# GOOD: CONFIG is a global instance of the Config class that is made available to the application
CONFIG = Config()

```

# 6. Unittests

**MUST**: Unittests must be implemented for all business logic.

**MUST**: Unittests must be stored in a `tests` directory.

**MUST**: Unittests must be implemented using the `pytest` package.

**MUST**: Each `.py` file should have a corresponding `_test.py` file that contains unittests.

**SHOULD**: The `tests` directory should contain a `conftest.py` file that contains common test fixtures.

**SHOULD**: Each function should have a corresponding unittest. Unittests should be named in the format `test_function_name`.

**SHOULD**: Unittests should mock database and other service connections rather than using real connections. Connection to live databases should only be used in integration tests. This ensures that unittests are fast and do not depend on external services.

**SHOULD**: Additional test data used to test business logic (PDF, CSV files etc) should be stored in a separate `tests/data` directory. This ensures that test files do not become cluttered with test data.

### AWS Applications

**MUST**: Any applications using AWS resources must use the `moto` package to mock AWS resources.

**MUST**: Any AWS resources that can be implemented using the `moto` package must be. This ensures that unittests can test interaction with AWS resources as well as possible.

**MUST**: When using DynamoDB, a `tests/data/table_schema.json` file must be provided that defines the schema of the table.

**MUST**: When using DynamoDB, a `tests/data/initial_db_items.json` file must be provided that defines the initial items in the table.

**MUST**: When using DynamoDB, a `mock_table` `pytest` fixture should be implemented in the `conftest.py` file. This should create the table using the schema defined in `tests/data/table_schema.json` and populate it with the items defined in `tests/data/initial_db_items.json`. See Example 3 for an implementation.

### Example 3

The following example illustrates a unittest implementation for a function that uses DynamoDB.

```python
# GOOD
# File: tests/confest.py
import json

import pytest
import boto3
from moto import mock_aws


@pytest.fixture
# GOOD: use moto to mock AWS resources
def mock_dynamodb_resource():
    with mock_aws():
        yield boto3.resource("dynamodb")

@pytest.fixture
# GOOD: create table using schema defined in tests/data/table_schema.json
# GOOD: populate table with items defined in tests/data/initial_db_items.json
def mock_table():
    with open("tests/data/table_schema.json", "r") as f:
        table_schema = json.load(f)

    table = mock_dynamodb_resource.create_table(**table_schema)

    with open("tests/data/initial_db_items.json", "r") as f:
        initial_db_items = json.load(f)

    for item in initial_db_items:
        table.put_item(Item=item)
    
    yield table

```

# 7. Logging

**MUST**: All applications must implement logging. Logging should be present at all levels of the application.

**MUST**: Logging must follow a structured logging format. All log messages should contain at minimum the timestamp, log level, and message.

**SHOULD**: Logging should be handled via the `structlog` package. See Example 4 for an implementation.

### Example 4

The following example illustrates a logging implementation.

```python
# GOOD
# File: main.py
import structlog # GOOD: use structlog for logging

LOGGER = structlog.get_logger()


def add(a: int, b: int) -> int:
    """Adds two numbers.

    Args:
        a (int): The first number.
        b (int): The second number.

    Returns:
        int: The sum of the two numbers.
    """

    # GOOD: implement logging at all levels of the application
    LOGGER.info("Adding two numbers", a=a, b=b)
    return a + b


def main():
    # GOOD: implement logging at all levels of the application
    LOGGER.info("Application started", application="my_app", version="1.0.0")

    add(1, 2)


if __name__ == "__main__":
    main()

```

The following example illustrates how logging should NOT be implemeted:

```python
# BAD
# File: main.py

def add(a: int, b: int) -> int:
    """Adds two numbers.

    Args:
        a (int): The first number.
        b (int): The second number.

    Returns:
        int: The sum of the two numbers.
    """

    # BAD: logging not implemented
    return a + b


def main():
    # BAD: logging should be implemented via structlog
    print("Application started version 1.0.0")
    add(1, 2)

```

# 8. Persistence Layers

**MUST**: Persistence layers must have their own dedicated file that contains all storage logic.

**SHOULD**: Prefer transactional operations that execute multiple database operations as a single unit of work. This ensures that database operations are atomic and consistent, and minimizes the risk of incomplete or inconsistent data.

**SHOULD**: Prefer a functional approach when implementing persistence layers.

**SHOULD**: Persistence layers should consist of a series of functions that take the database client as the first argument, then execute any required database operations.

**SHOULD**: Persistence layers should return domain models where multiple input and return values are required. See `Example 5` for an illustration.

**SHOULD**: PostgreSQL should be used by default if not otherwise specified.

### Example 5

The following example illustrates a persistence layer for a generic PostgreSQL database.

```python
# GOOD
# File: persistence.py
from uuid import uuid4

import psycopg 
from psycopg.rows import dict_row

from .config import CONFIG
from .models import User


def new_postgres_connection() -> tuple[psycopg.Connection, psycopg.Cursor]:
    """Creates a new connection to the PostgreSQL database based on the configuration.

    Returns:
        tuple[psycopg.Connection, psycopg.Cursor]: A tuple containing the connection and cursor.
    """

    # GOOD: disable autocommit in favor of transactions
    conn = psycopg.connect(
        host=CONFIG.POSTGRES_HOST,
        port=CONFIG.POSTGRES_PORT,
        user=CONFIG.POSTGRES_USER,
        password=CONFIG.POSTGRES_PASSWORD,
        dbname=CONFIG.POSTGRES_DB,
        autocommit=False,
    )
    # GOOD: use dict_row cursor factory
    return conn, conn.cursor(row_factory=dict_row)


# GOOD: persistence function takes database client as first argument
# GOOD: type hints for input and output values
def get_user_by_id(cursor: psycopg.Cursor, user_id: int) -> User | None:
    """Retrieves a user by their ID.

    Args:
        cursor (psycopg.Cursor): The database cursor.
        user_id (int): The ID of the user to retrieve.

    Returns:
        User | None: The user object if found, otherwise None.
    """

    cursor.execute(
        "SELECT * FROM users WHERE id = %s",
        (user_id,),
    )
    user = cursor.fetchone()
    # GOOD: return pydantic model
    return User(**user) if user else None


def create_user(cursor: psycopg.Cursor, username: str, email: str) -> str:
    """Creates a new user.

    Args:
        cursor (psycopg.Cursor): The database cursor.
        username (str): The username of the new user.
        email (str): The email of the new user.

    Returns:
        str: The ID of the newly created user.
    """

    user_id = str(uuid4()).replace("-", "")
    cursor.execute(
        "INSERT INTO users (id, username, email) VALUES (%s, %s, %s) RETURNING *",
        (user_id, username, email),
    )
    user = cursor.fetchone()
    return user_id

```

The following example illustrates how persistence layers should NOT be implemented:

```python
# BAD
# File: persistence.py
import psycopg


def new_postgres_connection() -> tuple[psycopg.Connection, psycopg.Cursor]:
    """Creates a new connection to the PostgreSQL database based on the configuration.

    Returns:
        tuple[psycopg.Connection, psycopg.Cursor]: A tuple containing the connection and cursor.
    """

    # BAD: autocommit should be disabled in favor of transactions
    conn = psycopg.connect(
        host=CONFIG.POSTGRES_HOST,
        port=CONFIG.POSTGRES_PORT,
        user=CONFIG.POSTGRES_USER,
        password=CONFIG.POSTGRES_PASSWORD,
        dbname=CONFIG.POSTGRES_DB,
        autocommit=True,
    )
    # BAD: postgresql connections should use dict_row cursor factory
    return conn, conn.cursor()


# BAD: prefer a functional approach over a class
class Persistence():

    def __init__(self):
        self.conn, self.cursor = new_postgres_connection()

    def get_user_by_id(self, user_id: int) -> User | None:
        return get_user_by_id(self.cursor, user_id)

    def create_user(self, username: str, email: str) -> str:
        return create_user(self.cursor, username, email)



def create_user(username: str, email: str) -> str:
    """Creates a new user.

    Args:
        username (str): The username of the new user.
        email (str): The email of the new user.

    Returns:
        str: The ID of the newly created user.
    """

    # BAD: persistence layers should use dependency injection.
    # Database connections must NOT be created inside the persistence layer.
    conn, cursor = new_postgres_connection()

    user_id = str(uuid4()).replace("-", "")
    cursor.execute(
        "INSERT INTO users (id, username, email) VALUES (%s, %s, %s) RETURNING *",
        (user_id, username, email),
    )
    user = cursor.fetchone()
    conn.commit()
    return user_id


# BAD: persistence functions should return pydantic models
def get_user(user_id: int):
    """Retrieves a user by their ID.

    Args:
        user_id (int): The ID of the user to retrieve.
    """

    # BAD: persistence layers should use dependency injection.
    # Database connections must NOT be created inside the persistence layer.
    conn, cursor = new_postgres_connection()
    cursor.execute(
        "SELECT * FROM users WHERE id = %s",
        (user_id,),
    )
    user = cursor.fetchone()
    return user if user else None

```

### PostgreSQL

**MUST**: PostgreSQL persistence layers must be implemented using the `psycopg` package. This enforces consistency across all applications.

**SHOULD**: Autocommit should be disabled by default in favor of transactions. This ensures that database operations are atomic and consistent and minimizes the risk of incomplete or inconsistent data.

**SHOULD**: `psycopg` connections should use a `dict_row` cursor factory. This ensures that database operations return a dictionary of column names and values that can be fed directly into `pydantic` models. See `Example 3` for an illustration.

### DynamoDB

**MUST**: DynamoDB persistence layers must be implemented using the `boto3` package. This enforces consistency across all applications.

**SHOULD**: Prefer using `boto3` table resources instead of clients. This ensures that responses from DynamoDB can be passed directly into `pydantic` models without requiring parsing of DynamoDB types. If a `boto3` client is required (such as when using `BatchGetItem`), extract the client from the table object using `table.meta.client`.

### Example 6

The following example illustrates how to structure DynamoDB persistence layers:

```python
# GOOD
# File: dynamodb.py
import boto3
from boto3.dynamodb.table import TableResource

from .models import User


def get_table(name: str) -> TableResource:
    """Retrieves a DynamoDB table resource.

    Args:
        name (str): The name of the table.

    Returns:
        TableResource: The DynamoDB table resource.
    """

    return boto3.resource("dynamodb").Table(name)

# GOOD: use table resource rather than client
# GOOD: return pydantic model
def get_user(table: TableResource, user_id: str) -> User | None:
    """Retrieves a user by their ID.

    Args:
        table (TableResource): The DynamoDB table resource.
        user_id (str): The ID of the user to retrieve.

    Returns:
        User | None: The user object if found, otherwise None.
    """

    response = table.get_item(
        Key={
            "PK": f"USER#{user_id}",
            "SK": "PROFILE",
        },
    )

    if "Item" in response:
        return User(**response["Item"])


# GOOD: batch get items is preferred over multiple get items
# GOOD: use table resources even if a client is required. extract client from table object using table.meta.client
def get_multiple_users(table: TableResource, ids: list[str]) -> list[User]:
    """Retrieves multiple users by their IDs.

    Args:
        table (TableResource): The DynamoDB table resource.
        ids (list[str]): A list of user IDs to retrieve.

    Returns:
        list[User]: A list of user objects.
    """

    keys = [{"PK": f"USER#{user_id}", "SK": "PROFILE"} for user_id in ids]

    response = table.meta.client.batch_get_item(
        RequestItems={
            table.name: {
                "Keys": keys,
            }
        }
    )
    return [User(**item) for item in response["Responses"][table.name]]
```

The following example illustrates how NOT to implement DynamoDB persistence layers:

```python
# BAD
# File: persistence.py

# BAD: using clients instead of table resources prevents direct use of pydantic models
def get_user(client: boto3.client, user_id: str) -> dict | None:
    """Retrieves a user by their ID.

    Args:
        client (boto3.client): The DynamoDB client.
        user_id (str): The ID of the user to retrieve.

    Returns:
        dict | None: The user object if found, otherwise None.
    """

    response = client.get_item(
        TableName="users",
        Key={
            "PK": f"USER#{user_id}",
            "SK": "PROFILE",
        },
    )

    # BAD: returning raw data instead of pydantic models makes it harder to enforce data consistency and validation
    if "Item" in response:
        return response["Item"]


def get_users() -> list[dict]:
    """Retrieves all users."""

    # BAD: using clients instead of table resources prevents direct use of pydantic models
    # BAD: dependency injection is not used
    client = boto3.client("dynamodb")

    # BAD: table scans are very slow, expensive, and should be avoided in 
    # favor of indexes
    response = client.scan(
        TableName="users",
    )   

    # BAD: returning raw data instead of pydantic models makes it harder to enforce data consistency and validation
    return response["Items"]
```

# 9. REST APIs

**MUST**: REST APIs must accept and return JSON data. Exceptions can be made for file uploads and responses where binary data is required. 

**MUST**: REST APIs must structure responses in a consistent manner. 

**MUST**: REST APIs must have a version prefix in the URL. This ensures that APIs can be versioned and that old APIs can be deprecated.

**MUST**: Error responses must contain an `error` field, and an optional `details` field. The `error` field must contain a generic error message i.e. "Internal Server Error", "Bad Request" etc. The `details` field should contain additional details about the error where applicable. 

**MUST**: Success responses must return data in a `data` field.

**MUST**: All REST APIs must have an associated `openapi.yaml` file that defines the API contract. This ensures that the API is well-documented.

**SHOULD**: REST API endpoints should follow a dependency injection pattern. Prefer initialization of dependencies within the endpoint handler rather than within business logic. See `Example 7` for an illustration.

**SHOULD**: REST APIs should be implemented using the `FastAPI` package.

**SHOULD**: REST APIs should be run using `uvicorn`.

**SHOULD**: Registration of endpoints should be kept minimal and only contain the basic logic for routing, creation of depedencies such as database clients, and error handling. All business logic should be implemented outside of the endpoint definition.

**SHOULD**: Database clients and other dependencies should be initialized within the endpoint handler, when the endpoint is invoked. Prefer a new database/client/service connection for each request as this avoids long-lived connections and reduces the risk of connection leaks. See `Example 7` for an illustration.

**SHOULD**: DTO datamodels should be defined separately from domain models using the `pydantic` package, and should include validation for all fields.

**SHOULD**: Each endpoint should have a `endpoint_name` function. It should return a `JSONResponse` instance that contains the HTTP response code, and body. See `Example 7` for an illustration.

**SHOULD**: `Depends` should be used to inject dependencies into endpoint handlers. This ensures that dependencies are initialized only once per request.

**SHOULD**: CORS should be enabled for REST APIs by default via the `fastapi.middleware.cors` package.

**SHOULD**: Each endpoint should have its own unittest.

**SHOULD**: When returning `pydantic` models, use the `model_dump` method with the `mode="json"` argument to ensure that the model is serialized to a JSON object.

### Example 7

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
from persistence import new_postgres_connection, create_user
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
    
    user = get_user(db, user.user_id)
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
    
    user_id = create_user(db, user.username, user.email)
    LOGGER.info("User created.", user_id=user_id)
    
    return JSONResponse(status_code=201, content={"data": user_id})


if __name__ == "__main__":
    import uvicorn
    
    uvicorn.run(APP, host="0.0.0.0", port=CONFIG.PORT)
```


The following example illustrates how REST APIs should NOT be structure:

```python
# GOOD
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
def create_user(user: User): # BAD: Return type is not specified.
    """Handles the creation of a new user.

    Args:
        user (User): The new user request object.
    """

    # BAD: Dependencies should be injected via Depends.
    db = new_postgres_connection()

    # BAD: logging should be implemented using structlog.
    print(f"Processing request to create new user: {user}")
    
    user_id = create_user(db, user.username, user.email)
    print(f"User created with id: {user_id}")

    # BAD: JSONResponse should be used to return a JSON response.
    return {"data": user_id}


if __name__ == "__main__":
    import uvicorn
    
    uvicorn.run(APP, host="0.0.0.0", port=CONFIG.PORT)
```


# 10. Dockerfiles

**MUST**: Dockerfiles must be provided for all applications. 

**MUST**: Dockerfiles must be implemented as multi-stage builds. 

**MUST**: Images must be built for AMD linux architecture. Use the `--platform linux/amd64` flag to specify the architecture when building the image. Additionally, the `--provenance=false` flag must be used to disable provenance.

**MUST**: Non-essential files should be excluded from the final image. This ensures that the final image is as small as possible.

**SHOULD**: Dockerfiles should consist of two stages. The first stage should run unittests, the second stage should build the application and run the application. The runtime image should be as small as possible. Prefer a smaller image over a larger image. See `Example 7` for an illustration.

**SHOULD**: The runtime image should be based on a slim image. Prefer debian based images over alpine based images. This avoids issues with missing C libraries and bindings when installing dependencies. See `Example 7` for an illustration.

### Example 7

The following example shows a minimal dockerfile for a python application that implements a `tests` and a `runtime` stage. Note that a smaller runtime image is used.

```dockerfile
# GOOD: Use bookworm as base image for tests
FROM python:3.13-bookworm AS tests

COPY requirements.txt ./
COPY tests/requirements.txt ./requirements-tests.txt

RUN pip install -U pip && \
    pip install -r requirements.txt && \
    pip install -r requirements-tests.txt

COPY src ./
COPY tests ./tests

CMD ["pytest", "-vv"]

# GOOD: Use slim as base image for runtime
FROM python:3.13-slim AS runtime

COPY requirements.txt ./

RUN pip install -U pip && \
    pip install -r requirements.txt

COPY src ./

CMD ["python", "src/main.py"]
```
