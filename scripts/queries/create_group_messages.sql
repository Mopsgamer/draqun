CREATE TABLE IF NOT EXISTS app_group_messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Message id',
    group_id BIGINT UNSIGNED NOT NULL COMMENT 'Group id',
    author_id BIGINT UNSIGNED NOT NULL COMMENT 'User id',
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL COMMENT 'Message create time',
    PRIMARY KEY (id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (author_id) REFERENCES app_users (id)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'Draqun messages';