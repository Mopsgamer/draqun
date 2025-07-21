package database

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
)

type ActionMembership struct {
	Db *goqu.Database

	UserId      uint64        `db:"user_id"`  // The user being acted upon.
	GroupId     uint64        `db:"group_id"` // The group where the action was performed.
	Description string        `db:"description"`
	IsJoin      types.BitBool `db:"is_join"` // True if the action is a join, false if it's a leave.
}

func (action ActionMembership) IsEmpty() bool {
	return action.UserId != 0 && action.GroupId != 0
}

func (action *ActionMembership) Insert() bool {
	return Insert(action.Db, string(TableMemberships), action) != nil
}

func (action ActionMembership) Update() bool {
	return Update(action.Db, TableMemberships, action, goqu.Ex{"user_id": action.UserId, "group_id": action.GroupId})
}

func (action *ActionMembership) FromId(userId, groupId uint64) bool {
	First(action.Db, TableMemberships, goqu.Ex{"user_id": userId, "group_id": groupId}, action)
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
