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

	Goqu = goqu.New("sqlite3", Sqlx)
	return nil
}
