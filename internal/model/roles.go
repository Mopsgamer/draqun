package model

type Role struct {
	GroupId  uint `db:"group_id"`
	UserId   uint `db:"user_id"`
	RightsId uint `db:"role_id"`
}
