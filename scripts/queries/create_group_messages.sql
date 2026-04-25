CREATE TABLE IF NOT EXISTS app_group_messages (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    author_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (author_id) REFERENCES app_users (id)
);