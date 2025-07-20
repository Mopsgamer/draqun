package database

import (
	"database/sql"

	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3/log"
)

func (db Database) RoleAssign(right model_database.RoleAssign) bool {
	return Insert(db, TableRoleAssigns, right) != nil
}

func (db Database) RoleCreate(role model_database.Role) *uint32 {
	id := uint32(*Insert(db, TableRoles, role))
	return &id
}

func (db Database) RoleById(roleId uint64) *model_database.Role {
	return First[model_database.Role](db, TableRoles, goqu.Ex{"id": roleId})
}

func (db Database) MemberRoleList(groupId, userId uint64) []model_database.Role {
	roleList := new([]model_database.Role)
	err := db.Goqu.From(TableRoles).Select(TableRoles+".*").
		LeftJoin(goqu.T(TableRoleAssigns), goqu.On(goqu.I(TableRoleAssigns+".role_id").Eq(goqu.I(TableRoles+".id")))).
		Where(goqu.Ex{TableRoleAssigns + ".group_id": groupId, TableRoleAssigns + ".user_id": userId}).
		ScanStructs(roleList)

	if err == sql.ErrNoRows {
		return *roleList
	}

	if err != nil {
		log.Error(err)
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
