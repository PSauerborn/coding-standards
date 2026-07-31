# [PY-037] DynamoDB mock_table Fixture

Statements: `[PY-029]` `[PY-031]` `[PY-033]` `[PY-034]` `[PY-035]` `[PY-036]` `[PY-037]`

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
