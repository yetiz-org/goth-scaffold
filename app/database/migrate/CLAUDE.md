# app/database/migrate

## File Naming

```
{YYYYMMDD}{NNN}_{action}_{target}.{up|down}.sql
```

Examples:

```
20260101001_create_users.up.sql
20260101001_create_users.down.sql
20260215001_add_index_to_users.up.sql
20260215001_add_index_to_users.down.sql
```

- `NNN` is a 3-digit sequence within the same date (001, 002, …)
- Every `.up.sql` **must** have a corresponding `.down.sql`
- **Never edit an existing migration** — always add a new one
- The "current schema" is the sum of all applied migrations, not any single file

## SQL Conventions

- MySQL: add a brief column `COMMENT` describing each column's purpose
- Use `IF NOT EXISTS` / `IF EXISTS` guards to make migrations idempotent where possible
- End every statement with `;`

```sql
-- Example up migration
CREATE TABLE IF NOT EXISTS `users`
(
    `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Primary key',
    `name`       VARCHAR(100)    NOT NULL DEFAULT '' COMMENT 'Display name',
    `created_at` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp (UTC)',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci
    COMMENT = 'Application users';

-- Example down migration
DROP TABLE IF EXISTS `users`;
```

## Cassandra Keyspaces

Place Cassandra CQL definitions under `keyspaces/` using the same date-prefix naming convention.

CQL-specific rules:

- Use `IF NOT EXISTS` on `CREATE TABLE` (supported by Cassandra).
- No inline column comments — CQL does not support `COMMENT` on columns.
- Always specify `WITH CLUSTERING ORDER BY` when range queries are expected on clustering keys.
- End every statement with `;`.

```cql
-- Example up migration
CREATE TABLE IF NOT EXISTS maintenance_logs (
    id         uuid,
    tenant_id  text,
    created_at timestamp,
    message    text,
    PRIMARY KEY ((tenant_id), created_at, id)
) WITH CLUSTERING ORDER BY (created_at DESC, id ASC);

-- Example down migration
DROP TABLE IF EXISTS maintenance_logs;
```
