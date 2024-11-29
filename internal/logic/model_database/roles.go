package model_database

type Role struct {
	GroupId  uint64 `db:"group_id"`
	UserId   uint64 `db:"user_id"`
	RightsId uint32 `db:"role_id"`
}
