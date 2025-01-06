package database

import (
	"restapp/internal/controller/model_database"

	"github.com/gofiber/fiber/v3/log"
)

func (db Database) RoleAssign(right model_database.RoleAssign) bool {
	query :=
		`INSERT INTO app_group_role_assigns (
			group_id,
			user_id,
			right_id
		)
    	VALUES (?, ?, ?)`
	_, err := db.Sql.Exec(query,
		right.GroupId,
		right.UserId,
		right.RightId,
	)

	if err != nil {
		log.Error(err)
		return false
	}

	return true
}

func (db Database) RoleCreate(role model_database.Role) *uint32 {
	query :=
		`INSERT INTO app_group_roles (
			name,
			color,
			perm_chat_read,
			perm_chat_write,
			perm_chat_delete,
			perm_kick,
			perm_ban,
			perm_change_group,
			perm_change_member
		)
    	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Sql.Exec(query,
		role.Name,
		role.Color,
		role.ChatRead,
		role.ChatWrite,
		role.ChatDelete,
		role.Kick,
		role.Ban,
		role.GroupChange,
		role.MemberChange,
	)

	if err != nil {
		log.Error(err)
		return nil
	}

	newId := uint32(db.Context().LastInsertId)
	return &newId
}

func (db Database) RoleById(roleId uint64) *model_database.Role {
	role := new(model_database.Role)
	query := `SELECT * FROM app_group_roles WHERE id = ?`
	err := db.Sql.Get(role, query, roleId)

	if err != nil {
		log.Error(err)
		return role
	}
	return role
}

func (db Database) MemberRoleList(groupId, userId uint64) []model_database.Role {
	roleList := new([]model_database.Role)
	query := `SELECT app_group_roles.*
	FROM app_group_roles
	LEFT JOIN app_group_role_assigns ON app_group_role_assigns.right_id = app_group_roles.id
	WHERE app_group_role_assigns.group_id = ? AND app_group_role_assigns.user_id = ?`
	err := db.Sql.Select(roleList, query, groupId, userId)

	if err != nil {
		log.Error(err)
		return *roleList
	}
	return *roleList
}

func (db Database) MemberRights(groupId, userId uint64) model_database.Role {
	roleList := db.MemberRoleList(groupId, userId)
	rights := model_database.RoleDefault
	if len(roleList) == 0 {
		return rights
	}

	rights.Merge(roleList...)
	return rights
}
