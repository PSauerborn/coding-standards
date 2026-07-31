# [PY-048] DynamoDB Persistence

Statements: `[PY-042]` `[PY-043]` `[PY-048]` `[PY-049]`

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
