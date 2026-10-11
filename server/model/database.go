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
	TableGroups        = "app_groups"
	TableMembers       = "app_group_members"
	TableUsers         = "app_users"
	TableMessages      = "app_group_messages"
	TableRoles         = "app_group_roles"
	TableRoleAssignees = "app_group_role_assignees"
	TableBans          = "app_group_action_bans"
	TableKicks         = "app_group_action_kicks"
	TableMemberships   = "app_group_action_memberships"
	TableRoleActions   = "app_group_action_roles"
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

	Goqu = goqu.New("sqlite3", Sqlx)
	return nil
}
