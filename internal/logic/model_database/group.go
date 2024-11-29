package model_database

import (
	"strconv"
	"time"
)

const (
	GroupModeDm      string = "dm"
	GroupModePrivate string = "private"
	GroupModePublic  string = "public"
)

type Group struct {
	Id        uint64 `db:"id"`
	CreatorId uint64 `db:"creator_id"`
	// Nick is customizable name. Can contain emojis and special characters.
	Nick string `db:"nickname"`
	// Name is a simple identificator, which can be used to create links to a specific groups or joining them by this name.
	Name string `db:"groupname"`
	// See: GroupModeDm, GroupModePrivate, GroupModePublic.
	Mode string `db:"groupmode"`
	// Optional hashed password string.
	Description string    `db:"description"`
	Password    *string   `db:"password"`
	Avatar      string    `db:"avatar"`
	CreatedAt   time.Time `db:"created_at"`
}

func (g Group) PagePath() string {
	return "/chat/groups/" + strconv.FormatUint(uint64(g.Id), 10)
}
