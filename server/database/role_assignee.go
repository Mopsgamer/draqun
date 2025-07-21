package database

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/gofiber/fiber/v3/log"
)

type RoleAssignee struct {
	Db *goqu.Database

	UserId uint64 `db:"user_id"`
	RoleId uint32 `db:"role_id"`
}

func NewRoleAssign(db *goqu.Database) RoleAssignee {
	return RoleAssignee{Db: db}
}

func (roleAssign *RoleAssignee) Insert() bool {
	return Insert(roleAssign.Db, TableRoleAssignees, roleAssign) != nil
}

func (roleAssign *RoleAssignee) Update() bool {
	return Update(roleAssign.Db, TableRoleAssignees, roleAssign, goqu.Ex{"role_id": roleAssign.RoleId, "user_id": roleAssign.UserId})
}

func (roleAssign *RoleAssignee) Delete() bool {
	return Delete(roleAssign.Db, TableRoleAssignees, goqu.Ex{"role_id": roleAssign.RoleId, "user_id": roleAssign.UserId})
}

func (roleAssign *RoleAssignee) Role() Role {
	member := NewRole(roleAssign.Db)
	found, err := roleAssign.Db.From(TableMembers).Select(TableMembers+".*").
		LeftJoin(goqu.I(TableRoles), goqu.On(
			goqu.I(TableRoles+".id").Eq(TableRoleAssignees+".role_id"),
		)).
		LeftJoin(goqu.I(TableRoleAssignees), goqu.On(
			goqu.I(TableRoleAssignees+".user_id").Eq(TableMembers+".user_id"),
		)).
		ScanStruct(member)

	if !found {
		log.Error(err)
	}

	return member
}

func (roleAssign *RoleAssignee) Member() Member {
	member := Member{Db: roleAssign.Db}
	found, err := roleAssign.Db.From(TableMembers).Select(TableMembers+".*").
		LeftJoin(goqu.I(TableRoles), goqu.On(
			goqu.I(TableRoles+".id").Eq(TableRoleAssignees+".role_id"),
		)).
		LeftJoin(goqu.I(TableRoleAssignees), goqu.On(
			goqu.I(TableRoleAssignees+".user_id").Eq(TableMembers+".user_id"),
		)).
		ScanStruct(member)

	if !found {
		log.Error(err)
	}

	return member
}
