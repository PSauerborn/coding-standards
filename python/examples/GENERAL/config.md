# [PY-023] Configuration Loading

Statements: `[PY-019]` `[PY-020]` `[PY-021]` `[PY-022]` `[PY-023]` `[PY-024]`

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
