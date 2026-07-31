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
examples:
- examples/GENERAL/data-models.md
- examples/GENERAL/config.md
- examples/GENERAL/dynamodb-fixture.md
- examples/GENERAL/logging.md
- examples/GENERAL/persistence-postgres.md
- examples/GENERAL/persistence-dynamodb.md
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

See `examples/GENERAL/data-models.md` for a reference data model implementation.

## 4. Configuration

`[PY-020]` **MUST**: Configuration must be handled via environment variables. This ensures that configuration is decoupled from code and can be easily changed without modifying code.

`[PY-021]` **MUST**: Configuration must be validated at application startup to ensure that all required variables are set. This ensures that the application fails fast if a required variable is not set.

`[PY-022]` **MUST**: Configuration settings must be defined in a dedicated `config.py` file.

`[PY-023]` **MUST**: Configuration must be parsed and validated using the `pydantic_settings` package. A `Config` class must be defined in the `config.py` file that implements the `BaseSettings` class.

`[PY-024]` **MUST**: A global instance of the `Config` class must be created and made available to the application. This ensures that configuration settings can be easily accessed from anywhere in the application, and that configuration settings are validated at application startup.

See `examples/GENERAL/config.md` for a reference configuration implementation.

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

`[PY-037]` **MUST**: When using DynamoDB, a `mock_table` `pytest` fixture must be implemented in the `conftest.py` file. This fixture must create the table using the schema defined in `tests/data/table_schema.json` and populate it with the items defined in `tests/data/initial_db_items.json`. See `examples/GENERAL/dynamodb-fixture.md` for an implementation.

## 6. Logging

`[PY-038]` **SHOULD**: Logging should be handled via the `structlog` package. See `examples/GENERAL/logging.md` for an implementation.

## 7. Persistence Layers

`[PY-039]` **MUST**: Persistence layers must have their own dedicated file that contains all storage logic.

`[PY-040]` **SHOULD**: Prefer transactional operations that execute multiple database operations as a single unit of work. This ensures that database operations are atomic and consistent, and minimizes the risk of incomplete or inconsistent data.

`[PY-041]` **SHOULD**: Prefer a functional approach when implementing persistence layers.

`[PY-042]` **SHOULD**: Persistence layers should consist of a series of functions that take the database client as the first argument, then execute any required database operations.

`[PY-043]` **SHOULD**: Persistence layers should return domain models where multiple input and return values are required. See `examples/GENERAL/persistence-postgres.md` for an illustration.

`[PY-044]` **SHOULD**: PostgreSQL should be used by default if not otherwise specified.

### PostgreSQL

`[PY-045]` **MUST**: PostgreSQL persistence layers must be implemented using the `psycopg` package. This enforces consistency across all applications.

`[PY-046]` **SHOULD**: Autocommit should be disabled by default in favor of transactions. This ensures that database operations are atomic and consistent and minimizes the risk of incomplete or inconsistent data.

`[PY-047]` **SHOULD**: `psycopg` connections should use a `dict_row` cursor factory. This ensures that database operations return a dictionary of column names and values that can be fed directly into `pydantic` models. See `examples/GENERAL/persistence-postgres.md` for an illustration.

### DynamoDB

`[PY-048]` **MUST**: DynamoDB persistence layers must be implemented using the `boto3` package. This enforces consistency across all applications. See `examples/GENERAL/persistence-dynamodb.md` for an illustration.

`[PY-049]` **SHOULD**: Prefer using `boto3` table resources instead of clients. This ensures that responses from DynamoDB can be passed directly into `pydantic` models without requiring parsing of DynamoDB types. If a `boto3` client is required (such as when using `BatchGetItem`), extract the client from the table object using `table.meta.client`. See `examples/GENERAL/persistence-dynamodb.md` for an illustration.
