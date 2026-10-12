CREATE TABLE IF NOT EXISTS app_group_action_roles (
    role_id INTEGER NOT NULL,
    creator_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('create', 'update', 'delete', 'assign', 'unassign')),
    description VARCHAR(510) NOT NULL,
    acted_at DATETIME NOT NULL,
    FOREIGN KEY (role_id) REFERENCES app_group_roles (id),
    FOREIGN KEY (creator_id) REFERENCES app_users (id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id)
);
