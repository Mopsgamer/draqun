package model

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
)

type ActionBan struct {
	TargetId    uint64      `db:"target_id"`  // The user being banned.
	CreatorId   uint64      `db:"creator_id"` // The user who created the ban.
	RevokerId   uint64      `db:"revoker_id"` // The user who unbanned the user.
	GroupId     uint64      `db:"group_id"`   // Nick is a customizable name.
	Description Description `db:"description"`
	ActedAt     TimePast    `db:"acted_at"`
	EndsAt      TimeFuture  `db:"ends_at"`
}

var _ Action = (*ActionBan)(nil)

func NewActionBanFromId(targetId uint64, groupId uint64) (ActionBan, error) {
	action := ActionBan{}
	return action, Last(TableBans, goqu.Ex{"target_id": targetId, "group_id": groupId}, goqu.C("target_id"), &action)
}

func (action ActionBan) Kind() string {
	return "ban"
}

func (action ActionBan) Validate() htmx.Alert {
	if !action.Description.IsValid() {
		return htmx.AlertFormatDescription
	}
	if !action.ActedAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}
	if !action.EndsAt.IsValid() {
		return htmx.AlertFormatFutureMoment
	}

	return nil
}

func (action ActionBan) IsEmpty() bool {
	return action.TargetId == 0 || action.CreatorId == 0 || action.GroupId == 0
}

func (action *ActionBan) Insert() error {
	return Insert(string(TableBans), action)
}

func (action ActionBan) Update() error {
	return Update(TableBans, action, goqu.Ex{"target_id": action.TargetId, "group_id": action.GroupId})
}

func (action ActionBan) Target() User {
	user, _ := NewUserFromId(action.TargetId)
	return user
}

func (action ActionBan) Creator() User {
	user, _ := NewUserFromId(action.CreatorId)
	return user
}

func (action ActionBan) Revoker() User {
	user, _ := NewUserFromId(action.RevokerId)
	return user
}
