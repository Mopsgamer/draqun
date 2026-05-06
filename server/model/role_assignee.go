package model

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type RoleAssignee struct {
	checkEmpty
	UserId uint64 `db:"user_id"`
	RoleId uint64 `db:"role_id"`
}

var _ Model = (*RoleAssignee)(nil)

func (roleAssign RoleAssignee) Validate() htmx.Alert {
	return nil
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

func (roleAssign *RoleAssignee) Role() Role {
	role := Role{}
	err := First(TableRoles, goqu.Ex{"id": roleAssign.RoleId}, &role)
	if err != nil {
		handleErr(err)
	}
	return role
}

func (roleAssign *RoleAssignee) Member() Member {
	role := roleAssign.Role()
	member, err := NewMemberFromId(role.GroupId, roleAssign.UserId)
	if err != nil {
		handleErr(err)
	}
	return member
}
