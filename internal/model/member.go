package model

import "github.com/jmoiron/sqlx/types"

type Member struct {
	GroupId  uint64        `db:"group_id"`
	UserId   uint64        `db:"user_id"`
	Nick     *string       `db:"membername"`
	IsOwner  types.BitBool `db:"is_owner"`
	IsBanned types.BitBool `db:"is_banned"`
}
