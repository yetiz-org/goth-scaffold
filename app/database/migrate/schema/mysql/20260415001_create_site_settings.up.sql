CREATE TABLE IF NOT EXISTS `site_settings`
(
    `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Setting ID',
    `category`        VARCHAR(100)    NOT NULL DEFAULT '' COMMENT 'Setting category',
    `key`             VARCHAR(100)    NOT NULL DEFAULT '' COMMENT 'Setting key',
    `value`           TEXT            NOT NULL COMMENT 'Setting value (JSON)',
    `default`         TINYINT(1)      NOT NULL DEFAULT 0 COMMENT 'Whether this is the default value',
    `effective_start` DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Effective start time',
    `effective_end`   DATETIME                 DEFAULT NULL COMMENT 'Effective end time (NULL means no expiry)',
    `description`     VARCHAR(500)             DEFAULT NULL COMMENT 'Description',
    `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Created time',
    `updated_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Updated time',
    `deleted_at`      DATETIME                 DEFAULT NULL COMMENT 'Deleted time (soft delete)',
    PRIMARY KEY (`id`),
    INDEX `idx_category_key_effective` (`category`, `key`, `effective_start`, `effective_end`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci
    COMMENT = 'Site settings';
