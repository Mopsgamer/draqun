CREATE TABLE IF NOT EXISTS app_users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'User id',
    nickname VARCHAR(255) NOT NULL COMMENT 'Customizable name',
    username VARCHAR(255) NOT NULL COMMENT 'Search-friendly changable identificator',
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(255) DEFAULT NULL,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255) DEFAULT NULL,
    created_at DATETIME NOT NULL COMMENT 'Account create time',
    last_seen DATETIME NOT NULL COMMENT 'Last seen time',
    PRIMARY KEY (id),
    UNIQUE (username),
    UNIQUE (email),
    UNIQUE (phone)
) ENGINE = InnoDB AUTO_INCREMENT = 1 DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = 'Restapp users';