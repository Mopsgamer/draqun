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
	Moniker     string    `db:"moniker"` // Nick is a customizable name.
	Name        string    `db:"name"`    // Name is a simple identificator, which can be used to create invite-links.
	Mode        GroupMode `db:"mode"`
	Password    string    `db:"password"` // Optional hashed password string.
	Description string    `db:"description"`
	Avatar      string    `db:"avatar"`
	CreatedAt   time.Time `db:"created_at"`
}
