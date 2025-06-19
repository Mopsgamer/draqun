package database

import (
	"database/sql"

	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3/log"
)

func (db Database) RoleAssign(right model_database.RoleAssign) bool {
	return Insert(db, "app_group_role_assigns", right) != nil
}

func (db Database) RoleCreate(role model_database.Role) *uint32 {
	id := uint32(*Insert(db, "app_group_roles", role))
	return &id
}

func (db Database) RoleById(roleId uint64) *model_database.Role {
	return First[model_database.Role](db, "app_group_roles", goqu.Ex{"id": roleId})
}

func (db Database) MemberRoleList(groupId, userId uint64) []model_database.Role {
	roleList := new([]model_database.Role)
	query := `SELECT app_group_roles.*
	FROM app_group_roles
	LEFT JOIN app_group_role_assigns ON app_group_role_assigns.right_id = app_group_roles.id
	WHERE app_group_role_assigns.group_id = ? AND app_group_role_assigns.user_id = ?`
	err := db.Sqlx.Select(roleList, query, groupId, userId)

	if err != nil {
		if err == sql.ErrNoRows {
			return *roleList
		}
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
