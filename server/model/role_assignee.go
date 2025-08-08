package model

import (
	"github.com/doug-martin/goqu/v9"
)

type RoleAssignee struct {
	Db *DB `db:"-"`

	UserId uint64 `db:"user_id"`
	RoleId uint32 `db:"role_id"`
}

func NewRoleAssign(db *DB) RoleAssignee {
	return RoleAssignee{Db: db}
}

func (roleAssign *RoleAssignee) Insert() error {
	return Insert0(roleAssign.Db, TableRoleAssignees, roleAssign)
}

func (roleAssign *RoleAssignee) Update() error {
	return Update(roleAssign.Db, TableRoleAssignees, roleAssign, goqu.Ex{"role_id": roleAssign.RoleId, "user_id": roleAssign.UserId})
}

func (roleAssign *RoleAssignee) Delete() bool {
	return Delete(roleAssign.Db, TableRoleAssignees, goqu.Ex{"role_id": roleAssign.RoleId, "user_id": roleAssign.UserId})
}

func (roleAssign *RoleAssignee) Role() Role {
	member := NewRole(roleAssign.Db)
	sql, args, err := roleAssign.Db.Goqu.From(TableMembers).Select(TableMembers+".*").
		LeftJoin(goqu.I(TableRoles), goqu.On(
			goqu.I(TableRoles+".id").Eq(TableRoleAssignees+".role_id"),
		)).
		LeftJoin(goqu.I(TableRoleAssignees), goqu.On(
			goqu.I(TableRoleAssignees+".user_id").Eq(TableMembers+".user_id"),
		)).
		ToSQL()
	if err != nil {
		handleErr(err)
		return member
	}

	err = roleAssign.Db.Sqlx.QueryRowx(sql, args...).StructScan(&member)
	if err != nil {
		handleErr(err)
		return member
	}

	return member
}

func (roleAssign *RoleAssignee) Member() Member {
	member := Member{Db: roleAssign.Db}
	sql, args, err := roleAssign.Db.Goqu.From(TableMembers).Select(TableMembers+".*").
		LeftJoin(goqu.I(TableRoles), goqu.On(
			goqu.I(TableRoles+".id").Eq(TableRoleAssignees+".role_id"),
		)).
		LeftJoin(goqu.I(TableRoleAssignees), goqu.On(
			goqu.I(TableRoleAssignees+".user_id").Eq(TableMembers+".user_id"),
		)).
		ToSQL()
	if err != nil {
		handleErr(err)
		return member
	}

	err = roleAssign.Db.Sqlx.QueryRowx(sql, args...).StructScan(&member)
	if err != nil {
		handleErr(err)
		return member
	}

	return member
}
