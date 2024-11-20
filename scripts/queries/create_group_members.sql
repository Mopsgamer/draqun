CREATE TABLE IF NOT EXISTS app_group_members (
    group_id BIGINT UNSIGNED NOT NULL COMMENT 'Group id',
    user_id BIGINT UNSIGNED NOT NULL COMMENT 'User id',
    is_owner BIT NOT NULL,
    is_banned BIT NOT NULL,
    membername VARCHAR(255),
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (user_id) REFERENCES app_users (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'Restapp all groups members';