CREATE TABLE IF NOT EXISTS app_group_messages (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL DEFAULT 0,
    author_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    attachment_url TEXT NOT NULL DEFAULT '',
    attachment_name TEXT NOT NULL DEFAULT '',
    attachment_size INTEGER NOT NULL DEFAULT 0,
    attachment_type TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (author_id) REFERENCES app_users (id)
);
