# [PY-038] Structured Logging with structlog

Statements: `[PY-038]`

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
