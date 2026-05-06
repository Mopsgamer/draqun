CREATE TABLE IF NOT EXISTS app_group_roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    moniker TEXT NOT NULL,
    name TEXT NOT NULL,
    color INTEGER NOT NULL,
    perm_messages TEXT NOT NULL CHECK (perm_messages IN ('', 'hidden', 'read', 'write', 'delete')),
    perm_roles TEXT NOT NULL CHECK (perm_roles IN ('', 'disallow', 'allow')),
    perm_members TEXT NOT NULL CHECK (perm_members IN ('', 'read', 'invite', 'write', 'delete')),
    perm_group_change TEXT NOT NULL CHECK (perm_group_change IN ('', 'disallow', 'allow')),
    perm_admin TEXT NOT NULL CHECK (perm_admin IN ('', 'disallow', 'allow')),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    UNIQUE (group_id, name)
);