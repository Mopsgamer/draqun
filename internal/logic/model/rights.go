package model

import "github.com/jmoiron/sqlx/types"

type Rights struct {
	Id            uint32        `db:"id"`
	GroupId       uint64        `db:"group_id"`
	ChatRead      types.BitBool `db:"perm_chat_read"`
	ChatWrite     types.BitBool `db:"perm_chat_write"`
	ChatDelete    types.BitBool `db:"perm_chat_delete"`
	Kick          types.BitBool `db:"perm_kick"`
	Ban           types.BitBool `db:"perm_ban"`
	ChangeGroup   types.BitBool `db:"perm_change_group"`
	ChangeMemeber types.BitBool `db:"perm_change_member"`
}
