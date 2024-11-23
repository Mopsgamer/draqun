CREATE TABLE IF NOT EXISTS app_group_role_rights (
    group_id BIGINT UNSIGNED NOT NULL COMMENT 'Group id',
    id MEDIUMINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Role id',
    perm_chat_read BIT NOT NULL,
    perm_chat_write BIT NOT NULL,
    perm_chat_delete BIT NOT NULL,
    perm_kick BIT NOT NULL,
    perm_ban BIT NOT NULL,
    perm_change_group BIT NOT NULL,
    perm_change_member BIT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'Restapp all groups roles';