CREATE TABLE IF NOT EXISTS app_group_action_bans (
    target_id BIGINT UNSIGNED NOT NULL,
    issuer_id BIGINT UNSIGNED NOT NULL,
    group_id BIGINT UNSIGNED NOT NULL,
    acted_at DATETIME NOT NULL,
    ends_at DATETIME NOT NULL,
    description VARCHAR(510) NOT NULL,
    FOREIGN KEY (target_id) REFERENCES app_users (id),
    FOREIGN KEY (issuer_id) REFERENCES app_users (id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;