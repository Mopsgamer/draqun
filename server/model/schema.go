package model

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed queries/*.sql queries/migrate/*.sql
var queriesFS embed.FS

var createTableFiles = []string{
	"queries/create_users.sql",
	"queries/create_groups.sql",
	"queries/create_group_members.sql",
	"queries/create_group_roles.sql",
	"queries/create_group_role_assignees.sql",
	"queries/create_group_messages.sql",
	"queries/create_group_action_memberships.sql",
	"queries/create_group_action_kicks.sql",
	"queries/create_group_action_bans.sql",
	"queries/create_group_action_roles.sql",
	"queries/create_channels.sql",
	"queries/create_group_channels.sql",
	"queries/create_channel_relations.sql",
	"queries/create_user_channels.sql",
	"queries/create_direct_channels.sql",
	"queries/create_channel_permissions.sql",
	"queries/create_channel_assignees.sql",
	"queries/create_notifications.sql",
	"queries/create_user_ips.sql",
	"queries/create_blocked_ips.sql",
	"queries/create_user_logs.sql",
	"queries/create_user_friends.sql",
	"queries/create_permissions.sql",
	"queries/create_group_message_attachments.sql",
	"queries/create_server_config.sql",
}

type migrationEntry struct {
	version int
	file    string
}

var migrationsList = []migrationEntry{
	{version: 1, file: "queries/migrate/alter_group_roles.sql"},
	{version: 2, file: "queries/migrate/alter_group_messages.sql"},
	{version: 3, file: "queries/migrate/alter_channels.sql"},
	{version: 4, file: "queries/migrate/alter_users.sql"},
	{version: 5, file: "queries/migrate/create_user_friends.sql"},
	{version: 6, file: "queries/migrate/create_group_message_attachments.sql"},
}

func initSchema() error {
	for _, file := range createTableFiles {
		content, err := queriesFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read embedded sql %s: %w", file, err)
		}
		if _, err := Sqlx.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute embedded sql %s: %w", file, err)
		}
	}

	return runSchemaMigrations()
}

func runSchemaMigrations() error {
	var currentVersion int
	var userVersion int
	err := Sqlx.Get(&userVersion, "PRAGMA user_version")
	if err == nil {
		currentVersion = userVersion
	}

	for _, m := range migrationsList {
		if m.version > currentVersion {
			content, err := queriesFS.ReadFile(m.file)
			if err != nil {
				return fmt.Errorf("failed to read migration file %s: %w", m.file, err)
			}

			statements := strings.Split(string(content), ";")
			for _, stmt := range statements {
				stmt = strings.TrimSpace(stmt)
				if stmt == "" {
					continue
				}
				if _, err := Sqlx.Exec(stmt); err != nil {
					msg := strings.ToLower(err.Error())
					if strings.Contains(msg, "duplicate column name") ||
						strings.Contains(msg, "already exists") ||
						strings.Contains(msg, "no such table") {
						continue
					}
					return fmt.Errorf("failed to run migration statement in %s: %w", m.file, err)
				}
			}

			if _, err := Sqlx.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.version)); err != nil {
				return fmt.Errorf("failed to update user_version to %d: %w", m.version, err)
			}
		}
	}

	return nil
}
