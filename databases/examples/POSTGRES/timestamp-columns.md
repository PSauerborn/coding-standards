# [PG-008] Timestamp Columns and the `updated_at` Trigger

Statements: `[PG-007]` `[PG-008]` `[PG-010]`

Every table carries `created_at` and `updated_at` columns. Both default to the current
UTC timestamp on the server, and `updated_at` is maintained by a trigger so that it can
never drift from reality if an application forgets to set it.

Use `timestamptz`: PostgreSQL stores it as UTC internally, and `now()` therefore records
the correct instant regardless of the session timezone. Do not write
`DEFAULT (now() AT TIME ZONE 'utc')` on a `timestamptz` column — that strips the zone and
then reinterprets the value in the server's local timezone, which silently shifts the
timestamp on any server not set to UTC.

## The trigger function

Define the function once per schema, then attach it to each table.

```sql
-- GOOD: a single reusable trigger function for the whole schema.
CREATE OR REPLACE FUNCTION base.set_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

## Applying it to a table

```sql
CREATE TABLE base.user (
    id         uuid PRIMARY KEY,
    email      text NOT NULL UNIQUE,
    -- GOOD: server defaults, so the application never supplies these on insert.
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- GOOD: the trigger keeps updated_at accurate on every UPDATE.
CREATE TRIGGER user_set_updated_at
    BEFORE UPDATE ON base.user
    FOR EACH ROW
    EXECUTE FUNCTION base.set_updated_at();
```

```sql
-- BAD: no server defaults and no trigger. Both columns depend on every caller
-- remembering to set them, and any missed UPDATE leaves updated_at stale.
CREATE TABLE base.user (
    id         uuid PRIMARY KEY,
    email      text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

-- BAD: strips the timezone, then reinterprets the result as server-local time.
created_at timestamptz NOT NULL DEFAULT (now() AT TIME ZONE 'utc')
```

## Defining the same schema with SQLAlchemy and alembic

Per `[PG-010]`, schemas are created through `alembic`, using `sqlalchemy` to define
columns and other schema entities.

```python
import sqlalchemy as sa

metadata = sa.MetaData(schema="base")

user = sa.Table(
    "user",
    metadata,
    sa.Column("id", sa.Uuid, primary_key=True),
    sa.Column("email", sa.Text, nullable=False, unique=True),
    # server_default renders as DEFAULT now(), so the value is set by PostgreSQL.
    sa.Column(
        "created_at",
        sa.TIMESTAMP(timezone=True),
        nullable=False,
        server_default=sa.func.now(),
    ),
    sa.Column(
        "updated_at",
        sa.TIMESTAMP(timezone=True),
        nullable=False,
        server_default=sa.func.now(),
    ),
)
```

The trigger itself has no SQLAlchemy construct, so create it explicitly in the migration:

```python
from alembic import op


def upgrade() -> None:
    op.create_table(...)  # as defined above

    op.execute(
        """
        CREATE TRIGGER user_set_updated_at
            BEFORE UPDATE ON base.user
            FOR EACH ROW
            EXECUTE FUNCTION base.set_updated_at();
        """
    )


def downgrade() -> None:
    op.execute("DROP TRIGGER IF EXISTS user_set_updated_at ON base.user;")
    op.drop_table("user", schema="base")
```
