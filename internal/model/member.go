package model

type Member struct {
	GroupId  uint    `db:"group_id"`
	UserId   uint    `db:"user_id"`
	Nick     *string `db:"membername"`
	IsOwner  bool    `db:"is_owner"`
	IsBanned bool    `db:"is_banned"`
}
