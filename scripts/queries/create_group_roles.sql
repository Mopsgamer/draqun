CREATE TABLE IF NOT EXISTS app_group_roles (
    id MEDIUMINT UNSIGNED NOT NULL AUTO_INCREMENT,
    moniker VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    color INT UNSIGNED NOT NULL,
    perm_msg ENUM('locked', 'read', 'write', 'delete') NOT NULL,
    perm_memb_kick BIT NOT NULL,
    perm_memb_ban BIT NOT NULL,
    perm_memb_change BIT NOT NULL,
    perm_group_cosmetic BIT NOT NULL,
    perm_admin BIT NOT NULL,
    PRIMARY KEY (id)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;