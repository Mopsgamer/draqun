CREATE TABLE IF NOT EXISTS app_group_action_memberships (
    user_id BIGINT UNSIGNED NOT NULL,
    group_id BIGINT UNSIGNED NOT NULL,
    acted_at DATETIME NOT NULL,
    is_join BIT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES app_users (id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;