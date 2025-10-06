package model

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type RoleAssignee struct {
	UserId uint64 `db:"user_id"`
	RoleId uint32 `db:"role_id"`
}

var _ Model = (*RoleAssignee)(nil)

func (roleAssign RoleAssignee) Validate() htmx.Alert {
	return nil
}

func (roleAssign RoleAssignee) IsEmpty() bool {
	return roleAssign.UserId == 0 || roleAssign.RoleId == 0
}

func (roleAssign *RoleAssignee) Insert() error {
	return Insert(TableRoleAssignees, roleAssign)
}

func (roleAssign RoleAssignee) Update() error {
	return Update(TableRoleAssignees, roleAssign, goqu.Ex{"role_id": roleAssign.RoleId, "user_id": roleAssign.UserId})
}

func (roleAssign RoleAssignee) Delete() error {
	return Delete(TableRoleAssignees, goqu.Ex{"role_id": roleAssign.RoleId, "user_id": roleAssign.UserId})
}

// FIXME: should use group
func (roleAssign *RoleAssignee) Role() Role {
	role := Role{}
	sql, args, err := Goqu.From(TableRoles).Select(goqu.C("*").Table(TableRoles)).
		LeftJoin(goqu.I(TableRoles), goqu.On(
			goqu.C("id").Table(TableRoles).Eq(goqu.C("role_id").Table(TableRoleAssignees)),
		)).
		LeftJoin(goqu.I(TableRoleAssignees), goqu.On(
			goqu.C("user_id").Table(TableRoleAssignees).Eq(goqu.C("user_id").Table(TableRoles)),
		)).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return role
	}

	err = Sqlx.QueryRowx(sql, args...).StructScan(&role)
	if err != nil {
		handleErr(err)
		return role
	}

	return role
}

// FIXME: should use group
func (roleAssign *RoleAssignee) Member() Member {
	member := Member{}
	sql, args, err := Goqu.From(TableMembers).Select(goqu.C("*").Table(TableMembers)).
		LeftJoin(goqu.I(TableRoles), goqu.On(
			goqu.C("id").Table(TableRoles).Eq(goqu.C("role_id").Table(TableRoleAssignees)),
		)).
		LeftJoin(goqu.I(TableRoleAssignees), goqu.On(
			goqu.C("user_id").Table(TableRoleAssignees).Eq(goqu.C("user_id").Table(TableMembers)),
		)).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return member
	}

	err = Sqlx.QueryRowx(sql, args...).StructScan(&member)
	if err != nil {
		handleErr(err)
		return member
	}

	return member
}
