CREATE TABLE IF NOT EXISTS app_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    creator_id BIGINT UNSIGNED NOT NULL,
    owner_id BIGINT UNSIGNED NOT NULL,
    moniker VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    mode ENUM('dm', 'private', 'public') NOT NULL,
    password VARCHAR(255) NOT NULL,
    description VARCHAR(510) NOT NULL,
    avatar VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    is_deleted BIT NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (creator_id) REFERENCES app_users (id),
    UNIQUE (name)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;