CREATE TABLE IF NOT EXISTS "site_setting_tags"
(
    "id"              BIGSERIAL       NOT NULL,
    "site_setting_id" BIGINT          NOT NULL,
    "name"            VARCHAR(100)    NOT NULL DEFAULT '',
    "created_at"      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY ("id")
);
CREATE INDEX IF NOT EXISTS "idx_site_setting_id" ON "site_setting_tags" ("site_setting_id");
