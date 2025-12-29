-- 数据库初始化脚本
-- 生成时间: " + time.Now().Format("2006-01-02 15:04:05") + "

-- 创建用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(50) DEFAULT NULL COMMENT '用户名',
  `email` VARCHAR(100) DEFAULT NULL COMMENT '邮箱',
  `telephone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
  `password` VARCHAR(100) DEFAULT NULL COMMENT '密码',
  `nickname` VARCHAR(50) DEFAULT NULL COMMENT '昵称',
  `head_image` VARCHAR(255) DEFAULT NULL COMMENT '头像',
  `openid` VARCHAR(100) DEFAULT NULL COMMENT 'openid',
  `github_url` VARCHAR(255) DEFAULT NULL COMMENT 'github地址',
  `github_login_id` VARCHAR(50) DEFAULT NULL COMMENT 'github登录id',
  `avatar_url` VARCHAR(255) DEFAULT NULL COMMENT 'github头像',
  `intro` TEXT DEFAULT NULL COMMENT '个人简介',
  `website` VARCHAR(255) DEFAULT NULL COMMENT '个人网站',
  `register_ip` VARCHAR(20) DEFAULT NULL COMMENT '注册ip',
  `register_time` DATETIME DEFAULT NULL COMMENT '注册时间',
  `last_login_time` DATETIME DEFAULT NULL COMMENT '最后登录时间',
  `is_admin` BOOLEAN DEFAULT FALSE COMMENT '是否管理员',
  `out_time` DATETIME DEFAULT NULL COMMENT '退出时间',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  INDEX `idx_email` (`email`),
  INDEX `idx_github_login_id` (`github_login_id`),
  INDEX `idx_openid` (`openid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建文章表
CREATE TABLE IF NOT EXISTS `posts` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `title` TEXT NOT NULL COMMENT '文章标题',
  `body` LONGTEXT NOT NULL COMMENT '文章内容',
  `view` INT DEFAULT 0 COMMENT '浏览量',
  `is_published` BOOLEAN DEFAULT FALSE COMMENT '是否发布',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  INDEX `idx_is_published` (`is_published`),
  INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建评论表
CREATE TABLE IF NOT EXISTS `comments` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `content` TEXT NOT NULL COMMENT '评论内容',
  `post_id` BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  `read_state` BOOLEAN DEFAULT FALSE COMMENT '是否已读',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  INDEX `idx_post_id` (`post_id`),
  INDEX `idx_user_id` (`user_id`),
  INDEX `idx_read_state` (`read_state`),
  INDEX `idx_created_at` (`created_at`),
  FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`) ON DELETE CASCADE,
  FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建标签表
CREATE TABLE IF NOT EXISTS `tags` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(50) NOT NULL COMMENT '标签名称',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE INDEX `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建文章标签关联表
CREATE TABLE IF NOT EXISTS `post_tags` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `post_id` BIGINT UNSIGNED NOT NULL COMMENT '文章ID',
  `tag_id` BIGINT UNSIGNED NOT NULL COMMENT '标签ID',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE INDEX `uk_post_tag` (`post_id`, `tag_id`),
  INDEX `idx_post_id` (`post_id`),
  INDEX `idx_tag_id` (`tag_id`),
  FOREIGN KEY (`post_id`) REFERENCES `posts` (`id`) ON DELETE CASCADE,
  FOREIGN KEY (`tag_id`) REFERENCES `tags` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建链接表
CREATE TABLE IF NOT EXISTS `links` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `name` VARCHAR(100) NOT NULL COMMENT '链接名称',
  `url` VARCHAR(255) NOT NULL COMMENT '链接地址',
  `sort` INT DEFAULT 0 COMMENT '排序',
  `view` INT DEFAULT 0 COMMENT '点击量',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
  INDEX `idx_url` (`url`),
  INDEX `idx_sort` (`sort`),
  INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建页面表
CREATE TABLE IF NOT EXISTS `pages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `title` VARCHAR(255) NOT NULL COMMENT '页面标题',
  `body` LONGTEXT NOT NULL COMMENT '页面内容',
  `slug` VARCHAR(50) NOT NULL COMMENT '页面标识',
  `view` INT DEFAULT 0 COMMENT '浏览量',
  `is_published` BOOLEAN DEFAULT TRUE COMMENT '是否发布',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  UNIQUE INDEX `idx_slug` (`slug`),
  INDEX `idx_is_published` (`is_published`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建ES同步状态表
CREATE TABLE IF NOT EXISTS `es_sync_statuses` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `last_sync_time` DATETIME NOT NULL COMMENT '最后同步时间',
  `last_post_id` BIGINT UNSIGNED NOT NULL COMMENT '最后同步的文章ID',
  `total_synced` INT DEFAULT 0 COMMENT '已同步总数',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建SMMS文件表
CREATE TABLE IF NOT EXISTS `smms_files` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `file_name` VARCHAR(255) NOT NULL COMMENT '文件名',
  `store_name` VARCHAR(255) NOT NULL COMMENT '存储名',
  `size` INT NOT NULL COMMENT '文件大小',
  `width` INT NOT NULL COMMENT '宽度',
  `height` INT NOT NULL COMMENT '高度',
  `hash` VARCHAR(255) NOT NULL COMMENT '文件哈希',
  `delete` VARCHAR(255) NOT NULL COMMENT '删除链接',
  `url` VARCHAR(500) NOT NULL COMMENT '文件URL',
  `path` VARCHAR(255) NOT NULL COMMENT '文件路径',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  INDEX `idx_hash` (`hash`),
  INDEX `idx_url` (`url`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 创建订阅者表
CREATE TABLE IF NOT EXISTS `subscribers` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `email` VARCHAR(255) NOT NULL COMMENT '邮箱地址',
  `verify_state` BOOLEAN DEFAULT FALSE COMMENT '验证状态',
  `subscribe_state` BOOLEAN DEFAULT TRUE COMMENT '订阅状态',
  `out_time` DATETIME DEFAULT NULL COMMENT '退出时间',
  `secret_key` VARCHAR(255) NOT NULL COMMENT '密钥',
  `signature` VARCHAR(255) NOT NULL COMMENT '签名',
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
  UNIQUE INDEX `idx_email` (`email`),
  INDEX `idx_verify_state` (`verify_state`),
  INDEX `idx_subscribe_state` (`subscribe_state`),
  INDEX `idx_signature` (`signature`),
  INDEX `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;