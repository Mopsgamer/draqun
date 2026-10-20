CREATE TABLE IF NOT EXISTS app_channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    moniker VARCHAR(255) NOT NULL,
    description VARCHAR(510) NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'text',
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    closed_reason VARCHAR(32) NOT NULL DEFAULT '',
    author_id INTEGER NOT NULL DEFAULT 0,
    duplicate_of_id INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME,
    created_at DATETIME NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT 0
);
