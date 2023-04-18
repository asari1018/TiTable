-- Table for tasks
DROP TABLE IF EXISTS `tasks`;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `classes`;
DROP TABLE IF EXISTS `user_info`;

CREATE TABLE `tasks` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `title` varchar(50) NOT NULL,
    `class` varchar(50) NOT NULL,
    `is_done` boolean NOT NULL DEFAULT b'0',
    `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `deadline_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `task_level` bigint(20) NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;

CREATE TABLE `users` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `name` varchar(50) NOT NULL,
    `email_id` varchar(50) NOT NULL,
    `user_auth` varchar(50) NOT NULL,
    `password` varchar(256) NOT NULL,
    `last_time` datetime NOT NULL,
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;

CREATE TABLE `classes` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `class` varchar(50) NOT NULL,
    `uid` bigint(20) NOT NULL,
    `comment` varchar(256) NOT NULL DEFAULT 'hoge',
    `start` datetime NOT NULL,
    `end` datetime NOT NULL,
    `url` varchar(256) NOT NULL DEFAULT 'hoge',
    `x` bigint(20) NOT NULL,
    `y` bigint(20) NOT NULL,
    `length` bigint(20) NOT NULL,
    PRIMARY KEY (`id`)
) DEFAULT CHARSET=utf8mb4;

CREATE TABLE `user_info` (
    `task_id` bigint(20) NOT NULL,
    `user_id` bigint(20) NOT NULL,
    PRIMARY KEY (`task_id`, `user_id`)
) DEFAULT CHARSET=utf8mb4;
