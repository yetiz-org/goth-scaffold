CREATE TABLE `example`
(
    `id`         bigint(20) unsigned NOT NULL AUTO_INCREMENT,
    `created_at` int(10) unsigned NOT NULL DEFAULT '0',
    `updated_at` int(10) unsigned NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
