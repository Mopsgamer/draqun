package model

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type ActionRole struct {
	checkEmpty
	RoleId      uint64      `db:"role_id"`
	CreatorId   uint64      `db:"creator_id"`
	GroupId     uint64      `db:"group_id"`
	Action      string      `db:"action"` // "create", "update", "delete"
	Description Description `db:"description"`
	ActedAt     TimePast    `db:"acted_at"`
}

var _ Action = (*ActionRole)(nil)

func (action ActionRole) Kind() string {
	return "role"
}

func (action ActionRole) Validate() htmx.Alert {
	if !action.Description.IsValid() {
		return htmx.AlertFormatDescription
	}
	if !action.ActedAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}
	return nil
}

func (action *ActionRole) Insert() error {
	return Insert(TableRoleActions, action)
}

func (action ActionRole) Update() error {
	return Update(TableRoleActions, action, goqu.Ex{"role_id": action.RoleId, "group_id": action.GroupId})
}

func (action ActionRole) Creator() User {
	user, _ := NewUserFromId(action.CreatorId)
	return user
}

func (action ActionRole) Role() Role {
	role, _ := NewRoleFromId(action.RoleId, action.GroupId)
	return role
}
