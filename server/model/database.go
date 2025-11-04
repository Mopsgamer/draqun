package model

import (
	"fmt"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/jmoiron/sqlx"
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
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		environment.DBUser,
		environment.DBPassword,
		environment.DBHost,
		environment.DBPort,
		environment.DBName,
	)

	var err error
	Sqlx, err = sqlx.Open("mysql", connectionString)
	if err != nil {
		return err
	}
	if err := Sqlx.Ping(); err != nil {
		return err
	}

	Goqu = goqu.New("mysql", Sqlx)
	return nil
}
