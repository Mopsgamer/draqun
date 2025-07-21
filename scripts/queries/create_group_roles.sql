CREATE TABLE IF NOT EXISTS app_group_roles (
    id MEDIUMINT UNSIGNED NOT NULL AUTO_INCREMENT,
    group_id BIGINT UNSIGNED NOT NULL,
    moniker VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    color INT UNSIGNED NOT NULL,
    perm_messages ENUM('', 'hidden', 'read', 'write', 'delete') NOT NULL,
    perm_roles ENUM('', 'disallow', 'allow') NOT NULL,
    perm_members ENUM('', 'read', 'invite', 'write', 'delete') NOT NULL,
    perm_group_change ENUM('', 'disallow', 'allow') NOT NULL,
    perm_admin ENUM('', 'disallow', 'allow') NOT NULL,
    PRIMARY KEY (id, group_id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    UNIQUE (name)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;