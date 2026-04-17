CREATE TABLE IF NOT EXISTS "site_settings"
(
    "id"              BIGSERIAL       NOT NULL,
    "category"        VARCHAR(100)    NOT NULL DEFAULT '',
    "key"             VARCHAR(100)    NOT NULL DEFAULT '',
    "value"           TEXT            NOT NULL,
    "default"         BOOLEAN         NOT NULL DEFAULT FALSE,
    "effective_start" TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "effective_end"   TIMESTAMP                DEFAULT NULL,
    "description"     VARCHAR(500)             DEFAULT NULL,
    "created_at"      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at"      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "deleted_at"      TIMESTAMP                DEFAULT NULL,
    PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "idx_category_key_effective" ON "site_settings" ("category", "key", "effective_start", "effective_end");
