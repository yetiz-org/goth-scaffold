CREATE TABLE IF NOT EXISTS `site_setting_tags`
(
    `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Tag ID',
    `site_setting_id` BIGINT UNSIGNED NOT NULL COMMENT 'Site setting ID (foreign key)',
    `name`            VARCHAR(100)    NOT NULL DEFAULT '' COMMENT 'Tag name',
    `created_at`      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Created time',
    PRIMARY KEY (`id`),
    INDEX `idx_site_setting_id` (`site_setting_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci
    COMMENT = 'Site setting tags (example for lazy/eager associations)';
