CREATE TABLE IF NOT EXISTS app_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Group id',
    creator_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'User id',
    nickname VARCHAR(255) NOT NULL COMMENT 'Customizable name',
    groupname VARCHAR(255) NOT NULL COMMENT 'Search-friendly changable identificator',
    groupmode ENUM('dm', 'private', 'public') NOT NULL,
    password VARCHAR(255) DEFAULT NULL,
    avatar VARCHAR(255) DEFAULT NULL,
    created_at DATETIME NOT NULL COMMENT 'Group create time',
    PRIMARY KEY (id),
    FOREIGN KEY (creator_id) REFERENCES app_users (id)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'Restapp groups';