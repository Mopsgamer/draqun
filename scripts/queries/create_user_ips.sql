CREATE TABLE IF NOT EXISTS app_user_ips (
    user_id INTEGER NOT NULL,
    ip VARCHAR(64) NOT NULL,
    last_seen_at DATETIME NOT NULL,
    PRIMARY KEY (user_id, ip),
    FOREIGN KEY (user_id) REFERENCES app_users (id)
);
