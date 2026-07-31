# [PY-043] PostgreSQL Persistence

Statements: `[PY-039]` `[PY-041]` `[PY-042]` `[PY-043]` `[PY-045]` `[PY-046]` `[PY-047]`

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
