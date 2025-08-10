package model

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type ActionKick struct {
	Db *DB `db:"-"`

	TargetId    uint64      `db:"target_id"`  // The user being banned.
	CreatorId   uint64      `db:"creator_id"` // The user who created the ban.
	GroupId     uint64      `db:"group_id"`   // The group from which the user is kicked.
	Description Description `db:"description"`
	ActedAt     TimePast    `db:"acted_at"`
}

var _ Action = (*ActionKick)(nil)

func (action ActionKick) Kind() string {
	return "kick"
}

func (action ActionKick) Validate() htmx.Alert {
	if !action.Description.IsValid() {
		return htmx.AlertFormatDescription
	}
	if !action.ActedAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}

	return nil
}

func (action ActionKick) IsEmpty() bool {
	return action.TargetId == 0 || action.CreatorId == 0 || action.GroupId == 0
}

func (action *ActionKick) Insert() error {
	return Insert(action.Db, string(TableKicks), action)
}

func (action ActionKick) Update() error {
	return Update(action.Db, TableKicks, action, goqu.Ex{"target_id": action.TargetId, "group_id": action.GroupId, "creator_id": action.CreatorId})
}

func (action *ActionKick) FromId(targetId, creatorId, groupId uint64) bool {
	Last(action.Db, TableKicks, goqu.Ex{"target_id": targetId, "group_id": groupId, "creator_id": creatorId}, goqu.I(TableKicks+".target_id"), action)
	return action.IsEmpty()
}

func (action ActionKick) Target() User {
	user := User{Db: action.Db}
	user.FromId(action.TargetId)
	return user
}

func (action ActionKick) Creator() User {
	user := User{Db: action.Db}
	user.FromId(action.CreatorId)
	return user
}

func (action ActionKick) Group() Group {
	group := Group{Db: action.Db}
	group.FromId(action.GroupId)
	return group
}
