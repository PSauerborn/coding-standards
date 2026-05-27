---
title: Python Code Standards
description: General standards for writing Python applications.
scope:
- '*.py'
parent: GENERAL.md
topics:
- python
- pydantic
- pytest
- structlog
---

# Python Code Standards

## 1. Versions and Tooling

`[PY-001]` **MUST**: Python version 3.13 or higher must be used.

`[PY-002]` **MUST**: When in doubt, follow PEP 8.

`[PY-003]` **SHOULD**: All code should be formatted using `black`.

`[PY-004]` **SHOULD**: All code should be linted using `flake8`.

## 2. Syntax, Naming & Style

`[PY-005]` **MUST**: All file and function names must be snake_case. This ensures adherence to PEP 8.

`[PY-006]` **MUST**: All functions must have a docstring that clearly describes the purpose of the function, its parameters, and its return values. The first word of every docstring should be the name of the function. This ensures that the docstring is easily searchable.

`[PY-007]` **MUST**: All docstrings must be formatted using Google style.

`[PY-008]` **MUST**: Type hints must be provided for all function parameters and return values.

`[PY-009]` **MUST**: Use spaces for indentation instead of tabs.

`[PY-010]` **SHOULD**: Prefer a functional approach when possible.

`[PY-011]` **SHOULD**: Prefer a single return type per function. This ensures that function complexity is kept minimal.

`[PY-012]` **SHOULD**: Source code should be placed in a `src` directory. A single `main.py` or `app.py` file should serve as the entry point at the root of the `src` directory.

## 3. Data Models and Validation

`[PY-013]` **MUST**: Data models must be defined in a dedicated `models.py` file.

`[PY-014]` **MUST**: Data models must be implemented using `pydantic`.

`[PY-015]` **SHOULD**: Data models should be used throughout the application to group related data. This ensures that the code base is more type-safe, and makes the code more readable.

`[PY-016]` **SHOULD**: An alias should be defined for fields using the `pydantic` `Field` class if applicable. Only add aliases if it is necessary to ensure that the model is compatible with either the API or the database schema.

`[PY-017]` **SHOULD**: DTOs and domain models should be defined separately. This ensures that API logic is decoupled from database/storage logic.

`[PY-018]` **SHOULD**: DTOs should have strict validation rules to ensure that they are valid before being used in business logic. This ensures that external data is validated before being used in business logic.

`[PY-019]` **SHOULD**: Prefer length-constrained strings over just `str` types in datamodels. `None` values should be preferred to empty strings. Empty strings are ambiguous, and it is not clear whether they should be considered as "empty" or "not set".

### Example 1

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

## 4. Configuration

`[PY-020]` **MUST**: Configuration must be handled via environment variables. This ensures that configuration is decoupled from code and can be easily changed without modifying code.

`[PY-021]` **MUST**: Configuration must be validated at application startup to ensure that all required variables are set. This ensures that the application fails fast if a required variable is not set.

`[PY-022]` **MUST**: Configuration settings must be defined in a dedicated `config.py` file.

`[PY-023]` **MUST**: Configuration must be parsed and validated using the `pydantic_settings` package. A `Config` class must be defined in the `config.py` file that implements the `BaseSettings` class.

`[PY-024]` **MUST**: A global instance of the `Config` class must be created and made available to the application. This ensures that configuration settings can be easily accessed from anywhere in the application, and that configuration settings are validated at application startup.

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

## 5. Unittests

`[PY-025]` **MUST**: Unittests must be implemented for all business logic.

`[PY-026]` **MUST**: Unittests must be stored in a `tests` directory.

`[PY-027]` **MUST**: Unittests must be implemented using the `pytest` package.

`[PY-028]` **SHOULD**: Each `.py` file should have a corresponding `test_.py` file that contains unittests.

`[PY-029]` **SHOULD**: The `tests` directory should contain a `conftest.py` file that contains common test fixtures.

