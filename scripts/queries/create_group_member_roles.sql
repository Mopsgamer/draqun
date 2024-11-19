CREATE TABLE IF NOT EXISTS app_group_member_roles (
    group_id BIGINT UNSIGNED NOT NULL COMMENT 'Group id',
    user_id BIGINT UNSIGNED NOT NULL COMMENT 'User id',
    role_id MEDIUMINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Role id',
    PRIMARY KEY (group_id, user_id, role_id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (user_id) REFERENCES app_users (id),
    FOREIGN KEY (role_id) REFERENCES app_group_roles (id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'Restapp all groups member roles';