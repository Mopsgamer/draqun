package database

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3/log"
)

type RoleAssign struct {
	Db *goqu.Database

	UserId uint64 `db:"user_id"`
	RoleId uint32 `db:"role_id"`
}

func NewRoleAssign(db *goqu.Database) RoleAssign {
	return RoleAssign{Db: db}
}

func (roleAssign *RoleAssign) Insert() bool {
	return Insert(roleAssign.Db, TableRoleAssigns, roleAssign) != nil
}

func (roleAssign *RoleAssign) Update() bool {
	return Update(roleAssign.Db, TableRoleAssigns, roleAssign, goqu.Ex{"role_id": roleAssign.RoleId, "user_id": roleAssign.UserId})
}

func (roleAssign *RoleAssign) Delete() bool {
	return Delete(roleAssign.Db, TableRoleAssigns, goqu.Ex{"role_id": roleAssign.RoleId, "user_id": roleAssign.UserId})
}

func (roleAssign *RoleAssign) Role() Role {
	member := NewRole(roleAssign.Db)
	found, err := roleAssign.Db.From(TableMembers).Select(TableMembers+".*").
		LeftJoin(goqu.I(TableRoles), goqu.On(
			goqu.I(TableRoles+".id").Eq(goqu.I(TableRoleAssigns+".role_id")),
		)).
		LeftJoin(goqu.I(TableRoleAssigns), goqu.On(
			goqu.I(TableRoleAssigns+".user_id").Eq(goqu.I(TableMembers+".user_id")),
		)).
		ScanStruct(member)

	if !found {
		log.Error(err)
	}

	return member
}

func (roleAssign *RoleAssign) Member() Member {
	member := Member{Db: roleAssign.Db}
	found, err := roleAssign.Db.From(TableMembers).Select(TableMembers+".*").
		LeftJoin(goqu.I(TableRoles), goqu.On(
			goqu.I(TableRoles+".id").Eq(goqu.I(TableRoleAssigns+".role_id")),
		)).
		LeftJoin(goqu.I(TableRoleAssigns), goqu.On(
			goqu.I(TableRoleAssigns+".user_id").Eq(goqu.I(TableMembers+".user_id")),
		)).
		ScanStruct(member)

	if !found {
		log.Error(err)
	}

	return member
}
