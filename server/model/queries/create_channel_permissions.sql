CREATE TABLE IF NOT EXISTS app_channel_permissions (
    channel_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (channel_id, role_id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id),
    FOREIGN KEY (role_id) REFERENCES app_group_roles (id)
);