`[PY-030]` **SHOULD**: Each function should have a corresponding unittest. Unittests should be named in the format `test_function_name`.

`[PY-031]` **SHOULD**: Unittests should mock database and other service connections rather than using real connections. Connection to live databases should only be used in integration tests. This ensures that unittests are fast and do not depend on external services.

`[PY-032]` **SHOULD**: Additional test data used to test business logic (PDF, CSV files etc) should be stored in a separate `tests/data` directory. This ensures that test files do not become cluttered with test data.

### AWS Applications

`[PY-033]` **MUST**: Any applications using AWS resources must use the `moto` package to mock AWS resources.

`[PY-034]` **MUST**: Any AWS resources that can be implemented using the `moto` package must be. This ensures that unittests can test interaction with AWS resources as well as possible.

`[PY-035]` **MUST**: When using DynamoDB, a `tests/data/table_schema.json` file must be provided that defines the schema of the table.

`[PY-036]` **MUST**: When using DynamoDB, a `tests/data/initial_db_items.json` file must be provided that defines the initial items in the table.

`[PY-037]` **MUST**: When using DynamoDB, a `mock_table` `pytest` fixture must be implemented in the `conftest.py` file. This fixture must create the table using the schema defined in `tests/data/table_schema.json` and populate it with the items defined in `tests/data/initial_db_items.json`. See Example 3 for an implementation.

### Example 3

The following example illustrates a unittest implementation for a function that uses DynamoDB.

```python
# GOOD
# File: tests/conftest.py
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
def mock_table(mock_dynamodb_resource):
    with open("tests/data/table_schema.json", "r") as f:
        table_schema = json.load(f)

    table = mock_dynamodb_resource.create_table(**table_schema)

    with open("tests/data/initial_db_items.json", "r") as f:
        initial_db_items = json.load(f)

    for item in initial_db_items:
        table.put_item(Item=item)
    
    yield table

```

## 6. Logging

`[PY-038]` **SHOULD**: Logging should be handled via the `structlog` package. See Example 4 for an implementation.

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

The following example illustrates how logging should NOT be implemented:

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

## 7. Persistence Layers

`[PY-039]` **MUST**: Persistence layers must have their own dedicated file that contains all storage logic.

`[PY-040]` **SHOULD**: Prefer transactional operations that execute multiple database operations as a single unit of work. This ensures that database operations are atomic and consistent, and minimizes the risk of incomplete or inconsistent data.

`[PY-041]` **SHOULD**: Prefer a functional approach when implementing persistence layers.

`[PY-042]` **SHOULD**: Persistence layers should consist of a series of functions that take the database client as the first argument, then execute any required database operations.

`[PY-043]` **SHOULD**: Persistence layers should return domain models where multiple input and return values are required. See `Example 5` for an illustration.

`[PY-044]` **SHOULD**: PostgreSQL should be used by default if not otherwise specified.

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

`[PY-045]` **MUST**: PostgreSQL persistence layers must be implemented using the `psycopg` package. This enforces consistency across all applications.

`[PY-046]` **SHOULD**: Autocommit should be disabled by default in favor of transactions. This ensures that database operations are atomic and consistent and minimizes the risk of incomplete or inconsistent data.

`[PY-047]` **SHOULD**: `psycopg` connections should use a `dict_row` cursor factory. This ensures that database operations return a dictionary of column names and values that can be fed directly into `pydantic` models. See `Example 3` for an illustration.

### DynamoDB

`[PY-048]` **MUST**: DynamoDB persistence layers must be implemented using the `boto3` package. This enforces consistency across all applications.

`[PY-049]` **SHOULD**: Prefer using `boto3` table resources instead of clients. This ensures that responses from DynamoDB can be passed directly into `pydantic` models without requiring parsing of DynamoDB types. If a `boto3` client is required (such as when using `BatchGetItem`), extract the client from the table object using `table.meta.client`.

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

