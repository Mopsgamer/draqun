CREATE TABLE IF NOT EXISTS app_group_action_memberships (
    user_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    acted_at DATETIME NOT NULL,
    is_join INTEGER NOT NULL DEFAULT 0 CHECK (is_join IN (0, 1)),
    FOREIGN KEY (user_id) REFERENCES app_users (id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id)
);