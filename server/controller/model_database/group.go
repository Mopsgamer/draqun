package model_database

import (
	"time"
)

type GroupMode int

const (
	GroupModeDm GroupMode = iota
	GroupModePrivate
	GroupModePublic
)

type Group struct {
	Id          uint64    `db:"id"`
	CreatorId   uint64    `db:"creator_id"`
	Nick        string    `db:"nickname"`    // Nick is a customizable name.
	Name        string    `db:"groupname"`   // Name is a simple identificator, which can be used to create invite-links.
	Mode        GroupMode `db:"groupmode"`   // See: GroupModeDm, GroupModePrivate, GroupModePublic.
	Description string    `db:"description"` // Optional hashed password string.
	Password    *string   `db:"password"`
	Avatar      string    `db:"avatar"`
	CreatedAt   time.Time `db:"created_at"`
}
