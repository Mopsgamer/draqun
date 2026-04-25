CREATE TABLE IF NOT EXISTS app_group_action_bans (
    target_id INTEGER NOT NULL,
    creator_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    revoker_id INTEGER NOT NULL,
    description VARCHAR(510) NOT NULL,
    acted_at DATETIME NOT NULL,
    ends_at DATETIME NOT NULL,
    FOREIGN KEY (target_id) REFERENCES app_users (id),
    FOREIGN KEY (creator_id) REFERENCES app_users (id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id)
);