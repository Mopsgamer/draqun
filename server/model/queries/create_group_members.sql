CREATE TABLE IF NOT EXISTS app_group_members (
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    moniker TEXT NOT NULL,
    first_time_joined_at DATETIME NOT NULL,
    is_deleted INTEGER NOT NULL DEFAULT 0 CHECK (is_deleted IN (0, 1)),
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (user_id) REFERENCES app_users (id)
);