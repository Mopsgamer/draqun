package model

import (
	"github.com/Mopsgamer/draqun/server/htmx"
	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx/types"
)

type ActionMembership struct {
	UserId  uint64        `db:"user_id"`  // The user being acted upon.
	GroupId uint64        `db:"group_id"` // The group where the action was performed.
	ActedAt TimePast      `db:"acted_at"` // The time when the action was performed.
	IsJoin  types.BitBool `db:"is_join"`  // True if the action is a join, false if it's a leave.
}

func NewActionMembershipFromId(userId, creatorId, groupId uint64) (ActionMembership, error) {
	action := ActionMembership{}
	return action, Last(TableMemberships, goqu.Ex{"user_id": userId, "group_id": groupId}, goqu.I(TableMemberships+".user_id"), &action)
}

var _ Action = (*ActionMembership)(nil)

func (action ActionMembership) Kind() string {
	return "membership"
}

func (action ActionMembership) Validate() htmx.Alert {
	if !action.ActedAt.IsValid() {
		return htmx.AlertFormatPastMoment
	}

	return nil
}

func (action ActionMembership) IsEmpty() bool {
	return action.UserId == 0 || action.GroupId == 0
}

func (action *ActionMembership) Insert() error {
	return Insert(string(TableMemberships), action)
}

func (action ActionMembership) Update() error {
	return Update(TableMemberships, action, goqu.Ex{"user_id": action.UserId, "group_id": action.GroupId})
}

func (action ActionMembership) User() User {
	user, _ := NewUserFromId(action.UserId)
	return user
}

func (action ActionMembership) Group() Group {
	group, _ := NewGroupFromId(action.GroupId)
	return group
}
