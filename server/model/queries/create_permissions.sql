CREATE TABLE IF NOT EXISTS app_perms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL REFERENCES app_groups(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    level TEXT NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS app_perm_channels (
    channel_id INTEGER NOT NULL REFERENCES app_channels(id) ON DELETE CASCADE,
    perm_id INTEGER NOT NULL REFERENCES app_perms(id) ON DELETE CASCADE,
    PRIMARY KEY (channel_id, perm_id)
);

CREATE TABLE IF NOT EXISTS app_perm_users (
    user_id INTEGER NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,
    perm_id INTEGER NOT NULL REFERENCES app_perms(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, perm_id)
);

CREATE TABLE IF NOT EXISTS app_perm_roles (
    role_id INTEGER NOT NULL REFERENCES app_group_roles(id) ON DELETE CASCADE,
    perm_id INTEGER NOT NULL REFERENCES app_perms(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, perm_id)
);
