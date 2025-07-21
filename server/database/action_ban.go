package database

import (
	"time"

	"github.com/doug-martin/goqu/v9"
)

type ActionBan struct {
	Db *goqu.Database

	TargetId    uint64    `db:"target_id"`  // The user being banned.
	CreatorId   uint64    `db:"creator_id"` // The user who created the ban.
	RevokerId   uint64    `db:"revoker_id"` // The user who unbanned the user.
	GroupId     uint64    `db:"group_id"`   // Nick is a customizable name.
	Description string    `db:"description"`
	ActedAt     time.Time `db:"acted_at"`
	EndsAt      time.Time `db:"ends_at"`
}

func (action ActionBan) IsEmpty() bool {
	return action.TargetId != 0 && action.CreatorId != 0 && action.GroupId != 0
}

func (action *ActionBan) Insert() bool {
	return Insert(action.Db, string(TableBans), action) != nil
}

func (action ActionBan) Update() bool {
	return Update(action.Db, TableBans, action, goqu.Ex{"target_id": action.TargetId, "group_id": action.GroupId, "creator_id": action.CreatorId})
}

func (action *ActionBan) FromId(targetId, creatorId, groupId uint64) bool {
	First(action.Db, TableBans, goqu.Ex{"target_id": targetId, "creator_id": creatorId, "group_id": groupId}, action)
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
