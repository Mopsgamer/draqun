CREATE TABLE IF NOT EXISTS app_group_roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    moniker TEXT NOT NULL,
    name TEXT NOT NULL,
    color INTEGER NOT NULL,
    UNIQUE (group_id, name),
    FOREIGN KEY (group_id) REFERENCES app_groups (id)
);
