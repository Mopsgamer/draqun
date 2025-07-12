package model_database

import (
	"time"
)

type GroupMode string

const (
	GroupModeDm      GroupMode = "dm"
	GroupModePrivate GroupMode = "private"
	GroupModePublic  GroupMode = "public"
)

type Group struct {
	Id          uint64    `db:"id"`
	CreatorId   uint64    `db:"creator_id"`
	Nick        string    `db:"nickname"`  // Nick is a customizable name.
	Name        string    `db:"groupname"` // Name is a simple identificator, which can be used to create invite-links.
	Mode        GroupMode `db:"groupmode"`
	Description string    `db:"description"`
	Password    *string   `db:"password"` // Optional hashed password string.
	Avatar      string    `db:"avatar"`
	CreatedAt   time.Time `db:"created_at"`
}
