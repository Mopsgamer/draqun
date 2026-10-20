CREATE TABLE IF NOT EXISTS app_group_role_assignees (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES app_users (id),
    FOREIGN KEY (role_id) REFERENCES app_group_roles (id)
);