CREATE TABLE `users` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `firstname` varchar(128),
  `lastname` varchar(128),
  `email` varchar(256) UNIQUE NOT NULL,
  `image` varchar(1024),
  `password` varchar(128),
  `type` ENUM ('user', 'superadmin', 'demo') DEFAULT "user",
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE `password_reset` (
  `id` bigint PRIMARY KEY,
  `token` varchar(256) UNIQUE
);

CREATE TABLE `user_verifications` (
  `id` bigint PRIMARY KEY,
  `token` varchar(256) UNIQUE,
  `status` ENUM ('verified', 'not_verified') NOT NULL DEFAULT "not_verified"
);

CREATE TABLE `refresh_sessions` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `refresh_token` varchar(256) UNIQUE NOT NULL,
  `ua` varchar(256) NOT NULL,
  `fingerprint` varchar(256) NOT NULL,
  `expires_at` timestamp NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE `organizations` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `image` varchar(1024),
  `name` varchar(128) NOT NULL,
  `city` varchar(128) NOT NULL,
  `population` int NOT NULL,
  `address` varchar(2048) NOT NULL,
  `invoiceAddress` varchar(2048) NOT NULL,
  `plan` ENUM ('basic', 'conference', 'school', 'pro') DEFAULT "basic",
  `status` ENUM ('paid', 'not_paid', 'blocked') DEFAULT "not_paid",
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE `organizations_users` (
  `organization_id` bigint NOT NULL,
  `user_id` bigint NOT NULL,
  `role` ENUM ('owner', 'admin', 'editor', 'user') NOT NULL,
  PRIMARY KEY (`organization_id`, `user_id`)
);

CREATE TABLE `addresses` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `organization_id` bigint NOT NULL,
  `firstname` varchar(128) NOT NULL,
  `lastname` varchar(128) NOT NULL,
  `email` varchar(256) NOT NULL,
  `phone` varchar(64) NOT NULL,
  `phone_extra` varchar(64),
  `role` varchar(256),
  `info` longtext,
  `sort` int DEFAULT 1
);

CREATE TABLE `languages` (
  `tag` varchar(2) PRIMARY KEY
);

CREATE TABLE `pages` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `organization_id` bigint,
  `parent_id` bigint,
  `language_tag` varchar(2) NOT NULL,
  `type` ENUM ('section', 'content', 'file') NOT NULL,
  `theme` ENUM ('deal_with', 'e_restore', 'precautions', 'e_avoid', 'e_gfs', 'e_school') NOT NULL,
  `status` ENUM ('hidden', 'visible', 'deleted') NOT NULL,
  `title` varchar(256) NOT NULL,
  `image` varchar(1024),
  `image_hover` varchar(1024),
  `sort` int DEFAULT 1,
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE `default_pages` (
  `id` bigint,
  `plan` ENUM ('basic', 'conference', 'school', 'pro'),
  PRIMARY KEY (`id`, `plan`)
);

CREATE TABLE `blocks` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `page_id` bigint NOT NULL,
  `title` varchar(256) NOT NULL,
  `content` longtext,
  `readmore` longtext,
  `image` varchar(1024),
  `image_hover` varchar(1024),
  `type` ENUM ('default', 'accordion', 'link') NOT NULL,
  `sort` int DEFAULT 1
);

CREATE TABLE `files` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `path` varchar(1024),
  `created_at` timestamp DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE `user_verifications` ADD FOREIGN KEY (`id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `password_reset` ADD FOREIGN KEY (`id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `refresh_sessions` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `organizations_users` ADD FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `organizations_users` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `addresses` ADD FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `pages` ADD FOREIGN KEY (`organization_id`) REFERENCES `organizations` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `pages` ADD FOREIGN KEY (`parent_id`) REFERENCES `pages` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `pages` ADD FOREIGN KEY (`language_tag`) REFERENCES `languages` (`tag`) ON DELETE NO ACTION ON UPDATE CASCADE;

ALTER TABLE `files` ADD FOREIGN KEY (`id`) REFERENCES `pages` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `default_pages` ADD FOREIGN KEY (`id`) REFERENCES `pages` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

ALTER TABLE `blocks` ADD FOREIGN KEY (`page_id`) REFERENCES `pages` (`id`) ON DELETE CASCADE ON UPDATE CASCADE;

