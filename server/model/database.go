package model

import (
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite" // <--- The "Pure Go" secret to cross-compilation
)

var Goqu *goqu.Database
var Sqlx *sqlx.DB

type DB = sqlx.DB

// SQL table name.
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

// Initialize the DB wrapper.
func LoadDB() error {
	// Simple local file path
	connectionString := "app_data.db"

	var err error
	// Use "sqlite" (the modernc driver name)
	Sqlx, err = sqlx.Open("sqlite", connectionString)
	if err != nil {
		return err
	}

	// Performance Tip: SQLite works best with one connection for writes
	Sqlx.SetMaxOpenConns(1)

	if err := Sqlx.Ping(); err != nil {
		return err
	}

	// Goqu uses "sqlite3" for its dialect name
	Goqu = goqu.New("sqlite3", Sqlx)
	return nil
}
