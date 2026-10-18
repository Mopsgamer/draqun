CREATE TABLE IF NOT EXISTS app_blocked_ips (
    ip VARCHAR(64) PRIMARY KEY,
    creator_id INTEGER NOT NULL,
    reason VARCHAR(255) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL
);
