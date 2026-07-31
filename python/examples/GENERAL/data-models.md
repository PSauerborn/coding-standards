# [PY-014] Data Model Implementation

Statements: `[PY-013]` `[PY-014]` `[PY-015]` `[PY-016]` `[PY-017]` `[PY-019]`

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
