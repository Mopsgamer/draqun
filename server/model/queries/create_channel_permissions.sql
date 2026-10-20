CREATE TABLE IF NOT EXISTS app_channel_permissions (
    channel_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    perm_messages VARCHAR(32) NOT NULL DEFAULT 'inherit',
    perm_roles VARCHAR(32) NOT NULL DEFAULT 'inherit',
    perm_members VARCHAR(32) NOT NULL DEFAULT 'inherit',
    perm_group_change VARCHAR(32) NOT NULL DEFAULT 'inherit',
    perm_admin VARCHAR(32) NOT NULL DEFAULT 'inherit',
    inherit BOOLEAN NOT NULL DEFAULT 1,
    PRIMARY KEY (channel_id, role_id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id),
    FOREIGN KEY (role_id) REFERENCES app_group_roles (id)
);
