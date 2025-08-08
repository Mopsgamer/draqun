package model

import (
	"time"

	"github.com/doug-martin/goqu/v9"
)

type ActionBan struct {
	Db *DB `db:"-"`

	TargetId    uint64    `db:"target_id"`  // The user being banned.
	CreatorId   uint64    `db:"creator_id"` // The user who created the ban.
	RevokerId   uint64    `db:"revoker_id"` // The user who unbanned the user.
	GroupId     uint64    `db:"group_id"`   // Nick is a customizable name.
	Description string    `db:"description"`
	ActedAt     time.Time `db:"acted_at"`
	EndsAt      time.Time `db:"ends_at"`
}

var _ Action = (*ActionBan)(nil)

func (action ActionBan) Kind() string {
	return "ban"
}

func (action ActionBan) IsEmpty() bool {
	return action.TargetId != 0 && action.CreatorId != 0 && action.GroupId != 0
}

func (action *ActionBan) Insert() error {
	return Insert0(action.Db, string(TableBans), action)
}

func (action ActionBan) Update() error {
	return Update(action.Db, TableBans, action, goqu.Ex{"target_id": action.TargetId, "group_id": action.GroupId})
}

func (action *ActionBan) FromId(targetId, groupId uint64) bool {
	Last(action.Db, TableBans, goqu.Ex{"target_id": targetId, "group_id": groupId}, goqu.I(TableBans+".target_id"), action)
	return action.IsEmpty()
}

func (action ActionBan) Target() User {
	user := User{Db: action.Db}
	user.FromId(action.TargetId)
	return user
}

func (action ActionBan) Creator() User {
	user := User{Db: action.Db}
	user.FromId(action.CreatorId)
	return user
}

func (action ActionBan) Revoker() User {
	user := User{Db: action.Db}
	user.FromId(action.RevokerId)
	return user
}
