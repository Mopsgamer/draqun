package database

import (
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
)

type ActionMembership struct {
	Db *DB `db:"-"`

	UserId  uint64        `db:"user_id"`  // The user being acted upon.
	GroupId uint64        `db:"group_id"` // The group where the action was performed.
	ActedAt time.Time     `db:"acted_at"` // The time when the action was performed.
	IsJoin  types.BitBool `db:"is_join"`  // True if the action is a join, false if it's a leave.
}

func (action ActionMembership) Kind() string {
	return "membership"
}

func (action ActionMembership) IsEmpty() bool {
	return action.UserId != 0 && action.GroupId != 0
}

func (action *ActionMembership) Insert() bool {
	return Insert(action.Db, string(TableMemberships), action) != 0
}

func (action ActionMembership) Update() bool {
	return Update(action.Db, TableMemberships, action, goqu.Ex{"user_id": action.UserId, "group_id": action.GroupId})
}

func (action *ActionMembership) FromId(userId, groupId uint64) bool {
	Last(action.Db, TableMemberships, goqu.Ex{"user_id": userId, "group_id": groupId}, goqu.I(TableMemberships+".user_id"), action)
	return action.IsEmpty()
}

func (action ActionMembership) User() User {
	user := User{Db: action.Db}
	user.FromId(action.UserId)
	return user
}

func (action ActionMembership) Group() Group {
	group := Group{Db: action.Db}
	group.FromId(action.GroupId)
	return group
}
