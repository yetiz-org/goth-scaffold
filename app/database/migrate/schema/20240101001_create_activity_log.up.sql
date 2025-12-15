CREATE TABLE `activity_log`
(
    `id`           int unsigned                                                  NOT NULL AUTO_INCREMENT,
    `log_name`     varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci      DEFAULT NULL,
    `description`  varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
    `subject_id`   int                                                                DEFAULT NULL,
    `subject_type` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci      DEFAULT NULL,
    `causer_id`    int                                                                DEFAULT NULL,
    `causer_type`  varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci      DEFAULT NULL,
    `properties`   text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci,
    `created_at`   timestamp                                                     NULL DEFAULT NULL,
    `updated_at`   timestamp                                                     NULL DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;
