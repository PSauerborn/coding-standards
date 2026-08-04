# [PG-005] Link Tables for Shared Resources

Statements: `[PG-005]` `[PG-006]`

The following example illustrates how link (junction) tables are used to associate
multiple entities with a shared, generic resource, rather than adding entity-specific
foreign keys to the shared table.

Here, both `base.user` and `base.job` need to own documents. Instead of adding
`user_id` and `job_id` columns to `base.document`, each owning entity links to the
generic `base.document` table through its own link table.

```sql
-- The shared, generic resource. It has NO knowledge of who owns it:
-- no user_id or job_id columns.
CREATE TABLE base.document (
    id          uuid PRIMARY KEY,
    file_name   text NOT NULL,
    storage_key text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE base.document IS
    'Generic file resource. Ownership is expressed by link tables, never by columns here.';
COMMENT ON COLUMN base.document.storage_key IS
    'Key of the object in the blob store; unique per document and never reused.';

-- GOOD: a link table maps users to their documents.
CREATE TABLE base.user_document (
    user_id     uuid NOT NULL REFERENCES base.user (id),
    document_id uuid NOT NULL REFERENCES base.document (id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, document_id)
);

COMMENT ON TABLE base.user_document IS
    'Links users to the documents they own. One row per (user, document) pair.';

-- GOOD: a separate link table maps jobs to their documents.
CREATE TABLE base.job_document (
    job_id      uuid NOT NULL REFERENCES base.job (id),
    document_id uuid NOT NULL REFERENCES base.document (id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (job_id, document_id)
);

COMMENT ON TABLE base.job_document IS
    'Links jobs to the documents they produced. One row per (job, document) pair.';
```

This keeps `base.document` reusable by any number of owning entities, and each link
table enforces referential integrity through foreign keys (see `[PG-004]`). Every table
also carries a `COMMENT ON`, so an operator inspecting the schema can tell which entity
a link table associates without reading the application code.

For contrast, the anti-pattern below couples the shared table to specific entities and
does not scale — every new owner type requires another nullable column on
`base.document`:

```sql
-- BAD: entity-specific foreign keys on the shared table.
-- Adding a third owner type means adding yet another nullable column.
CREATE TABLE base.document (
    id          uuid PRIMARY KEY,
    user_id     uuid REFERENCES base.user (id),  -- BAD
    job_id      uuid REFERENCES base.job (id),   -- BAD
    file_name   text NOT NULL,
    storage_key text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);
```

The same pattern applies to many-to-many relationships: a link table with a composite
primary key over the two foreign keys models the association between both entities.
