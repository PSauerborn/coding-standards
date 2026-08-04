# [PG-009] Timestamp Columns and the `updated_at` Trigger

Statements: `[PG-006]` `[PG-008]` `[PG-009]` `[PG-011]`

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

COMMENT ON FUNCTION base.set_updated_at() IS
    'Trigger function that stamps updated_at with the current UTC time on every UPDATE.';
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

-- GOOD: the table, its columns and the trigger all carry comments.
COMMENT ON TABLE base.user IS
    'Application users. Keyed on an internally generated ID, never an external one.';
COMMENT ON COLUMN base.user.created_at IS
    'UTC timestamp of row creation. Set by the server default; never supplied by callers.';
COMMENT ON COLUMN base.user.updated_at IS
    'UTC timestamp of the last update. Maintained by the user_set_updated_at trigger.';
COMMENT ON TRIGGER user_set_updated_at ON base.user IS
    'Keeps base.user.updated_at current on every UPDATE.';
```

```sql
-- BAD: no server defaults and no trigger. Both columns depend on every caller
-- remembering to set them, and any missed UPDATE leaves updated_at stale. No
-- COMMENT ON either, so an operator has to read the application to learn what
-- the table holds.
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

Per `[PG-011]`, schemas are created through `alembic`, using `sqlalchemy` to define
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
        comment=(
            "UTC timestamp of row creation. Set by the server default; never supplied "
            "by callers."
        ),
    ),
    sa.Column(
        "updated_at",
        sa.TIMESTAMP(timezone=True),
        nullable=False,
        server_default=sa.func.now(),
        comment=(
            "UTC timestamp of the last update. Maintained by the user_set_updated_at "
            "trigger."
        ),
    ),
    # comment= emits COMMENT ON TABLE, so the description lives with the schema
    # definition and stays in step with it.
    comment="Application users. Keyed on an internally generated ID, never an external one.",
)
```

`comment=` on `sa.Table` and `sa.Column` is rendered by alembic as `COMMENT ON TABLE` and
`COMMENT ON COLUMN`, and later edits are picked up as `alter_column(..., comment=...)` /
`create_table_comment(...)` operations. Keeping the text on the construct is what makes the
comments stay up to date; hand-written `op.execute("COMMENT ON ...")` calls drift.

The trigger and its comment have no SQLAlchemy construct, so create them explicitly in the
migration:

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
    op.execute(
        """
        COMMENT ON TRIGGER user_set_updated_at ON base.user IS
            'Keeps base.user.updated_at current on every UPDATE.';
        """
    )


def downgrade() -> None:
    op.execute("DROP TRIGGER IF EXISTS user_set_updated_at ON base.user;")
    op.drop_table("user", schema="base")
```
