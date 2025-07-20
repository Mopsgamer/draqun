package model_database

type RoleAssign struct {
	GroupId uint64 `db:"group_id"`
	UserId  uint64 `db:"user_id"`
	RoleId  uint32 `db:"role_id"`
}
