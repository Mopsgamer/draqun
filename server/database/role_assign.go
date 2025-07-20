package database

import (
	"github.com/Mopsgamer/draqun/server/model_database"
	"github.com/doug-martin/goqu/v9"
)

type RoleAssign struct {
	Db *goqu.Database
	*model_database.RoleAssign
}

func NewRoleAssign(db *goqu.Database) RoleAssign {
	return RoleAssign{Db: db}
}

func (roleAssign *RoleAssign) Insert() bool {
	return Insert(roleAssign.Db, TableRoleAssigns, roleAssign) != nil
}

func (roleAssign *RoleAssign) Update() bool {
	return Update(roleAssign.Db, TableRoleAssigns, roleAssign, goqu.Ex{"group_id": roleAssign.GroupId, "user_id": roleAssign.UserId, "role_id": roleAssign.RoleId})
}

func (roleAssign *RoleAssign) Delete(groupId, userId uint64) bool {
	return Delete(roleAssign.Db, TableRoleAssigns, goqu.Ex{"group_id": groupId, "user_id": userId})
}
