package database

import (
	"fmt"

	"github.com/Mopsgamer/draqun/server/environment"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"

	"github.com/gofiber/fiber/v3/log"
)

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
func InitDB() (*goqu.Database, error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		environment.DBUser,
		environment.DBPassword,
		environment.DBHost,
		environment.DBPort,
		environment.DBName,
	)

	connection, err := sqlx.Open("mysql", connectionString)
	if err != nil {
		return nil, err
	}

	if err := connection.Ping(); err != nil {
		return nil, err
	}

	goquConnection := goqu.New("mysql", connection)
	log.Info("Database connected successfully. Hope she is set up manually or by 'deno task init'.")
	return goquConnection, nil
}
