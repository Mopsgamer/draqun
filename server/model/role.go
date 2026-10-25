package model

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type Role struct {
	checkEmpty
	Id       uint64  `db:"id" goqu:"skipinsert"`
	GroupId  uint64  `db:"group_id"`
	Position uint    `db:"position"`
	Name     Name    `db:"name"`
	Moniker  Moniker `db:"moniker"`
	Color    Color   `db:"color"`
}

var _ Model = (*Role)(nil)

func (role Role) Validate() htmx.Alert {
	if !role.Name.IsValid() {
		return htmx.AlertFormatName
	}
	if !role.Moniker.IsValid() {
		return htmx.AlertFormatMoniker
	}
	if !role.Color.IsValid() {
		return htmx.AlertFormatColor
	}
	return nil
}

func NewRoleFromId(id uint64, groupId uint64) (Role, error) {
	role := Role{}
	return role, First(TableRoles, goqu.Ex{"id": id, "group_id": groupId}, &role)
}

func NewRoleFromName(name Name, groupId uint64) (Role, error) {
	role := Role{}
	return role, First(TableRoles, goqu.Ex{"name": name, "group_id": groupId}, &role)
}

const roleNameEveryone Name = "@everyone"

func NewRoleEveryone(groupId uint64) Role {
	return Role{
		GroupId: groupId,
		Name:    roleNameEveryone,
		Moniker: "everyone",
	}
}

func (role *Role) Insert() error {
	return InsertId(TableRoles, role, &role.Id)
}

func (role Role) Update() error {
	return Update(TableRoles, role, goqu.Ex{"id": role.Id, "group_id": role.GroupId})
}

func (role Role) Delete() error {
	return Delete(TableRoles, goqu.Ex{"id": role.Id, "group_id": role.GroupId})
}

func (role Role) Assignees() []User {
	userList := []User{}
	sql, args, err := Goqu.Select(TableUsers+".*").From(TableRoleAssignees).
		Join(goqu.I(TableUsers), goqu.On(goqu.I(TableUsers+".id").Eq(goqu.I(TableRoleAssignees+".user_id")))).
		Join(goqu.I(TableMembers), goqu.On(goqu.I(TableMembers+".user_id").Eq(goqu.I(TableRoleAssignees+".user_id")))).
		Where(goqu.Ex{
			TableRoleAssignees + ".role_id": role.Id,
			TableMembers + ".group_id":      role.GroupId,
			TableMembers + ".is_deleted":    false,
			TableUsers + ".is_deleted":      false,
		}).
		Prepared(true).ToSQL()
	if err != nil {
		handleErr(err)
		return userList
	}

	err = Sqlx.Select(&userList, sql, args...)
	if err != nil {
		handleErr(err)
	}

	for i := range userList {
		userList[i].SetFullfilled()
	}
	return userList
}
