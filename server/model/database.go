package model

import (
	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var Goqu *goqu.Database
var Sqlx *sqlx.DB

type DB = sqlx.DB

const (
	TableGroups             = "app_groups"
	TableMembers            = "app_group_members"
	TableUsers              = "app_users"
	TableMessages           = "app_group_messages"
	TableRoles              = "app_group_roles"
	TableRoleAssignees      = "app_group_role_assignees"
	TableBans               = "app_group_action_bans"
	TableKicks              = "app_group_action_kicks"
	TableMemberships        = "app_group_action_memberships"
	TableRoleActions        = "app_group_action_roles"
	TableChannels           = "app_channels"
	TableGroupChannels      = "app_group_channels"
	TableChannelRelations   = "app_channel_relations"
	TableUserChannels       = "app_user_channels"
	TableDirectChannels     = "app_direct_channels"
	TableChannelPermissions = "app_channel_permissions"
	TableChannelAssignees   = "app_channel_assignees"
	TableNotifications      = "app_notifications"
	TableUserIPs            = "app_user_ips"
	TableBlockedIPs         = "app_blocked_ips"
	TableUserLogs           = "app_user_logs"
)

func LoadDB() error {
	var err error
	Sqlx, err = sqlx.Open("sqlite", environment.DBPath)
	if err != nil {
		return err
	}

	Sqlx.SetMaxOpenConns(1)

	if err := Sqlx.Ping(); err != nil {
		return err
	}

	_, _ = Sqlx.Exec(`ALTER TABLE app_group_roles ADD COLUMN position INTEGER NOT NULL DEFAULT 0;`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_group_messages ADD COLUMN channel_id INTEGER NOT NULL DEFAULT 0;`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_group_messages ADD COLUMN attachment_url TEXT NOT NULL DEFAULT '';`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_group_messages ADD COLUMN attachment_name TEXT NOT NULL DEFAULT '';`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_group_messages ADD COLUMN attachment_size INTEGER NOT NULL DEFAULT 0;`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_group_messages ADD COLUMN attachment_type TEXT NOT NULL DEFAULT '';`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_channels ADD COLUMN status VARCHAR(32) NOT NULL DEFAULT 'open';`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_channels ADD COLUMN closed_reason VARCHAR(32) NOT NULL DEFAULT '';`)
	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    moniker VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    avatar VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    last_seen_at DATETIME NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT 0,
    is_suspended BOOLEAN NOT NULL DEFAULT 0,
    suspension_reason VARCHAR(510) NOT NULL DEFAULT '',
    is_locked BOOLEAN NOT NULL DEFAULT 0,
    locked_password VARCHAR(255) NOT NULL DEFAULT '',
    last_ip VARCHAR(64) NOT NULL DEFAULT '',
    UNIQUE (name),
    UNIQUE (email)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    creator_id INTEGER NOT NULL,
    owner_id INTEGER NOT NULL,
    moniker VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    mode VARCHAR(32) NOT NULL DEFAULT 'public',
    password VARCHAR(255) NOT NULL DEFAULT '',
    description VARCHAR(510) NOT NULL DEFAULT '',
    avatar VARCHAR(255) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT 0,
    UNIQUE (name),
    FOREIGN KEY (creator_id) REFERENCES app_users (id),
    FOREIGN KEY (owner_id) REFERENCES app_users (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_group_members (
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    moniker VARCHAR(255) NOT NULL DEFAULT '',
    first_time_joined_at DATETIME NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (user_id) REFERENCES app_users (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_group_roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL,
    moniker VARCHAR(255) NOT NULL,
    color INTEGER NOT NULL DEFAULT 0,
    perm_messages VARCHAR(32) NOT NULL DEFAULT '',
    perm_roles VARCHAR(32) NOT NULL DEFAULT '',
    perm_members VARCHAR(32) NOT NULL DEFAULT '',
    perm_group_change VARCHAR(32) NOT NULL DEFAULT '',
    perm_admin VARCHAR(32) NOT NULL DEFAULT '',
    FOREIGN KEY (group_id) REFERENCES app_groups (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_group_role_assignees (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES app_users (id),
    FOREIGN KEY (role_id) REFERENCES app_group_roles (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_group_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL DEFAULT 0,
    author_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    attachment_url TEXT NOT NULL DEFAULT '',
    attachment_name TEXT NOT NULL DEFAULT '',
    attachment_size INTEGER NOT NULL DEFAULT 0,
    attachment_type TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (author_id) REFERENCES app_users (id)
);`)

	_, _ = Sqlx.Exec(`ALTER TABLE app_channels ADD COLUMN author_id INTEGER NOT NULL DEFAULT 0;`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_channels ADD COLUMN duplicate_of_id INTEGER NOT NULL DEFAULT 0;`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_channels ADD COLUMN updated_at DATETIME;`)

	_, _ = Sqlx.Exec(`ALTER TABLE app_users ADD COLUMN is_suspended BOOLEAN NOT NULL DEFAULT 0;`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_users ADD COLUMN suspension_reason VARCHAR(510) NOT NULL DEFAULT '';`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_users ADD COLUMN is_locked BOOLEAN NOT NULL DEFAULT 0;`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_users ADD COLUMN locked_password VARCHAR(255) NOT NULL DEFAULT '';`)
	_, _ = Sqlx.Exec(`ALTER TABLE app_users ADD COLUMN last_ip VARCHAR(64) NOT NULL DEFAULT '';`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_user_ips (
    user_id INTEGER NOT NULL,
    ip VARCHAR(64) NOT NULL,
    last_seen_at DATETIME NOT NULL,
    PRIMARY KEY (user_id, ip),
    FOREIGN KEY (user_id) REFERENCES app_users (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_blocked_ips (
    ip VARCHAR(64) PRIMARY KEY,
    creator_id INTEGER NOT NULL,
    reason VARCHAR(255) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_user_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    actor_id INTEGER NOT NULL,
    action VARCHAR(64) NOT NULL,
    details VARCHAR(510) NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES app_users (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_group_action_bans (
    group_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,
    creator_id INTEGER NOT NULL,
    revoker_id INTEGER NOT NULL DEFAULT 0,
    description VARCHAR(510) NOT NULL DEFAULT '',
    acted_at DATETIME NOT NULL,
    ends_at DATETIME NOT NULL,
    PRIMARY KEY (group_id, target_id, acted_at),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (target_id) REFERENCES app_users (id),
    FOREIGN KEY (creator_id) REFERENCES app_users (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_group_action_kicks (
    group_id INTEGER NOT NULL,
    target_id INTEGER NOT NULL,
    creator_id INTEGER NOT NULL,
    description VARCHAR(510) NOT NULL DEFAULT '',
    acted_at DATETIME NOT NULL,
    PRIMARY KEY (group_id, target_id, acted_at),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (target_id) REFERENCES app_users (id),
    FOREIGN KEY (creator_id) REFERENCES app_users (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_group_action_memberships (
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    is_join BOOLEAN NOT NULL,
    acted_at DATETIME NOT NULL,
    PRIMARY KEY (group_id, user_id, acted_at),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (user_id) REFERENCES app_users (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_group_action_roles (
    role_id INTEGER NOT NULL,
    creator_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('create', 'update', 'delete', 'assign', 'unassign')),
    description VARCHAR(510) NOT NULL,
    acted_at DATETIME NOT NULL,
    FOREIGN KEY (role_id) REFERENCES app_group_roles (id),
    FOREIGN KEY (creator_id) REFERENCES app_users (id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    moniker VARCHAR(255) NOT NULL,
    description VARCHAR(510) NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'text',
    status VARCHAR(32) NOT NULL DEFAULT 'open',
    closed_reason VARCHAR(32) NOT NULL DEFAULT '',
    author_id INTEGER NOT NULL DEFAULT 0,
    duplicate_of_id INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME,
    created_at DATETIME NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT 0
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_group_channels (
    group_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, channel_id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_channel_relations (
    parent_id INTEGER NOT NULL,
    child_id INTEGER NOT NULL,
    relation_type VARCHAR(32) NOT NULL DEFAULT 'subchannel',
    PRIMARY KEY (parent_id, child_id),
    FOREIGN KEY (parent_id) REFERENCES app_channels (id),
    FOREIGN KEY (child_id) REFERENCES app_channels (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_user_channels (
    user_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, channel_id),
    FOREIGN KEY (user_id) REFERENCES app_users (id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_direct_channels (
    user1_id INTEGER NOT NULL,
    user2_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    PRIMARY KEY (user1_id, user2_id),
    FOREIGN KEY (user1_id) REFERENCES app_users (id),
    FOREIGN KEY (user2_id) REFERENCES app_users (id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_channel_permissions (
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
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_channel_assignees (
    channel_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    PRIMARY KEY (channel_id, user_id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id),
    FOREIGN KEY (user_id) REFERENCES app_users (id)
);`)

	_, _ = Sqlx.Exec(`CREATE TABLE IF NOT EXISTS app_notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    actor_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    channel_id INTEGER NOT NULL,
    message_id INTEGER NOT NULL,
    type VARCHAR(32) NOT NULL DEFAULT 'mention',
    title VARCHAR(255) NOT NULL,
    content VARCHAR(510) NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES app_users (id),
    FOREIGN KEY (actor_id) REFERENCES app_users (id),
    FOREIGN KEY (group_id) REFERENCES app_groups (id),
    FOREIGN KEY (channel_id) REFERENCES app_channels (id),
    FOREIGN KEY (message_id) REFERENCES app_group_messages (id)
);`)

	Goqu = goqu.New("sqlite3", Sqlx)
	return nil
}
