CREATE TABLE IF NOT EXISTS app_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    creator_id INTEGER NOT NULL,
    owner_id INTEGER NOT NULL,
    moniker TEXT NOT NULL,
    name TEXT NOT NULL,
    mode TEXT NOT NULL DEFAULT 0 CHECK (mode IN ('dm', 'private', 'public')),
    password TEXT NOT NULL,
    description TEXT NOT NULL,
    avatar TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    is_deleted INTEGER NOT NULL DEFAULT 0 CHECK (is_deleted IN (0, 1)),
    FOREIGN KEY (creator_id) REFERENCES app_users (id),
    UNIQUE (name)
);